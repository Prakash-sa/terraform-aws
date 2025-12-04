# Helm Template Files - YAML Lint Errors Are Expected

The YAML files in this directory contain Helm/Go template syntax (`{{- }}`), which causes 
VS Code's YAML linter to report errors. **These are false positives** and can be safely ignored.

The templates are valid and will work correctly when processed by Helm.

To verify the templates are valid, use:
```bash
helm lint ./deploy/helm/app
helm template ./deploy/helm/app
```

## Valid Helm Syntax Examples

These patterns will show YAML errors but are correct:
- `{{- include "app.labels" . | nindent 4 }}`
- `{{- if .Values.autoscaling.enabled }}`
- `{{ .Values.image.repository }}`

The templates render to proper YAML when processed by Helm.
