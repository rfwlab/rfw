import Capacitor
import Foundation

enum SSCConfigurationError: LocalizedError, Equatable {
    case missingAllowedOrigins
    case invalidAllowedOrigin(String)
    case invalidOriginHeader
    case invalidEndpoint
    case endpointNotAllowed
    case insecureEndpoint
    case invalidOutboundQueueSize
    case invalidMessageLimit

    var errorDescription: String? {
        switch self {
        case .missingAllowedOrigins:
            return "RFWSSC.allowedOrigins must contain at least one exact WSS origin"
        case .invalidAllowedOrigin:
            return "RFWSSC.allowedOrigins contains an invalid origin"
        case .invalidOriginHeader:
            return "RFWSSC.origin must be an origin without credentials, path, query, or fragment"
        case .invalidEndpoint:
            return "RFWSSC received an invalid WebSocket endpoint"
        case .endpointNotAllowed:
            return "RFWSSC endpoint is not in allowedOrigins"
        case .insecureEndpoint:
            return "RFWSSC requires WSS except for explicitly enabled loopback development"
        case .invalidOutboundQueueSize:
            return "RFWSSC.outboundQueueSize must be between 1 and 4096"
        case .invalidMessageLimit:
            return "RFWSSC.maxMessageBytes must be between 1024 and 16777216"
        }
    }
}

struct SSCPluginConfiguration {
    static let defaultOutboundQueueSize = 256
    static let defaultMaxMessageBytes = 8 * 1024 * 1024
    static let defaultConnectTimeoutMilliseconds = 10_000

    let allowedOrigins: Set<String>
    let originHeader: String
    let outboundQueueSize: Int
    let maxMessageBytes: Int
    let connectTimeout: TimeInterval
    let allowInsecureLocalhost: Bool

    init(pluginConfig: PluginConfig) throws {
        let configuredOrigins = pluginConfig.getArray("allowedOrigins")?.compactMap { $0 as? String } ?? []
        try self.init(
            allowedOrigins: configuredOrigins,
            originHeader: pluginConfig.getString("origin", "capacitor://localhost") ?? "capacitor://localhost",
            outboundQueueSize: pluginConfig.getInt("outboundQueueSize", Self.defaultOutboundQueueSize),
            maxMessageBytes: pluginConfig.getInt("maxMessageBytes", Self.defaultMaxMessageBytes),
            connectTimeoutMilliseconds: pluginConfig.getInt(
                "connectTimeoutMilliseconds",
                Self.defaultConnectTimeoutMilliseconds
            ),
            allowInsecureLocalhost: pluginConfig.getBoolean("allowInsecureLocalhost", false)
        )
    }

    init(
        allowedOrigins: [String],
        originHeader: String = "capacitor://localhost",
        outboundQueueSize: Int = Self.defaultOutboundQueueSize,
        maxMessageBytes: Int = Self.defaultMaxMessageBytes,
        connectTimeoutMilliseconds: Int = Self.defaultConnectTimeoutMilliseconds,
        allowInsecureLocalhost: Bool = false
    ) throws {
        guard !allowedOrigins.isEmpty else {
            throw SSCConfigurationError.missingAllowedOrigins
        }
        var normalized = Set<String>()
        for candidate in allowedOrigins {
            guard let origin = Self.normalizedWebSocketOrigin(candidate) else {
                throw SSCConfigurationError.invalidAllowedOrigin(candidate)
            }
            normalized.insert(origin)
        }
        guard Self.isValidOriginHeader(originHeader) else {
            throw SSCConfigurationError.invalidOriginHeader
        }
        guard (1...4096).contains(outboundQueueSize) else {
            throw SSCConfigurationError.invalidOutboundQueueSize
        }
        guard (1024...(16 * 1024 * 1024)).contains(maxMessageBytes) else {
            throw SSCConfigurationError.invalidMessageLimit
        }

        self.allowedOrigins = normalized
        self.originHeader = originHeader
        self.outboundQueueSize = outboundQueueSize
        self.maxMessageBytes = maxMessageBytes
        self.connectTimeout = TimeInterval(max(connectTimeoutMilliseconds, 1)) / 1000
        self.allowInsecureLocalhost = allowInsecureLocalhost
    }

    func request(endpoint: String) throws -> URLRequest {
        guard
            let components = URLComponents(string: endpoint),
            components.user == nil,
            components.password == nil,
            components.fragment == nil,
            let scheme = components.scheme?.lowercased(),
            let host = components.host?.lowercased(),
            let url = components.url
        else {
            throw SSCConfigurationError.invalidEndpoint
        }

        if scheme != "wss" {
            guard scheme == "ws", allowInsecureLocalhost, Self.isLoopback(host) else {
                throw SSCConfigurationError.insecureEndpoint
            }
        }
        guard let origin = Self.normalizedWebSocketOrigin(from: components), allowedOrigins.contains(origin) else {
            throw SSCConfigurationError.endpointNotAllowed
        }

        var request = URLRequest(url: url)
        request.timeoutInterval = connectTimeout
        request.setValue(originHeader, forHTTPHeaderField: "Origin")
        return request
    }

    private static func normalizedWebSocketOrigin(_ raw: String) -> String? {
        guard
            let components = URLComponents(string: raw),
            components.user == nil,
            components.password == nil,
            components.query == nil,
            components.fragment == nil,
            components.path.isEmpty || components.path == "/"
        else {
            return nil
        }
        return normalizedWebSocketOrigin(from: components)
    }

    private static func normalizedWebSocketOrigin(from components: URLComponents) -> String? {
        guard
            let scheme = components.scheme?.lowercased(),
            scheme == "ws" || scheme == "wss",
            let host = components.host?.lowercased()
        else {
            return nil
        }
        let port = components.port.map { ":\($0)" } ?? ""
        return "\(scheme)://\(host)\(port)"
    }

    private static func isValidOriginHeader(_ raw: String) -> Bool {
        guard
            let components = URLComponents(string: raw),
            components.user == nil,
            components.password == nil,
            components.query == nil,
            components.fragment == nil,
            components.path.isEmpty,
            let scheme = components.scheme?.lowercased(),
            ["capacitor", "https", "http"].contains(scheme),
            components.host != nil
        else {
            return false
        }
        return true
    }

    private static func isLoopback(_ host: String) -> Bool {
        host == "localhost" || host == "127.0.0.1" || host == "::1"
    }
}
