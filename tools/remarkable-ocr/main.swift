import Foundation
import Vision
import AppKit

struct OCRResult: Codable {
    let text: String
    let x: Double
    let y: Double
    let width: Double
    let height: Double
    let confidence: Float
}

guard CommandLine.arguments.count > 1 else {
    fputs("Usage: remarkable-ocr <image-path>\n", stderr)
    exit(1)
}

let imagePath = CommandLine.arguments[1]
guard let image = NSImage(contentsOfFile: imagePath),
      let cgImage = image.cgImage(forProposedRect: nil, context: nil, hints: nil) else {
    fputs("Error: could not load image at \(imagePath)\n", stderr)
    exit(1)
}

let request = VNRecognizeTextRequest()
request.recognitionLevel = .accurate
request.usesLanguageCorrection = true

let handler = VNImageRequestHandler(cgImage: cgImage, options: [:])
try handler.perform([request])

guard let observations = request.results else {
    print("[]")
    exit(0)
}

let imageHeight = Double(cgImage.height)
let imageWidth = Double(cgImage.width)

var results: [OCRResult] = []
for observation in observations {
    guard let candidate = observation.topCandidates(1).first else { continue }
    let box = observation.boundingBox
    results.append(OCRResult(
        text: candidate.string,
        x: box.origin.x * imageWidth,
        y: (1 - box.origin.y - box.height) * imageHeight,
        width: box.width * imageWidth,
        height: box.height * imageHeight,
        confidence: candidate.confidence
    ))
}

let encoder = JSONEncoder()
encoder.outputFormatting = .prettyPrinted
let data = try encoder.encode(results)
print(String(data: data, encoding: .utf8)!)
