-- Exercise platform selection without a graphical session.
local root = assert(arg[1])
local temporary = assert(arg[2])
local getenv = os.getenv
os.getenv = function(name)
  if name == "XDG_CONFIG_HOME" then return temporary end
  return getenv(name)
end
for _, platform in ipairs({ "x86_64-unknown-linux-gnu", "aarch64-apple-darwin" }) do
  local wezterm = {
    target_triple = platform,
    home_dir = temporary,
    config_builder = function() return {} end,
    font = function(name) return name end,
    on = function() end,
    add_to_config_reload_watch_list = function() end,
    mux = {},
    action = {
      CopyTo = function(value) return value end,
      PasteFrom = function(value) return value end,
      SendString = function(value) return value end,
    },
  }
  package.loaded.wezterm = wezterm
  local config = dofile(root .. "/configs/wezterm/wezterm.lua")
  local macos = platform:find("apple%-darwin") ~= nil
  assert(config.keys[1].mods == (macos and "CMD" or "CTRL|SHIFT"))
  assert(config.keys[2].mods == config.keys[1].mods)
  assert(config.keys[3].mods == "SHIFT")
  assert(config.macos_window_background_blur == (macos and 50 or nil))
  if macos then
    assert(config.native_macos_fullscreen_mode == false)
  else
    assert(config.native_macos_fullscreen_mode == nil)
  end
end
print("WezTerm config passed: Linux/macOS options and clipboard shortcuts")
