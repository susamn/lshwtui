# lshwtui — agent context

`lshwtui` is a lazygit-inspired terminal UI for `lshw`. It parses
`lshw -json` into a tree and lets you browse it with vim-style keys instead
of scrolling raw text.

## Package layout

| Package | Purpose | Status |
|---|---|---|
| `main.go` (root) | CLI entrypoint, flag parsing | not yet created |
| `internal/lshw` | Data model + capture: `Node`, JSON parsing, live/sudo/snapshot sources | exists |
| `internal/theme` | Glyphs and lipgloss styles, including the `--no-glyphs` fallback | not yet created |
| `internal/ui` | Bubble Tea model: tree pane, detail pane, search, help overlay | not yet created |

**Rule: `internal/lshw` must stay UI-agnostic.** It knows nothing about
Bubble Tea, lipgloss, or terminal rendering — only `map[string]any` decoding,
tree construction, and shelling out to `lshw`. Rendering, keybindings, and
layout belong in `internal/ui` and `internal/theme`.

## Build

This machine has a broken `GOROOT` env var. Every `go` invocation must strip
it:

```sh
env -u GOROOT go build -o lshwtui .
env -u GOROOT go test ./...
env -u GOROOT go vet ./...
```

Plain `go build`/`go test` will fail or behave unpredictably here — don't
drop the `env -u GOROOT` prefix in scripts, CI notes, or ad-hoc commands.

## `lshw.Node` design — do not "clean this up"

`lshw.Node` (`internal/lshw/node.go`) wraps the raw decoded
`map[string]any` instead of a fixed struct, and exposes typed accessors
(`Str`, `Num`, `Map`, `List`) on top of it. This is deliberate, not
unfinished:

- lshw's JSON schema is loose — which keys appear depends on device class,
  and new keys show up across lshw versions.
- A rigid struct would silently drop fields nobody anticipated when it was
  written, with no way to tell the difference between "field absent on this
  device" and "field not in the struct."
- The raw map is always available (e.g. for the raw-JSON view / yank
  feature), and the accessor layer gives call sites a typed API without
  giving up that flexibility.

Do not replace `Node.Raw` with a generated or hand-written struct. If a new
typed accessor is needed, add a method next to the existing ones in
`node.go` rather than reshaping the underlying storage.

## Privilege handling

`lshw` must never be assumed to run as root. Unprivileged runs return a
partial tree; treat that as the default case, not an edge case. Root-only
behavior is opt-in via explicit re-exec (`source.go`'s `PrivilegedCapture`,
`Escalator`), never assumed or silently retried.

## Commit convention

Small, single-purpose commits with conventional-commit prefixes
(`feat(scope): ...`, `fix(scope): ...`, `chore: ...`, `docs: ...`). See
existing history for the pattern, e.g.:

```
feat(lshw): add capture, privileged re-run and snapshot sources
feat(lshw): parse lshw JSON into a parent-linked tree
feat(lshw): add Node model with schema-tolerant field accessors
chore: init repo with MIT license and gitignore
```

Keep unrelated changes (docs, a new accessor, a UI keybinding) in separate
commits even when done in the same session.
