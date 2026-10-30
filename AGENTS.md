# Agent Instructions

Follow the project-specific requirements in [`ANGENTS.md`](ANGENTS.md).

## Graphify

- Use Graphify when it helps with broad codebase exploration, architecture questions, symbol relationships, or change-impact analysis.
- Before relying on graph data, check whether `graphify-out/graph.json` exists and is current. Refresh it from the repository root with `graphify extract . --code-only` when it is missing or stale.
- Run `graphify cluster-only .` when a community-level report is useful; this generates `graphify-out/GRAPH_REPORT.md`.
- Treat Graphify as a navigation aid, not a substitute for reading relevant source files, tests, or configuration directly.
- Do not run Graphify for every small edit or for tasks that are better answered by a focused text search or direct file inspection.

## Tests

- Write or update test cases when a change adds or alters behavior that can be tested.
- Run the relevant tests after changes and report any tests that could not be run.
