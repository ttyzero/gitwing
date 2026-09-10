package gitinfo

import (
	"context"
	"strings"
	"time"
)

// HeatHorizon is how far back we count daily commits. Two years is enough
// to fill a tall pane as a GitHub-style contribution wallpaper.
const HeatHorizon = 730

// CommitHeat returns commit counts per day for the last `days` days.
// Index 0 is today, 1 is yesterday. Missing days are zero. `days` is
// capped at HeatHorizon.
func CommitHeat(root string, days int) []int {
	if days < 1 {
		days = 1
	}
	if days > HeatHorizon {
		days = HeatHorizon
	}
	counts := make([]int, days)
	if root == "" {
		return counts
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	out, err := git(ctx, root, "log", "--all", "--pretty=format:%as", "--since="+since)
	if err != nil {
		return counts
	}
	today := startOfDay(time.Now())
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		t, err := time.ParseInLocation("2006-01-02", line, time.Local)
		if err != nil {
			continue
		}
		d := int(today.Sub(startOfDay(t)).Hours() / 24)
		if d >= 0 && d < days {
			counts[d]++
		}
	}
	return counts
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// HeatLevel maps a daily commit count onto 0–4, GitHub-style.
func HeatLevel(n int) int {
	switch {
	case n <= 0:
		return 0
	case n == 1:
		return 1
	case n <= 3:
		return 2
	case n <= 6:
		return 3
	default:
		return 4
	}
}
