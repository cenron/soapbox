# soapbox — Project Conventions

> Read `docs/design-principles.md` at the start of every task. It contains the core engineering philosophy and locked sections that must not be modified unless explicitly asked.

---

## MANDATORY: Module dependency gate

**YOU MUST READ `docs/plan.md` BEFORE DOING ANY WORK ON ANY MODULE.**

This is a collaborative project with strict module dependencies. Violating the build order will break the project for the other developer. There are no exceptions.

### Rules — these are not suggestions, they are hard requirements:

1. **READ `docs/plan.md` FIRST.** Before writing a single line of code, before creating a branch, before even thinking about implementation — read the plan and check module statuses.

2. **NEVER start a module whose dependencies are not marked `complete`.** If the plan says module X depends on module Y, and Y is not `complete`, you CANNOT work on X. Period. Do not ask "can I just start the parts that don't need Y?" No. Do not ask "can I stub it out?" No. Do not rationalize your way around this. The answer is no.

3. **NEVER modify another module's code.** You are working on auth? Do not touch `internal/users/`. You are working on feed? Do not touch `internal/posts/`. Each module is owned by whoever is building it. If you need something from another module, it must already be exposed via the bus. If it isn't, you cannot work on this module yet.

4. **NEVER modify another module's event contracts.** The publisher owns the event schema. If you are consuming `posts.created` and it doesn't have a field you need, you do NOT add it. You stop and tell the user: "This module depends on a field that doesn't exist in the event contract. The posts module owner needs to add it."

5. **If all available modules are blocked**, tell the user explicitly: "All remaining modules depend on [X] which is currently in progress. I can work on bug fixes, test coverage, documentation, or technical debt instead." Do NOT silently start a blocked module.

6. **When you complete a module**, update its status in `docs/plan.md` to `complete` BEFORE pushing. This is not optional. The other developer's Claude Code session reads this file to know what's available.

### How to check if you can start a module:

```
1. Read docs/plan.md
2. Find the module the user wants to work on
3. Check its "Dependencies" line
4. For EACH dependency, check if its status is "complete"
5. If ANY dependency is NOT complete → STOP. Tell the user. Do not proceed.
6. If ALL dependencies are complete → proceed
```

### What to say when blocking:

> "I cannot start work on [module]. It depends on [dependency] which is currently [status]. The plan requires all dependencies to be complete before starting a module. Would you like to work on [list available modules], or would you prefer to work on bug fixes, tests, or technical debt?"

Do not soften this. Do not offer workarounds. Do not say "but I could get started on the parts that don't depend on it." The module is blocked. Say so clearly.

### What to say when the user insists:

> "I understand you want to start [module], but the build order exists to prevent integration failures between collaborators. If I start this module before [dependency] is complete, the event contracts and query interfaces may not match, which will cause breakage when both branches merge. I strongly recommend waiting or working on something else."

If the user explicitly overrides after this warning, comply but add a comment at the top of every file you create: `// WARNING: Started before dependency [X] was complete. Verify contracts before merging.`

---

## What this is

A pre-2022 Twitter clone — chronological microblogging platform. Built as a modular monolith designed to scale into microservices when needed.

---

## MANDATORY: Module boundary enforcement

Modules are isolated. This is the foundational architectural principle of this project. Every violation creates a hidden coupling that will break when modules are split into separate services.

- **`internal/<module>/` directories are sovereign territory.** Only the developer actively building that module may modify files in its directory.
- **No cross-module imports.** If you find yourself writing `import "soapbox/internal/posts"` from inside `internal/feed/`, STOP. You are violating the boundary. Use the bus.
- **No cross-schema queries.** If you find yourself writing a SQL query that joins across schemas (e.g., `posts.posts JOIN users.profiles`), STOP. You are violating the boundary. Use the bus.
- **`internal/shared/` is the only shared code.** If two modules need the same utility, it goes in shared. But shared NEVER contains business logic — only infrastructure (bus, db, http, cache, registry, types).

If you catch yourself about to violate any of these, stop and tell the user what you were about to do and why it's wrong.

---

## Never guess

- **NEVER guess** about APIs, framework behavior, best practices, or library usage. If you are not 100% certain, look up the latest official documentation FIRST.
- If documentation is unavailable or ambiguous, **ask the user** before proceeding. Do not assume, improvise, or rely on potentially outdated training knowledge.
- This applies to: API signatures, framework conventions, config options, CLI flags, library methods, and any behavior you cannot verify from the codebase or docs.
- Wrong guesses compound into hard-to-find bugs. **When in doubt, look it up or ask.**

### High-risk identifiers — always verify from docs:

- Model names / model IDs
- API endpoint paths and versions
- SDK method signatures and config object fields
- Auth mechanisms (API key vs service account vs OAuth)

### Debugging library internals:

When a bug involves library internals, **read the installed source code first** — find it in the project's dependency directory (`node_modules/`, Go module cache via `go env GOMODCACHE`). Never state how a library works based on memory alone.

## Code style

### Simplicity above all

- Prefer the simplest solution. No abstractions for hypothetical scenarios.
- Only change what's needed. Don't refactor, add comments, or "improve" surrounding code unless asked.
- Three similar lines of code is better than a premature abstraction.

### Consistency is non-negotiable

- Once a pattern is established, propagate it to all similar code. No mixed generations of style.
- Read before writing. Understand existing patterns before changing anything.
- Match the project's existing naming, formatting, and conventions.
- Follow the language/framework's idiomatic best practices — they take priority over personal style.

### Code reads like a narrative

- Orchestrator methods describe *what* happens; helper methods handle *how*.
- Extract even small operations into named methods when it makes the calling code read like an outline.
- Guard clauses at the top, happy path flows straight down at the lowest indentation.
- Use blank lines to group related logic into visual blocks — code reads as grouped chunks, not a wall of text.

### Small files, clear boundaries

- One logical unit per file — extract early, not when it "gets big enough."
- If you can name the concept, it gets its own file.

### Delete, don't comment

- Delete dead code. Don't comment it out, rename to `_unused`, or add `// removed` markers.
- No backwards-compatibility hacks for removed code.
- No security vulnerabilities — no SQL injection, XSS, command injection, or other OWASP top 10 issues.
- Validate at system boundaries only (user input, external APIs). Trust internal code.

## Architecture

Follow SOLID principles. Keep coupling low and cohesion high.

### Design patterns

- **Dependency injection over globals.** Structs receive dependencies via constructor. Don't pull from global state or import singletons.
- **Bundle config into dedicated objects.** Group related settings into immutable config objects and inject those — don't pass raw values or read env vars deep in the call stack.
- **Composition at the entry point.** All wiring and construction happens in `cmd/web/main.go`. Use a factory function to construct the dependency graph from config. Makes it clear what depends on what.
- **Composition over inheritance.** Prefer composing objects from smaller parts. Use inheritance only for genuine is-a relationships.
- **Separate what's yours from what's theirs.** External services (S3, OAuth providers, email) get interfaces and dedicated provider folders. Provider-specific code never mixes into business logic.
- **One implementation, one location.** Shared abstractions live in `internal/shared/`. When a second module needs something that currently lives in one module, move it to shared — don't duplicate it. Extending an existing abstraction is always preferable to creating a parallel one.
- **Factory + config for wiring.** Bundle all settings into a config object, and use a factory function to construct the dependency graph from it.

### Project structure

```
soapbox/
├── cmd/web/main.go              # Composition root — wires modules
├── internal/
│   ├── shared/                  # Infrastructure only (bus, db, http, cache, registry, types)
│   │   ├── bus/                 # Bus interface + implementation
│   │   ├── registry/            # Registry interface + implementation
│   │   ├── cache/               # Cache interface + implementation
│   │   ├── db/                  # Connection pool, migrations, transactions
│   │   ├── http/                # Response helpers, middleware, pagination
│   │   └── types/               # Common types (IDs, timestamps)
│   ├── auth/                    # Auth module (credentials, sessions, roles, OAuth)
│   ├── users/                   # Users module (profiles, follows)
│   ├── posts/                   # Posts module (posts, likes, reposts, hashtags, link previews)
│   ├── feed/                    # Feed module (timeline assembly)
│   ├── notifications/           # Notifications module
│   ├── media/                   # Media module (S3 uploads)
│   └── moderation/              # Moderation module (reports, blocks, mutes, admin)
├── web/                         # React SPA (Vite + shadcn/ui + Tailwind)
├── build/
│   ├── Dockerfile               # Single image, all binaries
│   └── entrypoint.sh            # APP_MODE selects binary
└── docker-compose.yml           # Dev infra (Postgres, MinIO, Mailpit)
```

## Testing

- Run existing tests after changes to verify nothing broke.
- Write tests for new functionality unless told otherwise.
- Match the existing test framework and patterns.
- **NEVER update a test just to make it pass.** If a test fails, the code is guilty until proven innocent. Verify the actual behavior is correct first:
  - Web apps: run in-browser validation before touching the test.
  - APIs: make real requests to confirm the response is correct.
- Only update a test when you have confirmed the new behavior is genuinely correct and the test expectation was wrong or outdated.

## Verification loop

Before considering any task complete:

1. Typecheck (if applicable)
2. Lint
3. Run affected tests
4. Before PRs: run the full test suite
5. **After major refactors:** full consistency pass across all touched files:
   - Mixed patterns (old style vs new style living side by side)
   - Naming inconsistencies (e.g. camelCase in one file, snake_case in another)
   - Orphaned imports, dead code, or leftover references to removed code
   - Architectural violations (logic in the wrong layer, broken module boundaries)
   - No Frankenstein code — the codebase should read like one person wrote it

Don't just run the steps — prove it works. Ask: "Would a staff engineer approve this?"

## Git

**Trunk-based development:**

- `main` is the trunk — always shippable, never broken.
- All work happens on short-lived feature branches cut from `main`.
- One branch per module or logical unit of work.
- Open a PR only when work is complete and ready to merge.
- Branch naming: `phase/<name>`, `feat/<name>`, `fix/<name>`.
- Commit messages: concise, imperative mood, focused on "why" not "what."
- Stage specific files — never `git add .` or `git add -A`.
- Don't commit or push unless asked.

### MANDATORY: Pre-PR consistency check

**Before creating any PR, you MUST run a full code consistency check.** This is not optional. Do not create the PR until every item passes.

1. **Run all tests** — `make test`. Every test must pass. No exceptions.
2. **Run linter** — `make lint`. Zero warnings, zero errors.
3. **Run typecheck** — frontend: `npx tsc --noEmit`, backend: `go vet ./...`.
4. **Module boundary audit** — grep the entire module directory for imports of other `internal/` modules. If any exist, fix them before proceeding.
   ```bash
   # Backend: check for cross-module imports
   grep -r "soapbox/internal/" internal/<your-module>/ | grep -v "soapbox/internal/shared/"
   # Frontend: check for cross-module imports
   grep -r "from.*modules/" web/src/modules/<your-module>/ | grep -v "from.*shared/"
   ```
5. **Schema boundary audit** — grep for SQL joins across schemas.
   ```bash
   grep -rni "JOIN.*\." internal/<your-module>/ | grep -v "<your-schema>\."
   ```
6. **Naming consistency** — scan all new/modified files for mixed naming conventions (camelCase vs snake_case, inconsistent variable names, different patterns for the same concept).
7. **Dead code check** — no unused imports, no commented-out code, no orphaned functions.
8. **Pattern consistency** — compare your module's patterns against existing completed modules. Error handling, response formatting, test structure, and file organization must match.
9. **Design spec compliance** — verify your module implements all endpoints, events, and queries listed in `docs/specs/2026-03-21-soapbox-design.md` for this module. Nothing missing, nothing extra.
10. **Plan status update** — update `docs/plan.md` to mark this module as `complete`.

If any check fails, fix the issue and re-run ALL checks. Only create the PR when everything passes clean.

## Tech stack

### Go (backend)

- Use **air** for hot reloading during development (`air` watches for file changes and rebuilds automatically).
- Set up a `.air.toml` config at the project root to configure build commands, watched directories, and excluded paths.

When debugging library internals in Go, find the source in the module cache:
```bash
find $(go env GOMODCACHE) -path "*<module>*" -name "*.go" | xargs grep -l "<symbol>"
```

### Node/TypeScript (frontend)

- Use the project's package manager (`npm`, `pnpm`, or `yarn`) consistently — don't mix.
- `npm run dev` / `pnpm dev` — start the dev server
- `npm test` / `pnpm test` — run tests
- `npm run lint` / `pnpm lint` — run linter

When debugging library internals in Node, find the source in node_modules:
```bash
find node_modules -name "*.js" -path "*<package>*" | head -20
```

## Communication

- Be concise. Lead with the action, not the reasoning.
- Don't summarize what you just did — the diff speaks for itself.
- Ask when uncertain, especially for destructive or irreversible actions.

## Commands

<!-- Filled in once the project has build/run/test commands -->

## Key files

<!-- Updated as the project grows -->

## Lessons learned

This project maintains a living document at `.claude/lessons_learned.md`.

- **Read** it at the start of every task.
- **Update immediately** — after ANY correction from the user, after hitting an unexpected error, discovering a framework quirk, or finding a non-obvious solution. Write a rule that prevents the same mistake.
- **Reference** it before making decisions — the answer may already be there from a past session.
- Never let it go stale. If a lesson no longer applies, remove it.

## Environment variables

See `.env.example` for all vars.

## Project documentation

- Design spec: `docs/specs/2026-03-21-soapbox-design.md`
- Implementation plan: `docs/plan.md`
- Design principles: `docs/design-principles.md`
- Architecture decisions: `docs/decisions/`
