// swift-tools-version: 5.9

import PackageDescription

let package = Package(
    name: "RfwlabCapacitorSsc",
    platforms: [.iOS(.v15)],
    products: [
        .library(name: "RfwlabCapacitorSsc", targets: ["RFWSSCPlugin"])
    ],
    dependencies: [
        .package(url: "https://github.com/ionic-team/capacitor-swift-pm.git", from: "8.0.0")
    ],
    targets: [
        .target(
            name: "RFWSSCPlugin",
            dependencies: [
                .product(name: "Capacitor", package: "capacitor-swift-pm"),
                .product(name: "Cordova", package: "capacitor-swift-pm")
            ],
            path: "ios/Sources/RFWSSCPlugin"
        ),
        .testTarget(
            name: "RFWSSCPluginTests",
            dependencies: ["RFWSSCPlugin"],
            path: "ios/Tests/RFWSSCPluginTests"
        )
    ]
)
