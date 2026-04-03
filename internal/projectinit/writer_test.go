package projectinit_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/projectinit"
)

func TestWriter_Write_MinimalTemplate(t *testing.T) {
	dir := t.TempDir()

	w := &projectinit.Writer{TargetDir: dir, Force: false}
	vars := model.TemplateVars{
		ProjectName:    "test-project",
		AgentConfigDir: ".claude",
	}

	if err := w.Write(model.TemplateMinimal, vars); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// stack_config.json must exist and contain the project name
	stackPath := filepath.Join(dir, ".claude", "stack_config.json")
	content, err := os.ReadFile(stackPath)
	if err != nil {
		t.Fatalf("stack_config.json not created: %v", err)
	}
	if !contains(string(content), "test-project") {
		t.Errorf("stack_config.json does not contain project name")
	}

	// MANDATORY_CHECKS.md must exist
	checksPath := filepath.Join(dir, ".claude", "MANDATORY_CHECKS.md")
	if _, err := os.Stat(checksPath); err != nil {
		t.Fatalf("MANDATORY_CHECKS.md not created: %v", err)
	}
}

func TestWriter_Write_FastAPIPythonTemplate(t *testing.T) {
	dir := t.TempDir()

	w := &projectinit.Writer{TargetDir: dir, Force: false}
	vars := model.TemplateVars{
		ProjectName:    "my-api",
		AgentConfigDir: ".claude",
	}

	if err := w.Write(model.TemplateFastAPIPython, vars); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	expectedFiles := []string{
		filepath.Join(dir, ".claude", "stack_config.json"),
		filepath.Join(dir, ".claude", "MANDATORY_CHECKS.md"),
		filepath.Join(dir, ".claude", "agents", "architect.md"),
		filepath.Join(dir, ".claude", "agents", "backend_developer.md"),
		filepath.Join(dir, ".claude", "agents", "security_expert.md"),
		filepath.Join(dir, ".claude", "agents", "test_engineer.md"),
		filepath.Join(dir, ".claude", "validators", "check_stack.sh"),
		filepath.Join(dir, ".claude", "swarms", "create_module.yaml"),
	}

	for _, path := range expectedFiles {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file not created: %s", path)
		}
	}
}

func TestWriter_Write_ForceOverwrite(t *testing.T) {
	dir := t.TempDir()
	vars := model.TemplateVars{ProjectName: "proj", AgentConfigDir: ".claude"}

	w := &projectinit.Writer{TargetDir: dir, Force: false}
	if err := w.Write(model.TemplateMinimal, vars); err != nil {
		t.Fatalf("first Write() error: %v", err)
	}

	// Second write without force should fail
	if err := w.Write(model.TemplateMinimal, vars); err == nil {
		t.Error("expected error on second write without --force, got nil")
	}

	// With force should succeed
	w.Force = true
	if err := w.Write(model.TemplateMinimal, vars); err != nil {
		t.Errorf("Write() with force error: %v", err)
	}
}

func TestWriter_Write_CheckStackShIsExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod +x not applicable on Windows")
	}

	dir := t.TempDir()
	vars := model.TemplateVars{ProjectName: "proj", AgentConfigDir: ".claude"}

	w := &projectinit.Writer{TargetDir: dir, Force: false}
	if err := w.Write(model.TemplateFastAPIPython, vars); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	shPath := filepath.Join(dir, ".claude", "validators", "check_stack.sh")
	info, err := os.Stat(shPath)
	if err != nil {
		t.Fatalf("check_stack.sh not created: %v", err)
	}

	if info.Mode()&0111 == 0 {
		t.Error("check_stack.sh is not executable")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
