package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/epuerta9/clankstamp/internal/replay"
)

// setupRepoWithStamp creates a fresh git repo, runs `init` + `create
// --from-worktree`, and returns the run id of the created stamp. Tests for
// list/show/validate share this fixture so they exercise the same end-to-end
// shape that real users see.
func setupRepoWithStamp(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	repoDir := t.TempDir()
	gitInit(t, repoDir)
	writeFile(t, repoDir, "app.go", "package main\n\nfunc main() {}\n")
	gitCmd(t, repoDir, "add", ".")
	gitCmd(t, repoDir, "commit", "-m", "initial")
	writeFile(t, repoDir, "app.go", "package main\n\nfunc main() { println(\"hi\") }\n")
	t.Chdir(repoDir)

	var stdout, stderr bytes.Buffer
	if err := runInit(nil, &stdout, &stderr); err != nil {
		t.Fatalf("runInit: %v\nstderr: %s", err, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if err := runCreate(
		[]string{"--from-worktree", "--title", "say hi", "--task", "BD-1", "--harness", "claude-code"},
		&stdout, &stderr,
	); err != nil {
		t.Fatalf("runCreate: %v\nstderr: %s", err, stderr.String())
	}

	store := replay.OpenStore(repoDir)
	ids, err := store.ListRuns()
	if err != nil || len(ids) != 1 {
		t.Fatalf("expected 1 run, got %v err=%v", ids, err)
	}
	return ids[0]
}

func TestList_Table(t *testing.T) {
	runID := setupRepoWithStamp(t)
	var stdout, stderr bytes.Buffer
	if err := runList(nil, &stdout, &stderr); err != nil {
		t.Fatalf("runList: %v\nstderr: %s", err, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"RUN_ID", "STATUS", "HUNKS", "STEPS", "TITLE", runID, "say hi", "ready"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q:\n%s", want, out)
		}
	}
}

func TestList_JSON(t *testing.T) {
	runID := setupRepoWithStamp(t)
	var stdout, stderr bytes.Buffer
	if err := runList([]string{"--json"}, &stdout, &stderr); err != nil {
		t.Fatalf("runList --json: %v\nstderr: %s", err, stderr.String())
	}
	var entries []listEntry
	if err := json.Unmarshal(stdout.Bytes(), &entries); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, stdout.String())
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	got := entries[0]
	if got.RunID != runID || got.Title != "say hi" || got.Status != "ready" {
		t.Errorf("entry mismatch: %+v", got)
	}
	if got.Hunks != 1 {
		t.Errorf("expected 1 hunk, got %d", got.Hunks)
	}
}

func TestList_EmptyStoreEmitsArrayJSON(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	repoDir := t.TempDir()
	gitInit(t, repoDir)
	t.Chdir(repoDir)

	var stdout, stderr bytes.Buffer
	if err := runList([]string{"--json"}, &stdout, &stderr); err != nil {
		t.Fatalf("runList: %v", err)
	}
	if strings.TrimSpace(stdout.String()) != "[]" {
		t.Errorf("expected []; got: %q", stdout.String())
	}
}

func TestShow_Text(t *testing.T) {
	runID := setupRepoWithStamp(t)
	var stdout, stderr bytes.Buffer
	if err := runShow([]string{runID}, &stdout, &stderr); err != nil {
		t.Fatalf("runShow: %v\nstderr: %s", err, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"Title:        say hi",
		"Status:       ready",
		"Task:         BD-1",
		"Harness:      claude-code",
		"Hunks:        1",
		"Tour steps:   0",
		"(no tour generated yet",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("show missing %q:\n%s", want, out)
		}
	}
}

func TestShow_JSON(t *testing.T) {
	runID := setupRepoWithStamp(t)
	var stdout, stderr bytes.Buffer
	if err := runShow([]string{"--json", runID}, &stdout, &stderr); err != nil {
		t.Fatalf("runShow --json: %v\nstderr: %s", err, stderr.String())
	}
	var p showPayload
	if err := json.Unmarshal(stdout.Bytes(), &p); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, stdout.String())
	}
	if p.Manifest == nil || p.Manifest.RunID != runID {
		t.Errorf("manifest mismatch: %+v", p.Manifest)
	}
	if len(p.Hunks) != 1 {
		t.Errorf("expected 1 hunk, got %d", len(p.Hunks))
	}
	if len(p.Tour) != 0 {
		t.Errorf("expected empty tour, got %d", len(p.Tour))
	}
	if len(p.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(p.Events))
	}
}

func TestShow_UnknownRunIDFails(t *testing.T) {
	_ = setupRepoWithStamp(t)
	var stdout, stderr bytes.Buffer
	err := runShow([]string{"run_nonexistent"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for nonexistent run id")
	}
}

func TestValidate_OK(t *testing.T) {
	_ = setupRepoWithStamp(t)
	var stdout, stderr bytes.Buffer
	if err := runValidate(nil, &stdout, &stderr); err != nil {
		t.Fatalf("runValidate: %v\nstderr: %s\nstdout: %s", err, stderr.String(), stdout.String())
	}
	if !strings.Contains(stdout.String(), "ok  ") {
		t.Errorf("expected `ok` in output:\n%s", stdout.String())
	}
}

func TestValidate_DetectsCorruptManifest(t *testing.T) {
	runID := setupRepoWithStamp(t)
	store, _ := openProjectStore()
	manifestPath := filepath.Join(store.RunDir(runID), "manifest.json")

	// Corrupt the manifest: zero out title and status.
	if err := os.WriteFile(manifestPath, []byte(`{"schema_version":"0.1","run_id":"`+runID+`"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	err := runValidate(nil, &stdout, &stderr)
	if err == nil {
		t.Fatalf("expected validation to fail; stdout: %s", stdout.String())
	}
	out := stdout.String()
	for _, want := range []string{"FAIL", "title is empty", "status"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestValidate_JSON(t *testing.T) {
	runID := setupRepoWithStamp(t)
	var stdout, stderr bytes.Buffer
	if err := runValidate([]string{"--json", runID}, &stdout, &stderr); err != nil {
		t.Fatalf("runValidate: %v\nstderr: %s", err, stderr.String())
	}
	var reports []validateReport
	if err := json.Unmarshal(stdout.Bytes(), &reports); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, stdout.String())
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if !reports[0].OK {
		t.Errorf("expected ok=true; got %+v", reports[0])
	}
	if reports[0].RunID != runID {
		t.Errorf("run_id mismatch: got %q want %q", reports[0].RunID, runID)
	}
}
