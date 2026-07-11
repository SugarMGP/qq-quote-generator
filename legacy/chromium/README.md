# Chromium renderer baseline

The complete pre-refactor Chromium implementation is preserved by immutable
Git commit `49853a6`. It contains `renderer.go`, `pool.go`, and `template.go`.

`go run ./cmd/visual-regression` materializes that exact commit in a temporary
directory, injects the same embedded Noto font used by the SVG renderer, builds
both versions, submits the same JSON payload, and writes:

- `chromium.png`
- `resvg.png`
- `diff.png`
- `report.json`

The old browser code is therefore reproducible without being compiled into the
production binary or retained in the production dependency graph.
