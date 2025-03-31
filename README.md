# Dotfiles

Этот репозиторий содержит конфигурационные файлы (dotfiles) для настройки окружения.

## Список файлов

- `.zshrc` — основной конфиг для Zsh.
- `alacritty.toml` — конфигурация для Alacritty.
- `aliases.zsh` — пользовательские алиасы для Zsh.
- `get_keyboard_layout_linux.sh` — скрипт для определения текущей раскладки клавиатуры в Linux.
- `get_keyboard_layout_mac.sh` — скрипт для определения текущей раскладки клавиатуры в macOS.
- `tmux.conf` — конфигурация для Tmux.
- `wezterm.lua` — конфигурация для WezTerm.

## Пути установки

Для корректной работы файлы должны находиться в следующих местах:

```
~/.config/alacritty/alacritty.toml
~/.wezterm.lua
~/.tmux.conf
~/.oh-my-zsh/aliases/aliases.zsh
```

## Установка

1. Клонируй репозиторий:
   ```bash
   git clone https://github.com/your-username/dotfiles.git ~/dotfiles
   ```
2. Создай символические ссылки на файлы:
   ```bash
   ln -s ~/dotfiles/.zshrc ~/.zshrc
   ln -s ~/dotfiles/alacritty.toml ~/.config/alacritty/alacritty.toml
   ln -s ~/dotfiles/aliases.zsh ~/.oh-my-zsh/aliases/aliases.zsh
   ln -s ~/dotfiles/tmux.conf ~/.tmux.conf
   ln -s ~/dotfiles/wezterm.lua ~/.wezterm.lua
   ```

## Дополнительно

- Для работы `.zshrc` может потребоваться [Oh My Zsh](https://ohmyz.sh/).
- Для `tmux.conf` рекомендуется [Tmux Plugin Manager](https://github.com/tmux-plugins/tpm).
- Для `wezterm.lua` требуется [WezTerm](https://wezfurlong.org/wezterm/).

## Лицензия

Этот репозиторий распространяется под лицензией MIT. См. файл [LICENSE](LICENSE) для деталей.

