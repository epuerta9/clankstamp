package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/epuerta9/clankstamp/internal/replay"
)

// issue is one validation finding for a stamp. Keeping severity + a stable
// path/message shape means tooling (and the future `clankstamp doctor`) can
// post-process the JSON output.
type issue struct {
	RunID    string `json:"run_id"`
	Severity string `json:"severity"` // "error" | "warning"
	Path     string `json:"path"`     // e.g. "manifest.json", "events.jsonl:7"
	Message  string `json:"message"`
}

// validateReport is the per-run roll-up emitted as JSON.
type validateReport struct {
	RunID    string  `json:"run_id"`
	OK       bool    `json:"ok"`
	Issues   []issue `json:"issues"`
	Errors   int     `json:"errors"`
	Warnings int     `json:"warnings"`
}

// runValidate implements `clankstamp validate [<run_id>...]`.
//
// We validate against our Go types (which match the JSON schemas under
// embedded/schemas/). A future improvement is to plug in a real JSON Schema
// validator and consume the embedded schemas directly.
func runValidate(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	strict := fs.Bool("strict", false, "treat warnings as errors")
	asJSON := fs.Bool("json", false, "emit JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	store, err := openProjectStore()
	if err != nil {
		return err
	}
	if !store.Exists() {
		return errors.New("no .clankstamp/ store in this repo (run `clankstamp init`)")
	}

	var ids []string
	if fs.NArg() > 0 {
		ids = fs.Args()
	} else {
		ids, err = store.ListRuns()
		if err != nil {
			return err
		}
	}
	if len(ids) == 0 {
		fmt.Fprintln(stdout, "no stamps to validate")
		return nil
	}

	reports := make([]validateReport, 0, len(ids))
	totalErrors := 0
	totalWarnings := 0
	for _, id := range ids {
		r := validateRun(store, id)
		reports = append(reports, r)
		totalErrors += r.Errors
		totalWarnings += r.Warnings
	}

	if *asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(reports); err != nil {
			return err
		}
	} else {
		renderValidate(stdout, reports)
	}

	if totalErrors > 0 || (*strict && totalWarnings > 0) {
		return fmt.Errorf("%d error(s), %d warning(s)", totalErrors, totalWarnings)
	}
	return nil
}

func validateRun(store *replay.Store, runID string) validateReport {
	r := validateRun_inner(store, runID)
	for _, iss := range r.Issues {
		switch iss.Severity {
		case "error":
			r.Errors++
		case "warning":
			r.Warnings++
		}
	}
	r.OK = r.Errors == 0
	return r
}

// validateRun_inner does the actual checks; the outer wrapper computes counts.
func validateRun_inner(store *replay.Store, runID string) validateReport {
	r := validateReport{RunID: runID}
	add := func(sev, path, msg string) {
		r.Issues = append(r.Issues, issue{RunID: runID, Severity: sev, Path: path, Message: msg})
	}

	runDir := store.RunDir(runID)
	if st, err := os.Stat(runDir); err != nil || !st.IsDir() {
		add("error", "", fmt.Sprintf("run directory %s missing", runDir))
		return r
	}

	// --- manifest.json ---
	m, err := store.LoadManifest(runID)
	if err != nil {
		add("error", "manifest.json", err.Error())
		return r
	}
	if m.SchemaVersion == "" {
		add("error", "manifest.json", "schema_version is empty")
	} else if m.SchemaVersion != replay.SchemaVersion {
		add("warning", "manifest.json",
			fmt.Sprintf("schema_version %q does not match this binary's %q", m.SchemaVersion, replay.SchemaVersion))
	}
	if m.RunID == "" {
		add("error", "manifest.json", "run_id is empty")
	} else if m.RunID != runID {
		add("error", "manifest.json", fmt.Sprintf("run_id %q does not match directory name %q", m.RunID, runID))
	}
	if m.Title == "" {
		add("error", "manifest.json", "title is empty")
	}
	if !validStatus(m.Status) {
		add("error", "manifest.json", fmt.Sprintf("status %q is not in the enum", m.Status))
	}
	if m.CreatedAt.IsZero() {
		add("error", "manifest.json", "created_at is zero")
	}
	if m.Repo.Root == "" {
		add("error", "manifest.json", "repo.root is empty")
	}
	if m.Repo.BaseCommit == "" {
		add("warning", "manifest.json", "repo.base_commit is empty (replay durability is reduced)")
	}
	if m.Artifacts.Events == "" {
		add("error", "manifest.json", "artifacts.events is empty")
	}
	if m.Artifacts.Tour == "" {
		add("error", "manifest.json", "artifacts.tour is empty")
	}
	if m.Artifacts.PatchFull != "" {
		if _, err := os.Stat(filepath.Join(runDir, m.Artifacts.PatchFull)); err != nil {
			add("error", m.Artifacts.PatchFull, "referenced patch file is missing")
		}
	}

	// --- events.jsonl ---
	events, err := store.LoadEvents(runID)
	if err != nil {
		add("error", "events.jsonl", err.Error())
	} else {
		for i, e := range events {
			line := fmt.Sprintf("events.jsonl:%d", i+1)
			if !validEventType(e.Type) {
				add("error", line, fmt.Sprintf("unknown event type %q", e.Type))
			}
			if e.Ts.IsZero() {
				add("error", line, "ts is zero")
			}
		}
	}

	// --- tour.jsonl ---
	tour, err := store.LoadTour(runID)
	if err != nil {
		add("error", "tour.jsonl", err.Error())
	} else {
		for i, s := range tour {
			line := fmt.Sprintf("tour.jsonl:%d", i+1)
			if s.Type != "tour.step" {
				add("error", line, fmt.Sprintf(`type must be "tour.step", got %q`, s.Type))
			}
			if s.StepID == "" {
				add("error", line, "step_id is empty")
			}
			if s.Order <= 0 {
				add("error", line, "order must be > 0")
			}
			if s.Title == "" {
				add("error", line, "title is empty")
			}
			if s.Summary == "" {
				add("warning", line, "summary is empty")
			}
		}
	}

	// --- hunks.jsonl ---
	hunks, err := store.LoadHunks(runID)
	if err != nil {
		add("error", "hunks.jsonl", err.Error())
	} else {
		for i, h := range hunks {
			line := fmt.Sprintf("hunks.jsonl:%d", i+1)
			if h.HunkID == "" {
				add("error", line, "hunk_id is empty")
			}
			if h.File == "" {
				add("error", line, "file is empty")
			}
			if h.OldStart < 0 || h.NewStart < 0 || h.OldLines < 0 || h.NewLines < 0 {
				add("error", line, "negative range value")
			}
		}
	}

	return r
}

func validStatus(s replay.Status) bool {
	switch s {
	case replay.StatusDraft, replay.StatusReady, replay.StatusReviewing,
		replay.StatusAccepted, replay.StatusPartiallyAccepted, replay.StatusRejected, replay.StatusStale:
		return true
	}
	return false
}

var knownEventTypes = map[string]bool{
	"run.started":  true,
	"run.finished": true,
	"plan.loaded":  true,
	"file.read":    true,
	"file.write":   true,
	"file.delete":  true,
	"command.run":  true,
	"test.run":     true,
	"checkpoint":   true,
	"note":         true,
}

func validEventType(t string) bool { return knownEventTypes[t] }

func renderValidate(w io.Writer, reports []validateReport) {
	for _, r := range reports {
		status := "ok"
		if !r.OK {
			status = "FAIL"
		}
		fmt.Fprintf(w, "%s  %s  (%d errors, %d warnings)\n", status, r.RunID, r.Errors, r.Warnings)
		for _, iss := range r.Issues {
			path := iss.Path
			if path == "" {
				path = "-"
			}
			fmt.Fprintf(w, "  %-7s %s: %s\n", iss.Severity, path, iss.Message)
		}
	}
}
