# Lessons Learned

<!-- Entry format:
## [YYYY-MM-DD] Short title
**What happened:** Description
**Takeaway:** The rule or insight
-->

## [2026-03-21] TypeScript erasableSyntaxOnly blocks parameter properties
**What happened:** Vite's React TS template enables `erasableSyntaxOnly` in tsconfig.app.json. Using `constructor(public readonly x: string)` syntax fails with TS1294.
**Takeaway:** Always use explicit field declarations + constructor assignment instead of parameter properties in this project. E.g., declare `readonly x: string` as a field, then `this.x = x` in the constructor.

## [2026-03-21] Response body can only be consumed once in tests
**What happened:** `mockFetch.mockResolvedValue(new Response(...))` returns the same Response object for every call. After the first `res.json()`, the body is consumed. Subsequent calls to `res.json()` throw.
**Takeaway:** When testing multiple fetch calls that read the body, use `mockResolvedValueOnce` for each call, or avoid calling the API twice in the same test.

## [2026-03-21] shadcn buttonVariants export triggers react-refresh lint error
**What happened:** shadcn generates `export { Button, buttonVariants }` — the `cva()` call result is a `CallExpression`, not covered by `allowConstantExport`.
**Takeaway:** Add `buttonVariants` (and similar shadcn non-component exports) to `allowExportNames` in the eslint config. This will recur with each new shadcn component that exports variants.

## [2026-03-21] shadcn init places files based on components.json aliases
**What happened:** Default shadcn init puts components in `src/components/ui/` and utils in `src/lib/utils.ts`. Our project uses `src/shared/ui/` and `src/shared/lib/utils.ts`.
**Takeaway:** Update `components.json` aliases immediately after init — before adding any components — to point `ui` to `@/shared/ui`, `lib` to `@/shared/lib`, `hooks` to `@/shared/hooks`. Future `npx shadcn add` commands will then place files correctly.
