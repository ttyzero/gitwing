package ui

import (
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/ttyzero/commons/connect"
	"github.com/ttyzero/commons/files"
	"github.com/ttyzero/commons/theme"
	"github.com/ttyzero/gitwing/internal/github"
	"github.com/ttyzero/gitwing/internal/gitinfo"
)

const (
	tickEvery      = 2 * time.Second
	reconnectEvery = 3 * time.Second
)

// Options configure the pane.
type Options struct {
	Theme      theme.Spec
	Channel    string
	Path       string
	Borderless bool
}

type model struct {
	spec       theme.Spec
	pal        theme.Palette
	borderless bool

	width, height int

	channel string

	sess     *connect.Session
	busOK    bool
	busErr   string
	lastDial time.Time

	path string
	ev   files.Event

	git    gitinfo.Info
	gitErr string

	gh     github.Repo
	ghErr  string
	ghLoad bool
	ghKey  string

	ghClient *github.Client

	heat    []int
	heatKey string
}

type inspectMsg struct {
	path string
	info gitinfo.Info
	err  error
}
type ghMsg struct {
	key  string
	repo github.Repo
	err  error
}
type connectedMsg struct {
	sess *connect.Session
}
type connectErr struct{ err error }
type tickMsg struct{}
type heatMsg struct {
	key  string
	days []int
}

// New starts with cwd (or Path) so the pane is useful before any
// publisher speaks.
func New(opts Options) model {
	path := opts.Path
	if path == "" {
		path, _ = os.Getwd()
	}
	ch := opts.Channel
	if ch == "" {
		ch = "files"
	}
	return model{
		spec:       opts.Theme,
		pal:        theme.For(opts.Theme),
		borderless: opts.Borderless,
		width:      20,
		height:     24,
		channel:    ch,
		path:       path,
		ghClient:   github.New(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
		inspectCmd(m.path),
		connectCmd(m.channel),
		tickCmd(tickEvery),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.spec = m.spec.ApplyOSC(msg.IsDark())
		m.pal = theme.For(m.spec)
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			if m.sess != nil {
				_ = m.sess.Close()
			}
			return m, tea.Quit
		case "r":
			if m.git.Owner != "" {
				m.ghClient.Forget(m.git.Owner, m.git.Repo)
				m.ghKey = ""
			}
			cmds := []tea.Cmd{inspectCmd(m.path)}
			if !m.busOK {
				cmds = append(cmds, connectCmd(m.channel))
			}
			return m, tea.Batch(cmds...)
		}

	case connectedMsg:
		if m.sess != nil {
			_ = m.sess.Close()
		}
		m.sess = msg.sess
		m.busOK = true
		m.busErr = ""
		return m, tea.Batch(
			connect.Wait("files", m.sess.Files),
			connect.Wait(theme.Channel, m.sess.Theme),
		)

	case connectErr:
		m.busOK = false
		m.busErr = "no bus"
		m.lastDial = time.Now()
		return m, nil

	case connect.Line:
		wait := connect.Wait(msg.Channel, m.sess.Chan(msg.Channel))
		if msg.Channel == theme.Channel {
			if c, ok := theme.ParseBus(msg.Body); ok {
				return m, tea.Batch(wait, connect.ApplyLook(c, &m.spec, &m.pal, &m.borderless))
			}
			return m, wait
		}
		ev, err := files.Parse(msg.Body)
		if err != nil || ev.Path == "" {
			return m, wait
		}
		m.ev = ev
		m.path = ev.Path
		return m, tea.Batch(inspectCmd(m.path), wait)

	case connect.Closed:
		if msg.Channel == "files" {
			m.busOK = false
			m.busErr = "bus closed"
			if m.sess != nil {
				_ = m.sess.Close()
				m.sess = nil
			}
		}
		return m, nil

	case inspectMsg:
		if msg.path != m.path {
			return m, nil
		}
		if msg.err != nil {
			m.git = gitinfo.Info{}
			m.gitErr = "not a repo"
			m.clearGitHub()
			m.heat = nil
			m.heatKey = ""
			return m, nil
		}
		m.git = msg.info
		m.gitErr = ""
		return m, tea.Batch(m.syncGitHub(), m.syncHeat())

	case heatMsg:
		if msg.key != m.heatKey {
			return m, nil
		}
		m.heat = msg.days
		return m, nil

	case ghMsg:
		want := m.git.Owner + "/" + m.git.Repo
		if msg.key != want {
			return m, nil
		}
		m.ghLoad = false
		if msg.err != nil {
			m.gh = github.Repo{}
			m.ghErr = github.Short(msg.err)
			return m, nil
		}
		m.gh = msg.repo
		m.ghErr = ""
		return m, nil

	case tickMsg:
		cmds := []tea.Cmd{tickCmd(tickEvery), inspectCmd(m.path)}
		if !m.busOK && time.Since(m.lastDial) >= reconnectEvery {
			cmds = append(cmds, connectCmd(m.channel))
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m *model) clearGitHub() {
	m.gh = github.Repo{}
	m.ghErr = ""
	m.ghKey = ""
	m.ghLoad = false
}

// syncGitHub drops metadata from the previous project and fetches
// GitHub for the current slug. Same slug keeps the existing card.
func (m *model) syncGitHub() tea.Cmd {
	if m.git.Owner == "" || m.git.Repo == "" {
		m.clearGitHub()
		return nil
	}
	key := m.git.Owner + "/" + m.git.Repo
	if key == m.ghKey && (m.ghLive() || m.ghLoad) {
		return nil
	}
	m.gh = github.Repo{}
	m.ghErr = ""
	m.ghKey = key
	m.ghLoad = true
	return fetchCmd(m.ghClient, m.git.Owner, m.git.Repo)
}

// ghLive is true when cached GitHub metadata is for the current git remote.
func (m model) ghLive() bool {
	if m.gh.Name == "" || m.git.Owner == "" || m.git.Repo == "" {
		return false
	}
	want := m.git.Owner + "/" + m.git.Repo
	if m.gh.FullName != "" {
		return strings.EqualFold(m.gh.FullName, want)
	}
	return strings.EqualFold(m.gh.Name, m.git.Repo)
}

func (m *model) syncHeat() tea.Cmd {
	key := m.heatKeyOf()
	if key == "" {
		m.heat = nil
		m.heatKey = ""
		return nil
	}
	if key == m.heatKey && m.heat != nil {
		return nil
	}
	m.heat = nil
	m.heatKey = key
	root := m.git.Root
	n := heatNeed(m.width, m.height)
	return func() tea.Msg {
		return heatMsg{key: key, days: gitinfo.CommitHeat(root, n)}
	}
}

func (m model) heatKeyOf() string {
	if m.git.Root == "" {
		return ""
	}
	return m.git.Root + "@" + m.git.SHA
}

func inspectCmd(path string) tea.Cmd {
	return func() tea.Msg {
		info, err := gitinfo.Inspect(path)
		return inspectMsg{path: path, info: info, err: err}
	}
}

func fetchCmd(c *github.Client, owner, repo string) tea.Cmd {
	key := owner + "/" + repo
	return func() tea.Msg {
		r, err := c.Fetch(owner, repo)
		return ghMsg{key: key, repo: r, err: err}
	}
}

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return tickMsg{} })
}

func connectCmd(channel string) tea.Cmd {
	return func() tea.Msg {
		s, err := connect.Open("gitwing", channel)
		if err != nil {
			return connectErr{err}
		}
		return connectedMsg{sess: s}
	}
}
