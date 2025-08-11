package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"bubblerea/inspector"
	"github.com/mattn/go-isatty"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	isLoading bool
	table     table.Model
	items     []inspector.PortInfo
	lastErr   error
}

func initialModel() model {
	columns := []table.Column{
		{Title: "Proto", Width: 6},
		{Title: "Local", Width: 28},
		{Title: "Remote", Width: 28},
		{Title: "State", Width: 12},
		{Title: "PID", Width: 8},
		{Title: "Command", Width: 20},
		{Title: "User", Width: 14},
	}

	styles := table.Styles{
		Header: lipgloss.NewStyle().Foreground(lipgloss.Color("213")).Bold(true),
		Selected: lipgloss.NewStyle().Background(lipgloss.Color("63")).Foreground(lipgloss.Color("230")),
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
	)
	t.SetStyles(styles)

	return model{isLoading: true, table: t}
}

// messages

type portsMsg struct{
	items []inspector.PortInfo
	err error
}

type tickMsg time.Time

// commands

func loadPortsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()
		items, err := inspector.ListPorts(ctx)
		return portsMsg{items: items, err: err}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m model) Init() tea.Cmd {
	return tea.Batch(loadPortsCmd(), tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			m.isLoading = true
			return m, loadPortsCmd()
		}
	case portsMsg:
		m.isLoading = false
		m.lastErr = msg.err
		m.items = msg.items
		rows := make([]table.Row, 0, len(m.items))
		for _, it := range m.items {
			rows = append(rows, table.Row{
				it.Protocol,
				fmt.Sprintf("%s:%s", it.LocalAddress, it.LocalPort),
				fmt.Sprintf("%s:%s", it.RemoteAddress, it.RemotePort),
				it.State,
				fmt.Sprintf("%d", it.PID),
				it.Command,
				it.User,
			})
		}
		m.table.SetRows(rows)
		return m, nil
	case tickMsg:
		return m, tea.Batch(loadPortsCmd(), tickCmd())
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	status := ""
	if m.isLoading {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("178")).Render(" Loading…")
	} else if m.lastErr != nil {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(" Error: "+m.lastErr.Error())
	} else {
		status = lipgloss.NewStyle().Foreground(lipgloss.Color("108")).Render(" r: refresh • q: quit")
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("213")).Render("bubblerea - busy ports"),
		m.table.View(),
		status,
	)
}

func runOnce() int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	items, err := inspector.ListPorts(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	// simple header
	fmt.Println("Proto\tLocal\tRemote\tState\tPID\tCommand\tUser")
	for _, it := range items {
		fmt.Printf("%s\t%s:%s\t%s:%s\t%s\t%d\t%s\t%s\n",
			it.Protocol,
			it.LocalAddress, it.LocalPort,
			it.RemoteAddress, it.RemotePort,
			it.State,
			it.PID,
			it.Command,
			it.User,
		)
	}
	return 0
}

func main() {
	// Non-interactive mode: if not a TTY or --once provided
	args := os.Args[1:]
	if !isatty.IsTerminal(os.Stdout.Fd()) || !isatty.IsTerminal(os.Stdin.Fd()) {
		os.Exit(runOnce())
	}
	for _, a := range args {
		if a == "--once" || a == "-1" { os.Exit(runOnce()) }
	}

	if _, err := tea.NewProgram(initialModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error:", err)
	}
}