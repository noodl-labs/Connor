package output

import (
	"fmt"
	"io"

	"github.com/noodl-labs/ConnorLLM/services/runtime/internal/runtime/domain/entities"
)

// CompareView is the human-facing compare report model.
type CompareView struct {
	Version   string
	SuiteID   string
	Baseline  entities.RunSummary
	Candidate entities.RunSummary
	Result    entities.CompareResult
}

// PrintCompare renders a themed compare report (Lipgloss when TTY; plain for CI/pipes).
// Gate lines keep greppable PASS/FAIL prefixes (CI contract).
func PrintCompare(w io.Writer, view CompareView) {
	theme := NewTheme(w)
	printCompareHeader(w, view, theme)
	printCompareP95(w, view.Result.P95, theme)
	printComparePassRate(w, view.Result.PassRate, theme)
	printCompareFooter(w, view.Result.Passed, theme)
}

func printCompareHeader(w io.Writer, view CompareView, theme Theme) {
	fmt.Fprintf(w, "%s  %s\n", theme.render(theme.bold, "Connor"), view.Version)
	suite := view.SuiteID
	if suite == "" {
		suite = "compare"
	}
	fmt.Fprintf(w, "%s  %s\n", theme.render(theme.dim, "Compare"), suite)
	fmt.Fprintf(w, "%s  p95 %dms · pass rate %.0f%%\n",
		theme.render(theme.dim, "        baseline"),
		view.Baseline.P95Ms,
		view.Baseline.PassRate,
	)
	fmt.Fprintf(w, "%s  p95 %dms · pass rate %.0f%%\n\n",
		theme.render(theme.dim, "       candidate"),
		view.Candidate.P95Ms,
		view.Candidate.PassRate,
	)
}

func printCompareP95(w io.Writer, p95 entities.P95CompareResult, theme Theme) {
	if !p95.Checked {
		status := theme.render(theme.dim, "PASS")
		fmt.Fprintf(w, "%s  p95 %s %s\n",
			status,
			formatDelta(p95.DeltaPercent),
			theme.render(theme.dim, "(no threshold set)"),
		)
		return
	}
	if p95.Passed {
		fmt.Fprintf(w, "%s  p95 %s\n",
			theme.render(theme.pass, "PASS"),
			formatDelta(p95.DeltaPercent),
		)
		return
	}
	fmt.Fprintf(w, "%s  p95 %s  (threshold: %.0f%%)\n",
		theme.render(theme.fail, "FAIL"),
		formatDelta(p95.DeltaPercent),
		p95.Threshold,
	)
	if p95.Driver.Found {
		fmt.Fprintf(w, "      %s  %s  %s  %dms → %dms  (%s)\n",
			theme.render(theme.label, "driver"),
			p95.Driver.CaseID,
			p95.Driver.Model,
			p95.Driver.BaselineMs,
			p95.Driver.CandidateMs,
			formatDelta(p95.Driver.DeltaPercent),
		)
	}
}

func printComparePassRate(w io.Writer, pr entities.PassRateCompareResult, theme Theme) {
	if !pr.Checked {
		return
	}
	if pr.Passed {
		fmt.Fprintf(w, "%s  pass rate %.0f%%\n",
			theme.render(theme.pass, "PASS"),
			pr.CandidatePassRate,
		)
		return
	}
	fmt.Fprintf(w, "%s  pass rate %.0f%%  (threshold: %.0f%%)\n",
		theme.render(theme.fail, "FAIL"),
		pr.CandidatePassRate,
		pr.Threshold,
	)
}

func printCompareFooter(w io.Writer, passed bool, theme Theme) {
	fmt.Fprintln(w, theme.render(theme.dim, "────────────────────────────────────"))
	if passed {
		fmt.Fprintln(w, theme.render(theme.pass, "GATE PASSED — safe to merge"))
		fmt.Fprintln(w, theme.render(theme.dim, "exit 0"))
		return
	}
	fmt.Fprintln(w, theme.render(theme.fail, "GATE FAILED — do not merge"))
	fmt.Fprintln(w, theme.render(theme.dim, "exit 1"))
}

func formatDelta(pct float64) string {
	if pct >= 0 {
		return fmt.Sprintf("+%.0f%%", pct)
	}
	return fmt.Sprintf("%.0f%%", pct)
}
