package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/ttyzero/commons/theme"
	"github.com/ttyzero/gitwing/internal/ui"
)

var version = "dev"

func main() {
	fs := flag.NewFlagSet("gitwing", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	themeFlag := fs.String("theme", "", "auto|dark|light (TTYTHEME, then CLITHEME)")
	borderless := fs.Bool("borderless", false, "no rounded frame (TTYBORDERLESS=1)")
	socket := fs.String("socket", "", "ttybus socket path")
	bus := fs.String("bus", "", "named ttybus")
	channel := fs.String("channel", "files", "ttybus channel to subscribe")
	showVersion := fs.Bool("version", false, "print version")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `gitwing — git companion pane for ttybus

A tight vertical TUI. Sit it beside vim (~20 columns): it follows
whatever path is published on the files channel.

Usage:
  gitwing [flags] [path]

Flags:
`)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Private GitHub repos:
  $GITHUB_TOKEN, $GH_TOKEN, then gh auth token (gh auth login / keyring)

Theme (all companion panes honor this):
  --theme, $TTYTHEME, $CLITHEME, then OSC 11 / $COLORFGBG
  Values: auto (default), dark, light
  --borderless / $TTYBORDERLESS=1  no rounded frame

Bus:
  $TTYBUS_SOCKET, $TTYBUS_BUS, or --socket / --bus

Keys: q quit   r refresh
`)
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	if *showVersion {
		fmt.Println("gitwing", version)
		return
	}

	if *socket != "" {
		_ = os.Setenv("TTYBUS_SOCKET", *socket)
	}
	if *bus != "" {
		_ = os.Setenv("TTYBUS_BUS", *bus)
	}

	path := ""
	if rest := fs.Args(); len(rest) > 0 {
		path = rest[0]
	}

	m := ui.New(ui.Options{
		Theme:      theme.Resolve(*themeFlag),
		Channel:    *channel,
		Path:       path,
		Borderless: theme.Borderless(*borderless),
	})
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
