local wezterm = require("wezterm")
local act = wezterm.action
local mux = wezterm.mux
local config = wezterm.config_builder()

-- Read the same theme source of truth used by tmux/ghostty/alacritty.
local function read_theme_mode()
	local home = wezterm.home_dir
	local f = io.open(home .. "/.config/theme/mode", "r")
	if not f then
		return "dark"
	end
	local mode = f:read("*l") or "dark"
	f:close()
	return mode:match("^%s*(.-)%s*$") == "light" and "light" or "dark"
end

local theme_mode = read_theme_mode()
config.color_scheme = (theme_mode == "light") and "tokyonight_day" or "tokyonight_moon"

-- Non-native fullscreen on macOS (matches Ghostty's non-native fullscreen).
config.native_macos_fullscreen_mode = false

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
config.macos_window_background_blur = 50
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

-- Cursor (matches Ghostty's cursor-style = block, no blink, color #7aa2f7).
config.default_cursor_style = "SteadyBlock"
config.colors = {
	cursor_bg = "#7aa2f7",
	cursor_border = "#7aa2f7",
}

-- Key bindings
config.keys = {
	{ key = "c", mods = "CMD", action = act.CopyTo("Clipboard") },
	{ key = "v", mods = "CMD", action = act.PasteFrom("Clipboard") },
	{ key = "Enter", mods = "SHIFT", action = act.SendString("\x1b[13;2u") },
}

return config
