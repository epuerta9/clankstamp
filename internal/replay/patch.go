package replay

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// ParseHunks extracts a flat list of Hunk records from a unified diff.
//
// It recognizes the common `git diff` headers:
//   diff --git a/<path> b/<path>
//   --- a/<path>
//   +++ b/<path>
//   @@ -<old_start>[,<old_lines>] +<new_start>[,<new_lines>] @@ [<context>]
//
// The hunk's `Symbol` is filled from the trailing context tag on the @@ line
// when one is present (git emits the enclosing function/section name there).
//
// Hunk IDs are assigned sequentially: hunk_001, hunk_002, ...
func ParseHunks(unifiedDiff, patchRef string) []Hunk {
	var (
		hunks   []Hunk
		curFile string
		next    = 1
	)
	sc := bufio.NewScanner(strings.NewReader(unifiedDiff))
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "+++ b/"):
			curFile = strings.TrimPrefix(line, "+++ b/")
		case strings.HasPrefix(line, "+++ ") && !strings.HasPrefix(line, "+++ /dev/null"):
			curFile = strings.TrimPrefix(line, "+++ ")
		case strings.HasPrefix(line, "@@ "):
			h, ok := parseHunkHeader(line)
			if !ok {
				continue
			}
			h.HunkID = fmt.Sprintf("hunk_%03d", next)
			h.File = curFile
			h.PatchRef = patchRef
			hunks = append(hunks, h)
			next++
		}
	}
	return hunks
}

// parseHunkHeader parses a line like:
//   @@ -38,12 +40,37 @@ func authenticate(...)
// Returns false if the line doesn't match.
func parseHunkHeader(line string) (Hunk, bool) {
	rest := strings.TrimPrefix(line, "@@ ")
	end := strings.Index(rest, " @@")
	if end < 0 {
		return Hunk{}, false
	}
	ranges := rest[:end]
	tail := strings.TrimSpace(rest[end+3:])
	parts := strings.Fields(ranges)
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "-") || !strings.HasPrefix(parts[1], "+") {
		return Hunk{}, false
	}
	oldStart, oldLines := parseRange(parts[0][1:])
	newStart, newLines := parseRange(parts[1][1:])
	return Hunk{
		OldStart: oldStart,
		OldLines: oldLines,
		NewStart: newStart,
		NewLines: newLines,
		Symbol:   tail,
	}, true
}

// parseRange parses "<start>" or "<start>,<count>". Missing count defaults to 1.
func parseRange(s string) (start, count int) {
	if i := strings.Index(s, ","); i >= 0 {
		start, _ = strconv.Atoi(s[:i])
		count, _ = strconv.Atoi(s[i+1:])
		return
	}
	start, _ = strconv.Atoi(s)
	return start, 1
}
