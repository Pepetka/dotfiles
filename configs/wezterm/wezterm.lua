local wezterm = require("wezterm")
local act = wezterm.action
local mux = wezterm.mux
local config = wezterm.config_builder()
local is_macos = wezterm.target_triple:find("apple%-darwin") ~= nil
local clipboard_mods = is_macos and "CMD" or "CTRL|SHIFT"

-- The theme engine installs this bridge under the shared config root.
local config_root = os.getenv("XDG_CONFIG_HOME")
if not config_root or config_root == "" then
	config_root = wezterm.home_dir .. "/.config"
end
local theme_bridge = config_root .. "/theme/wezterm-bridge.lua"
wezterm.add_to_config_reload_watch_list(theme_bridge)
local load_theme = loadfile(theme_bridge)
if load_theme then
	load_theme()(wezterm, config)
end

-- Non-native fullscreen on macOS (matches Ghostty's non-native fullscreen).
if is_macos then
	config.native_macos_fullscreen_mode = false
	config.macos_window_background_blur = 50
end

-- Start in fullscreen.
wezterm.on("gui-startup", function(cmd)
	local _, _, window = mux.spawn_window(cmd or {})
	window:gui_window():toggle_fullscreen()
end)

-- Appearance
config.font = wezterm.font("JetBrainsMono Nerd Font")
config.font_size = 15
config.freetype_load_target = "Light"
config.freetype_render_target = "Normal"
config.harfbuzz_features = { "calt=1", "liga=1", "zero=1" }
config.window_padding = { left = 0, right = 0, top = 20, bottom = 0 }
config.window_background_opacity = 0.85
config.window_decorations = "NONE"
config.enable_tab_bar = false
config.use_fancy_tab_bar = false
config.hide_mouse_cursor_when_typing = true

-- Behavior
config.disable_default_key_bindings = true
config.window_close_confirmation = "NeverPrompt"
config.quit_when_all_windows_are_closed = true

-- Treat Option as Alt on both sides (matches Ghostty's macos-option-as-alt).
config.send_composed_key_when_left_alt_is_pressed = false
config.send_composed_key_when_right_alt_is_pressed = false

-- Cursor colors come from the current scheme.
config.default_cursor_style = "SteadyBlock"

-- Key bindings
config.keys = {
	{ key = "c", mods = clipboard_mods, action = act.CopyTo("Clipboard") },
	{ key = "v", mods = clipboard_mods, action = act.PasteFrom("Clipboard") },
	{ key = "Enter", mods = "SHIFT", action = act.SendString("\x1b[13;2u") },
}

return config
