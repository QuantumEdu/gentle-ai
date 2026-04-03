package hooks_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/hooks"
)

const sampleRulesYAML = `
version: 1

hooks:
  PreToolUse:
    - matcher: "tool_name=Bash"
      command: "validators/check_stack.sh"
      block_on: "exit_code == 2"
    - matcher: "tool_name=Write"
      command: "validators/check_forbidden.sh"
      block_on: "exit_code == 2"
  SessionStart:
    - command: "scripts/load_context.sh"
`

func TestLoader_Load_ParsesHooks(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "rules.yaml"), []byte(sampleRulesYAML), 0644); err != nil {
		t.Fatal(err)
	}

	loader := hooks.Loader{ConfigDir: ".claude"}
	cfg, err := loader.Load(dir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}

	preHooks := hooks.HooksFor(cfg, hooks.EventPreToolUse)
	if len(preHooks) != 2 {
		t.Fatalf("expected 2 PreToolUse hooks, got %d", len(preHooks))
	}
	if preHooks[0].Command != "validators/check_stack.sh" {
		t.Errorf("unexpected command: %s", preHooks[0].Command)
	}
	if preHooks[0].BlockOn != hooks.BlockOnExitCode2 {
		t.Errorf("unexpected block_on: %s", preHooks[0].BlockOn)
	}

	sessionHooks := hooks.HooksFor(cfg, hooks.EventSessionStart)
	if len(sessionHooks) != 1 {
		t.Fatalf("expected 1 SessionStart hook, got %d", len(sessionHooks))
	}
}

func TestLoader_Load_MissingFile_ReturnsEmpty(t *testing.T) {
	dir := t.TempDir()

	loader := hooks.Loader{ConfigDir: ".claude"}
	cfg, err := loader.Load(dir)
	if err != nil {
		t.Fatalf("Load() with missing file should not error, got: %v", err)
	}
	if len(cfg.Hooks) != 0 {
		t.Errorf("expected empty hooks, got %d", len(cfg.Hooks))
	}
}

func TestMatchesHook_EmptyMatcher_MatchesAll(t *testing.T) {
	h := hooks.Hook{Command: "script.sh", Matcher: ""}
	if !hooks.MatchesHook(h, "Bash") {
		t.Error("empty matcher should match any tool name")
	}
	if !hooks.MatchesHook(h, "Write") {
		t.Error("empty matcher should match any tool name")
	}
}

func TestMatchesHook_SpecificTool_MatchesOnly(t *testing.T) {
	h := hooks.Hook{Command: "script.sh", Matcher: "tool_name=Bash"}
	if !hooks.MatchesHook(h, "Bash") {
		t.Error("matcher tool_name=Bash should match Bash")
	}
	if hooks.MatchesHook(h, "Write") {
		t.Error("matcher tool_name=Bash should not match Write")
	}
}

func TestMatchesHook_MultipleTools_MatchesAny(t *testing.T) {
	h := hooks.Hook{Command: "script.sh", Matcher: "tool_name=Bash||tool_name=Write"}
	if !hooks.MatchesHook(h, "Bash") {
		t.Error("should match Bash")
	}
	if !hooks.MatchesHook(h, "Write") {
		t.Error("should match Write")
	}
	if hooks.MatchesHook(h, "Read") {
		t.Error("should not match Read")
	}
}
