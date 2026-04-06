# SSH Menu Modernization Design

## Overview

Full modernization of ssh-menu: restructure into focused packages, add a two-pane UI with fuzzy search, pinned hosts stored in SSH config comments, non-blocking config validation, comprehensive tests, and cleanup of dead code and bad defaults.

## Package Structure

```
ssh-menu/
├── main.go                     # Entry point, CLI flags, SSH execution
├── internal/
│   ├── config/
│   │   ├── parser.go           # SSH config file parsing
│   │   ├── parser_test.go
│   │   ├── colors.go           # Color config from SSH config comments
│   │   └── colors_test.go
│   ├── host/
│   │   ├── host.go             # Host type definition
│   │   ├── filter.go           # Fuzzy matching and filtering
│   │   ├── filter_test.go
│   │   ├── groups.go           # Group logic, menu numbering
│   │   ├── groups_test.go
│   │   ├── validate.go         # SSH config validation
│   │   └── validate_test.go
│   ├── ui/
│   │   ├── model.go            # Top-level Bubble Tea model
│   │   ├── hostlist.go         # Left pane: scrollable host list
│   │   ├── detail.go           # Right pane: host detail panel
│   │   ├── viewbar.go          # Group tab bar
│   │   └── keys.go             # Key bindings
│   └── theme/
│       ├── theme.go            # Consolidated color system
│       └── theme_test.go
```

Current `internal/config.go` splits into `config/parser.go` (SSH parsing) and `theme/theme.go` (colors), eliminating the color parsing duplication between `config.go` and `theme.go`. Current `ui.go` splits into focused components for the two-pane layout. Current `host.go` splits into type definition, filtering, grouping, and validation.

## Two-Pane UI Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│ SSH Menu                    ↑/↓ Navigate • ←/→ View • Enter Select │
│ [ All ] • Servers • Databases • Ungrouped                          │
│ Filter: prod                                                       │
├──────────────────────────────────────┬──────────────────────────────┤
│  ▸  1) web-prod-01                   │  web-prod-01                │
│     2) web-prod-02                   │  Production web server      │
│     3) api-prod-01                   │                             │
│     4) db-prod-01                    │  Host: 10.0.1.50            │
│                                      │  User: deploy               │
│                                      │  Port: 22                   │
│                                      │  Key:  ~/.ssh/prod_ed25519  │
│                                      │  IP:   203.0.113.50         │
│                                      │                             │
│                                      │  Groups: Servers, Production│
│                                      │                             │
│                                      │  ★ Pinned                   │
│                                      │                             │
│                                      │  ⚠ Identity file not found  │
└──────────────────────────────────────┴──────────────────────────────┘
```

### Layout behavior

- Left pane ~55% width, right pane ~45%, adjusts on terminal resize.
- Host list scrolls when it exceeds available height, keeping cursor visible (viewport-style scrolling).
- Detail panel updates instantly as cursor moves.
- Pinned hosts sort to top of every view.
- Validation warnings show in detail panel with yellow `⚠`.

### Detail panel contents

- Host alias (bold, accent color)
- Description text from `# Menu:` comment
- Connection details: HostName, User, Port, IdentityFile, IP (only if explicitly set in config)
- Groups list
- Pin status (`★ Pinned` if pinned)
- Validation warnings

### Graceful degradation

If terminal width < 60 columns, collapse to single-pane mode (enhanced version of current behavior with scrolling). Detail panel hides rather than wrapping awkwardly.

## Fuzzy Search

Score-based fuzzy matching. Each host gets a score, results sorted best-first.

### Scoring rules

- Exact substring match scores highest.
- Consecutive character matches score higher than scattered.
- Matches at word boundaries (after `-`, `_`, `.`) score higher than mid-word.
- Earlier matches in the string score higher.

### Fields searched (priority order)

1. ShortName (host alias) — highest weight
2. Description text
3. LongName (hostname)
4. IP address
5. Group names

### Behavior

- Pure numeric input still filters by menu number prefix (preserving quick numeric selection).
- All non-numeric input uses fuzzy matching.
- Results re-sort by score on each keystroke; best match at top.
- Single remaining match + Enter connects immediately (current behavior preserved).
- Cursor resets to top on each keystroke.

### Implementation

Internal `filter.go` with `Score(query, text string) int` and `Match(query string, host Host) int`. No external dependency — fzf-style scoring is ~80 lines of Go.

## Favorites / Pinned Hosts

### Storage

A `# Pinned` comment in the SSH config:

```ssh-config
# Menu 1: Production web server
# Group: Servers
# Pinned
Host web-prod-01
    HostName 10.0.1.50
    User deploy
```

### Behavior

- Pinned hosts sort to top of every view, maintaining relative menu number order among themselves.
- Non-pinned hosts appear after, also in menu number order.
- `★` indicator in list and detail panel.
- `p` key toggles pin — writes/removes `# Pinned` line in the SSH config file.

### Config file editing

- Parser tracks which file each host was parsed from (stored on the Host struct as `SourceFile string`).
- Pin toggle writes to the same file the host was originally parsed from (handles `config.d/` correctly).
- Locates the host block and adds/removes the `# Pinned` line.
- Inserts after the last `# Group:` or `# Menu:` comment for that host.
- Preserves all other formatting, whitespace, and comments.

## SSH Config Validation

### What gets validated

- **Identity file exists** — checks path on disk (expanding `~`).
- **Duplicate host aliases** — flags same alias in multiple config files.
- **Empty hostname** — warns if `Host` entry has no `HostName` and alias isn't an IP/FQDN.

### What does NOT get validated

- No network reachability checks.
- No permission checks on key files.
- No validation of SSH options we don't parse.

### How warnings surface

- Validation runs once at startup. Results stored on each Host as `[]Warning`.
- Warnings display only in the detail panel with yellow `⚠`.
- Warnings don't block connection.

### Warning type

```go
type Warning struct {
    Level   string // "warn" for now
    Message string // e.g., "Identity file not found: ~/.ssh/missing_key"
}
```

## Error Handling & Code Quality

### Remove os.Exit from library code

`AssignMenuNumbers` returns an error on duplicate menu numbers instead of calling `os.Exit(1)`. All exit logic moves to `main.go`.

### Remove hardcoded defaults

Stop setting `User: "root"` and `Port: "22"` during parsing. Fields are empty strings when not explicitly configured. SSH handles defaults; the detail panel shows only what's in the config.

### Remove dead code

- `ShowHelp()` — unused
- `clearScreen()` — unnecessary with `tea.WithAltScreen()`
- `GroupHosts()` — never called
- `ConnectTimeout`, `ServerAliveInterval`, `ServerAliveCountMax` on Host — parsed but never used

### Key bindings

All key bindings defined in `ui/keys.go` using Bubble Tea's `key.Binding` system instead of raw `tea.KeyMsg` switches. Full binding list:

- `↑`/`↓` — navigate hosts
- `←`/`→` — switch group views
- `Tab` — next group view
- `Enter` — connect to selected/only host
- `Esc` — quit
- `Backspace` — delete filter character
- `p` — toggle pin
- Any other rune — fuzzy filter input

## Testing Strategy

- **`config/parser_test.go`** — parse sample SSH configs from strings, verify host extraction, comment parsing (`# Menu`, `# Group`, `# Pinned`, `# IP`), multi-file merging, edge cases (empty files, wildcard hosts, Include directives).
- **`host/filter_test.go`** — fuzzy scoring: exact beats partial, boundary beats mid-word, numeric prefix preserved, multi-field search, empty query returns all.
- **`host/groups_test.go`** — menu number assignment, duplicate detection returns error, group extraction, pin sorting order.
- **`host/validate_test.go`** — identity file checks with temp files, duplicate alias detection, empty hostname detection.
- **`theme/theme_test.go`** — env var overrides config file, config file overrides defaults, fallback to defaults.
