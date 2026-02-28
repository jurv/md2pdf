package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/julien/md2pdf/internal/deps"
	"github.com/spf13/cobra"
)

type doctorFlags struct {
	JSON      bool
	PDFEngine string
}

func newDoctorCmd(_ *GlobalOptions) *cobra.Command {
	flags := &doctorFlags{PDFEngine: "xelatex"}
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check required and optional runtime dependencies",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDoctor(flags)
		},
	}

	cmd.Flags().BoolVar(&flags.JSON, "json", false, "Emit dependency report as JSON")
	cmd.Flags().StringVar(&flags.PDFEngine, "pdf-engine", "xelatex", "Default PDF engine to validate")

	return cmd
}

func runDoctor(flags *doctorFlags) error {
	statuses := deps.CollectDoctorStatuses(flags.PDFEngine)
	missingRequired := false
	for _, status := range statuses {
		if status.Required && !status.Available {
			missingRequired = true
		}
	}

	if flags.JSON {
		payload := map[string]any{
			"engine":           flags.PDFEngine,
			"missing_required": missingRequired,
			"dependencies":     statuses,
		}
		blob, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return runtimeError("failed to format JSON report", err)
		}
		fmt.Println(string(blob))
	} else {
		writer := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
		fmt.Fprintln(writer, "NAME\tREQUIRED\tSTATUS\tVERSION\tDETAIL")
		for _, status := range statuses {
			state := "missing"
			if status.Available {
				state = "ok"
			}
			fmt.Fprintf(writer, "%s\t%t\t%s\t%s\t%s\n",
				status.Name,
				status.Required,
				state,
				status.Version,
				status.Message,
			)
		}
		writer.Flush()
	}

	if missingRequired {
		return dependencyError("one or more required dependencies are missing", nil)
	}
	return nil
}
