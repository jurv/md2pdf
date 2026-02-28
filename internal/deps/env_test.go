package deps

import "testing"

func TestDepsHeadlessEnvNoDuplicate(t *testing.T) {
	in := []string{"JAVA_TOOL_OPTIONS=-Djava.awt.headless=true"}
	out := withHeadlessJavaEnv(in)
	if len(out) != 1 {
		t.Fatalf("unexpected env size: %d", len(out))
	}
	if out[0] != in[0] {
		t.Fatalf("expected unchanged env, got %q", out[0])
	}
}
