# Agent Changes

## Dynamic Graph Engine Prototype

- Added a new `workflow/dynamic` package defining a `NodeGenerator` interface. `DefaultGenerator` can be replaced to provide custom step generation logic.
- Introduced `FileGenerator` which reads a YAML file mapping node names to additional `WorkflowStep`s. The file path is taken from the `DYNAMIC_STEPS_PATH` environment variable.
- `workflow/controller/operator.go` now calls the generator when a node completes. Generated steps are executed sequentially and added as children of the completed node.

This remains an experimental feature. A full dynamic engine would require deeper integration with the workflow controller and API changes.

### Example Workflows

`examples/dynamic-file-generator.yaml` uses a simple file-based mapping defined in `examples/dynamic-steps.yaml`.

`examples/dynamic-template.yaml` shows the same approach using a `WorkflowTemplate`.

Run:

```bash
export DYNAMIC_STEPS_PATH=examples/dynamic-steps.yaml
argo submit --watch examples/dynamic-file-generator.yaml
# or
# argo submit --watch --from workflowtemplate/dynamic-template
```


### Development Notes

Running lint or vet locally requires a placeholder UI build. Execute `make ui/dist/app/index.html` before `go vet` to avoid missing file errors.
