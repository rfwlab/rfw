import Capacitor
import Foundation

final class NativeSSCConnection: NSObject, URLSessionWebSocketDelegate {
    private let identifier: String
    private let request: URLRequest
    private let callback: CAPPluginCall
    private let maxMessageBytes: Int
    private let onFinish: (String) -> Void
    private let stateQueue: DispatchQueue

    private var session: URLSession?
    private var task: URLSessionWebSocketTask?
    private var outbound: BoundedOutboundQueue
    private var closed = false

    init(
        identifier: String,
        request: URLRequest,
        callback: CAPPluginCall,
        outboundQueueSize: Int,
        maxMessageBytes: Int,
        onFinish: @escaping (String) -> Void
    ) {
        self.identifier = identifier
        self.request = request
        self.callback = callback
        self.maxMessageBytes = maxMessageBytes
        self.onFinish = onFinish
        self.stateQueue = DispatchQueue(label: "dev.rfw.capacitor.ssc.\(identifier)")
        self.outbound = BoundedOutboundQueue(capacity: outboundQueueSize)
        super.init()
    }

    func start() {
        stateQueue.async { [weak self] in
            guard let self, !self.closed else { return }
            let configuration = URLSessionConfiguration.default
            configuration.httpCookieStorage = HTTPCookieStorage.shared
            configuration.httpShouldSetCookies = true
            configuration.urlCache = nil
            let session = URLSession(configuration: configuration, delegate: self, delegateQueue: nil)
            let task = session.webSocketTask(with: self.request)
            self.session = session
            self.task = task
            task.resume()
            self.receiveNext()
        }
    }

    func enqueue(_ text: String) {
        stateQueue.async { [weak self] in
            guard let self, !self.closed else { return }
            guard self.outbound.enqueue(text) else {
                self.finish(
                    event: [
                        "type": "error",
                        "message": "native SSC outbound queue overflow"
                    ],
                    cancelCode: .policyViolation,
                    cancelReason: "outbound queue overflow"
                )
                return
            }
            self.drainOutbound()
        }
    }

    func close(code: URLSessionWebSocketTask.CloseCode = .normalClosure, reason: String = "connection closed") {
        stateQueue.async { [weak self] in
            self?.finish(
                event: ["type": "close", "code": code.rawValue, "reason": reason],
                cancelCode: code,
                cancelReason: reason
            )
        }
    }

    private func drainOutbound() {
        guard !closed, let task, let text = outbound.next() else { return }
        task.send(.string(text)) { [weak self] error in
            self?.stateQueue.async {
                guard let self, !self.closed else { return }
                self.outbound.complete()
                if let error {
                    self.finish(
                        event: ["type": "error", "message": "native SSC write failed"],
                        cancelCode: .goingAway,
                        cancelReason: error.localizedDescription
                    )
                    return
                }
                self.drainOutbound()
            }
        }
    }

    private func receiveNext() {
        guard !closed, let task else { return }
        task.receive { [weak self] result in
            self?.stateQueue.async {
                guard let self, !self.closed else { return }
                switch result {
                case .success(let message):
                    self.receive(message)
                    self.receiveNext()
                case .failure:
                    self.finish(
                        event: ["type": "error", "message": "native SSC read failed"],
                        cancelCode: .goingAway,
                        cancelReason: "read failed"
                    )
                }
            }
        }
    }

    private func receive(_ message: URLSessionWebSocketTask.Message) {
        let event: JSObject
        let byteCount: Int
        switch message {
        case .string(let text):
            byteCount = text.lengthOfBytes(using: .utf8)
            event = ["type": "message", "encoding": "text", "data": text]
        case .data(let data):
            byteCount = data.count
            event = ["type": "message", "encoding": "base64", "data": data.base64EncodedString()]
        @unknown default:
            finish(
                event: ["type": "error", "message": "native SSC returned an unsupported frame"],
                cancelCode: .unsupportedData,
                cancelReason: "unsupported frame"
            )
            return
        }

        guard byteCount <= maxMessageBytes else {
            finish(
                event: ["type": "error", "message": "native SSC inbound frame exceeds maxMessageBytes"],
                cancelCode: .messageTooBig,
                cancelReason: "message too large"
            )
            return
        }
        callback.resolve(event)
    }

    private func finish(
        event: JSObject,
        cancelCode: URLSessionWebSocketTask.CloseCode,
        cancelReason: String
    ) {
        guard !closed else { return }
        closed = true
        outbound.removeAll()
        task?.cancel(with: cancelCode, reason: cancelReason.data(using: .utf8))
        task = nil
        session?.invalidateAndCancel()
        session = nil
        callback.keepAlive = false
        callback.resolve(event)
        onFinish(identifier)
    }

    func urlSession(
        _ session: URLSession,
        webSocketTask: URLSessionWebSocketTask,
        didOpenWithProtocol protocol: String?
    ) {
        stateQueue.async { [weak self] in
            guard let self, !self.closed, webSocketTask === self.task else { return }
            self.callback.resolve(["type": "open"])
        }
    }

    func urlSession(
        _ session: URLSession,
        webSocketTask: URLSessionWebSocketTask,
        didCloseWith closeCode: URLSessionWebSocketTask.CloseCode,
        reason: Data?
    ) {
        let text = reason.flatMap { String(data: $0, encoding: .utf8) } ?? ""
        stateQueue.async { [weak self] in
            guard let self, !self.closed, webSocketTask === self.task else { return }
            self.finish(
                event: ["type": "close", "code": closeCode.rawValue, "reason": text],
                cancelCode: closeCode,
                cancelReason: text
            )
        }
    }

    func urlSession(
        _ session: URLSession,
        task: URLSessionTask,
        didCompleteWithError error: Error?
    ) {
        guard error != nil else { return }
        stateQueue.async { [weak self] in
            guard let self, !self.closed, task === self.task else { return }
            self.finish(
                event: ["type": "error", "message": "native SSC connection failed"],
                cancelCode: .goingAway,
                cancelReason: "connection failed"
            )
        }
    }

    func urlSession(
        _ session: URLSession,
        task: URLSessionTask,
        willPerformHTTPRedirection response: HTTPURLResponse,
        newRequest request: URLRequest,
        completionHandler: @escaping (URLRequest?) -> Void
    ) {
        // A redirect could move an authenticated native handshake outside the
        // exact allowlist after validation. WebSocket endpoints are therefore
        // non-redirecting; the application must configure the canonical URL.
        completionHandler(nil)
        stateQueue.async { [weak self] in
            guard let self, !self.closed, task === self.task else { return }
            self.finish(
                event: ["type": "error", "message": "native SSC redirect denied"],
                cancelCode: .policyViolation,
                cancelReason: "redirect denied"
            )
        }
    }
}
