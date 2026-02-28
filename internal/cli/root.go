package cli

import (
	"github.com/spf13/cobra"
)

var version = "dev"

type GlobalOptions struct {
	Verbose bool
	Debug   bool
}

func Execute() (int, error) {
	cmd := NewRootCmd()
	err := cmd.Execute()
	if err != nil {
		return errorExitCode(err), err
	}
	return ExitOK, nil
}

func NewRootCmd() *cobra.Command {
	opts := &GlobalOptions{}

	rootCmd := &cobra.Command{
		Use:           "md2pdf",
		Short:         "Generate PDFs from Markdown with layered YAML configuration",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       version,
	}

	rootCmd.PersistentFlags().BoolVar(&opts.Verbose, "verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&opts.Debug, "debug", false, "Enable debug-level output")

	rootCmd.AddCommand(newBuildCmd(opts))
	rootCmd.AddCommand(newDoctorCmd(opts))
	rootCmd.AddCommand(newMergeCmd(opts))
	rootCmd.AddCommand(newCompressCmd(opts))
	rootCmd.AddCommand(newInitCmd(opts))

	return rootCmd
}
