# Progress

Status: pre-alpha — data model and capture layer exist; no UI yet.

## Done

- Repo initialized with MIT license and `.gitignore`
- Go module set up (`lshwtui`)
- `lshw.Node` model with schema-tolerant field accessors (`internal/lshw/node.go`)
- `lshw.Parse` — decodes `lshw -json` output into a parent-linked tree,
  handling both single-object and array top-level shapes (`internal/lshw/parse.go`)
- Capture/snapshot sources: live capture, root re-run via sudo/doas/pkexec,
  snapshot save/load (`internal/lshw/source.go`)

## In progress

- `internal/theme` — glyphs and lipgloss styles
- `internal/ui` — Bubble Tea model (tree pane, detail pane, status bar)
- `main.go` — CLI entrypoint and flag parsing

## Planned

- Tree pane: collapsible system > bus > device rendering with indent guides
- Detail pane: field list plus `configuration` / `capabilities` sections
- Navigation keys: `j`/`k`, `g`/`G`, `ctrl+d`/`ctrl+u`, `h`/`l`, `enter`/`space`, `tab`
- Incremental search: `/`, `n`, `N`, `esc`
- Raw-JSON view toggle (`x`)
- Yank to clipboard: node detail (`y`) and raw JSON (`Y`)
- Reload (`r`) and privileged reload (`R`) wired into the UI
- Snapshot save from the UI (`s`)
- Help overlay (`?`)
- CLI flags: `-f`/`--file`, `--no-glyphs`, `--version`
- `--no-glyphs` fallback rendering in `internal/theme`

## Decisions

| Date | Decision | Rationale |
|---|---|---|
| 2026-09-05 | Go + Bubble Tea over gocui or Textual | Single static binary, no runtime dependency, active ecosystem (lipgloss) for styling a two-pane layout |
| 2026-09-05 | Two-pane tree + detail layout | Mirrors lazygit's model: keep the navigable structure and the selected item's detail both on screen at once, instead of paging through one flat view |
| 2026-09-05 | Unprivileged by default, in-app root re-run (`R`) | `lshw` only sees the full tree as root; forcing sudo on every launch is unnecessary friction when a user just wants a quick look |
| 2026-09-05 | v1 scope: search + snapshot + raw JSON + yank | Smallest feature set that makes the tool usable standalone (find a device, capture/share a report, inspect exact fields, get data out) without scope creep into e.g. diffing two snapshots |
