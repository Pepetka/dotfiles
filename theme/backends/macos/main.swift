import AppKit
import CoreFoundation
import Foundation

// A LaunchAgent's effectiveAppearance may differ from the desktop. Read the
// global preference, refreshed before each observation.
func systemMode() -> String {
  CFPreferencesSynchronize(
    kCFPreferencesAnyApplication, kCFPreferencesCurrentUser, kCFPreferencesAnyHost)
  let style = CFPreferencesCopyValue(
    "AppleInterfaceStyle" as CFString, kCFPreferencesAnyApplication,
    kCFPreferencesCurrentUser, kCFPreferencesAnyHost)
  return (style as? String) == "Dark" ? "dark" : "light"
}

@main
struct MacOSBackend {
  static func main() {
    switch CommandLine.arguments.dropFirst().first {
    case "read":
      print(systemMode())
    case "watch":
      var previous = ""
      let publish = {
        let mode = systemMode()
        if mode != previous {
          previous = mode
          FileHandle.standardOutput.write(Data((mode + "\n").utf8))
        }
      }
      publish()
      let changed = { (_: Notification) in
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.1, execute: publish)
      }
      let appearance = DistributedNotificationCenter.default().addObserver(
        forName: Notification.Name("AppleInterfaceThemeChangedNotification"),
        object: nil, queue: .main, using: changed)
      let workspace = NSWorkspace.shared.notificationCenter
      let wake = workspace.addObserver(
        forName: NSWorkspace.didWakeNotification, object: nil, queue: .main, using: changed)
      let session = workspace.addObserver(
        forName: NSWorkspace.sessionDidBecomeActiveNotification,
        object: nil, queue: .main, using: changed)
      let timer = Timer.scheduledTimer(withTimeInterval: 1, repeats: true) { _ in publish() }
      withExtendedLifetime([appearance, wake, session, timer]) {
        NSApplication.shared.setActivationPolicy(.prohibited)
        NSApplication.shared.run()
      }
    default:
      FileHandle.standardError.write(Data("Usage: macos-backend {read|watch}\n".utf8))
      exit(1)
    }
  }
}
