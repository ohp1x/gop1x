package exec

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const gracefulTimeout = 5 * time.Second

type Runner struct {
	DryRun  bool
	Verbose bool
}

func isRoot() bool {
	return os.Geteuid() == 0
}

type Result struct {
	Command    string
	Args       []string
	Stdout     string
	Stderr     string
	ExitCode   int
	Duration   time.Duration
	Err        error
	Cancelled  bool
}

func (r *Runner) Run(name string, args ...string) *Result {
	res := &Result{Command: name, Args: args}

	if r.DryRun {
		return res
	}

	start := time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var stdout, stderr bytes.Buffer

	if r.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	errCh := make(chan error, 1)
	if err := cmd.Start(); err != nil {
		res.ExitCode = 1
		res.Err = fmt.Errorf("%s: %w", name, err)
		return res
	}

	go func() { errCh <- cmd.Wait() }()

	select {
	case err := <-errCh:
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
	case <-sigCh:
		res.Cancelled = true
		r.terminateProcess(cmd)
		<-errCh
		res.Duration = time.Since(start)
		res.Stdout = stdout.String()
		res.Stderr = stderr.String()
		res.ExitCode = 130
		res.Err = fmt.Errorf("%s: interrupted", name)
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	errCh := make(chan error, 1)
	if err := cmd.Start(); err != nil {
		close(done)
		res.ExitCode = 1
		res.Err = fmt.Errorf("%s: %w", name, err)
		return res
	}

	go func() { errCh <- cmd.Wait() }()

	select {
	case err := <-errCh:
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
	case <-sigCh:
		close(done)
		res.Cancelled = true
		r.terminateProcess(cmd)
		<-errCh
		res.Duration = time.Since(start)
		res.Stdout = stdout.String()
		res.Stderr = stderr.String()
		res.ExitCode = 130
		res.Err = fmt.Errorf("%s: interrupted", name)
	}

	return res
}

func (r *Runner) RunElevated(label, name string, args ...string) *Result {
	if isRoot() {
		return r.RunWithSpinner(label, name, args...)
	}
	sudoArgs := append([]string{name}, args...)
	res := &Result{Command: "sudo", Args: sudoArgs}

	if r.DryRun {
		return res
	}

	start := time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, "sudo", sudoArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	errCh := make(chan error, 1)
	if err := cmd.Start(); err != nil {
		res.ExitCode = 1
		res.Err = fmt.Errorf("sudo %s: %w", name, err)
		return res
	}

	go func() { errCh <- cmd.Wait() }()

	select {
	case err := <-errCh:
		res.Duration = time.Since(start)
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				res.ExitCode = exitErr.ExitCode()
			} else {
				res.ExitCode = 1
			}
			res.Err = fmt.Errorf("sudo %s failed (exit %d)", name, res.ExitCode)
		}
	case <-sigCh:
		res.Cancelled = true
		r.terminateProcess(cmd)
		<-errCh
		res.Duration = time.Since(start)
		res.ExitCode = 130
		res.Err = fmt.Errorf("sudo %s: interrupted", name)
	}

	return res
}

func (r *Runner) terminateProcess(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err == nil {
		syscall.Kill(-pgid, syscall.SIGTERM)
	} else {
		cmd.Process.Signal(syscall.SIGTERM)
	}

	done := make(chan struct{})
	go func() {
		cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(gracefulTimeout):
		if pgid, err := syscall.Getpgid(cmd.Process.Pid); err == nil {
			syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			cmd.Process.Kill()
		}
		<-done
	}
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
