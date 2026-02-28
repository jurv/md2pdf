package render

import "testing"

func TestWithHeadlessJavaEnvAddsVariable(t *testing.T) {
	in := []string{"PATH=/usr/bin"}
	out := withHeadlessJavaEnv(in)

	found := false
	for _, item := range out {
		if item == "JAVA_TOOL_OPTIONS=-Djava.awt.headless=true" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected JAVA_TOOL_OPTIONS to be injected, got %v", out)
	}
}

func TestWithHeadlessJavaEnvAppendsFlag(t *testing.T) {
	in := []string{"JAVA_TOOL_OPTIONS=-Xmx512m"}
	out := withHeadlessJavaEnv(in)

	if len(out) != 1 {
		t.Fatalf("unexpected env size: %d", len(out))
	}
	want := "JAVA_TOOL_OPTIONS=-Xmx512m -Djava.awt.headless=true"
	if out[0] != want {
		t.Fatalf("expected %q, got %q", want, out[0])
	}
}
