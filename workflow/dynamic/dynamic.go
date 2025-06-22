package dynamic

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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

// ContainerGenerator runs a Docker container to produce workflow steps. The
// container's stdout must contain a YAML array of `WorkflowStep` objects.
// The workflow and node names are provided via WORKFLOW_NAME and NODE_NAME
// environment variables.
type ContainerGenerator struct {
	Image   string
	Command []string
}

// Generate executes the container and parses the resulting steps.
func (c ContainerGenerator) Generate(wf *wfv1.Workflow, nodeName string) ([]wfv1.WorkflowStep, error) {
	if c.Image == "" {
		return nil, nil
	}
	args := []string{"run", "--rm", c.Image}
	args = append(args, c.Command...)
	cmd := exec.Command("docker", args...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("WORKFLOW_NAME=%s", wf.Name),
		fmt.Sprintf("NODE_NAME=%s", nodeName),
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var steps []wfv1.WorkflowStep
	if err := yaml.Unmarshal(out, &steps); err != nil {
		return nil, err
	}
	return steps, nil
}

// DefaultGenerator is the global node generator used by the workflow
// controller. It can be replaced at startup to provide custom implementations.
var DefaultGenerator NodeGenerator = noopGenerator{}

func init() {
	if image := os.Getenv("DYNAMIC_GENERATOR_IMAGE"); image != "" {
		cmdStr := os.Getenv("DYNAMIC_GENERATOR_COMMAND")
		var cmd []string
		if cmdStr != "" {
			cmd = strings.Split(cmdStr, " ")
		}
		DefaultGenerator = ContainerGenerator{Image: image, Command: cmd}
	} else if os.Getenv("DYNAMIC_STEPS_PATH") != "" {
		DefaultGenerator = FileGenerator{}
	}
}
