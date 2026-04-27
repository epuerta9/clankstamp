// Package replay defines the on-disk types that make up a clankstamp stamp.
//
// A stamp lives at .clankstamp/runs/<run_id>/ and consists of a manifest plus
// JSONL streams (events, tour steps, hunks, evidence, review state) and any
// captured patches/test output. See PRD.md §10 and §11 for the canonical
// layout; the JSON schemas under embedded/schemas/ are authoritative.
package replay

import "time"

// SchemaVersion is the on-disk schema version this binary writes.
const SchemaVersion = "0.1"

// Status enumerates the lifecycle states a stamp can be in. See PRD §16.
type Status string

const (
	StatusDraft              Status = "draft"
	StatusReady              Status = "ready"
	StatusReviewing          Status = "reviewing"
	StatusAccepted           Status = "accepted"
	StatusPartiallyAccepted  Status = "partially_accepted"
	StatusRejected           Status = "rejected"
	StatusStale              Status = "stale"
)

// Manifest is the top-level metadata file (manifest.json) for a stamp.
type Manifest struct {
	SchemaVersion string         `json:"schema_version"`
	RunID         string         `json:"run_id"`
	Title         string         `json:"title"`
	Status        Status         `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at,omitempty"`
	Repo          RepoInfo       `json:"repo"`
	Task          *TaskInfo      `json:"task,omitempty"`
	Harness       *HarnessInfo   `json:"harness,omitempty"`
	Artifacts     ArtifactPaths  `json:"artifacts"`
}

// RepoInfo describes the git context the stamp was captured against.
type RepoInfo struct {
	Root           string `json:"root"`
	Remote         string `json:"remote,omitempty"`
	BaseCommit     string `json:"base_commit"`
	HeadCommit     string `json:"head_commit,omitempty"`
	DirtyAtStart   bool   `json:"dirty_at_start"`
	DirtyAtFinish  bool   `json:"dirty_at_finish"`
}

// TaskInfo is optional task-tracker linkage.
type TaskInfo struct {
	Source string `json:"source,omitempty"`
	ID     string `json:"id,omitempty"`
	Title  string `json:"title,omitempty"`
}

// HarnessInfo identifies the agent harness (claude-code, pi, opencode, ...).
type HarnessInfo struct {
	Name      string `json:"name,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// ArtifactPaths records the relative filenames inside the stamp directory.
type ArtifactPaths struct {
	Events    string `json:"events"`
	Tour      string `json:"tour"`
	Hunks     string `json:"hunks,omitempty"`
	Evidence  string `json:"evidence,omitempty"`
	Review    string `json:"review,omitempty"`
	PatchFull string `json:"patch_full,omitempty"`
}

// Event is one JSONL record in events.jsonl.
type Event struct {
	Type      string    `json:"type"`
	Ts        time.Time `json:"ts"`
	RunID     string    `json:"run_id,omitempty"`
	Title     string    `json:"title,omitempty"`
	Path      string    `json:"path,omitempty"`
	Command   string    `json:"command,omitempty"`
	ExitCode  *int      `json:"exit_code,omitempty"`
	OutputRef string    `json:"output_ref,omitempty"`
	Source    string    `json:"source,omitempty"`
	ID        string    `json:"id,omitempty"`
	Status    string    `json:"status,omitempty"`
	Note      string    `json:"note,omitempty"`
}

// Hunk is one JSONL record in hunks.jsonl.
type Hunk struct {
	HunkID    string `json:"hunk_id"`
	StepID    string `json:"step_id,omitempty"`
	File      string `json:"file"`
	OldStart  int    `json:"old_start"`
	OldLines  int    `json:"old_lines"`
	NewStart  int    `json:"new_start"`
	NewLines  int    `json:"new_lines"`
	Symbol    string `json:"symbol,omitempty"`
	PatchRef  string `json:"patch_ref,omitempty"`
}

// TourStep is one JSONL record in tour.jsonl.
type TourStep struct {
	Type             string       `json:"type"`
	StepID           string       `json:"step_id"`
	Order            int          `json:"order"`
	Title            string       `json:"title"`
	Kind             string       `json:"kind,omitempty"`
	Summary          string       `json:"summary"`
	Why              string       `json:"why,omitempty"`
	Risk             string       `json:"risk,omitempty"`
	Files            []FileRef    `json:"files,omitempty"`
	PatchRef         string       `json:"patch_ref,omitempty"`
	ReviewQuestions  []string     `json:"review_questions,omitempty"`
	EvidenceRefs     []string     `json:"evidence_refs,omitempty"`
	// Connections links this step to other steps in the tour. The renderer
	// uses these to show "tests step 3", "uses step 2", etc., so the user
	// can see how the change hangs together as a whole rather than as a
	// pile of isolated edits.
	Connections []StepConnection `json:"connections,omitempty"`
}

// StepConnection points from one tour step to another, with a verb that
// describes the relationship. Common kinds: "tests", "uses", "extends",
// "supersedes", "depends-on".
type StepConnection struct {
	ToStepID string `json:"to_step_id"`
	Kind     string `json:"kind,omitempty"`
	Label    string `json:"label,omitempty"`
}

// FileRef points at a region of a file referenced by a tour step.
type FileRef struct {
	Path      string `json:"path"`
	LineStart int    `json:"line_start,omitempty"`
	LineEnd   int    `json:"line_end,omitempty"`
	Symbol    string `json:"symbol,omitempty"`
}
