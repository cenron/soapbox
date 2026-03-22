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

## What This Is
A pre-2022 Twitter clone — chronological microblogging platform

## Tech Stack
# Go Conventions

- Use **air** for hot reloading during development (`air` watches for file changes and rebuilds automatically).
- Set up a `.air.toml` config at the project root to configure build commands, watched directories, and excluded paths.

When debugging library internals in Go, find the source in the module cache:
```bash
find $(go env GOMODCACHE) -path "*<module>*" -name "*.go" | xargs grep -l "<symbol>"
```

# Node/TypeScript Conventions

- Use the project's package manager (`npm`, `pnpm`, or `yarn`) consistently — don't mix.
- `npm run dev` / `pnpm dev` — start the dev server
- `npm test` / `pnpm test` — run tests
- `npm run lint` / `pnpm lint` — run linter

When debugging library internals in Node, find the source in node_modules:
```bash
find node_modules -name "*.js" -path "*<package>*" | head -20
```

## MANDATORY: Module boundary enforcement

Modules are isolated. This is the foundational architectural principle of this project. Every violation creates a hidden coupling that will break when modules are split into separate services.

- **`internal/<module>/` directories are sovereign territory.** Only the developer actively building that module may modify files in its directory.
- **No cross-module imports.** If you find yourself writing `import "soapbox/internal/posts"` from inside `internal/feed/`, STOP. You are violating the boundary. Use the bus.
- **No cross-schema queries.** If you find yourself writing a SQL query that joins across schemas (e.g., `posts.posts JOIN users.profiles`), STOP. You are violating the boundary. Use the bus.
- **`internal/shared/` is the only shared code.** If two modules need the same utility, it goes in shared. But shared NEVER contains business logic — only infrastructure (bus, db, http, cache, registry, types).

If you catch yourself about to violate any of these, stop and tell the user what you were about to do and why it's wrong.

## Commands
<!-- Filled in once the project has build/run/test commands -->

## Architecture
<!-- Project structure diagram -->
<!-- Data flow if applicable -->

## Code Patterns
<!-- Project-specific conventions not covered by docs/design-principles.md -->

## Key Files
<!-- Updated as the project grows -->

## Lessons Learned

This project maintains a living document at `.claude/lessons_learned.md`. See `docs/design-principles.md` § Git & Workflow for the full protocol.

## Environment Variables
See `.env.example` for all vars.

## Project Documentation
- Design principles: `docs/design-principles.md`
- Architecture decisions: `docs/decisions/`
- Design specs: `docs/specs/`
