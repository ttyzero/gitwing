package ui

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
)

func compactInt(n int) string {
	switch {
	case n < 1000:
		return fmt.Sprintf("%d", n)
	case n < 10_000:
		return trim1(float64(n)/1000) + "k"
	case n < 1_000_000:
		return fmt.Sprintf("%dk", n/1000)
	case n < 10_000_000:
		return trim1(float64(n)/1_000_000) + "m"
	default:
		return fmt.Sprintf("%dm", n/1_000_000)
	}
}

func trim1(f float64) string {
	s := fmt.Sprintf("%.1f", f)
	return strings.TrimSuffix(s, ".0")
}

func ellipsize(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return runewidth.Truncate(s, max, "…")
}

func wrap(s string, width, maxLines int) []string {
	s = strings.TrimSpace(s)
	if s == "" || width <= 0 || maxLines <= 0 {
		return nil
	}
	var lines []string
	var cur strings.Builder
	curW := 0
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		lines = append(lines, cur.String())
		cur.Reset()
		curW = 0
	}
	for _, word := range strings.Fields(s) {
		ww := runewidth.StringWidth(word)
		if curW == 0 {
			if ww <= width {
				cur.WriteString(word)
				curW = ww
				continue
			}
			lines = append(lines, ellipsize(word, width))
			if len(lines) >= maxLines {
				return lines
			}
			continue
		}
		if curW+1+ww <= width {
			cur.WriteByte(' ')
			cur.WriteString(word)
			curW += 1 + ww
			continue
		}
		flush()
		if len(lines) >= maxLines {
			return lines
		}
		if ww <= width {
			cur.WriteString(word)
			curW = ww
		} else {
			lines = append(lines, ellipsize(word, width))
			if len(lines) >= maxLines {
				return lines
			}
		}
	}
	flush()
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		last := lines[maxLines-1]
		lines[maxLines-1] = ellipsize(last, width)
	}
	return lines
}

func hairline(width int) string {
	if width < 1 {
		return ""
	}
	return strings.Repeat("─", width)
}
