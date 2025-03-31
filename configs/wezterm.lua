local wezterm = require("wezterm")
local act = wezterm.action
local config = wezterm.config_builder()

local function create_tab_switch_bindings()
	local keys = {}

	for i = 1, 9 do
		table.insert(keys, {
			key = tostring(i),
			mods = "LEADER",
			action = wezterm.action.ActivateTab(i - 1),
		})
	end

	return keys
end

local themes = { "Drakula", "Catppuccin Mocha" }

local function toggle_theme(window)
	local overrides = window:get_config_overrides() or {}
	local current_theme = overrides.color_scheme or wezterm.config_builder().color_scheme or themes[1]
	local next_theme = (current_theme == themes[1]) and themes[2] or themes[1]

	window:set_config_overrides({
		color_scheme = next_theme,
	})
end

wezterm.on("toggle-theme", function(window)
	toggle_theme(window)
end)

config.window_padding = {
	left = 0,
	right = 0,
	top = 30,
	bottom = 0,
}

config.default_prog = { "C:/Program Files/Git/bin/bash.exe" }
config.font = wezterm.font("JetBrainsMono Nerd Font")
config.freetype_load_target = "Light"
config.freetype_render_target = "Normal"
config.harfbuzz_features = { "calt=0", "liga=0", "size=120" }
config.font_size = 12

config.window_decorations = "RESIZE"
config.color_scheme = themes[1]

config.window_background_opacity = 0.75
config.macos_window_background_blur = 50

config.initial_rows = 40
config.initial_cols = 140

wezterm.on("update-right-status", function(window, _)
	local key = window:active_key_table()
	local leader = window:leader_is_active()

	if key then
		key = "TABLE: " .. key .. " | "
	elseif leader then
		key = "LEADER | "
	else
		key = ""
	end

	local time = wezterm.strftime("%H:%M")
	window:set_right_status(wezterm.format({
		{ Text = key .. time },
	}))
end)

config.status_update_interval = 1000
config.use_fancy_tab_bar = false
config.hide_mouse_cursor_when_typing = true

config.leader = { key = "a", mods = "CTRL", timeout_milliseconds = 10000 }
config.keys = {
	{
		key = "c",
		mods = "CTRL",
		action = act.CopyTo("Clipboard"),
	},
	{
		key = "v",
		mods = "CTRL",
		action = act.PasteFrom("Clipboard"),
	},
	{
		key = "\\",
		mods = "LEADER",
		action = act.SplitHorizontal({ domain = "CurrentPaneDomain" }),
		timeout_milliseconds = 1000,
	},
	{
		key = "-",
		mods = "LEADER",
		action = act.SplitVertical({ domain = "CurrentPaneDomain" }),
		timeout_milliseconds = 1000,
	},
	{
		key = "r",
		mods = "LEADER",
		action = act.ActivateKeyTable({
			name = "resize_pane",
			one_shot = false,
		}),
	},
	{
		key = "a",
		mods = "LEADER",
		action = act.ActivateKeyTable({
			name = "activate_pane",
			one_shot = false,
		}),
	},
	{
		key = "r",
		mods = "CTRL|ALT",
		action = wezterm.action.PromptInputLine({
			description = "Введите новое имя для вкладки",
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
		action = wezterm.action.PromptInputLine({
			description = "Введите название вкладки",
			action = wezterm.action_callback(function(window, pane, line)
				window:perform_action(wezterm.action.SpawnTab("DefaultDomain"), pane)

				if line then
					window:active_tab():set_title(line)
				end
			end),
		}),
	},
	{
		key = "x",
		mods = "LEADER",
		action = wezterm.action.CloseCurrentTab({ confirm = true }),
	},
	{
		key = "n",
		mods = "LEADER",
		action = wezterm.action.ActivateTabRelative(1),
	},
	{
		key = "p",
		mods = "LEADER",
		action = wezterm.action.ActivateTabRelative(-1),
	},
	{
		key = "t",
		mods = "CTRL|SHIFT",
		action = wezterm.action.EmitEvent("toggle-theme"),
	},
	table.unpack(create_tab_switch_bindings()),
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

		{ key = "Escape", action = "PopKeyTable" },
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

		{ key = "Escape", action = "PopKeyTable" },
	},
}

if wezterm.target_triple:find("apple") then
	config.enable_tab_bar = false
	config.disable_default_key_bindings = true
	config.leader = nil
	config.key_tables = {}
	config.keys = {
		{
			key = "Enter",
			mods = "ALT",
			action = wezterm.action.ToggleFullScreen,
		},
		{
			key = "c",
			mods = "CMD",
			action = act.CopyTo("Clipboard"),
		},
		{
			key = "v",
			mods = "CMD",
			action = act.PasteFrom("Clipboard"),
		},
		{
			key = "t",
			mods = "CMD|SHIFT",
			action = wezterm.action.EmitEvent("toggle-theme"),
		},
	}
	config.font_size = 14
	config.default_prog = nil
end

return config
