---
name: scaffolding-only-means-scaffolding
description: When asked to scaffold, do not add implementation code — only the bare minimum structure
type: feedback
---

When the user asks for "scaffolding", provide only the minimal skeleton (project init, dependency install, empty entrypoints). Do not add routes, components, API layers, styles, or any implementation logic.

**Why:** User explicitly stopped me from going beyond scaffold into implementation work (routes, API wrappers, styled nav, page components).

**How to apply:** If the task says "scaffold" or "setup", stop after the project structure compiles/runs empty. Wait for explicit instructions before adding any feature code.
