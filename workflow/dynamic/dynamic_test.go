package dynamic

import (
	"os"
	"testing"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
)

func TestFileGenerator(t *testing.T) {
	content := "step:\n  - name: dyn\n    template: whalesay\n"
	f, err := os.CreateTemp("", "dynsteps*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()

	g := FileGenerator{Path: f.Name()}
	steps, err := g.Generate(&wfv1.Workflow{}, "step")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(steps))
	}
	if steps[0].Name != "dyn" {
		t.Fatalf("unexpected step name: %s", steps[0].Name)
	}
}
