package cli

import (
	"fmt"
	"os"

	"github.com/julien/md2pdf/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type initFlags struct {
	Profile string
	Output  string
	Force   bool
}

func newInitCmd(_ *GlobalOptions) *cobra.Command {
	flags := &initFlags{Profile: "default", Output: "md2pdf.yaml"}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a starter configuration file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInit(flags)
		},
	}

	cmd.Flags().StringVar(&flags.Profile, "profile", "default", "Profile: default, report, meeting")
	cmd.Flags().StringVarP(&flags.Output, "output", "o", "md2pdf.yaml", "Config output path")
	cmd.Flags().BoolVar(&flags.Force, "force", false, "Overwrite output file if it already exists")

	return cmd
}

func runInit(flags *initFlags) error {
	cfg, err := profileConfig(flags.Profile)
	if err != nil {
		return userError("invalid profile", err)
	}
	if _, err := os.Stat(flags.Output); err == nil && !flags.Force {
		return userError("output file already exists", fmt.Errorf("use --force to overwrite %s", flags.Output))
	}

	blob, err := yaml.Marshal(cfg)
	if err != nil {
		return runtimeError("failed to serialize configuration", err)
	}
	if err := os.WriteFile(flags.Output, blob, 0o644); err != nil {
		return runtimeError("failed to write configuration file", err)
	}

	fmt.Printf("Configuration file created: %s\n", flags.Output)
	return nil
}

func profileConfig(profile string) (config.Config, error) {
	cfg := config.Default()
	switch profile {
	case "default":
		return cfg, nil
	case "report":
		cfg.TOC.Mode = "on"
		cfg.TOC.Title = "Contents"
		cfg.TOC.Depth = 3
		cfg.Style.Colors.Primary = "#1F4E79"
		return cfg, nil
	case "meeting":
		cfg.TOC.Mode = "off"
		cfg.Metadata.Subject = "Meeting Notes"
		return cfg, nil
	default:
		return config.Config{}, fmt.Errorf("unknown profile %q", profile)
	}
}
