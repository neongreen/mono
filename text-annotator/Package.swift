// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "TextAnnotator",
    platforms: [
        .macOS(.v12)
    ],
    products: [
        .executable(
            name: "TextAnnotator",
            targets: ["TextAnnotator"]
        )
    ],
    targets: [
        .executableTarget(
            name: "TextAnnotator",
            path: "Sources"
        )
    ]
)
