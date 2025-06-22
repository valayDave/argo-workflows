package dynamic

import (
	"os"

	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"gopkg.in/yaml.v3"
)

// NodeGenerator defines an interface for objects that can generate additional
// workflow steps at runtime based on the current workflow and node name.
type NodeGenerator interface {
	Generate(wf *wfv1.Workflow, nodeName string) ([]wfv1.WorkflowStep, error)
}

// noopGenerator implements NodeGenerator and returns no new steps. It is used
// as the default generator when dynamic graph execution is not configured.
type noopGenerator struct{}

// Generate always returns no steps.
func (noopGenerator) Generate(wf *wfv1.Workflow, nodeName string) ([]wfv1.WorkflowStep, error) {
	return nil, nil
}

// FileGenerator loads dynamic steps from a YAML file. The YAML file should map
// node names to arrays of workflow steps. If Path is empty, the file path is
// read from the DYNAMIC_STEPS_PATH environment variable.
type FileGenerator struct {
	Path string
}

// Generate reads the mapping file and returns any steps for the given node.
func (f FileGenerator) Generate(wf *wfv1.Workflow, nodeName string) ([]wfv1.WorkflowStep, error) {
	path := f.Path
	if path == "" {
		path = os.Getenv("DYNAMIC_STEPS_PATH")
	}
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m := map[string][]wfv1.WorkflowStep{}
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m[nodeName], nil
}

// DefaultGenerator is the global node generator used by the workflow
// controller. It can be replaced at startup to provide custom implementations.
var DefaultGenerator NodeGenerator = noopGenerator{}

func init() {
	if os.Getenv("DYNAMIC_STEPS_PATH") != "" {
		DefaultGenerator = FileGenerator{}
	}
}
