import Capacitor
import Foundation
import UIKit

@objc(RFWSSCPlugin)
public final class RFWSSCPlugin: CAPPlugin, CAPBridgedPlugin {
    public let identifier = "RFWSSCPlugin"
    public let jsName = "RFWSSC"
    public let pluginMethods: [CAPPluginMethod] = [
        CAPPluginMethod(name: "connect", returnType: CAPPluginReturnCallback),
        CAPPluginMethod(name: "send", returnType: CAPPluginReturnNone),
        CAPPluginMethod(name: "close", returnType: CAPPluginReturnNone)
    ]

    private let connectionsLock = NSLock()
    private var connections: [String: NativeSSCConnection] = [:]
    private var observers: [NSObjectProtocol] = []
    private var applicationActive = true

    public override func load() {
        setApplicationActive(UIApplication.shared.applicationState == .active)
        observers.append(
            NotificationCenter.default.addObserver(
                forName: UIApplication.didEnterBackgroundNotification,
                object: nil,
                queue: nil
            ) { [weak self] _ in
                self?.setApplicationActive(false)
                self?.closeAll(reason: "app backgrounded")
            }
        )
        observers.append(
            NotificationCenter.default.addObserver(
                forName: UIApplication.didBecomeActiveNotification,
                object: nil,
                queue: nil
            ) { [weak self] _ in
                self?.setApplicationActive(true)
            }
        )
    }

    deinit {
        for observer in observers {
            NotificationCenter.default.removeObserver(observer)
        }
        closeAll(reason: "plugin released")
    }

    @objc func connect(_ call: CAPPluginCall) {
        guard
            let id = call.getString("id"), !id.isEmpty,
            let endpoint = call.getString("url"), !endpoint.isEmpty
        else {
            call.reject("RFWSSC connect requires non-empty id and url")
            return
        }

        do {
            let configuration = try SSCPluginConfiguration(pluginConfig: getConfig())
            let request = try configuration.request(endpoint: endpoint)
            call.keepAlive = true
            let connection = NativeSSCConnection(
                identifier: id,
                request: request,
                callback: call,
                outboundQueueSize: configuration.outboundQueueSize,
                maxMessageBytes: configuration.maxMessageBytes
            ) { [weak self] finishedID in
                self?.removeConnection(finishedID)
            }

            connectionsLock.lock()
            let duplicate = connections[id] != nil
            let inactive = !applicationActive
            if !duplicate && !inactive {
                connections[id] = connection
            }
            connectionsLock.unlock()

            guard !inactive else {
                call.keepAlive = false
                call.reject("RFWSSC is unavailable while the application is inactive")
                return
            }
            guard !duplicate else {
                call.keepAlive = false
                call.reject("RFWSSC connection id already exists")
                return
            }
            connection.start()
        } catch {
            call.keepAlive = false
            call.reject(error.localizedDescription)
        }
    }

    @objc func send(_ call: CAPPluginCall) {
        guard let id = call.getString("id"), let data = call.getString("data") else { return }
        connection(id)?.enqueue(data)
    }

    @objc func close(_ call: CAPPluginCall) {
        guard let id = call.getString("id") else { return }
        connection(id)?.close()
    }

    private func connection(_ id: String) -> NativeSSCConnection? {
        connectionsLock.lock()
        defer { connectionsLock.unlock() }
        return connections[id]
    }

    private func removeConnection(_ id: String) {
        connectionsLock.lock()
        connections.removeValue(forKey: id)
        connectionsLock.unlock()
    }

    private func setApplicationActive(_ active: Bool) {
        connectionsLock.lock()
        applicationActive = active
        connectionsLock.unlock()
    }

    private func closeAll(reason: String) {
        connectionsLock.lock()
        let active = Array(connections.values)
        connectionsLock.unlock()
        for connection in active {
            connection.close(code: .goingAway, reason: reason)
        }
    }
}
