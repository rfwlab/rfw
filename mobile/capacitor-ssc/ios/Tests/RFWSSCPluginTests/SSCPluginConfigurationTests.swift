import XCTest
@testable import RFWSSCPlugin

final class SSCPluginConfigurationTests: XCTestCase {
    func testBuildsAllowlistedWSSRequestWithNativeOrigin() throws {
        let configuration = try SSCPluginConfiguration(
            allowedOrigins: ["wss://api.example.com"],
            originHeader: "capacitor://localhost"
        )

        let request = try configuration.request(endpoint: "wss://api.example.com/ssc?tenant=one")

        XCTAssertEqual(request.url?.absoluteString, "wss://api.example.com/ssc?tenant=one")
        XCTAssertEqual(request.value(forHTTPHeaderField: "Origin"), "capacitor://localhost")
    }

    func testRejectsEndpointOutsideExactOriginAllowlist() throws {
        let configuration = try SSCPluginConfiguration(
            allowedOrigins: ["wss://api.example.com"]
        )

        XCTAssertThrowsError(try configuration.request(endpoint: "wss://evil.example/ssc")) { error in
            XCTAssertEqual(error as? SSCConfigurationError, .endpointNotAllowed)
        }
        XCTAssertThrowsError(try configuration.request(endpoint: "wss://api.example.com.evil.invalid/ssc")) { error in
            XCTAssertEqual(error as? SSCConfigurationError, .endpointNotAllowed)
        }
    }

    func testRejectsCredentialsAndInsecureRemoteEndpoints() throws {
        let configuration = try SSCPluginConfiguration(
            allowedOrigins: ["wss://api.example.com", "ws://api.example.com"]
        )

        XCTAssertThrowsError(try configuration.request(endpoint: "wss://user:pass@api.example.com/ssc")) { error in
            XCTAssertEqual(error as? SSCConfigurationError, .invalidEndpoint)
        }
        XCTAssertThrowsError(try configuration.request(endpoint: "ws://api.example.com/ssc")) { error in
            XCTAssertEqual(error as? SSCConfigurationError, .insecureEndpoint)
        }
    }

    func testInsecureLoopbackRequiresExplicitOptIn() throws {
        let denied = try SSCPluginConfiguration(allowedOrigins: ["ws://127.0.0.1:8080"])
        XCTAssertThrowsError(try denied.request(endpoint: "ws://127.0.0.1:8080/ws")) { error in
            XCTAssertEqual(error as? SSCConfigurationError, .insecureEndpoint)
        }

        let allowed = try SSCPluginConfiguration(
            allowedOrigins: ["ws://127.0.0.1:8080"],
            allowInsecureLocalhost: true
        )
        XCTAssertNoThrow(try allowed.request(endpoint: "ws://127.0.0.1:8080/ws"))
    }

    func testRequiresAWellFormedNativeConfiguration() {
        XCTAssertThrowsError(try SSCPluginConfiguration(allowedOrigins: [])) { error in
            XCTAssertEqual(error as? SSCConfigurationError, .missingAllowedOrigins)
        }
        XCTAssertThrowsError(
            try SSCPluginConfiguration(
                allowedOrigins: ["wss://api.example.com/path"]
            )
        ) { error in
            XCTAssertEqual(
                error as? SSCConfigurationError,
                .invalidAllowedOrigin("wss://api.example.com/path")
            )
        }
        XCTAssertThrowsError(
            try SSCPluginConfiguration(
                allowedOrigins: ["wss://api.example.com"],
                originHeader: "capacitor://localhost/path"
            )
        ) { error in
            XCTAssertEqual(error as? SSCConfigurationError, .invalidOriginHeader)
        }
    }

    func testRejectsUnsafeQueueAndFrameLimits() {
        XCTAssertThrowsError(
            try SSCPluginConfiguration(
                allowedOrigins: ["wss://api.example.com"],
                outboundQueueSize: 0
            )
        ) { error in
            XCTAssertEqual(error as? SSCConfigurationError, .invalidOutboundQueueSize)
        }
        XCTAssertThrowsError(
            try SSCPluginConfiguration(
                allowedOrigins: ["wss://api.example.com"],
                maxMessageBytes: 17 * 1024 * 1024
            )
        ) { error in
            XCTAssertEqual(error as? SSCConfigurationError, .invalidMessageLimit)
        }
    }
}
