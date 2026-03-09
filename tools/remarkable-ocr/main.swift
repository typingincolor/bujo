import Foundation
import Vision
import AppKit

struct OCRCandidate: Codable {
    let text: String
    let confidence: Float
}

struct OCRResult: Codable {
    let text: String
    let x: Double
    let y: Double
    let width: Double
    let height: Double
    let confidence: Float
    let candidates: [OCRCandidate]?
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
request.recognitionLanguages = ["en"]

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
    let topCandidates = observation.topCandidates(5)
    guard let best = topCandidates.first else { continue }
    let box = observation.boundingBox

    let candidates: [OCRCandidate]? = topCandidates.count > 1 ? topCandidates.map {
        OCRCandidate(text: $0.string, confidence: $0.confidence)
    } : nil

    results.append(OCRResult(
        text: best.string,
        x: box.origin.x * imageWidth,
        y: (1 - box.origin.y - box.height) * imageHeight,
        width: box.width * imageWidth,
        height: box.height * imageHeight,
        confidence: best.confidence,
        candidates: candidates
    ))
}

let encoder = JSONEncoder()
encoder.outputFormatting = .prettyPrinted
let data = try encoder.encode(results)
print(String(data: data, encoding: .utf8)!)
