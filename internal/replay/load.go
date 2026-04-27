package replay

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// ListRuns returns the run ids present under .clankstamp/runs/, sorted
// lexicographically (which is also chronological because run ids start with a
// UTC timestamp).
func (s *Store) ListRuns() ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(s.Root, "runs"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			ids = append(ids, e.Name())
		}
	}
	sort.Strings(ids)
	return ids, nil
}

// LoadManifest reads <run>/manifest.json into a Manifest.
func (s *Store) LoadManifest(runID string) (*Manifest, error) {
	path := filepath.Join(s.RunDir(runID), "manifest.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &m, nil
}

// LoadEvents reads <run>/events.jsonl. A missing file is treated as an empty
// list rather than an error — events.jsonl is created lazily.
func (s *Store) LoadEvents(runID string) ([]Event, error) {
	return readJSONL[Event](filepath.Join(s.RunDir(runID), "events.jsonl"))
}

// LoadTour reads <run>/tour.jsonl. Missing file → empty list.
func (s *Store) LoadTour(runID string) ([]TourStep, error) {
	return readJSONL[TourStep](filepath.Join(s.RunDir(runID), "tour.jsonl"))
}

// LoadHunks reads <run>/hunks.jsonl. Missing file → empty list.
func (s *Store) LoadHunks(runID string) ([]Hunk, error) {
	return readJSONL[Hunk](filepath.Join(s.RunDir(runID), "hunks.jsonl"))
}

// readJSONL reads a JSONL file into a slice of T. A missing file returns
// (nil, nil); each malformed line returns an error annotated with the line
// number.
func readJSONL[T any](path string) ([]T, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []T
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	lineno := 0
	for sc.Scan() {
		lineno++
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var rec T
		if err := json.Unmarshal(line, &rec); err != nil {
			return out, fmt.Errorf("%s:%d: %w", path, lineno, err)
		}
		out = append(out, rec)
	}
	return out, sc.Err()
}
