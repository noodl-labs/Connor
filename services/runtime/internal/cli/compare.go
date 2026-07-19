package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/noodl-labs/ConnorLLM/services/runtime/internal/cli/output"
	"github.com/noodl-labs/ConnorLLM/services/runtime/internal/runtime/domain/entities"
)

func newCompareCmd() *cobra.Command {
	var maxP95Regression float64
	var minPassRate float64

	cmd := &cobra.Command{
		Use:   "compare baseline.json candidate.json",
		Short: "Compare two run.json artifacts (baseline vs candidate)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			baseline, err := loadRunArtifact(args[0])
			if err != nil {
				return exitCompareUsage(err)
			}
			candidate, err := loadRunArtifact(args[1])
			if err != nil {
				return exitCompareUsage(err)
			}

			var maxP95 *float64
			if cmd.Flags().Changed("max-p95-regression") {
				maxP95 = &maxP95Regression
			}
			var minPR *float64
			if cmd.Flags().Changed("min-pass-rate") {
				minPR = &minPassRate
			}

			result, err := entities.CompareRuns(baseline, candidate, maxP95, minPR)
			if err != nil {
				return exitCompareUsage(err)
			}

			output.PrintCompare(os.Stdout, output.CompareView{
				Version:   Version,
				SuiteID:   candidate.SuiteID,
				Baseline:  baseline.Summary,
				Candidate: candidate.Summary,
				Result:    result,
			})
			if !result.Passed {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().Float64Var(
		&maxP95Regression, "max-p95-regression", 0,
		"Fail if p95 latency regression exceeds this percent (e.g. 20)",
	)
	cmd.Flags().Float64Var(
		&minPassRate, "min-pass-rate", 0,
		"Fail if candidate pass rate is below this percent (e.g. 95)",
	)

	return cmd
}

func loadRunArtifact(path string) (entities.RunArtifact, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return entities.RunArtifact{}, fmt.Errorf("connor: read %s: %w", path, err)
	}
	return entities.ParseRunArtifactJSON(data)
}

// exitCompareUsage prints err and exits with code 2 (RFC 0001 §2).
func exitCompareUsage(err error) error {
	fmt.Fprintf(os.Stderr, "connor compare: %v\n", err)
	os.Exit(2)
	return nil
}
