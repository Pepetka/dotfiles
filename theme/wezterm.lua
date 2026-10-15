-- A bridge only: policy is resolved by the engine, never by the GUI appearance.
return function(wezterm, config)
	local root = os.getenv("XDG_CONFIG_HOME")
	if not root or root == "" then
		root = (os.getenv("HOME") or wezterm.home_dir) .. "/.config"
	end
	local path = root .. "/theme/wezterm-theme.lua"
	wezterm.add_to_config_reload_watch_list(path)
	config.automatically_reload_config = true
	local ok, theme = pcall(dofile, path)
	if ok and type(theme) == "table" and type(theme.color_scheme) == "string" then
		config.color_scheme = theme.color_scheme
	else
		config.color_scheme = "tokyonight_moon"
	end
	-- Keep the Day cursor visible, including applications using a thin bar.
	if config.color_scheme == "tokyonight_day" then
		config.colors = {
			cursor_bg = "#3760bf",
			cursor_fg = "#e1e2e7",
			cursor_border = "#3760bf",
		}
	else
		config.colors = nil
	end
end
