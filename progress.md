# Progress

Status: v1 feature-complete — builds, tests pass, all planned keybindings wired.
Not yet exercised in a real terminal session by a human.

## Done

- Repo initialized with MIT license and `.gitignore`
- Go module set up (`lshwtui`), `go` directive relaxed to 1.24
- `lshw.Node` model with schema-tolerant field accessors (`internal/lshw/node.go`)
- `lshw.Parse` — decodes `lshw -json` into a parent-linked tree, handling both
  single-object and array top-level shapes (`internal/lshw/parse.go`)
- Capture/snapshot sources: live capture, root re-run via sudo/doas/pkexec,
  snapshot save/load (`internal/lshw/source.go`)
- `internal/theme` — per-class Nerd Font glyphs with ASCII fallbacks, adaptive
  colour palette, pane and status styles
- `internal/ui/tree.go` — expand/collapse state, flattened rows, indent guides
- `internal/ui/search.go` — incremental search across every field including
  `configuration` and `capabilities`, matches in tree order
- `internal/ui/detail.go` — detail lines as pure data, byte/Hz humanisation
- `internal/ui/clipboard.go` — wl-copy/xclip/xsel/pbcopy with an OSC 52 fallback
- `internal/ui/model.go`, `update.go`, `view.go`, `help.go` — the Bubble Tea app
- `main.go` — CLI with `-f`/`--file`, `--no-glyphs`, `--version`
- `Makefile` — build, run, test, vet, fmt, check, install, snapshot, clean
- Tests: `internal/lshw` (parsing, accessors, snapshot round-trip) and
  `internal/ui` (model driven by a redacted real capture in `testdata/`)

## In progress

- Nothing. Awaiting a real terminal run and feedback.

## Planned

- Verify glyph rendering and colours in a live terminal with a Nerd Font
- A remote has not been created yet; nothing is pushed
- Possible follow-ups, explicitly out of v1 scope: diffing two snapshots,
  filtering the tree to matches only, exporting a device as YAML/text report

## Decisions

| Date | Decision | Rationale |
|---|---|---|
| 2026-09-05 | Go + Bubble Tea over gocui or Textual | Single static binary, no runtime dependency, active ecosystem (lipgloss) for styling a two-pane layout |
| 2026-09-05 | Two-pane tree + detail layout | Mirrors lazygit's model: keep the navigable structure and the selected item's detail both on screen at once, instead of paging through one flat view |
| 2026-09-05 | Unprivileged by default, in-app root re-run (`R`) | `lshw` only sees the full tree as root; forcing sudo on every launch is unnecessary friction when a user just wants a quick look |
| 2026-09-05 | v1 scope: search + snapshot + raw JSON + yank | Smallest feature set that makes the tool usable standalone (find a device, capture/share a report, inspect exact fields, get data out) without scope creep into e.g. diffing two snapshots |
