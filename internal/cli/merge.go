package cli

import (
	"fmt"
	"strings"

	"github.com/julien/md2pdf/internal/pdf"
	"github.com/spf13/cobra"
)

type mergeFlags struct {
	Output string
}

func newMergeCmd(_ *GlobalOptions) *cobra.Command {
	flags := &mergeFlags{}
	cmd := &cobra.Command{
		Use:   "merge <file1.pdf> <file2.pdf> ...",
		Short: "Merge multiple PDF files into a single output",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.Output == "" {
				return userError("missing required output path", fmt.Errorf("use --output"))
			}
			if err := pdf.Merge(flags.Output, args); err != nil {
				if strings.Contains(err.Error(), "install pdftk or pdfunite") {
					return dependencyError("pdf merge backend missing", err)
				}
				return runtimeError("merge failed", err)
			}
			fmt.Printf("Merged PDF written to: %s\n", flags.Output)
			return nil
		},
	}
	cmd.Flags().StringVarP(&flags.Output, "output", "o", "", "Output PDF path")
	return cmd
}
