package ui

import (
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/mattn/go-runewidth"
	"github.com/ttyzero/commons/theme"
	"github.com/ttyzero/gitwing/internal/gitinfo"
)

func TestHeatCellWidth(t *testing.T) {
	if w := runewidth.StringWidth(heatCell); w != 1 {
		t.Fatalf("heatCell %q width %d, want 1", heatCell, w)
	}
}

func TestDaysAgoAtToday(t *testing.T) {
	dow := int(time.Now().Weekday())
	w, h := 20, 7
	// Rightmost column, today's weekday row, single band → today.
	got := daysAgoAt(w-1, dow, w, h, dow)
	if got != 0 {
		t.Fatalf("today cell daysAgo=%d", got)
	}
	// Future weekday in the current week.
	if dow < 6 {
		fut := daysAgoAt(w-1, dow+1, w, h, dow)
		if fut >= 0 {
			t.Fatalf("future cell should be negative, got %d", fut)
		}
	}
}

func TestHeatLevel(t *testing.T) {
	if gitinfo.HeatLevel(0) != 0 || gitinfo.HeatLevel(1) != 1 || gitinfo.HeatLevel(8) != 4 {
		t.Fatal("buckets")
	}
}

func TestRenderHeatFits(t *testing.T) {
	m := model{
		pal:    theme.New(true),
		width:  20,
		height: 24,
		heat:   []int{3, 0, 1, 0, 8},
	}
	out := m.renderHeat(20, 24)
	lines := splitLines(out)
	if len(lines) != 24 {
		t.Fatalf("rows %d", len(lines))
	}
	for i, ln := range lines {
		if w := lipgloss.Width(ln); w != 20 {
			t.Errorf("row %d width %d", i, w)
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}
