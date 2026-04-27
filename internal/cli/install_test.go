package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillInstall_Standard_PathOverride(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "clankstamp")
	var stdout, stderr bytes.Buffer
	err := runSkillInstall(
		[]string{"--target", "standard", "--path", dest},
		&stdout, &stderr,
	)
	if err != nil {
		t.Fatalf("runSkillInstall: %v\nstderr: %s", err, stderr.String())
	}

	skillBytes, err := os.ReadFile(filepath.Join(dest, "SKILL.md"))
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}
	skill := string(skillBytes)
	if !strings.Contains(skill, "name: clankstamp") {
		t.Errorf("SKILL.md missing `name: clankstamp` frontmatter:\n%s", snippet(skill, 200))
	}
	if !strings.Contains(skill, "clankstamp record file.write") {
		t.Errorf("SKILL.md missing recommended-flow snippet")
	}
	if !strings.Contains(stdout.String(), "installed skill standard") {
		t.Errorf("unexpected stdout: %s", stdout.String())
	}
}

func snippet(s string, n int) string {
	if len(s) < n {
		return s
	}
	return s[:n]
}

func TestSkillInstall_ClaudeCode_DryRun(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "clankstamp")
	var stdout, stderr bytes.Buffer
	err := runSkillInstall(
		[]string{"--claude", "--path", dest, "--dry-run"},
		&stdout, &stderr,
	)
	if err != nil {
		t.Fatalf("runSkillInstall: %v\nstderr: %s", err, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "would install skill claude-code") {
		t.Errorf("dry-run header missing: %s", out)
	}
	if !strings.Contains(out, "+ SKILL.md") {
		t.Errorf("dry-run did not list SKILL.md: %s", out)
	}
	// Crucially: nothing should be written.
	if _, err := os.Stat(dest); err == nil {
		t.Errorf("dry-run created %s", dest)
	}
}

func TestSkillInstall_ForceRequiredForNonEmptyDest(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "clankstamp")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "stale.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := runSkillInstall(
		[]string{"--target", "standard", "--path", dest},
		&stdout, &stderr,
	)
	if err == nil || !strings.Contains(err.Error(), "non-empty") {
		t.Fatalf("expected non-empty-dest error, got: %v", err)
	}

	// With --force it should succeed.
	stdout.Reset()
	stderr.Reset()
	if err := runSkillInstall(
		[]string{"--target", "standard", "--path", dest, "--force"},
		&stdout, &stderr,
	); err != nil {
		t.Fatalf("runSkillInstall --force: %v\nstderr: %s", err, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(dest, "SKILL.md")); err != nil {
		t.Errorf("SKILL.md not written after --force: %v", err)
	}
}

func TestSkillInstall_UnknownTarget(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runSkillInstall(
		[]string{"--target", "bogus", "--path", t.TempDir()},
		&stdout, &stderr,
	)
	if err == nil || !strings.Contains(err.Error(), "unknown skill target") {
		t.Fatalf("expected unknown-target error, got: %v", err)
	}
}

func TestNvimInstall_PathOverride(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "clankstamp.nvim")
	var stdout, stderr bytes.Buffer
	err := runNvimInstall(
		[]string{"--path", dest},
		&stdout, &stderr,
	)
	if err != nil {
		t.Fatalf("runNvimInstall: %v\nstderr: %s", err, stderr.String())
	}

	// PRD §9.2 requires `plugin/clankstamp.lua` and `lua/clankstamp/init.lua`.
	for _, rel := range []string{
		"plugin/clankstamp.lua",
		"lua/clankstamp/init.lua",
	} {
		b, err := os.ReadFile(filepath.Join(dest, rel))
		if err != nil {
			t.Errorf("missing %s: %v", rel, err)
			continue
		}
		if !strings.Contains(string(b), "clankstamp") {
			t.Errorf("%s does not mention clankstamp", rel)
		}
	}

	// The plugin entry should register at least the :Clankstamp command.
	plugin, _ := os.ReadFile(filepath.Join(dest, "plugin/clankstamp.lua"))
	if !strings.Contains(string(plugin), "nvim_create_user_command") {
		t.Errorf("plugin/clankstamp.lua missing user-command registration")
	}

	if !strings.Contains(stdout.String(), "installed Neovim plugin") {
		t.Errorf("unexpected stdout: %s", stdout.String())
	}
}

func TestNvimInstall_DryRun(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "clankstamp.nvim")
	var stdout, stderr bytes.Buffer
	if err := runNvimInstall([]string{"--path", dest, "--dry-run"}, &stdout, &stderr); err != nil {
		t.Fatalf("runNvimInstall: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "+ plugin/clankstamp.lua") || !strings.Contains(out, "+ lua/clankstamp/init.lua") {
		t.Errorf("dry-run output missing expected files:\n%s", out)
	}
	if _, err := os.Stat(dest); err == nil {
		t.Errorf("dry-run created %s", dest)
	}
}

