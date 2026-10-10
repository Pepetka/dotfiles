package adapters

import "dotfiles/theme/internal/engine"

// Adding an application only requires an adapter and registration here. The
// engine's policy, watcher and publisher have no application-specific branches.
func All() []engine.Adapter {
	return []engine.Adapter{Neovim{}, Palette{"alacritty", "toml"}, WezTerm{}, Ghostty{}, Tmux{}, Prompt{}}
}
