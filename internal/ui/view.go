package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.WindowTitle = windowTitle(m)
	return v
}

func windowTitle(m model) string {
	name := m.git.Repo
	if name == "" {
		name = m.git.Folder
	}
	if name == "" {
		return "gitwing"
	}
	return "gitwing · " + name
}

func (m model) render() string {
	w, h := m.width, m.height
	if w < 12 {
		w = 12
	}
	if h < 8 {
		h = 8
	}
	st := newStyles(m.pal, w, h, m.borderless)
	inner := innerWidth(w, m.borderless)
	innerH := innerHeight(h, m.borderless)

	var b strings.Builder
	write := func(s string) {
		if s == "" {
			return
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(s)
	}

	write(m.headerLine(st, inner))

	if m.git.Root == "" && m.path == "" {
		write(st.muted.Render("listening"))
		write(st.accent.Render("files"))
		if m.busErr != "" {
			write(st.bad.Render(ellipsize(m.busErr, inner)))
		} else if !m.busOK {
			write(st.subtle.Render("no bus yet"))
		} else {
			write(st.subtle.Render("pub a path"))
		}
	} else {
		for _, line := range m.repoBlock(st, inner) {
			write(line)
		}
		write(st.rule.Render(hairline(inner)))
		for _, line := range m.branchBlock(st, inner) {
			write(line)
		}
		if m.git.File.InRepo && !m.git.File.IsDir {
			write(st.rule.Render(hairline(inner)))
			for _, line := range m.fileBlock(st, inner) {
				write(line)
			}
		}
		if m.gitErr != "" {
			write(st.bad.Render(ellipsize(m.gitErr, inner)))
		}
	}

	help := st.subtle.Render("q  r")
	body := b.String()
	used := 0
	if body != "" {
		used = strings.Count(body, "\n") + 1
	}
	// Fourth row: contribution graph fills leftover space above help.
	remain := innerH - used - 1
	if m.git.Root != "" && remain >= 3 {
		write(m.renderHeat(inner, remain))
		body = b.String()
		used = strings.Count(body, "\n") + 1
	}
	if used < innerH {
		body += strings.Repeat("\n", innerH-used) + help
	}
	return st.box.Render(body)
}

func (m model) headerLine(st styles, inner int) string {
	dot := st.subtle.Render("·")
	if m.busOK {
		dot = st.ok.Render("●")
	} else if m.busErr != "" {
		dot = st.bad.Render("●")
	}
	title := st.title.Render("gitwing")
	left := title + " " + dot
	leftW := lipgloss.Width(title) + 1 + 1
	if inner-leftW < 0 {
		return ellipsize(title, inner)
	}
	return left
}

func (m model) repoBlock(st styles, inner int) []string {
	live := m.ghLive()
	name := m.git.Repo
	if name == "" {
		name = m.git.Folder
	}
	if live && m.gh.Name != "" {
		name = m.gh.Name
	}
	if name == "" {
		name = filepath.Base(m.path)
	}
	if name == "" || name == "." {
		return nil
	}

	var lines []string
	label := ellipsize(name, inner)
	if live && m.gh.HTMLURL != "" {
		lines = append(lines, st.link.Hyperlink(m.gh.HTMLURL).Render(label))
	} else {
		lines = append(lines, st.text.Bold(true).Render(label))
	}

	if live {
		star := st.warn.Render("★") + st.text.Render(compactInt(m.gh.Stars))
		fork := st.accent.Render("⑂") + st.text.Render(compactInt(m.gh.Forks))
		stats := star + "  " + fork
		if lipgloss.Width(stats) > inner {
			stats = star + " " + fork
		}
		lines = append(lines, stats)
		if lang := strings.TrimSpace(m.gh.Language); lang != "" {
			lines = append(lines, st.muted.Render(ellipsize(lang, inner)))
		}
		descLines := wrap(m.gh.Description, inner, descLinesFor(m.height))
		for _, ln := range descLines {
			lines = append(lines, st.muted.Render(ln))
		}
	} else if m.ghLoad {
		lines = append(lines, st.subtle.Render("github…"))
	} else if m.ghErr != "" && m.git.Owner != "" {
		lines = append(lines, st.subtle.Render(ellipsize(m.ghErr, inner)))
	} else if m.git.Owner != "" {
		slug := ellipsize(m.git.Owner+"/"+m.git.Repo, inner)
		lines = append(lines, st.muted.Render(slug))
	}
	return lines
}

func descLinesFor(h int) int {
	switch {
	case h >= 22:
		return 3
	case h >= 18:
		return 2
	case h >= 14:
		return 1
	default:
		return 0
	}
}

func (m model) branchBlock(st styles, inner int) []string {
	if m.git.Root == "" {
		if m.gitErr != "" {
			return []string{st.muted.Render("not a repo")}
		}
		return nil
	}
	br := m.git.Branch
	if br == "" {
		br = "HEAD"
	}
	if m.git.Detached {
		br = "(" + br + ")"
	}
	lines := []string{st.text.Render(ellipsize(br, inner))}

	var status string
	if m.git.Dirty {
		status = st.warn.Render("●") + " " + st.muted.Render("dirty")
	} else {
		status = st.ok.Render("○") + " " + st.muted.Render("clean")
	}
	lines = append(lines, status)

	if m.git.Ahead > 0 || m.git.Behind > 0 {
		ab := st.ok.Render(fmt.Sprintf("↑%d", m.git.Ahead)) + " " +
			st.warn.Render(fmt.Sprintf("↓%d", m.git.Behind))
		lines = append(lines, ab)
	}
	if m.git.SHA != "" {
		lines = append(lines, st.subtle.Render(ellipsize(m.git.SHA, inner)))
	}
	return lines
}

func (m model) fileBlock(st styles, inner int) []string {
	f := m.git.File
	name := f.Name
	if name == "" {
		name = filepath.Base(m.path)
	}
	lines := []string{st.text.Render(ellipsize(name, inner))}

	letter := fileLetter(f.Status, f.Untracked)
	var bits []string
	if letter != "" {
		sty := st.muted
		switch letter {
		case "M", "?":
			sty = st.warn
		case "A":
			sty = st.ok
		case "D", "U":
			sty = st.bad
		}
		bits = append(bits, sty.Bold(true).Render(letter))
	}
	if f.Untracked {
		bits = append(bits, st.warn.Render("new"))
	} else if f.Added == 0 && f.Deleted == 0 && letter == "" {
		bits = append(bits, st.ok.Render("clean"))
	} else {
		if f.Added > 0 || f.Deleted > 0 || letter == "M" {
			bits = append(bits, st.plus.Render(fmt.Sprintf("+%d", f.Added)))
			bits = append(bits, st.minus.Render(fmt.Sprintf("−%d", f.Deleted)))
		}
	}
	if len(bits) > 0 {
		lines = append(lines, strings.Join(bits, " "))
	}
	if f.Rel != "" && f.Rel != f.Name && m.height >= 18 {
		lines = append(lines, st.subtle.Render(ellipsize(f.Rel, inner)))
	}
	if m.ev.Line > 0 {
		lines = append(lines, st.subtle.Render(fmt.Sprintf(":%d", m.ev.Line)))
	}
	return lines
}

func fileLetter(status string, untracked bool) string {
	if untracked {
		return "?"
	}
	status = strings.TrimSpace(status)
	if status == "" {
		return ""
	}
	for i := 0; i < len(status); i++ {
		c := status[i]
		if c != ' ' && c != '?' {
			return string(c)
		}
	}
	return ""
}
