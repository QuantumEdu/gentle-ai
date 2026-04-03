package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/cli"
)

func TestRunInit_ListTemplates(t *testing.T) {
	var buf bytes.Buffer
	err := cli.RunInit([]string{"--list"}, &buf)
	if err != nil {
		t.Fatalf("RunInit --list error: %v", err)
	}

	output := buf.String()
	expectedTemplates := []string{"fastapi-python", "react-nextjs", "go-backend", "go-tui", "minimal"}
	for _, tmpl := range expectedTemplates {
		if !strings.Contains(output, tmpl) {
			t.Errorf("expected template %q in list output, not found", tmpl)
		}
	}
}

func TestRunInit_MissingTemplate(t *testing.T) {
	var buf bytes.Buffer
	err := cli.RunInit([]string{}, &buf)
	if err == nil {
		t.Error("expected error when no template specified, got nil")
	}
	if !strings.Contains(err.Error(), "template required") {
		t.Errorf("expected 'template required' error, got: %v", err)
	}
}

func TestRunInit_UnknownTemplate(t *testing.T) {
	var buf bytes.Buffer
	err := cli.RunInit([]string{"--template", "does-not-exist"}, &buf)
	if err == nil {
		t.Error("expected error for unknown template, got nil")
	}
}

func TestRunInit_DryRun(t *testing.T) {
	var buf bytes.Buffer
	err := cli.RunInit([]string{"--template", "minimal", "--dry-run"}, &buf)
	if err != nil {
		t.Fatalf("RunInit dry-run error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", output)
	}
	if !strings.Contains(output, "stack_config.json") {
		t.Errorf("expected stack_config.json in dry-run output, got: %s", output)
	}
}

func TestRunInit_PositionalTemplate(t *testing.T) {
	var buf bytes.Buffer
	// Positional arg (no --template flag) should work
	err := cli.RunInit([]string{"minimal", "--dry-run"}, &buf)
	if err != nil {
		t.Fatalf("RunInit positional template error: %v", err)
	}
}
