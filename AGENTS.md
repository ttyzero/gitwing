# Agents guide for gitwing

Git companion pane for [ttybus](https://github.com/ttyzero/ttybus). Subscribes
to `files`, shows GitHub + local git + the selected file. Designed for a
~20-column vertical split beside vim.

## Stack

- Go 1.25, module `github.com/ttyzero/gitwing`
- Bubble Tea v2 + Lip Gloss v2
- Shared look: `github.com/ttyzero/commons/theme` + `commons/connect`
- ttybus client via commons/connect
- stdlib `flag`. No cobra.

```sh
make test
make build
```

## Layout

- `cmd/gitwing` — flags, program start
- `internal/gitinfo` — local git snapshot
- `internal/github` — REST metadata, cached
- `internal/ui` — Bubble Tea model + compact view
- `contrib/vim` — publisher

Shared theme, borders, and the `theme` bus live in
[ttyzero/commons](https://github.com/ttyzero/commons).

## Invariants

- Follow `files` events; do not require the editor to know about gitwing.
- Stay readable at 20 columns. Do not paint a panel background — the
  terminal theme shows through.
- Honor `$TTYTHEME` then `$CLITHEME` then OSC 11 then `$COLORFGBG`.
- Subscribe to `theme` and apply `THEME` / `BORDERS` live.
- Payloads are single-line; path or JSON.
