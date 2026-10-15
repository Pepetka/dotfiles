package adapters

import (
	"context"
	"dotfiles/theme/internal/engine"
	"fmt"
	"os"
)

type Palette struct{ Application, Extension string }

func (p Palette) Name() string { return p.Application }

func style(mode engine.Mode) string {
	if mode == engine.Light {
		return "day"
	}
	return "moon"
}

func palette(c engine.Config, mode engine.Mode, application, extension, name string) ([]engine.Artifact, error) {
	data, err := os.ReadFile(c.Resource(application, "tokyonight-"+style(mode)+"."+extension))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty %s palette", application)
	}
	return []engine.Artifact{{Name: name, Data: data}}, nil
}

func (p Palette) Render(c engine.Config, r engine.RenderRequest) ([]engine.Artifact, error) {
	return palette(c, r.Mode, p.Application, p.Extension, p.Application+"-theme."+p.Extension)
}

func (p Palette) Apply(_ context.Context, _ engine.Config, r engine.ApplyRequest) (engine.AdapterState, error) {
	return engine.AdapterState{State: "published", Mode: r.Mode}, nil
}

type Neovim struct{}

func (Neovim) Name() string                                                          { return "nvim" }
func (Neovim) Render(engine.Config, engine.RenderRequest) ([]engine.Artifact, error) { return nil, nil }
func (Neovim) Apply(_ context.Context, _ engine.Config, r engine.ApplyRequest) (engine.AdapterState, error) {
	return engine.AdapterState{State: "published", Mode: r.Mode}, nil
}

type WezTerm struct{}

func (WezTerm) Name() string { return "wezterm" }
func (WezTerm) Render(_ engine.Config, r engine.RenderRequest) ([]engine.Artifact, error) {
	return []engine.Artifact{{Name: "wezterm-theme.lua", Data: []byte("return { color_scheme = \"tokyonight_" + style(r.Mode) + "\" }\n")}}, nil
}

func (WezTerm) Apply(_ context.Context, _ engine.Config, r engine.ApplyRequest) (engine.AdapterState, error) {
	return engine.AdapterState{State: "published", Mode: r.Mode}, nil
}

// No mode-dependent prompt configuration: emitted base ANSI colors follow the
// terminal palette. Zsh sources it directly from the repository.
type Prompt struct{}

func (Prompt) Name() string                                                          { return "p10k" }
func (Prompt) Render(engine.Config, engine.RenderRequest) ([]engine.Artifact, error) { return nil, nil }
func (Prompt) Apply(_ context.Context, _ engine.Config, r engine.ApplyRequest) (engine.AdapterState, error) {
	return engine.AdapterState{State: "ansi", Mode: r.Mode}, nil
}
