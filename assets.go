// Package clankstamp is the root library package for the clankstamp project.
//
// Its job is to expose the assets that ship inside the binary — the embedded
// SKILL.md drafts, the Neovim Lua plugin, and the JSON schemas — so they can
// be installed onto disk by `clankstamp skill install` and `clankstamp nvim
// install`. Keeping the //go:embed directive at the repo root means the
// human-readable source of those assets lives in `embedded/` next to PRD.md
// rather than buried inside an internal package.
package clankstamp

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed all:embedded
var embedded embed.FS

// SkillTarget identifies which SKILL.md variant to install.
type SkillTarget string

const (
	SkillTargetStandard   SkillTarget = "standard"
	SkillTargetClaudeCode SkillTarget = "claude-code"
)

// SkillFS returns a filesystem rooted at the SKILL.md's parent directory for
// the requested target. The returned FS contains a `SKILL.md` at its root
// (and any future supporting files alongside it).
func SkillFS(target SkillTarget) (fs.FS, error) {
	switch target {
	case SkillTargetStandard, SkillTargetClaudeCode:
		return fs.Sub(embedded, "embedded/skills/"+string(target)+"/clankstamp")
	default:
		return nil, fmt.Errorf("unknown skill target %q (expected: standard, claude-code)", target)
	}
}

// NvimFS returns a filesystem rooted at the Neovim plugin tree. The returned
// FS has the plugin/ and lua/ directories at its root, ready to be copied
// directly into a `clankstamp.nvim/` install path.
func NvimFS() (fs.FS, error) {
	return fs.Sub(embedded, "embedded/nvim")
}

// SchemaFS returns a filesystem rooted at the embedded JSON schemas. Used by
// the future `clankstamp validate` command.
func SchemaFS() (fs.FS, error) {
	return fs.Sub(embedded, "embedded/schemas")
}
