# Svelte UI Experiment — Feature Specification

## Motivation

The current frontend is a single `index.html` file (~29KB) using Alpine.js + Tailwind/DaisyUI from CDNs.
It works, but as the UI grows, the monolithic inline approach creates friction:

- **All state lives in one `app()` function** (~230 lines) with 30+ reactive properties — hard to reason about.
- **No component boundaries** — the config card, templates card, jobs table, and three processing forms are all interleaved in one file. Changes to one section risk breaking another.
- **No build step** means no tree-shaking, no minification, no dead-code elimination. The full Alpine.js and DaisyUI libraries ship to the browser even though we use a fraction.
- **Lazy-loading is hand-rolled** — `if (templatesOpen && tmpl.resume === undefined)` scattered through methods. A component lifecycle would handle this naturally.
- **No TypeScript** — API response shapes are implicit, making refactors fragile.

Svelte addresses all of these while compiling down to a small, framework-free JS bundle that can still be embedded in the Go binary via `embed`.

---

## Current UI Features (parity target)

### 1. Configuration Card
- Backend selection (DeepSeek / Kimi K2.5)
- Model selection (per-backend)
- API key input with show/hide toggle
- Port configuration
- System prompt textarea
- Task list textarea
- Single save button for all config

### 2. Templates Card
- Resume template textarea (lazy-loaded)
- Cover letter template textarea (lazy-loaded)
- Per-template save buttons with flash feedback

### 3. Process URL Card
- URL input + Generate button
- Success/error feedback

### 4. Process Batch Card
- Multi-line URL textarea + Generate All button
- Per-URL success/failure results list

### 5. Process Local Text Card
- Raw text textarea + Generate button
- Success/error feedback

### 6. Applications Table
- Sortable columns: Date, Company, Role, Score (color-coded badge), Status (dropdown)
- Expandable rows with inline resume/cover editing + save
- Delete button per row
- Lazy file loading on first expand

---

## What Svelte Improves

### Component isolation
| Current (Alpine) | Svelte |
|---|---|
| One 600-line HTML file | `ConfigCard.svelte`, `TemplatesCard.svelte`, `ProcessUrl.svelte`, `ProcessBatch.svelte`, `ProcessLocal.svelte`, `JobsTable.svelte`, `JobRow.svelte` |
| Shared `app()` scope — any method can touch any state | Each component owns its state; cross-component communication via props/stores |

### Reactive state
| Current | Svelte |
|---|---|
| Manual `flash(prop)` with `setTimeout` | `$effect` or transition directives (`fade`, `slide`) |
| Hand-rolled lazy loading checks | `onMount` lifecycle + `{#await}` blocks |
| `x-show` / `x-model` string directives | Compiled reactivity — `$state`, `bind:value`, `{#if}` |

### Type safety
- Svelte 5 has first-class TypeScript support.
- API response types (`Config`, `PromptConfig`, `Job`, `BatchResult`) defined once and shared across components.
- Catches shape mismatches at build time instead of runtime.

### Bundle size & embedding
- Svelte compiles to vanilla JS — no runtime shipped. The current Alpine.js CDN load (~43KB gzip) becomes ~10-15KB for the compiled app.
- Vite builds to a single `index.html` with inlined CSS/JS — drop-in replacement for `//go:embed web/index.html`.
- Tree-shaking removes unused DaisyUI utilities automatically.

### Developer experience
- Hot module replacement during development (Vite dev server proxied to Go backend).
- Component-scoped styles — no class name collisions.
- `{#each}`, `{#if}`, `{#await}` blocks replace verbose Alpine attribute soup.

---

## Proposed Component Tree

```
App.svelte
├── ConfigCard.svelte
│   ├── BackendSelect.svelte
│   ├── ApiKeyInput.svelte
│   └── PromptEditor.svelte
├── TemplatesCard.svelte
├── ProcessUrl.svelte
├── ProcessBatch.svelte
├── ProcessLocal.svelte
└── JobsTable.svelte
    └── JobRow.svelte (per row, expandable)
```

## Shared modules

```
lib/
├── api.ts          — typed fetch wrappers for all endpoints
├── types.ts        — Config, PromptConfig, Job, BatchResult, Template interfaces
├── stores.ts       — Svelte stores for config + jobs (shared across components)
└── flash.ts        — reusable flash/notification utility
```

---

## Proposed Stack

| Layer | Choice | Rationale |
|---|---|---|
| Framework | Svelte 5 (runes) | Compiled reactivity, small output, TS support |
| Build tool | Vite | Fast dev server, single-file production build |
| CSS | Tailwind v4 + DaisyUI | Keeps visual parity, utility-first |
| Language | TypeScript | Type-safe API layer |
| Embedding | `//go:embed` on `web/dist/index.html` | Same pattern as current, just built output |

---

## Migration Plan

### Phase 1 — Scaffold & dev tooling
- `npm create svelte@latest` in a `ui/` directory
- Vite config: proxy `/api/*` to Go backend during development
- Tailwind + DaisyUI setup
- Verify `go:embed` works with Vite build output

### Phase 2 — API layer + types
- Define TypeScript interfaces matching all JSON shapes
- Typed fetch wrappers with error handling
- Svelte stores for config and jobs

### Phase 3 — Component migration (feature parity)
- Migrate each card/section as an independent component
- Maintain exact same API contract — zero backend changes
- Visual parity with current DaisyUI theme

### Phase 4 — Svelte-native improvements
- Replace hand-rolled lazy loading with `{#await}` blocks
- Add Svelte transitions (`fade`, `slide`) for card expand/collapse
- Form validation with reactive declarations
- Keyboard shortcuts (Ctrl+Enter to submit)

### Phase 5 — New capabilities (post-parity)
- Dark mode toggle (DaisyUI `data-theme` swap via store)
- Table search/filter (reactive `$derived` on jobs array)
- Table sorting by column
- Pagination or virtual scrolling for large job lists
- Toast notification system (component-based, not `setTimeout`)

---

## Backend Changes Required

**None.** The Go backend serves a single embedded HTML file and exposes a JSON API.
Svelte compiles to a single HTML file with inlined assets — the `//go:embed` directive and all handlers remain unchanged.

---

## Success Criteria

1. Full feature parity with current Alpine.js UI
2. All existing API endpoints work without modification
3. Production build produces a single embeddable HTML file
4. Bundle size is smaller than current CDN-loaded approach
5. Components are independently testable
6. TypeScript catches at least one shape mismatch that currently exists silently
