---
name: clankstamp
description: Create a clankstamp replay artifact (a "stamp") after meaningful code changes so a human can replay what changed, why it changed, where it changed, and what needs review. Trigger when modifying code, config, tests, migrations, dependencies, or architecture.
---

# clankstamp Skill (Claude Code)

This is the Claude Code variant of the clankstamp skill. Behavior is identical to the open `SKILL.md` standard variant — see the rules and recommended flow below.

## When to use

Trigger this skill when you make meaningful changes to:

- Source code (functions, classes, modules)
- Configuration files (auth, routing, feature flags)
- Database migrations
- Tests (added, removed, or changed)
- Dependencies (added, upgraded, removed)
- Architecture decisions (new services, refactors)

Skip for: typo fixes, comment-only changes, formatting-only changes.

## Rules

- Do **not** store hidden chain-of-thought.
- **Do** store concise rationale, evidence, assumptions, risks, and review questions.
- Prefer exact file paths, symbols, hunks, and test evidence.
- Order tour steps by **human comprehension**, not file path.
- Keep each step short.
- Highlight risky changes: auth, billing, permissions, migrations, destructive operations, dependencies, config.
- Always note whether tests were added or run.
- Always include git base/head metadata when available.
- Always write JSONL artifacts under `.clankstamp/`.

## Recommended flow

```bash
# Before changes (preferred)
clankstamp start --title "<task title>" --task "<task id>"

# During work
clankstamp record file.read <path>
clankstamp record file.write <path>
clankstamp record test.run -- <test command>

# After changes
clankstamp finalize --auto-tour

# Or post-hoc from working tree
clankstamp create --from-worktree --title "<task title>" --auto-tour
```

## Tour step requirements

Each step answers: **what changed, why, where, what evidence, what to review.**
