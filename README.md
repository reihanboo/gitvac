# gitvac

TUI tool for Windows to quickly find and clean up build junk/folders like `node_modules` under `C:\Users\reihan\git`. 

## Quick Start

```powershell
go run ./cmd/gitvac

go run ./cmd/gitvac --scan

go run ./cmd/gitvac --root C:\src
```

To build a standalone binary:

```powershell
go build -o gitvac.exe ./cmd/gitvac
```

Run `gitvac --help` for all options.

## Config

Optional config file lives at `%AppData%\gitvac\config.json`:

```json
{
  "root": "C:\\Users\\reihan\\git",
  "targets": ["node_modules", ".venv", "venv", "__pycache__", ".pytest_cache", ".mypy_cache", "target", "dist", "build", ".next", ".svelte-kit", ".turbo"],
  "excludes": [".git", "vendor"]
}

```

_Passing `--root` via CLI overrides the root setting here._

## Keys

- `j`/`k` or Up/Down to navigate
- `Space` to toggle selection
- `a`/`n` to select all / select none
- `d` to delete selected (`y` to confirm, `n`/`Esc` to cancel)
- `q`/`Ctrl+c` exit
