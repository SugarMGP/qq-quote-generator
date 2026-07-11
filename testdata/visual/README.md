# Visual regression fixtures

The fixtures are deterministic local files. Run:

```powershell
powershell -ExecutionPolicy Bypass -File native/resvg/build.ps1
go run ./cmd/visual-regression -fixture testdata/visual/messages.json -out testdata/visual/out/current
```

The comparator restores Chromium code from commit `49853a6`, uses one browser
page, and sends identical JSON containing identical data-URI images to both
services. Both renderers resolve the same system-font family list.

Verified on 2026-07-12:

- Chromium dimensions: 600×419
- resvg dimensions: 600×419
- exact changed pixels: 34,715 / 251,400
- exact difference ratio: 0.13808671439936357

The raw ratio intentionally has no heuristic pass threshold. Inspection of
`chromium.png`, `resvg.png`, and `diff.png` confirmed aligned card, row, avatar,
bubble, text-line, and message-image bounds. Remaining changes are rasterizer
sampling differences at glyph edges, rounded edges, and scaled image pixels.
