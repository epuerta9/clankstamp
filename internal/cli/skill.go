package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	clankstamp "github.com/epuerta9/clankstamp"
	"github.com/epuerta9/clankstamp/internal/git"
)

// runSkillInstall implements `clankstamp skill install`.
//
// Destination matrix:
//
//   --target=claude-code --scope=global   ~/.claude/skills/clankstamp/
//   --target=claude-code --scope=project  <repo>/.claude/skills/clankstamp/
//   --target=standard    --scope=global   ~/.local/share/agent-skills/clankstamp/
//   --target=standard    --scope=project  <repo>/.agent-skills/clankstamp/
//
// `--path <dir>` overrides the destination entirely (the SKILL.md and any
// supporting files are written under <dir>). This is what tests use.
func runSkillInstall(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("skill install", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		target  = fs.String("target", "", "skill target: claude-code | standard")
		scope   = fs.String("scope", "", "scope: global | project")
		path    = fs.String("path", "", "override destination directory (skill files written under this path)")
		force   = fs.Bool("force", false, "overwrite an existing non-empty destination")
		dryRun  = fs.Bool("dry-run", false, "print what would be installed without writing")
		// Aliases: --claude / --standard / --global / --project.
		claude   = fs.Bool("claude", false, "alias for --target=claude-code")
		standard = fs.Bool("standard", false, "alias for --target=standard")
		global   = fs.Bool("global", false, "alias for --scope=global")
		project  = fs.Bool("project", false, "alias for --scope=project")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *claude {
		*target = "claude-code"
	}
	if *standard {
		*target = "standard"
	}
	if *global {
		*scope = "global"
	}
	if *project {
		*scope = "project"
	}

	if *target == "" {
		return errors.New("--target is required (claude-code | standard)")
	}
	tgt := clankstamp.SkillTarget(*target)
	src, err := clankstamp.SkillFS(tgt)
	if err != nil {
		return err
	}

	dest, err := resolveSkillDest(tgt, *scope, *path)
	if err != nil {
		return err
	}

	plan, err := buildInstallPlan(src, dest)
	if err != nil {
		return err
	}
	if *dryRun {
		printPlan(stdout, fmt.Sprintf("would install skill %s to:", tgt), plan)
		return nil
	}
	if plan.Replace && !*force {
		return fmt.Errorf("destination %s is non-empty; pass --force to overwrite", plan.Dest)
	}
	if err := applyInstallPlan(src, plan); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "installed skill %s -> %s (%d files)\n", tgt, plan.Dest, len(plan.Files))
	return nil
}

func resolveSkillDest(target clankstamp.SkillTarget, scope, override string) (string, error) {
	if override != "" {
		return override, nil
	}
	if scope == "" {
		return "", errors.New("--scope is required (global | project), or pass --path")
	}
	switch scope {
	case "global":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		switch target {
		case clankstamp.SkillTargetClaudeCode:
			return filepath.Join(home, ".claude", "skills", "clankstamp"), nil
		case clankstamp.SkillTargetStandard:
			if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
				return filepath.Join(dataHome, "agent-skills", "clankstamp"), nil
			}
			return filepath.Join(home, ".local", "share", "agent-skills", "clankstamp"), nil
		}
	case "project":
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		repo, err := git.Discover(cwd)
		if err != nil {
			return "", fmt.Errorf("project scope must run inside a git repo: %w", err)
		}
		switch target {
		case clankstamp.SkillTargetClaudeCode:
			return filepath.Join(repo.Root, ".claude", "skills", "clankstamp"), nil
		case clankstamp.SkillTargetStandard:
			return filepath.Join(repo.Root, ".agent-skills", "clankstamp"), nil
		}
	default:
		return "", fmt.Errorf("unknown scope %q (expected: global, project)", scope)
	}
	return "", fmt.Errorf("unknown target %q for scope %q", target, scope)
}
