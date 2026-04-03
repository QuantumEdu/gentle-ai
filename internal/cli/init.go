package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/gentleman-programming/gentle-ai/internal/catalog"
	"github.com/gentleman-programming/gentle-ai/internal/model"
	"github.com/gentleman-programming/gentle-ai/internal/projectinit"
)

// InitFlags holds parsed flags for `gentle-ai init`.
type InitFlags struct {
	Template string
	Agent    string
	Project  string
	Force    bool
	DryRun   bool
	List     bool
}

// RunInit executes the `gentle-ai init` command.
func RunInit(args []string, stdout io.Writer) error {
	flags, err := parseInitFlags(args)
	if err != nil {
		return err
	}

	if flags.List {
		printTemplateList(stdout)
		return nil
	}

	if flags.Template == "" {
		return fmt.Errorf("template required — use --template <name> or --list to see options\n\nExample:\n  gentle-ai init --template fastapi-python")
	}

	opts := projectinit.Options{
		TemplateID:  model.TemplateID(flags.Template),
		ProjectName: flags.Project,
		AgentID:     model.AgentID(flags.Agent),
		Force:       flags.Force,
		DryRun:      flags.DryRun,
	}

	result, err := projectinit.Run(opts)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprint(stdout, RenderInitResult(result))
	return nil
}

// RenderInitResult formats the init result for CLI output.
func RenderInitResult(result projectinit.Result) string {
	var sb strings.Builder

	if result.DryRun {
		sb.WriteString(fmt.Sprintf("DRY RUN — template: %s\n\n", result.TemplateID))
		sb.WriteString(fmt.Sprintf("Would create in %s:\n", result.TargetDir))
	} else {
		sb.WriteString(fmt.Sprintf("✓ Project initialized with template: %s\n\n", result.TemplateID))
		sb.WriteString(fmt.Sprintf("Created in %s:\n", result.TargetDir))
	}

	for _, f := range result.Files {
		sb.WriteString(fmt.Sprintf("  %s\n", f))
	}

	if !result.DryRun {
		sb.WriteString("\nNext steps:\n")
		sb.WriteString("  1. Review .claude/stack_config.json — adjust to your project\n")
		sb.WriteString("  2. Edit .claude/MANDATORY_CHECKS.md — add project-specific rules\n")
		sb.WriteString("  3. Run /sdd-init in your AI agent to register project context\n")
	}

	return sb.String()
}

// printTemplateList prints available templates to stdout.
func printTemplateList(stdout io.Writer) {
	templates := catalog.AllTemplates()
	fmt.Fprintf(stdout, "Available templates:\n\n")
	for _, t := range templates {
		fmt.Fprintf(stdout, "  %-20s %s\n    Stack: %s\n\n", t.ID, t.Description, t.Stack)
	}
}

// parseInitFlags parses args for the init subcommand.
func parseInitFlags(args []string) (InitFlags, error) {
	var flags InitFlags

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--template", "-t":
			if i+1 >= len(args) {
				return flags, fmt.Errorf("--template requires a value")
			}
			i++
			flags.Template = args[i]
		case "--agent", "-a":
			if i+1 >= len(args) {
				return flags, fmt.Errorf("--agent requires a value")
			}
			i++
			flags.Agent = args[i]
		case "--project", "-p":
			if i+1 >= len(args) {
				return flags, fmt.Errorf("--project requires a value")
			}
			i++
			flags.Project = args[i]
		case "--force", "-f":
			flags.Force = true
		case "--dry-run", "-n":
			flags.DryRun = true
		case "--list", "-l":
			flags.List = true
		default:
			// Positional: first non-flag arg is the template name
			if !strings.HasPrefix(args[i], "-") && flags.Template == "" {
				flags.Template = args[i]
			}
		}
	}

	return flags, nil
}
