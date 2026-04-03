package catalog

import "github.com/gentleman-programming/gentle-ai/internal/model"

// Template describes a project-level scaffold template available via `gentle-ai init`.
type Template struct {
	ID          model.TemplateID
	Name        string
	Description string
	Stack       string
}

var allTemplates = []Template{
	{
		ID:          model.TemplateFastAPIPython,
		Name:        "fastapi-python",
		Description: "FastAPI + SQLAlchemy 2.0 + Alembic + Pydantic + uv",
		Stack:       "Python",
	},
	{
		ID:          model.TemplateReactNextJS,
		Name:        "react-nextjs",
		Description: "Next.js 15 + TypeScript + Tailwind 4 + Zustand + TanStack Query",
		Stack:       "TypeScript",
	},
	{
		ID:          model.TemplateGoBackend,
		Name:        "go-backend",
		Description: "Go + Chi/Gin + PostgreSQL + sqlc",
		Stack:       "Go",
	},
	{
		ID:          model.TemplateGoTUI,
		Name:        "go-tui",
		Description: "Go + Bubbletea + Charm toolchain",
		Stack:       "Go",
	},
	{
		ID:          model.TemplateMinimal,
		Name:        "minimal",
		Description: "stack_config.json + MANDATORY_CHECKS.md only",
		Stack:       "Any",
	},
}

func AllTemplates() []Template {
	templates := make([]Template, len(allTemplates))
	copy(templates, allTemplates)
	return templates
}

func FindTemplate(id model.TemplateID) (Template, bool) {
	for _, t := range allTemplates {
		if t.ID == id {
			return t, true
		}
	}
	return Template{}, false
}
