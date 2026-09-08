package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "connor",
		Short: "ConnorLLM — LLM runtime reliability benchmarks",
	}
	root.AddCommand(newRunCmd())
	root.AddCommand(newCompareCmd())
	root.AddCommand(newInspectCmd())
	return root
}

func Execute() {
	if err := NewRoot().Execute(); err != nil {
		os.Exit(1)
	}
}
