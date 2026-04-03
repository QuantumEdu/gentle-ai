package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Loader reads and parses the project rules.yaml file.
// It uses a minimal YAML parser to avoid adding a dependency —
// the rules.yaml format is simple enough to parse with JSON-compatible logic
// once converted, but for now we use a hand-rolled subset parser.
type Loader struct {
	// ConfigDir is the agent config directory (e.g. ".claude").
	ConfigDir string
}

// Load reads rules.yaml from ConfigDir and returns the parsed RulesConfig.
// Returns an empty config (no error) if the file does not exist.
func (l *Loader) Load(projectRoot string) (RulesConfig, error) {
	path := filepath.Join(projectRoot, l.ConfigDir, "rules.yaml")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return RulesConfig{}, nil
	}
	if err != nil {
		return RulesConfig{}, fmt.Errorf("read rules.yaml: %w", err)
	}

	cfg, err := parseRulesYAML(data)
	if err != nil {
		return RulesConfig{}, fmt.Errorf("parse rules.yaml: %w", err)
	}

	return cfg, nil
}

// HooksFor returns hooks registered for the given event.
func HooksFor(cfg RulesConfig, event Event) []Hook {
	return cfg.Hooks[event]
}

// MatchesHook reports whether the hook's matcher applies to the given context.
// An empty matcher matches everything.
func MatchesHook(hook Hook, toolName string) bool {
	if hook.Matcher == "" {
		return true
	}

	// Supported matcher format: "tool_name=X" or "tool_name=X||tool_name=Y"
	parts := strings.Split(hook.Matcher, "||")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if kv := strings.SplitN(part, "=", 2); len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])
			if key == "tool_name" && val == toolName {
				return true
			}
		}
	}

	return false
}

// parseRulesYAML is a minimal parser for the rules.yaml subset used by gentle-ai.
// It converts YAML to JSON first using a line-by-line approach, then unmarshals.
// This avoids adding a YAML library dependency for a simple fixed-schema file.
//
// Supported structure:
//
//	version: 1
//	hooks:
//	  PreToolUse:
//	    - matcher: "tool_name=Bash"
//	      command: "validators/check_stack.sh"
//	      block_on: exit_code == 2
func parseRulesYAML(data []byte) (RulesConfig, error) {
	// We do a simple line-by-line parse into a map structure, then marshal/unmarshal.
	// This is intentionally minimal — rules.yaml has a fixed, known schema.
	type rawHook struct {
		Command string `json:"command"`
		Matcher string `json:"matcher"`
		BlockOn string `json:"block_on"`
	}

	type rawConfig struct {
		Version int                       `json:"version"`
		Hooks   map[string][]rawHook      `json:"hooks"`
	}

	raw, err := yamlToJSON(data)
	if err != nil {
		return RulesConfig{}, err
	}

	var rc rawConfig
	if err := json.Unmarshal(raw, &rc); err != nil {
		return RulesConfig{}, fmt.Errorf("unmarshal rules: %w", err)
	}

	cfg := RulesConfig{
		Version: rc.Version,
		Hooks:   make(map[Event][]Hook),
	}

	for eventStr, rawHooks := range rc.Hooks {
		event := Event(eventStr)
		hooks := make([]Hook, 0, len(rawHooks))
		for _, rh := range rawHooks {
			hooks = append(hooks, Hook{
				Command: rh.Command,
				Matcher: rh.Matcher,
				BlockOn: BlockPolicy(rh.BlockOn),
			})
		}
		cfg.Hooks[event] = hooks
	}

	return cfg, nil
}

// yamlToJSON converts the rules.yaml subset to JSON for unmarshalling.
// This handles only the specific structure of rules.yaml — not general YAML.
func yamlToJSON(data []byte) ([]byte, error) {
	// Delegate to a simple recursive descent parser for the known schema.
	// For a production implementation this would use gopkg.in/yaml.v3,
	// but we keep zero new dependencies for now.
	//
	// The current implementation uses json.Marshal on a manually built map.
	lines := strings.Split(string(data), "\n")
	result, _, err := parseYAMLMap(lines, 0, 0)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// parseYAMLMap parses a YAML mapping block starting at the given line and indent.
func parseYAMLMap(lines []string, start, indent int) (map[string]any, int, error) {
	result := make(map[string]any)
	i := start

	for i < len(lines) {
		line := lines[i]
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			i++
			continue
		}

		currentIndent := countIndent(line)
		if currentIndent < indent {
			break
		}

		trimmed := strings.TrimSpace(line)
		colonIdx := strings.Index(trimmed, ":")
		if colonIdx < 0 {
			i++
			continue
		}

		key := strings.TrimSpace(trimmed[:colonIdx])
		rest := strings.TrimSpace(trimmed[colonIdx+1:])

		if rest != "" {
			// Scalar value
			result[key] = unquote(rest)
			i++
		} else {
			// Nested structure: look ahead
			i++
			if i >= len(lines) {
				break
			}
			nextLine := lines[i]
			nextTrimmed := strings.TrimSpace(nextLine)
			nextIndent := countIndent(nextLine)

			if strings.HasPrefix(nextTrimmed, "-") {
				// Array of mappings
				arr, newI, err := parseYAMLArray(lines, i, nextIndent)
				if err != nil {
					return nil, i, err
				}
				result[key] = arr
				i = newI
			} else if nextIndent > indent {
				// Nested map
				nested, newI, err := parseYAMLMap(lines, i, nextIndent)
				if err != nil {
					return nil, i, err
				}
				result[key] = nested
				i = newI
			}
		}
	}

	return result, i, nil
}

// parseYAMLArray parses a YAML sequence block.
func parseYAMLArray(lines []string, start, indent int) ([]map[string]any, int, error) {
	var result []map[string]any
	i := start

	for i < len(lines) {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}

		currentIndent := countIndent(line)
		if currentIndent < indent {
			break
		}

		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "-") {
			break
		}

		// Start of a new list item
		item := make(map[string]any)
		rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
		if rest != "" {
			// Inline key: value on the same line as -
			if colonIdx := strings.Index(rest, ":"); colonIdx >= 0 {
				k := strings.TrimSpace(rest[:colonIdx])
				v := strings.TrimSpace(rest[colonIdx+1:])
				item[k] = unquote(v)
			}
		}
		i++

		// Collect additional key:value lines at deeper indent
		itemIndent := indent + 2
		for i < len(lines) {
			nextLine := lines[i]
			if strings.TrimSpace(nextLine) == "" {
				i++
				continue
			}
			nextIndent := countIndent(nextLine)
			if nextIndent < itemIndent {
				break
			}
			nextTrimmed := strings.TrimSpace(nextLine)
			if strings.HasPrefix(nextTrimmed, "-") {
				break
			}
			if colonIdx := strings.Index(nextTrimmed, ":"); colonIdx >= 0 {
				k := strings.TrimSpace(nextTrimmed[:colonIdx])
				v := strings.TrimSpace(nextTrimmed[colonIdx+1:])
				item[k] = unquote(v)
			}
			i++
		}

		result = append(result, item)
	}

	return result, i, nil
}

func countIndent(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else {
			break
		}
	}
	return count
}

// unquote strips surrounding quotes and returns the Go value.
// Integers are returned as int, everything else as string.
func unquote(s string) any {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	// Detect integer literals
	if isInteger(s) {
		n := 0
		for _, ch := range s {
			n = n*10 + int(ch-'0')
		}
		return n
	}
	return s
}

func isInteger(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
