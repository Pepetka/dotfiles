import AppKit
import Carbon
import Foundation

let ghosttyID = "com.mitchellh.ghostty"

func ghosttyApplications() -> [NSRunningApplication] {
  NSRunningApplication.runningApplications(withBundleIdentifier: ghosttyID)
}

func permission(prompt: Bool) -> OSStatus {
  var target = AEAddressDesc()
  let bytes = Array(ghosttyID.utf8)
  let result = bytes.withUnsafeBytes {
    AECreateDesc(DescType(typeApplicationBundleID), $0.baseAddress, $0.count, &target)
  }
  guard result == noErr else { return OSStatus(result) }
  defer { AEDisposeDesc(&target) }
  return AEDeterminePermissionToAutomateTarget(
    &target, AEEventClass(typeWildCard), AEEventID(typeWildCard), prompt)
}

func reply(_ path: String, state: String, error: String? = nil) {
  var result = ["state": state]
  if let error { result["error"] = error }
  do {
    let data = try JSONSerialization.data(withJSONObject: result, options: [.sortedKeys])
    try data.write(to: URL(fileURLWithPath: path), options: .atomic)
  } catch {
    FileHandle.standardError.write(Data("Ghostty reply: \(error)\n".utf8))
  }
}

@main
struct MacOSPlatform {
  static func main() {
    let args = Array(CommandLine.arguments.dropFirst())
    guard let command = args.first else { exit(1) }
    if command == "ghostty-state" {
      let identity = ghosttyApplications().map { String($0.processIdentifier) }.sorted().joined(separator: ",")
      print(identity)
      return
    }
    guard args.count == 2,
      command == "ghostty-reload" || command == "ghostty-authorize"
    else { exit(1) }
    let path = args[1]
    guard !ghosttyApplications().isEmpty else {
      reply(path, state: "not running")
      return
    }
    let result = permission(prompt: command == "ghostty-authorize")
    guard result == noErr else {
      reply(
        path, state: "permission required",
        error: "Ghostty Automation permission unavailable (\(result)). Run theme authorize with Ghostty open; check System Settings > Privacy & Security > Automation.")
      return
    }
    if command == "ghostty-authorize" {
      reply(path, state: "authorized")
      return
    }
    // No activate/focus/select action. Reload is app-wide; target one existing
    // surface and retry next cycle if the app is between window transitions.
    let source = """
      with timeout of 2 seconds
        tell application id "\(ghosttyID)"
          set surfaces to terminals
          if (count of surfaces) is 0 then return false
          return perform action "reload_config" on item 1 of surfaces
        end tell
      end timeout
      """
    var scriptError: NSDictionary?
    let output = NSAppleScript(source: source)?.executeAndReturnError(&scriptError)
    if let scriptError {
      reply(path, state: "error", error: "Ghostty reload: \(scriptError)")
    } else if output?.booleanValue == true {
      reply(path, state: "reload requested")
    } else {
      reply(path, state: "error", error: "Ghostty has no available terminal or rejected reload_config")
    }
  }
}
