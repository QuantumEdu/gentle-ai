package projectinit

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gentleman-programming/gentle-ai/internal/catalog"
	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// Options configures the Init operation.
type Options struct {
	// TemplateID selects which project template to scaffold.
	TemplateID model.TemplateID
	// ProjectName is used as {{.ProjectName}} in template files.
	// Defaults to the current directory name if empty.
	ProjectName string
	// AgentID controls which config directory is targeted (.claude/, .cursor/, etc.).
	// Defaults to claude-code (.claude/).
	AgentID model.AgentID
	// Force overwrites existing files without prompting.
	Force bool
	// DryRun lists files that would be created without writing them.
	DryRun bool
}

// Result holds the outcome of an Init operation.
type Result struct {
	TemplateID model.TemplateID
	TargetDir  string
	Files      []string
	DryRun     bool
}

// Run scaffolds the project template into the current working directory.
func Run(opts Options) (Result, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Result{}, fmt.Errorf("get working directory: %w", err)
	}

	// Resolve project name
	projectName := opts.ProjectName
	if projectName == "" {
		projectName = filepath.Base(cwd)
	}

	// Resolve agent config directory
	agentConfigDir := agentConfigDirFor(opts.AgentID)

	// Validate template exists
	tmpl, ok := catalog.FindTemplate(opts.TemplateID)
	if !ok {
		return Result{}, fmt.Errorf("unknown template %q — run 'gentle-ai init --list' to see available templates", opts.TemplateID)
	}

	vars := model.TemplateVars{
		ProjectName:    projectName,
		AgentConfigDir: agentConfigDir,
	}

	if opts.DryRun {
		files, err := listTemplateFiles(tmpl.ID)
		if err != nil {
			return Result{}, err
		}
		return Result{
			TemplateID: tmpl.ID,
			TargetDir:  filepath.Join(cwd, agentConfigDir),
			Files:      files,
			DryRun:     true,
		}, nil
	}

	w := &Writer{
		TargetDir: cwd,
		Force:     opts.Force,
	}

	if err := w.Write(tmpl.ID, vars); err != nil {
		return Result{}, fmt.Errorf("scaffold template: %w", err)
	}

	files, _ := listTemplateFiles(tmpl.ID)

	return Result{
		TemplateID: tmpl.ID,
		TargetDir:  filepath.Join(cwd, agentConfigDir),
		Files:      files,
	}, nil
}

// agentConfigDirFor returns the config subdirectory for a given agent.
// Defaults to .claude/ for unknown or unset agents.
func agentConfigDirFor(agent model.AgentID) string {
	switch agent {
	case model.AgentCursor:
		return ".cursor"
	case model.AgentOpenCode:
		return ".opencode"
	case model.AgentGeminiCLI:
		return ".gemini"
	case model.AgentCodex:
		return ".codex"
	default:
		return ".claude"
	}
}

// listTemplateFiles returns the relative paths of all files in a template.
func listTemplateFiles(id model.TemplateID) ([]string, error) {
	templateDir := "templates/" + string(id)
	var files []string

	entries, err := FS.ReadDir(templateDir)
	if err != nil {
		return nil, err
	}

	_ = entries
	// WalkDir collects all leaf files
	var walkFn func(string) error
	walkFn = func(dir string) error {
		entries, err := FS.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			path := dir + "/" + e.Name()
			if e.IsDir() {
				if err := walkFn(path); err != nil {
					return err
				}
			} else {
				// Strip "templates/<id>/" prefix
				rel := path[len(templateDir)+1:]
				files = append(files, rel)
			}
		}
		return nil
	}

	if err := walkFn(templateDir); err != nil {
		return nil, err
	}

	return files, nil
}
