package model

// TemplateID identifies a project-level scaffold template.
type TemplateID string

const (
	TemplateFastAPIPython TemplateID = "fastapi-python"
	TemplateReactNextJS   TemplateID = "react-nextjs"
	TemplateGoBackend     TemplateID = "go-backend"
	TemplateGoTUI         TemplateID = "go-tui"
	TemplateMinimal       TemplateID = "minimal"
)

// Template describes a project-level scaffold template.
type Template struct {
	ID          TemplateID
	Name        string
	Description string
	Stack       string
}

// TemplateVars holds variables substituted into template files during scaffolding.
type TemplateVars struct {
	ProjectName    string
	PackageManager string
	AgentConfigDir string // e.g. ".claude", ".cursor/agents"
}
