package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/epuerta9/clankstamp/internal/replay"
)

// listEntry is the per-stamp record emitted by `clankstamp list --json`. The
// nvim plugin will consume this format to populate the picker.
type listEntry struct {
	RunID     string `json:"run_id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	Hunks     int    `json:"hunks"`
	TourSteps int    `json:"tour_steps"`
}

// runList implements `clankstamp list`.
func runList(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	asJSON := fs.Bool("json", false, "emit JSON instead of a table (consumed by the nvim picker)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	store, err := openProjectStore()
	if err != nil {
		return err
	}
	if !store.Exists() {
		if *asJSON {
			fmt.Fprintln(stdout, "[]")
			return nil
		}
		fmt.Fprintln(stdout, "no stamps yet — run `clankstamp init` then `clankstamp create`")
		return nil
	}

	ids, err := store.ListRuns()
	if err != nil {
		return err
	}

	entries, err := loadListEntries(store, ids)
	if err != nil {
		return err
	}

	if *asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		// Always emit an array (never `null`) so the nvim plugin can json.decode safely.
		if entries == nil {
			entries = []listEntry{}
		}
		return enc.Encode(entries)
	}

	if len(entries) == 0 {
		fmt.Fprintln(stdout, "no stamps yet — run `clankstamp create --from-worktree --title \"...\"`")
		return nil
	}

	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "RUN_ID\tSTATUS\tHUNKS\tSTEPS\tTITLE")
	for _, e := range entries {
		fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%s\n", e.RunID, e.Status, e.Hunks, e.TourSteps, e.Title)
	}
	return tw.Flush()
}

func loadListEntries(store *replay.Store, ids []string) ([]listEntry, error) {
	var out []listEntry
	for _, id := range ids {
		m, err := store.LoadManifest(id)
		if err != nil {
			// A run dir without a parseable manifest is suspicious but not
			// fatal for `list`; surface it inline so the user can still see
			// the other stamps.
			out = append(out, listEntry{RunID: id, Title: "(manifest unreadable: " + err.Error() + ")"})
			continue
		}
		hunks, _ := store.LoadHunks(id)
		tour, _ := store.LoadTour(id)
		out = append(out, listEntry{
			RunID:     m.RunID,
			Title:     m.Title,
			Status:    string(m.Status),
			CreatedAt: m.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			Hunks:     len(hunks),
			TourSteps: len(tour),
		})
	}
	return out, nil
}
