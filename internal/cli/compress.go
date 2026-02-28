package cli

import (
	"fmt"
	"strings"

	"github.com/julien/md2pdf/internal/pdf"
	"github.com/spf13/cobra"
)

type compressFlags struct {
	Output  string
	Quality string
}

func newCompressCmd(_ *GlobalOptions) *cobra.Command {
	flags := &compressFlags{Quality: "printer"}
	cmd := &cobra.Command{
		Use:   "compress <input.pdf>",
		Short: "Compress a PDF file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.Output == "" {
				return userError("missing required output path", fmt.Errorf("use --output"))
			}
			if err := pdf.Compress(args[0], flags.Output, flags.Quality); err != nil {
				if strings.Contains(err.Error(), "ghostscript") {
					return dependencyError("ghostscript is required", err)
				}
				return runtimeError("compression failed", err)
			}
			fmt.Printf("Compressed PDF written to: %s\n", flags.Output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&flags.Output, "output", "o", "", "Output PDF path")
	cmd.Flags().StringVar(&flags.Quality, "quality", "printer", "Compression profile: screen, ebook, printer, prepress")
	return cmd
}
