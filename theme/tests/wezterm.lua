-- Test bridge behavior without a GUI or a WezTerm-specific Lua runtime.
local source = assert(arg[1])
local temporary = assert(arg[2])
local env = { XDG_CONFIG_HOME = temporary }
os.getenv = function(name)
	return env[name]
end
local watched = {}
local wezterm = {
	home_dir = temporary,
	add_to_config_reload_watch_list = function(path)
		watched[#watched + 1] = path
	end,
}
local path = temporary .. "/theme/wezterm-theme.lua"
local function publish(style)
	local file = assert(io.open(path .. ".next", "w"))
	file:write('return { color_scheme = "tokyonight_' .. style .. '" }\n')
	file:close()
	assert(os.rename(path .. ".next", path))
end
local bridge = dofile(source .. "/wezterm.lua")
local config = {}
bridge(wezterm, config)
assert(config.color_scheme == "tokyonight_moon" and config.automatically_reload_config)
publish("day")
bridge(wezterm, config)
assert(config.color_scheme == "tokyonight_day")
assert(config.colors.cursor_bg == "#6172b0" and config.colors.cursor_fg == "#e1e2e7")
assert(config.colors.cursor_border == config.colors.cursor_bg)
publish("moon")
bridge(wezterm, config)
assert(config.color_scheme == "tokyonight_moon", "Bridge retained a cached require result")
assert(config.colors == nil, "Day cursor override survived a switch to Moon")
assert(watched[1] == path)
print("WezTerm bridge passed: missing state, XDG/spaces and atomic replacements")
