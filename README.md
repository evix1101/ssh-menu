# SSH Menu

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/evix1101/ssh-menu)](https://goreportcard.com/report/github.com/evix1101/ssh-menu)

A beautiful terminal-based SSH connection manager that reads your SSH config and presents it as an interactive menu with filtering and group organization.

## Features

- 🚀 **Filtering**: Type to instantly filter hosts by name or number
- 📋 **Intuitive Navigation**: Arrow keys to navigate, left/right to switch between views
- 🗂️ **Group Organization**: Organize hosts into groups 
- 📂 **Config Integration**: Reads from `~/.ssh/config` and `~/.ssh/config.d/`
- 🎯 **Quick Selection**: Type host numbers or names for instant filtering
- 🌈 **Customizable Themes**: Full color customization support
- 🔒 **Secure**: Uses native SSH

## Screenshot

![SSH Menu Demo](demo.svg)

## Installation

### From Source

Requires Go 1.21 or later.

```bash
go install github.com/evix1101/ssh-menu@latest
```

### Binary Releases

Download the latest release for your platform from the [Releases page](https://github.com/evix1101/ssh-menu/releases).

Each release contains the binary named `ssh-menu` (or `ssh-menu.exe` on Windows) in an archive:
- `ssh-menu-linux-amd64.tar.gz` - Linux x86_64
- `ssh-menu-darwin-amd64.tar.gz` - macOS Intel
- `ssh-menu-darwin-arm64.tar.gz` - macOS Apple Silicon
- `ssh-menu-windows-amd64.zip` - Windows x86_64

#### Installation from release

```bash
# Linux/macOS - extract and install
tar -xzf ssh-menu-*.tar.gz
chmod +x ssh-menu
sudo mv ssh-menu /usr/local/bin/

# Or install to user directory
mkdir -p ~/.local/bin
mv ssh-menu ~/.local/bin/
# Add ~/.local/bin to your PATH if not already there
```

#### Security warnings

Since the binaries are not signed, you may encounter security warnings:

**macOS**: When you first run ssh-menu, you'll see "ssh-menu cannot be opened because it is from an unidentified developer"
- Go to System Settings → Privacy & Security
- Find the message about ssh-menu being blocked
- Click "Open Anyway"
- Or, from Terminal: `xattr -d com.apple.quarantine ssh-menu-darwin-*`

**Windows**: You may see "Windows protected your PC"
- Click "More info"
- Click "Run anyway"

## Quick Start

Just run the command to see the interactive menu:

```bash
ssh-menu
```

You can also directly connect to a host:

```bash
ssh-menu 3          # Connect to host with menu number 3
ssh-menu webserver  # Connect to host with name 'webserver'
```

## User Interface

### Navigation
- **↑/↓**: Navigate through hosts
- **←/→**: Switch between views (All hosts → Groups)
- **Type**: Filter hosts by typing numbers or letters
- **Enter**: Connect to selected host (auto-selects if only one match)
- **Esc**: Quit without connecting
- **Tab**: Alternative way to cycle through views

### Filtering
- Type **numbers** to filter by menu number (e.g., "1" shows hosts 1, 10-19, 100-199)
- Type **letters** to filter by hostname (case-insensitive prefix matching)
- Filtering works on both short names and full hostnames


## SSH Config Setup

For a host to appear in the menu, add a `# Menu:` comment in your SSH config:

```
Host myserver
    HostName server.example.com
    User admin
    Port 22
    # Menu: Production web server
    # IP: 203.0.113.10
    # Group: Production
```

You can specify an explicit menu number:

```
Host database
    HostName db.example.com
    User dbadmin
    # Menu 5: Primary database server
    # Group: Database
    # Group: Critical
```

### Host Groups

Organize hosts into groups for better management:

```
Host webserver
    HostName web1.example.com
    # Menu: Web server 1
    # Group: Frontend
    # Group: Production
```

## Usage Options

| Option | Description |
|--------|-------------|
| `-d` | Show detailed connection information |
| `-V` | Enable SSH verbose mode |
| `-s "opts"` | Pass additional SSH options |
| `-g <group>` | Filter hosts by group |
| `-l` | List all available groups |

### Examples

Connect with agent forwarding:
```bash
ssh-menu -s "-A" webserver
```

Connect through a jump host:
```bash
ssh-menu -s "-J jumphost.example.com" database
```

Filter by group with details:
```bash
ssh-menu -g Production -d
```

## Theme Customization

Customize colors using environment variables:

```bash
export SSH_MENU_COLOR_BACKGROUND="#1e1e2e"  # Dark background
export SSH_MENU_COLOR_FOREGROUND="#cdd6f4"  # Light text
export SSH_MENU_COLOR_BORDER="#9399b2"      # Border color
export SSH_MENU_COLOR_SELECTED="#a6e3a1"    # Selected item/active view
export SSH_MENU_COLOR_ACCENT="#89dceb"      # Titles and accents
export SSH_MENU_COLOR_DIMMED="#585b70"      # Help text and separators
```

Or add color settings to your SSH config:

```
# ColorBackground: #1e1e2e
# ColorForeground: #cdd6f4
# ColorBorder: #9399b2
# ColorSelected: #a6e3a1
# ColorAccent: #89dceb
# ColorDimmed: #585b70
```

### Integration with tmux
Add to `~/.tmux.conf`:
```
bind-key S split-window -h "ssh-menu"
```

## License

This project is licensed under the MIT License.

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the amazing TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) for beautiful terminal styling
