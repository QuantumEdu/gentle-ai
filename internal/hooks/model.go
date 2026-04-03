// Package hooks implements project-level hook definitions loaded from
// .claude/rules.yaml (or equivalent agent config dir).
package hooks

// Event names that gentle-ai project hooks support.
type Event string

const (
	EventPreToolUse       Event = "PreToolUse"
	EventPostToolUse      Event = "PostToolUse"
	EventTaskCompleted    Event = "TaskCompleted"
	EventSessionStart     Event = "SessionStart"
	EventPermissionRequest Event = "PermissionRequest"
)

// BlockPolicy controls when a hook causes the tool call to be blocked.
type BlockPolicy string

const (
	// BlockOnExitCode2 blocks the tool call when the hook script exits with code 2.
	BlockOnExitCode2 BlockPolicy = "exit_code == 2"
	// BlockOnStdoutContains blocks when stdout contains a specific prefix.
	BlockOnStdoutContains BlockPolicy = "stdout_contains=BLOCKED"
)

// Hook defines a single hook entry: a script to run on a given event,
// optionally filtered by a matcher expression.
type Hook struct {
	// Command is the path to the script relative to the project root.
	Command string
	// Matcher is an optional filter expression (e.g. "tool_name=Bash").
	// Empty matcher means the hook runs for all events of this type.
	Matcher string
	// BlockOn defines when the hook should block the triggering action.
	// Empty means the hook is advisory only (output shown, action not blocked).
	BlockOn BlockPolicy
}

// RulesConfig is the top-level structure parsed from .claude/rules.yaml.
type RulesConfig struct {
	Version int                 `yaml:"version"`
	Hooks   map[Event][]Hook    `yaml:"hooks"`
}
