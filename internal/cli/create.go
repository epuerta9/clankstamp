package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/epuerta9/clankstamp/internal/git"
	"github.com/epuerta9/clankstamp/internal/replay"
)

// runCreate handles `clankstamp create`.
//
// v0 only supports --from-worktree: capture the current working-tree diff
// against HEAD as a fresh stamp. --from-commit and --from-staged are accepted
// but currently return a friendly "not implemented" error pointing at the PRD.
func runCreate(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		title        = fs.String("title", "", "stamp title (required)")
		taskID       = fs.String("task", "", "task id (e.g. BD-42); used in the run id")
		taskSource   = fs.String("task-source", "", "task tracker name (beads, linear, github, ...)")
		harnessName  = fs.String("harness", os.Getenv("CLANKSTAMP_HARNESS"), "agent harness name (claude-code, pi, opencode, codex, ...)")
		fromWorktree = fs.Bool("from-worktree", false, "capture the working-tree diff vs HEAD")
		fromCommit   = fs.String("from-commit", "", "capture the diff of a single commit (NOT YET IMPLEMENTED)")
		fromStaged   = fs.Bool("from-staged", false, "capture the staged diff (NOT YET IMPLEMENTED)")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *fromCommit != "" || *fromStaged {
		return fmt.Errorf("--from-commit and --from-staged not implemented yet — use --from-worktree (PRD §12.2)")
	}
	if !*fromWorktree {
		return fmt.Errorf("create requires a source flag (e.g. --from-worktree)")
	}
	if *title == "" {
		return fmt.Errorf("--title is required")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	repo, err := git.Discover(cwd)
	if err != nil {
		return fmt.Errorf("create must run inside a git repo: %w", err)
	}
	store := replay.OpenStore(repo.Root)
	if err := store.EnsureInit(); err != nil {
		return err
	}

	now := time.Now().UTC()

	// Pick a run-id suffix: prefer the task id, fall back to the short HEAD sha.
	suffix := *taskID
	if suffix == "" {
		short, _ := repo.ShortSHA("HEAD")
		suffix = short
	}
	runID := replay.NewRunID(now, suffix)
	runDir := store.RunDir(runID)
	if err := os.MkdirAll(filepath.Join(runDir, "patches"), 0o755); err != nil {
		return err
	}

	headSHA, err := repo.HeadSHA()
	if err != nil {
		return err
	}
	dirty, err := repo.IsDirty()
	if err != nil {
		return err
	}
	remote, _ := repo.RemoteURL()

	diff, err := repo.WorktreeDiff()
	if err != nil {
		return fmt.Errorf("capture diff: %w", err)
	}
	patchPath := filepath.Join(runDir, "patches", "full.patch")
	if err := os.WriteFile(patchPath, []byte(diff), 0o644); err != nil {
		return err
	}

	// Parse hunks (best-effort; an empty diff yields zero hunks).
	hunks := replay.ParseHunks(diff, "patches/full.patch")
	hunksPath := filepath.Join(runDir, "hunks.jsonl")
	for _, h := range hunks {
		if err := replay.AppendJSONL(hunksPath, h); err != nil {
			return fmt.Errorf("write hunks.jsonl: %w", err)
		}
	}

	// Manifest.
	manifest := &replay.Manifest{
		SchemaVersion: replay.SchemaVersion,
		RunID:         runID,
		Title:         *title,
		Status:        replay.StatusReady,
		CreatedAt:     now,
		UpdatedAt:     now,
		Repo: replay.RepoInfo{
			Root:          repo.Root,
			Remote:        remote,
			BaseCommit:    headSHA,
			HeadCommit:    "", // worktree-based stamps have no committed head yet
			DirtyAtStart:  dirty,
			DirtyAtFinish: dirty,
		},
		Artifacts: replay.ArtifactPaths{
			Events:    "events.jsonl",
			Tour:      "tour.jsonl",
			Hunks:     "hunks.jsonl",
			Evidence:  "evidence.jsonl",
			Review:    "review.jsonl",
			PatchFull: "patches/full.patch",
		},
	}
	if *taskID != "" || *taskSource != "" {
		manifest.Task = &replay.TaskInfo{Source: *taskSource, ID: *taskID, Title: *title}
	}
	if *harnessName != "" {
		manifest.Harness = &replay.HarnessInfo{Name: *harnessName}
	}
	if err := replay.WriteManifest(runDir, manifest); err != nil {
		return err
	}

	// Bookend events: run.started + run.finished. No live capture in v0, so
	// these two records summarize the post-hoc creation.
	eventsPath := filepath.Join(runDir, "events.jsonl")
	if err := replay.AppendJSONL(eventsPath, replay.Event{
		Type: "run.started", Ts: now, RunID: runID, Title: *title,
	}); err != nil {
		return err
	}
	if err := replay.AppendJSONL(eventsPath, replay.Event{
		Type: "run.finished", Ts: now, RunID: runID, Status: string(replay.StatusReady),
	}); err != nil {
		return err
	}

	// Touch tour.jsonl so the tour-generation step has somewhere to append.
	if f, err := os.OpenFile(filepath.Join(runDir, "tour.jsonl"), os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
		_ = f.Close()
	}

	// Append to the global index for fast listing.
	indexPath := filepath.Join(store.Root, "index.jsonl")
	_ = replay.AppendJSONL(indexPath, map[string]any{
		"run_id":     runID,
		"title":      *title,
		"status":     string(replay.StatusReady),
		"created_at": now,
		"hunks":      len(hunks),
	})

	fmt.Fprintf(stdout, "created stamp %s\n", runID)
	fmt.Fprintf(stdout, "  path:   %s\n", runDir)
	fmt.Fprintf(stdout, "  hunks:  %d\n", len(hunks))
	fmt.Fprintf(stdout, "  status: %s\n", manifest.Status)
	if !dirty && len(hunks) == 0 {
		fmt.Fprintln(stdout, "  note:   working tree is clean — the stamp captures an empty diff")
	}
	return nil
}
