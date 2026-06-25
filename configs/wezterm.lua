local wezterm = require("wezterm")
local act = wezterm.action
local config = wezterm.config_builder()

-- Theme
local themes = { "Dracula", "Catppuccin Mocha" }

-- Helpers
local function tab_switch_bindings()
	local keys = {}
	for i = 1, 9 do
		table.insert(keys, {
			key = tostring(i),
			mods = "LEADER",
			action = act.ActivateTab(i - 1),
		})
	end
	return keys
end

local function toggle_theme(window)
	local overrides = window:get_config_overrides() or {}
	local current = overrides.color_scheme or config.color_scheme
	local next_theme = (current == themes[1]) and themes[2] or themes[1]
	window:set_config_overrides({ color_scheme = next_theme })
end

wezterm.on("toggle-theme", toggle_theme)

wezterm.on("update-right-status", function(window, _)
	local key = window:active_key_table()
	local leader = window:leader_is_active()
	local prefix = key and ("TABLE: " .. key .. " | ") or (leader and "LEADER | " or "")
	local time = wezterm.strftime("%H:%M")
	window:set_right_status(wezterm.format({ { Text = prefix .. time } }))
end)

-- Appearance
config.color_scheme = themes[1]
config.font = wezterm.font("JetBrainsMono Nerd Font")
config.font_size = 12
config.freetype_load_target = "Light"
config.freetype_render_target = "Normal"
config.harfbuzz_features = { "calt=0", "liga=0" }
config.window_padding = { left = 0, right = 0, top = 30, bottom = 0 }
config.window_background_opacity = 0.75
config.macos_window_background_blur = 50
config.window_decorations = "RESIZE"
config.use_fancy_tab_bar = false
config.hide_mouse_cursor_when_typing = true

-- Behavior
config.initial_rows = 40
config.initial_cols = 140

if wezterm.target_triple:find("windows") then
	config.default_prog = { "C:/Program Files/Git/bin/bash.exe" }
end

-- Key bindings
config.leader = { key = "a", mods = "CTRL", timeout_milliseconds = 10000 }

config.keys = {
	{ key = "c", mods = "CTRL", action = act.CopyTo("Clipboard") },
	{ key = "v", mods = "CTRL", action = act.PasteFrom("Clipboard") },
	{ key = "\\", mods = "LEADER", action = act.SplitHorizontal({ domain = "CurrentPaneDomain" }) },
	{ key = "-", mods = "LEADER", action = act.SplitVertical({ domain = "CurrentPaneDomain" }) },
	{ key = "r", mods = "LEADER", action = act.ActivateKeyTable({ name = "resize_pane", one_shot = false }) },
	{ key = "a", mods = "LEADER", action = act.ActivateKeyTable({ name = "activate_pane", one_shot = false }) },
	{
		key = "r",
		mods = "CTRL|ALT",
		action = act.PromptInputLine({
			description = "Rename tab:",
			action = wezterm.action_callback(function(window, _, line)
				if line then
					window:active_tab():set_title(line)
				end
			end),
		}),
	},
	{
		key = "c",
		mods = "LEADER",
		action = act.PromptInputLine({
			description = "New tab name:",
			action = wezterm.action_callback(function(window, pane, line)
				window:perform_action(act.SpawnTab("DefaultDomain"), pane)
				if line then
					window:active_tab():set_title(line)
				end
			end),
		}),
	},
	{ key = "x", mods = "LEADER", action = act.CloseCurrentTab({ confirm = true }) },
	{ key = "n", mods = "LEADER", action = act.ActivateTabRelative(1) },
	{ key = "p", mods = "LEADER", action = act.ActivateTabRelative(-1) },
	{ key = "t", mods = "CTRL|SHIFT", action = act.EmitEvent("toggle-theme") },
	table.unpack(tab_switch_bindings()),
}

config.key_tables = {
	resize_pane = {
		{ key = "LeftArrow", action = act.AdjustPaneSize({ "Left", 5 }) },
		{ key = "h", action = act.AdjustPaneSize({ "Left", 5 }) },
		{ key = "RightArrow", action = act.AdjustPaneSize({ "Right", 5 }) },
		{ key = "l", action = act.AdjustPaneSize({ "Right", 5 }) },
		{ key = "UpArrow", action = act.AdjustPaneSize({ "Up", 5 }) },
		{ key = "k", action = act.AdjustPaneSize({ "Up", 5 }) },
		{ key = "DownArrow", action = act.AdjustPaneSize({ "Down", 5 }) },
		{ key = "j", action = act.AdjustPaneSize({ "Down", 5 }) },
		{ key = "Escape", action = act.PopKeyTable },
	},
	activate_pane = {
		{ key = "LeftArrow", action = act.ActivatePaneDirection("Left") },
		{ key = "h", action = act.ActivatePaneDirection("Left") },
		{ key = "RightArrow", action = act.ActivatePaneDirection("Right") },
		{ key = "l", action = act.ActivatePaneDirection("Right") },
		{ key = "UpArrow", action = act.ActivatePaneDirection("Up") },
		{ key = "k", action = act.ActivatePaneDirection("Up") },
		{ key = "DownArrow", action = act.ActivatePaneDirection("Down") },
		{ key = "j", action = act.ActivatePaneDirection("Down") },
		{ key = "Escape", action = act.PopKeyTable },
	},
}

-- macOS overrides
if wezterm.target_triple:find("apple") then
	config.enable_tab_bar = false
	config.disable_default_key_bindings = true
	config.leader = nil
	config.key_tables = {}
	config.keys = {
		{ key = "Enter", mods = "ALT", action = act.ToggleFullScreen },
		{ key = "c", mods = "CMD", action = act.CopyTo("Clipboard") },
		{ key = "v", mods = "CMD", action = act.PasteFrom("Clipboard") },
		{ key = "t", mods = "CMD|SHIFT", action = act.EmitEvent("toggle-theme") },
	}
	config.font_size = 14
	config.default_prog = nil
end

return config
