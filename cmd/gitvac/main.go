package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gitvac/internal/config"
	"gitvac/internal/scanner"
	"gitvac/internal/tui"

	tea "charm.land/bubbletea/v2"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "gitvac:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("gitvac", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	root := flags.String("root", "", "directory to scan")
	scanOnly := flags.Bool("scan", false, "print candidates without opening the TUI")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	if *root == "" {
		*root = cfg.Root
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected argument: %s", flags.Arg(0))
	}
	absoluteRoot, err := filepath.Abs(*root)
	if err != nil {
		return err
	}
	if *scanOnly {
		candidates, errs := scanner.Scan(absoluteRoot, cfg.Targets, cfg.Excludes)
		for _, candidate := range candidates {
			fmt.Printf("%8s  %-16s  %s\n", sizeString(candidate.Size), candidate.Name, candidate.Rel)
		}
		for _, scanErr := range errs {
			fmt.Fprintln(os.Stderr, "warning:", scanErr)
		}
		return nil
	}

	program := tea.NewProgram(tui.New(absoluteRoot, cfg.Targets, cfg.Excludes))
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run TUI: %w", err)
	}
	return nil
}

func sizeString(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	units := []string{"KiB", "MiB", "GiB", "TiB", "PiB"}
	value := float64(bytes)
	for _, unit := range units {
		value /= 1024
		if value < 1024 || unit == units[len(units)-1] {
			return fmt.Sprintf("%.1f %s", value, unit)
		}
	}
	return fmt.Sprintf("%d B", bytes)
}
