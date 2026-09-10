# gitwing

A Git companion pane for [ttybus](https://github.com/ttyzero/ttybus).

Switch buffers in vim (or pick a file anywhere) and a neighbouring ~20-column
pane shows that file's repo on GitHub, the local branch, and whether the file
itself is dirty.

```
╭ gitwing ● ──╮
│ ttybus      │
│ ★12  ⑂3     │
│ Go          │
│─────────────│
│ main        │
│ ● dirty     │
│─────────────│
│ README.md   │
│ M +8 −2     │
│─────────────│
│ ■□■■■□□■■□  │
│ □■■■■□■□□■  │
│ ■□■■■■■□■■  │
│ □□■□■■■■■□  │
│ ■■■□■□■■■■  │
│ □■■■■■■□■□  │
│ ■□■■■□□■■■  │
│ q  r        │
╰─────────────╯
```

The fourth block is a GitHub-style contribution graph for this repo
(weeks across, Sunday–Saturday down). Cell colour is daily commit
intensity; **today** is tinted by the selected file (green clean, amber
dirty, pink untracked, rose deleted).

Built with [Bubble Tea](https://charm.land/) v2 and Lip Gloss v2. No
panel fill — the terminal theme shows through; we only tint type,
borders, accents, and the graph.

## Install

Go 1.25+ (Charm v2). `git` on `PATH`.

```sh
make build
./gitwing
```

GitHub stars/forks/description come from the API. Public repos work
unauthenticated. Private repos need a token — see below.

## Use it

Sit it beside an editor. tmux:

```sh
tmux split-window -h -l 22 gitwing
```

It subscribes to ttybus channel `files`. A path (optional `:line`) or a
one-line JSON object is enough:

```sh
ttybus pub files "$(pwd)/README.md"
ttybus pub files '{"path":"/code/foo.go","line":12,"origin":"vim"}'
```

Vim / Neovim — `contrib/vim/gitwing.vim` publishes `BufEnter`:

```vim
set runtimepath+=~/code/gitwing/contrib/vim
runtime gitwing.vim
```

Without a publisher, gitwing still inspects the current directory so the
pane is useful on its own.

Keys: `q` quit, `r` refresh.

## Private repositories

Unauthenticated GitHub calls see private repos as 404. The pane then
shows the local git name and a muted `private` instead of stars.

gitwing looks for a token, first match wins:

1. `$GITHUB_TOKEN`
2. `$GH_TOKEN`
3. `gh auth token` — your existing `gh auth login`, including the macOS
   keyring. If `gh` is not on `PATH` (common in a slim tmux pane), it
   also tries `/opt/homebrew/bin/gh` and `/usr/local/bin/gh`.
4. `oauth_token` in `$GH_CONFIG_DIR/hosts.yml` or `~/.config/gh/hosts.yml`
   (older gh that stored the token in the file)

```sh
gh auth login                 # once; gitwing reuses it
# or
export GITHUB_TOKEN=ghp_...   # or GH_TOKEN
```

The token needs `repo` scope to read private repositories. `gh auth status`
shows what you already have. gitwing never prints the token.

## Theme

Every companion pane (gitwing now; peek and buscope next) honors the same
override so a workspace stays one family:

```
--theme flag
  > $TTYTHEME          gitwing / peek / buscope
  > $CLITHEME          proposed cross-CLI standard (dark|light|auto)
  > OSC 11             live terminal background (Bubble Tea)
  > $COLORFGBG         cheap, static (rxvt, Konsole, iTerm)
  > dark
```

```sh
export TTYTHEME=auto          # default; follow the terminal
export TTYTHEME=nord
gitwing --theme catppuccin
gitwing --borderless          # or TTYBORDERLESS=1 for all three panes
```

Palettes: Charm `auto`/`dark`/`light`, plus Catppuccin, Dracula, Nord,
Gruvbox, Tokyo Night, Solarized, One Dark, Rosé Pine, Everforest,
Kanagawa. Switch live with [ttythemer](https://github.com/ttyzero/ttythemer).

Live look (all three panes subscribe to ttybus channel `theme`):

```sh
ttybus pub theme 'THEME nord'
ttybus pub theme 'THEME catppuccin BORDERS=0'
ttybus pub theme 'BORDERS=0'
```

### How other TUIs pick up a theme

There is no single standard yet. In practice:

| Signal | Who uses it | Notes |
|---|---|---|
| **OSC 11** (`\e]11;?\a`) | Neovim, fish, bat/delta, Charm / Bubble Tea v2, most modern TUIs | Queries the *current* background RGB; classify by luminance. Works over SSH. Best live signal. |
| **`$COLORFGBG`** | rxvt, Konsole, iTerm2, many fallbacks | `fg;bg` ANSI indices. Fast, no I/O, often stale after a theme switch. |
| **`$CLITHEME`** | [proposed](https://wiki.tau.garden/cli-theme/) | `dark` \| `light` \| `auto`. One override for every CLI. |
| App env (`OPENCLAW_THEME`, `HERMES_TUI_THEME`, …) | Individual tools | Same idea, different names — we use `$TTYTHEME` for this family and still honor `$CLITHEME`. |
| CSI `? 996 n` | A handful of terminals | Direct dark/light query. Too rare to rely on. |

Charm's Lip Gloss v2 dropped implicit `AdaptiveColor`. The supported path
with Bubble Tea is: `tea.RequestBackgroundColor` in `Init`, then
`BackgroundColorMsg.IsDark()` to pick a palette. That is OSC 11, done
correctly against the program's TTY (including SSH via Wish).

gitwing uses that stack, with `$TTYTHEME` / `$CLITHEME` as the unified
override so peek and buscope can match without each growing a private flag.

## Flags

```
gitwing [--theme auto|dark|light] [--bus NAME] [--socket PATH]
        [--channel files] [path]
```

`--bus` / `--socket` / `$TTYBUS_BUS` / `$TTYBUS_SOCKET` select a named
ttybus. The daemon is started on demand if `ttybus` is on `PATH`.

## Why

ttybus is a side channel beside the TTY. Editors publish `files`; gitwing
subscribes. Swap vim for fzf and the pane still follows. That is the demo:
tools that talk, not an IDE.

See [tty-ideas.md](../tty-ideas.md) for peek (preview) and buscope (bus
inspector), which share this theme.
