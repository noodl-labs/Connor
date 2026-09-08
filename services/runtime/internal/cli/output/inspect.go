package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/noodl-labs/ConnorLLM/services/runtime/internal/runtime/domain/entities"
)

const inspectNameWidth = 24

// InspectView is the human-facing inspect report model.
type InspectView struct {
	Version string
	Report  entities.InspectReport
}

// PrintInspect renders a themed inspect report (Lipgloss when TTY; plain for CI/pipes).
// Display only — no PASS/FAIL gate lines (RFC 0003 §7).
func PrintInspect(w io.Writer, view InspectView) {
	theme := NewTheme(w)
	fmt.Fprintf(w, "%s  %s\n", theme.render(theme.bold, "Connor"), view.Version)

	rep := view.Report
	if allHTTPOnly(rep.Cases) {
		printInspectHTTPOnly(w, rep, theme)
		return
	}

	suite := rep.SuiteID
	if suite == "" {
		suite = "inspect"
	}
	if len(rep.Cases) == 1 {
		printInspectAgentCase(w, rep.Cases[0], suite, theme)
		return
	}

	fmt.Fprintf(w, "%s  %s  (%d cases)\n\n", theme.render(theme.dim, "Inspect"), suite, len(rep.Cases))
	for i, c := range rep.Cases {
		if i > 0 {
			fmt.Fprintln(w)
		}
		if c.HTTPOnly {
			printInspectHTTPCase(w, c, theme)
			continue
		}
		printInspectAgentCase(w, c, c.ID, theme)
	}
}

func printInspectHTTPOnly(w io.Writer, rep entities.InspectReport, theme Theme) {
	suite := rep.SuiteID
	if suite == "" {
		suite = "inspect"
	}
	fmt.Fprintf(w, "%s  %s\n\n", theme.render(theme.dim, "Inspect"), suite)
	fmt.Fprintf(w, "%s  HTTP-only artifact — no trajectory\n\n", theme.render(theme.dim, "note"))
	for _, c := range rep.Cases {
		printInspectHTTPCase(w, c, theme)
	}
}

func printInspectHTTPCase(w io.Writer, c entities.InspectCase, theme Theme) {
	icon := theme.render(theme.fail, "✗")
	if c.Passed {
		icon = theme.render(theme.pass, "✓")
	}
	reason := c.Reason
	if reason == "" {
		reason = "—"
	}
	fmt.Fprintf(w, "%s  %-12s  %-*s  %4dms  %s\n",
		icon,
		c.ID,
		modelColWidth,
		truncate(c.Model, modelColWidth),
		c.LatencyMs,
		theme.render(theme.dim, reason),
	)
}

func printInspectAgentCase(w io.Writer, c entities.InspectCase, label string, theme Theme) {
	runID := c.RunID
	if runID == "" {
		runID = "—"
	}
	fmt.Fprintf(w, "%s     %s    %s\n\n", theme.render(theme.dim, "RUN"), runID, label)

	status := "FAILED"
	statusStyle := theme.fail
	if c.Passed {
		status = "PASSED"
		statusStyle = theme.pass
	}
	fmt.Fprintf(w, "%s%s\n", padLabel("Status"), theme.render(statusStyle, status))
	fmt.Fprintf(w, "%s%s\n", padLabel("Duration"), formatInspectDuration(c.DurationMs))
	fmt.Fprintf(w, "%s%d\n", padLabel("LLM calls"), c.LLMCalls)
	fmt.Fprintf(w, "%s%d\n", padLabel("Tool calls"), c.ToolCalls)
	fmt.Fprintf(w, "%s%d\n\n", padLabel("Tokens"), c.Tokens)

	fmt.Fprintf(w, "%s\n", theme.render(theme.bold, "Trajectory"))
	for _, n := range c.Tree {
		printInspectNode(w, n, theme)
	}
}

func printInspectNode(w io.Writer, n entities.InspectNode, theme Theme) {
	indent := strings.Repeat("   ", n.Depth)
	fmt.Fprintf(w, "%s%s  %-*s  %dms\n",
		indent,
		inspectNodeIcon(n, theme),
		inspectNameWidth,
		n.Name,
		n.DurationMs,
	)
	detailIndent := indent + "   "
	if n.Error != "" {
		fmt.Fprintf(w, "%s%s  error  %s\n", detailIndent, theme.render(theme.dim, "└─"), n.Error)
	}
	if n.Duplicate {
		fmt.Fprintf(w, "%s%s  duplicate name (%s call)\n",
			detailIndent,
			theme.render(theme.dim, "└─"),
			ordinal(n.CallIndex),
		)
	}
}

func inspectNodeIcon(n entities.InspectNode, theme Theme) string {
	if n.Status == entities.SpanStatusError {
		return theme.render(theme.fail, "✗")
	}
	if n.Duplicate {
		return theme.render(theme.warn, "⚠")
	}
	return theme.render(theme.pass, "✓")
}

func padLabel(s string) string {
	return fmt.Sprintf("%-13s", s)
}

func formatInspectDuration(ms int64) string {
	if ms >= 1000 {
		return fmt.Sprintf("%.1fs", float64(ms)/1000)
	}
	return fmt.Sprintf("%dms", ms)
}

func allHTTPOnly(cases []entities.InspectCase) bool {
	if len(cases) == 0 {
		return true
	}
	for _, c := range cases {
		if !c.HTTPOnly {
			return false
		}
	}
	return true
}

func ordinal(n int) string {
	switch n {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	default:
		return fmt.Sprintf("%dth", n)
	}
}
