package projectinit

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// Writer scaffolds a project template into the target directory.
type Writer struct {
	// TargetDir is the project root (where .claude/ will be created).
	TargetDir string
	// Force overwrites existing files without prompting.
	Force bool
}

// Write renders all template files for the given template ID into TargetDir.
// Note: embed.FS always uses forward slashes regardless of OS.
func (w *Writer) Write(tmplID model.TemplateID, vars model.TemplateVars) error {
	templateDir := "templates/" + string(tmplID)

	entries, err := fs.ReadDir(FS, templateDir)
	if err != nil {
		return fmt.Errorf("template %q not found: %w", tmplID, err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("template %q is empty", tmplID)
	}

	return fs.WalkDir(FS, templateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		return w.writeFile(path, templateDir, vars)
	})
}

// writeFile renders one template file and writes it to the target directory.
func (w *Writer) writeFile(embeddedPath, templateDir string, vars model.TemplateVars) error {
	// Compute destination path: strip templates/<id>/ prefix, prepend .claude/
	rel := strings.TrimPrefix(embeddedPath, templateDir+"/")
	destPath := filepath.Join(w.TargetDir, vars.AgentConfigDir, rel)

	if !w.Force {
		if _, err := os.Stat(destPath); err == nil {
			return fmt.Errorf("file already exists: %s (use --force to overwrite)", destPath)
		}
	}

	// Read raw template content from embed.FS
	raw, err := FS.ReadFile(embeddedPath)
	if err != nil {
		return fmt.Errorf("read template file %s: %w", embeddedPath, err)
	}

	// Render template variables
	rendered, err := renderTemplate(string(raw), vars)
	if err != nil {
		return fmt.Errorf("render template %s: %w", embeddedPath, err)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("create directory for %s: %w", destPath, err)
	}

	// Write the file
	if err := os.WriteFile(destPath, []byte(rendered), 0644); err != nil {
		return fmt.Errorf("write %s: %w", destPath, err)
	}

	// Make shell scripts executable
	if strings.HasSuffix(destPath, ".sh") {
		if err := os.Chmod(destPath, 0755); err != nil {
			return fmt.Errorf("chmod +x %s: %w", destPath, err)
		}
	}

	return nil
}

// renderTemplate substitutes TemplateVars into a template string.
func renderTemplate(content string, vars model.TemplateVars) (string, error) {
	tmpl, err := template.New("").Option("missingkey=zero").Parse(content)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", err
	}

	return buf.String(), nil
}
