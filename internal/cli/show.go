package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/epuerta9/clankstamp/internal/replay"
)

// showPayload is the JSON shape emitted by `clankstamp show --json`. The nvim
// plugin uses this to render the tour panel — keeping it stable matters.
type showPayload struct {
	Manifest *replay.Manifest  `json:"manifest"`
	Tour     []replay.TourStep `json:"tour"`
	Hunks    []replay.Hunk     `json:"hunks"`
	Events   []replay.Event    `json:"events,omitempty"`
}

// runShow implements `clankstamp show <run_id>`.
func runShow(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	fs.SetOutput(stderr)
	asJSON := fs.Bool("json", false, "emit JSON (consumed by the nvim tour panel)")
	withEvents := fs.Bool("events", false, "include events.jsonl in the output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: clankstamp show <run_id>")
	}
	runID := fs.Arg(0)

	store, err := openProjectStore()
	if err != nil {
		return err
	}
	m, err := store.LoadManifest(runID)
	if err != nil {
		return fmt.Errorf("load manifest for %s: %w", runID, err)
	}
	tour, err := store.LoadTour(runID)
	if err != nil {
		return err
	}
	hunks, err := store.LoadHunks(runID)
	if err != nil {
		return err
	}
	var events []replay.Event
	if *withEvents || *asJSON {
		events, _ = store.LoadEvents(runID)
	}

	if *asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(showPayload{Manifest: m, Tour: tour, Hunks: hunks, Events: events})
	}

	renderShow(stdout, m, tour, hunks)
	if *withEvents {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Events:")
		for _, e := range events {
			fmt.Fprintf(stdout, "  %s  %-14s  %s%s\n",
				e.Ts.UTC().Format("15:04:05"), e.Type, e.Path, e.Command)
		}
	}
	return nil
}

func renderShow(w io.Writer, m *replay.Manifest, tour []replay.TourStep, hunks []replay.Hunk) {
	fmt.Fprintf(w, "Title:        %s\n", m.Title)
	fmt.Fprintf(w, "Run ID:       %s\n", m.RunID)
	fmt.Fprintf(w, "Status:       %s\n", m.Status)
	fmt.Fprintf(w, "Created:      %s\n", m.CreatedAt.UTC().Format("2006-01-02 15:04:05 UTC"))
	fmt.Fprintf(w, "Repo:         %s\n", m.Repo.Root)
	if m.Repo.BaseCommit != "" {
		short := m.Repo.BaseCommit
		if len(short) > 7 {
			short = short[:7]
		}
		fmt.Fprintf(w, "Base commit:  %s\n", short)
	}
	if m.Task != nil && (m.Task.ID != "" || m.Task.Source != "") {
		fmt.Fprintf(w, "Task:         %s", m.Task.ID)
		if m.Task.Source != "" {
			fmt.Fprintf(w, " (%s)", m.Task.Source)
		}
		fmt.Fprintln(w)
	}
	if m.Harness != nil && m.Harness.Name != "" {
		fmt.Fprintf(w, "Harness:      %s\n", m.Harness.Name)
	}
	fmt.Fprintf(w, "Hunks:        %d\n", len(hunks))
	fmt.Fprintf(w, "Tour steps:   %d\n", len(tour))

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Tour:")
	if len(tour) == 0 {
		fmt.Fprintln(w, "  (no tour generated yet — agent skill writes tour.jsonl)")
		return
	}
	for _, s := range tour {
		fmt.Fprintf(w, "  %2d. %s", s.Order, s.Title)
		if s.Risk != "" {
			fmt.Fprintf(w, "  [risk:%s]", s.Risk)
		}
		fmt.Fprintln(w)
		if s.Summary != "" {
			fmt.Fprintf(w, "      %s\n", s.Summary)
		}
	}
}
