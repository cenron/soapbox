---
name: review-pr
description: Pull GitHub PR comments, triage by severity, assess if they should be fixed, and apply soapbox-specific review criteria (module boundaries, event contracts, schema isolation). Use when the user says "review PR", "check PR comments", "address feedback", or references a PR number.
---

# Review PR — Soapbox Project

You are a PR comment reviewer for the soapbox modular monolith. Your job is to pull all comments from a GitHub PR, evaluate each one against both general code quality AND soapbox-specific architectural rules, and present a triage before making any changes.

## Step 1: Identify the PR

- If the user provides a PR number (e.g., `/review-pr 12`), use that.
- If no number is given, find the PR for the current branch:
  ```bash
  gh pr view --json number,title,url,headRefName
  ```
- If no PR exists for the current branch, tell the user and stop.

## Step 2: Pull all comments

```bash
# Inline review comments (on specific lines of code)
gh api repos/{owner}/{repo}/pulls/{pr_number}/comments

# General conversation comments
gh api repos/{owner}/{repo}/issues/{pr_number}/comments
```

Parse out: `id`, `body`, `path`, `line`/`original_line`, `user.login`, `html_url`, `in_reply_to_id`.

Filter out:
- Your own previous replies
- Comments in resolved threads where a fix was already confirmed

## Step 3: Read project rules

Before triaging, read these files to understand what's enforced:
- `docs/plan.md` — module statuses and dependencies
- `CLAUDE.md` — module boundary rules, code style, pre-PR checklist
- `docs/design-principles.md` — locked architectural principles

## Step 4: Triage each comment

For each comment, classify it into one of these categories:

| Category | Action |
|----------|--------|
| **Bug / logic error** | Fix the code, add a test |
| **Edge case not handled** | Fix the code, add a test |
| **Module boundary violation** | Fix immediately — this is always critical in soapbox |
| **Event contract violation** | STOP — flag to user, do not fix (publisher owns the schema) |
| **Schema boundary violation** | Fix immediately — no cross-schema queries allowed |
| **Style / naming / formatting** | Fix the code, no test needed |
| **Swagger annotation missing** | Fix — all endpoints must have annotations |
| **Documentation / typo** | Fix it, no test needed |
| **Question / discussion** | Skip — don't fix, don't reply |
| **Already addressed** | Skip — code already handles it |
| **Disagrees with design** | Skip — flag to user for decision |

Assign a severity level:

| Severity | Meaning | Soapbox-specific examples |
|----------|---------|---------------------------|
| 🔴 **Critical** | Will cause bugs, data loss, or architectural violation | Cross-module imports, cross-schema JOINs, wrong event contracts, SQL injection, missing auth middleware |
| 🟡 **Moderate** | Correctness issue in edge cases, or degrades reliability | Unhandled edge cases, missing validation at boundaries, bus query not registered, missing swagger annotations |
| 🟢 **Low** | Cosmetic, cleanup, or minor improvement | Unused imports, typos, naming nits, style inconsistencies |

### Soapbox-specific checks

When reviewing comments, also proactively check the PR diff for these violations even if no comment mentions them:

1. **Cross-module imports** — any import of `internal/<other-module>/` that isn't `internal/core/`
2. **Cross-schema SQL** — any JOIN across schemas (e.g., `posts.posts JOIN users.profiles`)
3. **Event contract tampering** — modifying another module's published event structure
4. **Missing swagger annotations** — any HTTP handler without `@Summary`, `@Router`, etc.
5. **Module dependency gate** — check `docs/plan.md` to verify all dependencies were `complete` before this module was started

If you find violations that no reviewer caught, add them to the triage as self-identified issues.

## Step 5: Present triage

Present the full assessment to the user before making any changes:

```
## PR #12 Comment Review

### Will fix (3):
1. 🔴 @reviewer: "This imports internal/users from inside internal/posts" (handler.go:5) → remove cross-module import, use bus query instead
2. 🟡 @copilot: "Missing error handling on bus.Query" (service.go:42) → add error check + test
3. 🟢 @reviewer: "Inconsistent naming: userID vs userId" (types.go:10) → standardize to userID

### Self-identified (1):
4. 🟡 Missing swagger annotations on POST /posts endpoint (handler.go:28)

### Skipping (2):
5. @reviewer: "Should we use Redis for cache here?" → design question, needs your input
6. @reviewer: "LGTM" → no action needed

### Blocked (0):
(Event contract issues that need the other module's owner)

Proceed with fixes?
```

**Wait for the user to confirm or adjust before proceeding.**

## Step 6: Fix each issue

For each comment that needs fixing:

1. Read the relevant file to understand context
2. Make the minimal fix — don't refactor surrounding code
3. If the fix involves logic: write or update a test covering the edge case
4. If the fix is cosmetic: no test needed
5. Run `make test` after each fix

## Step 7: Reply on each fixed comment

```bash
# For inline review comments
gh api repos/{owner}/{repo}/pulls/{pr_number}/comments \
  -X POST \
  -f body="Fixed in $(git rev-parse --short HEAD) — [description]. Added test: \`test_name\`." \
  -F in_reply_to={comment_id}

# For conversation comments
gh api repos/{owner}/{repo}/issues/{pr_number}/comments \
  -X POST \
  -f body="Fixed in $(git rev-parse --short HEAD) — [description]."
```

## Step 8: Commit and summarize

1. Run `make test` and `make lint` — everything must pass
2. Stage only changed files (never `git add .`)
3. Commit referencing the PR:
   ```
   Address PR #N review feedback

   - [list of fixes]
   - Added N tests covering reported edge cases
   ```
4. Push to the branch
5. Print final summary:
   ```
   ## Summary
   - Fixed: N comments
   - Self-identified: N issues
   - Skipped: N comments (reasons)
   - Blocked: N comments (need other module owner)
   - Tests added: N
   - All tests passing
   ```

## Step 9: Update lessons learned

If any bugs reveal a repeatable pattern (not one-off typos), append to `.claude/lessons_learned.md`:

```markdown
## [YYYY-MM-DD] Short title
**What happened:** One-sentence description.
**Takeaway:** The rule to apply going forward.
```

## Important guidelines

- **Module boundaries are sacred.** Any cross-module import or cross-schema query is always 🔴 Critical, regardless of what the reviewer said.
- **Publisher owns the event contract.** If a comment asks you to modify another module's event schema, do NOT fix it. Flag it as blocked.
- **Don't fix what isn't broken.** If current code is correct, skip and flag.
- **Match existing patterns.** Check how similar code is structured in completed modules.
- **One fix at a time.** Apply sequentially for clean verification.
- **Tests prove the fix.** Target the specific edge case, not broad coverage.
- **Don't reply to design discussions.** Flag for user input.
