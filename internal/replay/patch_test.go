package replay

import "testing"

func TestParseHunks(t *testing.T) {
	const diff = `diff --git a/foo.go b/foo.go
index 1111111..2222222 100644
--- a/foo.go
+++ b/foo.go
@@ -1,3 +1,4 @@ package foo
 line a
-line b
+line bb
+line c
diff --git a/bar.go b/bar.go
index 3333333..4444444 100644
--- a/bar.go
+++ b/bar.go
@@ -10,5 +12,7 @@ func authenticate(req *Request)
 ctx := req.Context()
-old1
-old2
+new1
+new2
+new3
+new4
@@ -50 +60,2 @@
-x
+y
+z
`
	hunks := ParseHunks(diff, "patches/full.patch")
	if got, want := len(hunks), 3; got != want {
		t.Fatalf("len(hunks) = %d, want %d", got, want)
	}

	// First hunk: foo.go @@ -1,3 +1,4 @@
	h := hunks[0]
	if h.HunkID != "hunk_001" || h.File != "foo.go" {
		t.Errorf("hunk[0] id/file = %q/%q", h.HunkID, h.File)
	}
	if h.OldStart != 1 || h.OldLines != 3 || h.NewStart != 1 || h.NewLines != 4 {
		t.Errorf("hunk[0] ranges = -%d,%d +%d,%d", h.OldStart, h.OldLines, h.NewStart, h.NewLines)
	}
	if h.Symbol != "package foo" {
		t.Errorf("hunk[0] symbol = %q", h.Symbol)
	}

	// Second hunk: bar.go @@ -10,5 +12,7 @@
	h = hunks[1]
	if h.File != "bar.go" || h.OldStart != 10 || h.NewLines != 7 {
		t.Errorf("hunk[1] = %+v", h)
	}
	if h.Symbol != "func authenticate(req *Request)" {
		t.Errorf("hunk[1] symbol = %q", h.Symbol)
	}

	// Third hunk: single-line range form @@ -50 +60,2 @@ — old count defaults to 1.
	h = hunks[2]
	if h.OldStart != 50 || h.OldLines != 1 || h.NewStart != 60 || h.NewLines != 2 {
		t.Errorf("hunk[2] = %+v", h)
	}
}

func TestParseHunks_Empty(t *testing.T) {
	if got := ParseHunks("", "p"); len(got) != 0 {
		t.Errorf("empty diff produced %d hunks", len(got))
	}
}
