package deps

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Status struct {
	Name      string `json:"name"`
	Required  bool   `json:"required"`
	Available bool   `json:"available"`
	Path      string `json:"path,omitempty"`
	Version   string `json:"version,omitempty"`
	Message   string `json:"message,omitempty"`
}

type MissingDependencyError struct {
	Names []string
}

func (e MissingDependencyError) Error() string {
	return fmt.Sprintf("missing required dependencies: %s", strings.Join(e.Names, ", "))
}

func CollectDoctorStatuses(defaultEngine string) []Status {
	checks := []struct {
		name       string
		required   bool
		versionArg []string
		probe      bool
		note       string
	}{
		{name: "pandoc", required: true, versionArg: []string{"--version"}, probe: true, note: "Required for markdown to PDF conversion."},
		{name: "xelatex", required: defaultEngine == "xelatex", versionArg: []string{"--version"}, probe: true, note: "PDF engine (recommended default)."},
		{name: "lualatex", required: defaultEngine == "lualatex", versionArg: []string{"--version"}, probe: true, note: "Optional PDF engine."},
		{name: "pdflatex", required: defaultEngine == "pdflatex", versionArg: []string{"--version"}, probe: true, note: "Optional PDF engine."},
		{name: "pandoc-plantuml", required: false, versionArg: nil, probe: false, note: "Required when rendering PlantUML diagrams."},
		{name: "plantuml", required: false, versionArg: []string{"-version"}, probe: true, note: "Required when rendering PlantUML diagrams."},
		{name: "java", required: false, versionArg: []string{"-version"}, probe: true, note: "Runtime often required by PlantUML."},
		{name: "dot", required: false, versionArg: []string{"-V"}, probe: true, note: "Graphviz binary used by PlantUML."},
		{name: "pdftk", required: false, versionArg: []string{"--version"}, probe: true, note: "Used by the merge command."},
		{name: "gs", required: false, versionArg: []string{"--version"}, probe: true, note: "Used by the compress command."},
	}

	out := make([]Status, 0, len(checks))
	for _, check := range checks {
		status := inspectCommand(check.name, check.required, check.versionArg, check.probe)
		if status.Message == "" {
			status.Message = check.note
		}
		if check.name == "pandoc-plantuml" && status.Available {
			status.Message += " Version probing is intentionally disabled because this binary crashes when called with --version."
		}
		out = append(out, status)
	}
	return out
}

func EnsureBuildDependencies(engine string, needsPlantUML bool) error {
	required := []string{"pandoc", engine}

	if needsPlantUML {
		required = append(required, "pandoc-plantuml", "plantuml", "dot")
	}

	missing := make([]string, 0)
	for _, dep := range required {
		if !hasCommand(dep) {
			missing = append(missing, dep)
		}
	}
	if len(missing) > 0 {
		return MissingDependencyError{Names: missing}
	}
	return nil
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func inspectCommand(name string, required bool, versionArg []string, probeVersion bool) Status {
	status := Status{Name: name, Required: required}
	path, err := exec.LookPath(name)
	if err != nil {
		status.Available = false
		status.Message = "Command not found in PATH."
		return status
	}
	status.Available = true
	status.Path = path

	if !probeVersion {
		return status
	}

	version, err := commandVersion(name, versionArg...)
	if err == nil {
		status.Version = version
	}
	return status
}

func commandVersion(cmdName string, args ...string) (string, error) {
	cmd := exec.Command(cmdName, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stdout.Len() == 0 && stderr.Len() == 0 {
			return "", err
		}
	}
	joined := strings.TrimSpace(stdout.String())
	if joined == "" {
		joined = strings.TrimSpace(stderr.String())
	}
	if joined == "" {
		return "", nil
	}
	for _, line := range strings.Split(joined, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line, nil
		}
	}
	return "", nil
}
