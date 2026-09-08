package output

import (
	"io"
	"os"

	"charm.land/lipgloss/v2"
	"golang.org/x/term"
)

type Theme struct {
	enabled bool
	pass    lipgloss.Style
	fail    lipgloss.Style
	warn    lipgloss.Style
	dim     lipgloss.Style
	label   lipgloss.Style
	bold    lipgloss.Style
}

func NewTheme(w io.Writer) Theme {
	enabled := colorEnabled(w)
	return Theme{
		enabled: enabled,
		pass:    lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true), // vert
		fail:    lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true), // rouge
		warn:    lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true), // jaune
		dim:     lipgloss.NewStyle().Foreground(lipgloss.Color("8")),            // gris
		label:   lipgloss.NewStyle().Foreground(lipgloss.Color("6")),            // bleu labels
		bold:    lipgloss.NewStyle().Bold(true),
	}
}

func colorEnabled(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false // tests → bytes.Buffer
	}
	return term.IsTerminal(int(f.Fd()))
}

func (t Theme) render(style lipgloss.Style, s string) string {
	if !t.enabled {
		return s
	}
	return style.Render(s)
}
