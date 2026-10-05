package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsWithin(t *testing.T) {
	root := filepath.Join(t.TempDir(), "root")
	inside := filepath.Join(root, "project", "node_modules")
	outside := filepath.Join(filepath.Dir(root), "root-other")
	if !IsWithin(root, inside) {
		t.Fatal("expected child path to be inside root")
	}
	if IsWithin(root, outside) {
		t.Fatal("expected sibling path to be outside root")
	}
	if !IsWithin(root, root) {
		t.Fatal("expected root to be within itself")
	}
}

func TestIsTarget(t *testing.T) {
	targets := []string{"node_modules", ".venv"}
	if !IsTarget("NODE_MODULES", targets) {
		t.Fatal("matching should ignore case on Windows-style paths")
	}
	if IsTarget("vendor", targets) {
		t.Fatal("vendor should not match the default targets")
	}
	if IsTarget(filepath.Join("project", "node_modules"), targets) {
		t.Fatal("target matching should use a basename, not a path")
	}
}

func TestScanFindsAndSizesDirectories(t *testing.T) {
	root := t.TempDir()
	moduleDir := filepath.Join(root, "project", "node_modules")
	cacheDir := filepath.Join(root, "project", ".venv")
	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(moduleDir, "package.json"), []byte("12345"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "data"), []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}

	candidates, errs := Scan(root, []string{"node_modules", ".venv"}, nil)
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %v", errs)
	}
	if len(candidates) != 2 {
		t.Fatalf("got %d candidates, want 2", len(candidates))
	}
	if candidates[0].Name != "node_modules" || candidates[0].Size != 5 {
		t.Fatalf("unexpected largest candidate: %+v", candidates[0])
	}
	if candidates[1].Name != ".venv" || candidates[1].Size != 3 {
		t.Fatalf("unexpected candidate: %+v", candidates[1])
	}
}

func TestScanDoesNotReturnCandidatesInsideCandidates(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "project", "node_modules")
	nested := filepath.Join(parent, "target")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(parent, "package.json"), []byte("12345"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "cache"), []byte("abc"), 0o600); err != nil {
		t.Fatal(err)
	}

	candidates, errs := Scan(root, []string{"node_modules", "target"}, nil)
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %v", errs)
	}
	if len(candidates) != 1 || candidates[0].Name != "node_modules" || candidates[0].Size != 8 {
		t.Fatalf("unexpected candidates: %+v", candidates)
	}
}

func TestScanExcludesDiscoveryButCountsContainedBytes(t *testing.T) {
	root := t.TempDir()
	candidate := filepath.Join(root, "project", "node_modules")
	excludedCandidate := filepath.Join(root, ".git", "target")
	if err := os.MkdirAll(filepath.Join(candidate, ".git", "target"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(excludedCandidate, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(candidate, ".git", "data"), []byte("four"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(excludedCandidate, "data"), []byte("ignored"), 0o600); err != nil {
		t.Fatal(err)
	}

	candidates, errs := Scan(root, []string{"node_modules", "target"}, []string{".git"})
	if len(errs) != 0 {
		t.Fatalf("unexpected scan errors: %v", errs)
	}
	if len(candidates) != 1 || candidates[0].Name != "node_modules" || candidates[0].Size != 4 {
		t.Fatalf("unexpected candidates: %+v", candidates)
	}
}

func TestDeleteRefusesRootAndOutsidePaths(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := Delete(root, root); err == nil {
		t.Fatal("expected root deletion to be refused")
	}
	if err := Delete(root, outside); err == nil {
		t.Fatal("expected outside deletion to be refused")
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatalf("outside directory should remain: %v", err)
	}
}

func TestDeleteRemovesContainedCandidate(t *testing.T) {
	root := t.TempDir()
	candidate := filepath.Join(root, "project", "node_modules")
	if err := os.MkdirAll(candidate, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(candidate, "file"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Delete(root, candidate); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(candidate); !os.IsNotExist(err) {
		t.Fatalf("candidate still exists or unexpected stat error: %v", err)
	}
}
