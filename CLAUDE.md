# soapbox — Project Conventions

> Read `docs/design-principles.md` at the start of every task. It contains the core engineering philosophy and locked sections that must not be modified unless explicitly asked.

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
