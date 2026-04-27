// Package cli implements the clankstamp command-line dispatch.
//
// The CLI surface is intentionally small in v0: each subcommand is a stub that
// prints a "not implemented" notice. The contracts are pinned by PRD.md so
// implementations can be filled in incrementally without breaking the surface.
package cli

import (
	"fmt"
	"io"
	"os"
)

// Run dispatches the provided argv to a subcommand. It returns nil on success.
func Run(args []string, version string) error {
	if len(args) == 0 {
		printUsage(os.Stdout, version)
		return nil
	}

	cmd, rest := args[0], args[1:]

	switch cmd {
	case "-h", "--help", "help":
		printUsage(os.Stdout, version)
		return nil
	case "-v", "--version", "version":
		fmt.Fprintln(os.Stdout, version)
		return nil
	case "init":
		return runInit(rest, os.Stdout, os.Stderr)
	case "create":
		return runCreate(rest, os.Stdout, os.Stderr)
	case "finalize":
		return notImplemented("finalize", "close the active stamp and generate the tour")
	case "list":
		return notImplemented("list", "list stamps in .clankstamp/")
	case "show":
		return notImplemented("show", "print a stamp's manifest and tour")
	case "validate":
		return notImplemented("validate", "validate a stamp against the JSON schemas")
	case "doctor":
		return notImplemented("doctor", "diagnose the local clankstamp install")
	case "nvim":
		return dispatchNvim(rest)
	case "skill":
		return dispatchSkill(rest)
	case "git":
		return dispatchGit(rest)
	case "start", "event", "record":
		return notImplemented(cmd, "agent-facing live-capture command (see PRD §12.3)")
	default:
		return fmt.Errorf("unknown command %q (run `clankstamp help`)", cmd)
	}
}

func dispatchNvim(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("nvim subcommand required: install | uninstall | doctor")
	}
	switch args[0] {
	case "install":
		return runNvimInstall(args[1:], os.Stdout, os.Stderr)
	case "uninstall":
		return notImplemented("nvim uninstall", "remove the embedded Lua plugin")
	case "doctor":
		return notImplemented("nvim doctor", "diagnose the nvim plugin install")
	default:
		return fmt.Errorf("unknown nvim subcommand %q", args[0])
	}
}

func dispatchSkill(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("skill subcommand required: install | uninstall | doctor")
	}
	switch args[0] {
	case "install":
		return runSkillInstall(args[1:], os.Stdout, os.Stderr)
	case "uninstall":
		return notImplemented("skill uninstall", "remove the embedded SKILL.md")
	case "doctor":
		return notImplemented("skill doctor", "diagnose the skill install")
	default:
		return fmt.Errorf("unknown skill subcommand %q", args[0])
	}
}

func dispatchGit(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("git subcommand required: snapshot | patch | checkpoints")
	}
	return notImplemented("git "+args[0], "git integration helpers (see PRD §15)")
}

func notImplemented(name, hint string) error {
	return fmt.Errorf("%s: not implemented yet — %s", name, hint)
}

func printUsage(w io.Writer, version string) {
	fmt.Fprintf(w, `clankstamp %s — replay tours for AI-generated code changes

USAGE:
    clankstamp <command> [flags]

ARTIFACT COMMANDS:
    init                Scaffold .clankstamp/ in the current repo
    create              Create a stamp from a diff, commit, or worktree
    finalize            Close the active stamp and generate the tour
    list                List stamps in .clankstamp/
    show                Print a stamp's manifest and tour
    validate            Validate a stamp against the JSON schemas

INSTALL COMMANDS:
    nvim install        Install the embedded Neovim plugin
    skill install       Install the embedded agent skill (Claude Code or SKILL.md)

DIAGNOSTICS:
    doctor              Diagnose the local clankstamp install

See PRD.md for the full product spec.
`, version)
}
