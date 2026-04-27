---
name: clankstamp
description: Create a clankstamp replay artifact (a "stamp") after meaningful code changes so a human can replay what changed, why it changed, where it changed, and what needs review.
---

# clankstamp Skill

Use this skill when you make meaningful changes to code, configuration, tests, migrations, dependencies, or architecture.

Your job is to leave behind a **stamp** under `.clankstamp/` that helps the human quickly rebuild their mental model. A normal diff shows what changed. Chat may explain why. A clankstamp combines both into a guided replay tour anchored to exact files, hunks, evidence, and review questions.

## Rules

- Do **not** store hidden chain-of-thought.
- **Do** store concise rationale, evidence, assumptions, risks, and review questions.
- Prefer exact file paths, symbols, hunks, and test evidence.
- Order tour steps by **human comprehension**, not file path.
- Keep each step short.
- Highlight risky changes: auth, billing, permissions, migrations, destructive operations, dependencies, config.
- If a change is off-plan or surprising, create a tour step for it.
- Always note whether tests were added or run.
- Always include git base/head metadata when available.
- Always write JSONL artifacts under `.clankstamp/`.

## Recommended flow

Before changes:

```bash
clankstamp start --title "<task title>" --task "<task id if available>"
```

During work, record important activity:

```bash
clankstamp record file.read <path>
clankstamp record file.write <path>
clankstamp record test.run -- <test command>
```

After changes:

```bash
clankstamp finalize --auto-tour
```

If clankstamp was not started before work, create a post-hoc stamp:

```bash
clankstamp create --from-worktree --title "<task title>" --auto-tour
```

## Tour step requirements

Each tour step should answer:

1. What changed?
2. Why did it change?
3. Where is the code?
4. What evidence supports this?
5. What should the human review?

## Example: a good step

**Title:** Auth middleware now checks API keys

**Summary:** The middleware checks `X-API-Key` before falling back to JWT authentication.

**Why:** Machine clients need tenant-scoped authentication without user sessions.

**Files:**
- `app/auth/middleware.py`
- `app/auth/api_keys.py`

**Review:**
- Can this bypass route permissions?
- Is tenant context attached correctly?

**Evidence:**
- `validate_api_key()` added
- middleware calls `validate_api_key()`
- `tests/test_api_keys.py` covers valid and invalid keys
