package ui

import (
	"image/color"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/ttyzero/gitwing/internal/gitinfo"
)

// heatCell is a one-column square, like a GitHub contribution box.
const heatCell = "■"

// daysAgoAt maps a wallpaper cell onto "N days ago" using GitHub's
// contribution layout: columns are weeks (left = older), rows 0–6 are
// Sun–Sat, and further rows wrap into older bands. Negative means a
// future day in the current week (leave blank, as GitHub does).
func daysAgoAt(x, y, w, h, todayDOW int) int {
	if w <= 0 || h <= 0 || x < 0 || y < 0 || x >= w || y >= h {
		return -1
	}
	band := y / 7
	dow := y % 7
	bands := (h + 6) / 7
	week := x + band*w
	newest := bands*w - 1
	weeksAgo := newest - week
	return weeksAgo*7 + (todayDOW - dow)
}

func heatNeed(w, h int) int {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	n := ((h + 6) / 7) * w * 7
	if n > gitinfo.HeatHorizon {
		return gitinfo.HeatHorizon
	}
	if n < 7 {
		return 7
	}
	return n
}

func (m model) renderHeat(w, h int) string {
	todayDOW := int(time.Now().Weekday())
	var b strings.Builder
	b.Grow(h * (w + 1) * 16)
	for y := 0; y < h; y++ {
		if y > 0 {
			b.WriteByte('\n')
		}
		for x := 0; x < w; x++ {
			ago := daysAgoAt(x, y, w, h, todayDOW)
			b.WriteString(m.heatGlyph(ago))
		}
	}
	return b.String()
}

func (m model) heatGlyph(daysAgo int) string {
	if daysAgo < 0 {
		return " "
	}
	n := 0
	if daysAgo < len(m.heat) {
		n = m.heat[daysAgo]
	}
	lvl := gitinfo.HeatLevel(n)
	c := m.heatColor(daysAgo, lvl)
	return lipgloss.NewStyle().Foreground(c).Render(heatCell)
}

func (m model) heatColor(daysAgo, level int) color.Color {
	p := m.pal
	if daysAgo == 0 {
		f := m.git.File
		switch {
		case f.Untracked:
			return p.Accent
		case fileLetter(f.Status, false) == "D":
			return p.Danger
		case m.git.Dirty || f.Added > 0 || f.Deleted > 0 || fileLetter(f.Status, false) != "":
			return p.Warn
		case level > 0:
			return p.Heat[4]
		default:
			return p.Success
		}
	}
	if level < 0 {
		level = 0
	}
	if level > 4 {
		level = 4
	}
	return p.Heat[level]
}
