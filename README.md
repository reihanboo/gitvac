# gitclean

A Windows-first terminal utility for finding disposable project directories below `C:\Users\reihan\git` and deleting only those you explicitly select.

## Run

```powershell
go run ./cmd/gitclean

go run ./cmd/gitclean --root C:\src
go run ./cmd/gitclean --scan
```

Build the Windows executable with `go build -o gitclean.exe ./cmd/gitclean`. The executable is named `gitclean`; the Go module remains `gitvac`.

`--help` lists the CLI options. `--scan` prints matching paths and sizes without opening the TUI.

## Configuration

An optional JSON file can be placed at `%AppData%\gitvac\config.json` (the path returned by Go's `os.UserConfigDir()`). Defaults are used when the file does not exist. Example:

```json
{
  "root": "C:\\Users\\reihan\\git",
  "targets": ["node_modules", ".venv", "venv", "__pycache__", ".pytest_cache", ".mypy_cache", "target", "dist", "build", ".next", ".svelte-kit", ".turbo"],
  "excludes": [".git", "vendor"]
}
```

`vendor` is not a default target. `--root` overrides the configured root.

## Controls

- `j` / `k` or arrow keys: move
- `space`: select or deselect
- `a` / `n`: select all / none
- `d`: open the deletion confirmation; `y` confirms, `n` or `esc` cancels
- `q` / `ctrl+c`: quit
