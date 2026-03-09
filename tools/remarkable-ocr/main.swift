import Foundation
import Vision
import AppKit
import CoreImage

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
    fputs("Usage: remarkable-ocr <image-path> [--custom-words <file>]\n", stderr)
    exit(1)
}

let imagePath = CommandLine.arguments[1]
guard let image = NSImage(contentsOfFile: imagePath),
      let cgImage = image.cgImage(forProposedRect: nil, context: nil, hints: nil) else {
    fputs("Error: could not load image at \(imagePath)\n", stderr)
    exit(1)
}

var customWords: [String] = []
if let idx = CommandLine.arguments.firstIndex(of: "--custom-words"), idx + 1 < CommandLine.arguments.count {
    let wordsPath = CommandLine.arguments[idx + 1]
    if let content = try? String(contentsOfFile: wordsPath, encoding: .utf8) {
        customWords = content.components(separatedBy: .newlines).filter { !$0.isEmpty }
    }
}

let processedImage = cgImage

let request = VNRecognizeTextRequest()
request.recognitionLevel = .accurate
request.revision = VNRecognizeTextRequestRevision3
request.usesLanguageCorrection = true
request.recognitionLanguages = ["en"]
request.minimumTextHeight = 0.02
if !customWords.isEmpty {
    request.customWords = customWords
}

let handler = VNImageRequestHandler(cgImage: processedImage, options: [:])
try handler.perform([request])

guard let observations = request.results else {
    print("[]")
    exit(0)
}

let imageHeight = Double(processedImage.height)
let imageWidth = Double(processedImage.width)

var results: [OCRResult] = []
for observation in observations {
    let topCandidates = observation.topCandidates(10)
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
