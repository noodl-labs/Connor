package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/noodl-labs/ConnorLLM/services/runtime/internal/cli/output"
	"github.com/noodl-labs/ConnorLLM/services/runtime/internal/runtime/domain/entities"
)

func newInspectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect run.json",
		Short: "Explain a recorded run.json (no LLM, display only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			artifact, err := loadRunArtifact(args[0])
			if err != nil {
				return exitInspectUsage(err)
			}
			output.PrintInspect(os.Stdout, output.InspectView{
				Version: Version,
				Report:  entities.InspectViewFromArtifact(artifact),
			})
			return nil
		},
	}
}

// exitInspectUsage prints err and exits with code 2 (RFC 0003 §7).
func exitInspectUsage(err error) error {
	fmt.Fprintf(os.Stderr, "connor inspect: %v\n", err)
	os.Exit(2)
	return nil
}
