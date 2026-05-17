package exec

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Runner struct {
	DryRun  bool
	Verbose bool
}

type Result struct {
	Command  string
	Args     []string
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
	Err      error
}

func (r *Runner) Run(name string, args ...string) *Result {
	res := &Result{Command: name, Args: args}

	if r.DryRun {
		return res
	}

	start := time.Now()

	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer

	if r.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	err := cmd.Run()
	res.Duration = time.Since(start)
	res.Stdout = stdout.String()
	res.Stderr = stderr.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
		} else {
			res.ExitCode = 1
		}
		res.Err = fmt.Errorf("%s failed (exit %d): %s", name, res.ExitCode, res.Stderr)
	}

	return res
}

func (r *Runner) RunWithSpinner(label, name string, args ...string) *Result {
	res := &Result{Command: name, Args: args}

	if r.DryRun {
		return res
	}

	if r.Verbose {
		return r.Run(name, args...)
	}

	done := make(chan struct{})
	go func() {
		p := tea.NewProgram(newSpinnerModel(label), tea.WithOutput(os.Stderr))
		go func() {
			<-done
			p.Send(tea.Quit())
		}()
		p.Run()
	}()

	start := time.Now()
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	close(done)

	res.Duration = time.Since(start)
	res.Stdout = stdout.String()
	res.Stderr = stderr.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
		} else {
			res.ExitCode = 1
		}
		res.Err = fmt.Errorf("%s failed (exit %d): %s", name, res.ExitCode, res.Stderr)
	}

	return res
}

type spinnerModel struct {
	spinner spinner.Model
	label   string
}

func newSpinnerModel(label string) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return spinnerModel{spinner: s, label: label}
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.QuitMsg:
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m spinnerModel) View() string {
	return m.spinner.View() + " " + m.label
}
