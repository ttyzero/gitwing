package ui

import (
	"charm.land/lipgloss/v2"

	"github.com/ttyzero/commons/theme"
)

type styles struct {
	box    lipgloss.Style
	title  lipgloss.Style
	text   lipgloss.Style
	muted  lipgloss.Style
	subtle lipgloss.Style
	ok     lipgloss.Style
	bad    lipgloss.Style
	warn   lipgloss.Style
	accent lipgloss.Style
	link   lipgloss.Style
	plus   lipgloss.Style
	minus  lipgloss.Style
	rule   lipgloss.Style
}

func newStyles(p theme.Palette, width, height int, borderless bool) styles {
	// Lip Gloss v2 Width/Height include the border when one is set.
	w, h := width, height
	if w < 10 {
		w = 10
	}
	if h < 6 {
		h = 6
	}
	box := lipgloss.NewStyle().
		Width(w).
		Height(h).
		MaxHeight(h).
		Padding(0, 1)
	if !borderless {
		box = box.Border(lipgloss.RoundedBorder()).BorderForeground(p.Border)
	}
	return styles{
		box:    box,
		title:  lipgloss.NewStyle().Bold(true).Foreground(p.Accent),
		text:   lipgloss.NewStyle().Foreground(p.Text),
		muted:  lipgloss.NewStyle().Foreground(p.Muted),
		subtle: lipgloss.NewStyle().Foreground(p.Subtle),
		ok:     lipgloss.NewStyle().Foreground(p.Success),
		bad:    lipgloss.NewStyle().Foreground(p.Danger),
		warn:   lipgloss.NewStyle().Foreground(p.Warn),
		accent: lipgloss.NewStyle().Foreground(p.Accent2),
		link:   lipgloss.NewStyle().Bold(true).Foreground(p.Text).Underline(true),
		plus:   lipgloss.NewStyle().Foreground(p.Success),
		minus:  lipgloss.NewStyle().Foreground(p.Danger),
		rule:   lipgloss.NewStyle().Foreground(p.Subtle),
	}
}

func innerWidth(termWidth int, borderless bool) int {
	w := termWidth - 2 // padding
	if !borderless {
		w -= 2 // rounded border
	}
	if w < 4 {
		return 4
	}
	return w
}

func innerHeight(termHeight int, borderless bool) int {
	h := termHeight
	if !borderless {
		h -= 2
	}
	if h < 4 {
		return 4
	}
	return h
}
