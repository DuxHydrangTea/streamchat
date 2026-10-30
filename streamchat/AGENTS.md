# StreamChat Agent Instructions

Follow the inherited project requirements in [`../AGENTS.md`](../AGENTS.md) and [`../ANGENTS.md`](../ANGENTS.md). This file adds repository-specific guidance and does not replace those instructions.

## Impeccable

Use the Impeccable skill for frontend and visual-design work when it is available in the current agent environment. Before making UI changes:

- Run the Impeccable context command for the target surface and inspect the applicable command reference/playbook.
- For scoped refinements, treat the current implementation as the visual baseline: preserve its behavior and identity outside the requested scope, and use source code and rendered output as evidence.
- After UI changes, run the Impeccable detector on the changed targets when the CLI is available, then address relevant findings.
- Do not run the Impeccable installer unless explicitly requested for the current task.
- Do not create or rewrite `PRODUCT.md`, `DESIGN.md`, or surface briefs as a side effect. The project does not yet have confirmed product/design context files; do not invent personas, positioning, brand commitments, visual authority, or design decisions. Ask the user when a requested new surface or visual replacement needs those decisions.
- For narrow UI refinements that do not require missing product decisions, proceed from the confirmed request and the existing code without blocking on initialization.
