package main

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Output is a lightweight terminal output helper for the CLI. It replaces
// the previous *slog.Logger, which was never initialized and would panic on
// the first log call. Styles degrade automatically when the output is not a
// TTY or NO_COLOR is set.
type Output struct {
	w        io.Writer
	verbose  bool
	renderer *lipgloss.Renderer

	info lipgloss.Style
	dim  lipgloss.Style
}

func NewOutput(w io.Writer, verbose bool) *Output {
	if w == nil {
		w = os.Stdout
	}
	r := lipgloss.NewRenderer(w)
	return &Output{
		w:        w,
		verbose:  verbose,
		renderer: r,
		info:     r.NewStyle().Foreground(lipgloss.Color("#0d9488")).Bold(true),
		dim:      r.NewStyle().Foreground(lipgloss.Color("#6c757d")),
	}
}

func (o *Output) emit(s lipgloss.Style, format string, args ...any) {
	fmt.Fprintln(o.w, s.Render(fmt.Sprintf(format, args...)))
}

// Info prints a user-facing message (always shown).
func (o *Output) Info(format string, args ...any) {
	o.emit(o.info, format, args...)
}

// Verbose prints a detailed message only when verbose mode is enabled.
func (o *Output) Verbose(format string, args ...any) {
	if o.verbose {
		o.emit(o.dim, format, args...)
	}
}
