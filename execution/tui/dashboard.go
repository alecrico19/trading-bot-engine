package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"trading-bot/execution/internal/types"
)

type Engine interface {
	IsRunning() bool
	IsBreached() bool
	GetSummary() string
	GetPositions() []types.Position
	GetActiveSignals() []*types.Signal
	GetEquity() float64
	GetDailyStats() (pnl float64, trades int, wins int)
	GetStrategyNames() []string
	PauseAll()
	ResumeAll()
	KillAll()
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700")).
			Padding(0, 1)

	statusRunning = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).Bold(true).Render("● RUNNING")
	statusPaper = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00BFFF")).Bold(true).Render("● PAPER")
	statusHalted = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).Bold(true).Render("■ HALTED")
	statusPaused = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).Bold(true).Render("⏸ PAUSED")

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)

	green  = lipgloss.Color("#00FF00")
	red    = lipgloss.Color("#FF0000")
	blue   = lipgloss.Color("#00BFFF")
	yellow = lipgloss.Color("#FFA500")
	gray   = lipgloss.Color("#808080")
)

type tickMsg time.Time

type keyMap struct {
	Quit     key.Binding
	Pause    key.Binding
	Start    key.Binding
	Kill     key.Binding
	Refresh  key.Binding
	Help     key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.Pause, k.Start, k.Kill, k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Quit, k.Pause, k.Start, k.Kill, k.Refresh},
	}
}

var keys = keyMap{
	Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Pause:   key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "pause all")),
	Start:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "start all")),
	Kill:    key.NewBinding(key.WithKeys("k"), key.WithHelp("k", "kill all")),
	Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	Help:    key.NewBinding(key.WithKeys("h", "?"), key.WithHelp("h", "help")),
}

type Model struct {
	engine  Engine
	help    help.Model
	keys    keyMap
	width   int
	height  int
	showAll bool
	ready   bool
}

func New(engine Engine) *Model {
	help := help.New()
	help.ShowAll = false
	return &Model{
		engine: engine,
		help:   help,
		keys:   keys,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Pause):
			m.engine.PauseAll()
		case key.Matches(msg, m.keys.Start):
			m.engine.ResumeAll()
		case key.Matches(msg, m.keys.Kill):
			m.engine.KillAll()
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Refresh):
		}

	case tickMsg:
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return m, nil
}

func (m *Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	panels := []string{
		m.renderStatusBar(),
		m.renderMainView(),
		m.renderHelp(),
	}

	return lipgloss.JoinVertical(lipgloss.Left, panels...)
}

func (m *Model) renderStatusBar() string {
	status := statusPaused
	if m.engine.IsBreached() {
		status = statusHalted
	} else if m.engine.IsRunning() {
		status = statusPaper
	}

	equity := m.engine.GetEquity()
	dailyPnL, trades, wins := m.engine.GetDailyStats()

	pnlStr := fmt.Sprintf("$%.2f", dailyPnL)
	pnlStyle := lipgloss.NewStyle().Foreground(green)
	if dailyPnL < 0 {
		pnlStyle = lipgloss.NewStyle().Foreground(red)
	}

	left := lipgloss.JoinHorizontal(lipgloss.Left,
		titleStyle.Render("Trading Bot"),
		" "+status,
		fmt.Sprintf(" │ Equity: $%.2f", equity),
		fmt.Sprintf(" │ Daily: %s", pnlStyle.Render(pnlStr)),
		fmt.Sprintf(" │ Trades: %d/%d", wins, trades),
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color("#1a1a2e")).
		Padding(0, 1).
		Render(left)
}

func (m *Model) renderMainView() string {
	halfW := m.width/2 - 2

	topLeft := m.renderPositions(halfW)
	topRight := m.renderSignals(halfW)
	botLeft := m.renderOrderBook(halfW)
	botRight := m.renderStrategies(halfW)

	top := lipgloss.JoinHorizontal(lipgloss.Top, topLeft, topRight)
	bottom := lipgloss.JoinHorizontal(lipgloss.Top, botLeft, botRight)

	return lipgloss.JoinVertical(lipgloss.Left, top, bottom)
}

func (m *Model) renderPositions(width int) string {
	box := boxStyle.Width(width).Height(10)
	positions := m.engine.GetPositions()

	if len(positions) == 0 {
		return box.Render(lipgloss.NewStyle().Foreground(gray).Render("No active positions"))
	}

	var lines []string
	for _, p := range positions {
		pnlColor := green
		pnlSign := "+"
		if p.UnrealizedPnL < 0 {
			pnlColor = red
			pnlSign = ""
		}
		line := fmt.Sprintf("%s %s %.4f @ %.2f %s",
			p.Side, p.Symbol, p.Amount, p.AvgPrice,
			pnlSign+lipgloss.NewStyle().Foreground(pnlColor).Render(fmt.Sprintf("%.2f", p.UnrealizedPnL)),
		)
		lines = append(lines, line)
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(blue).Render("POSITIONS")
	return box.Render(lipgloss.JoinVertical(lipgloss.Left, append([]string{title}, lines...)...))
}

func (m *Model) renderSignals(width int) string {
	box := boxStyle.Width(width).Height(10)
	signals := m.engine.GetActiveSignals()

	title := lipgloss.NewStyle().Bold(true).Foreground(blue).Render(fmt.Sprintf("SIGNALS (%d)", len(signals)))

	if len(signals) == 0 {
		return box.Render(lipgloss.JoinVertical(lipgloss.Left, title, lipgloss.NewStyle().Foreground(gray).Render("No active signals")))
	}

	var lines []string
	for _, s := range signals {
		dirColor := green
		if s.Direction == types.SignalDirectionShort {
			dirColor = red
		}
		dir := lipgloss.NewStyle().Foreground(dirColor).Render(string(s.Direction))
		line := fmt.Sprintf("%s %s %.0f%% %s",
			dir, s.Symbol, s.Confidence*100,
			truncate(s.Reason, 30),
		)
		lines = append(lines, line)
	}

	return box.Render(lipgloss.JoinVertical(lipgloss.Left, append([]string{title}, lines...)...))
}

func (m *Model) renderOrderBook(width int) string {
	box := boxStyle.Width(width).Height(8)
	title := lipgloss.NewStyle().Bold(true).Foreground(blue).Render("ORDER BOOK")
	return box.Render(lipgloss.JoinVertical(lipgloss.Left, title, lipgloss.NewStyle().Foreground(gray).Render("(WebSocket feed active)")))
}

func (m *Model) renderStrategies(width int) string {
	box := boxStyle.Width(width).Height(8)
	title := lipgloss.NewStyle().Bold(true).Foreground(blue).Render("STRATEGIES")

	names := m.engine.GetStrategyNames()
	if len(names) == 0 {
		return box.Render(lipgloss.JoinVertical(lipgloss.Left, title, lipgloss.NewStyle().Foreground(gray).Render("No strategies loaded")))
	}

	running := m.engine.IsRunning() && !m.engine.IsBreached()
	var lines []string
	for _, name := range names {
		status := "⏸ PAUSED"
		if running {
			status = "▶ RUNNING"
		}
		lines = append(lines, fmt.Sprintf("  %s %s", padRight(name, 20), status))
	}

	return box.Render(lipgloss.JoinVertical(lipgloss.Left, append([]string{title}, lines...)...))
}

func (m *Model) renderHelp() string {
	if m.help.ShowAll {
		return m.help.View(m.keys)
	}
	return lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color("#1a1a2e")).
		Padding(0, 1).
		Render("[Q]uit  [P]ause All  [S]tart All  [K]ill All  [H]elp")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}
