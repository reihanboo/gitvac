package scanner

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Candidate struct {
	Path string
	Rel  string
	Name string
	Size int64
}

func Scan(root string, targets, excludes []string) ([]Candidate, []error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, []error{err}
	}
	absolute = filepath.Clean(absolute)
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return nil, []error{err}
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return nil, []error{err}
	}
	if !info.IsDir() {
		return nil, []error{fmt.Errorf("scan root is not a directory: %s", root)}
	}

	targetSet := nameSet(targets)
	excludeSet := nameSet(excludes)
	var candidates []Candidate
	var active []int
	var excluded []string
	var scanErrors []error

	err = filepath.WalkDir(resolved, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if !errors.Is(walkErr, os.ErrNotExist) {
				scanErrors = append(scanErrors, walkErr)
			}
			return nil
		}
		if path == resolved {
			return nil
		}

		for len(active) > 0 && !isWithin(candidates[active[len(active)-1]].Path, path) {
			active = active[:len(active)-1]
		}
		for len(excluded) > 0 && !isWithin(excluded[len(excluded)-1], path) {
			excluded = excluded[:len(excluded)-1]
		}

		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if entry.IsDir() {
			name := entry.Name()
			if containsName(excludeSet, name) {
				if len(active) == 0 {
					return filepath.SkipDir
				}
				excluded = append(excluded, path)
				return nil
			}
			if len(excluded) > 0 {
				return nil
			}
			if containsName(targetSet, name) {
				rel, relErr := filepath.Rel(resolved, path)
				if relErr != nil {
					scanErrors = append(scanErrors, relErr)
					return filepath.SkipDir
				}
				candidates = append(candidates, Candidate{Path: path, Rel: rel, Name: name})
				active = append(active, len(candidates)-1)
			}
			return nil
		}

		info, infoErr := entry.Info()
		if infoErr != nil {
			if !errors.Is(infoErr, os.ErrNotExist) {
				scanErrors = append(scanErrors, infoErr)
			}
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		for _, candidateIndex := range active {
			candidates[candidateIndex].Size += info.Size()
		}
		return nil
	})
	if err != nil {
		scanErrors = append(scanErrors, err)
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Size == candidates[j].Size {
			return strings.ToLower(candidates[i].Rel) < strings.ToLower(candidates[j].Rel)
		}
		return candidates[i].Size > candidates[j].Size
	})
	return candidates, scanErrors
}

func IsWithin(root, path string) bool {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	return isWithin(filepath.Clean(rootAbs), filepath.Clean(pathAbs))
}

func IsStrictlyWithin(root, path string) bool {
	return IsWithin(root, path) && !samePath(root, path)
}

func IsTarget(name string, targets []string) bool {
	return containsName(nameSet(targets), name)
}

func Delete(root, candidate string) error {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	candidateAbs, err := filepath.Abs(candidate)
	if err != nil {
		return err
	}
	rootAbs = filepath.Clean(rootAbs)
	candidateAbs = filepath.Clean(candidateAbs)
	resolvedRoot, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return fmt.Errorf("resolve root: %w", err)
	}
	if !IsStrictlyWithin(rootAbs, candidateAbs) && !IsStrictlyWithin(resolvedRoot, candidateAbs) {
		return fmt.Errorf("refusing to delete path outside root: %s", candidate)
	}
	resolvedCandidate, err := filepath.EvalSymlinks(candidateAbs)
	if err != nil {
		return fmt.Errorf("resolve candidate: %w", err)
	}
	if !IsStrictlyWithin(resolvedRoot, resolvedCandidate) {
		return fmt.Errorf("refusing to delete path outside root: %s", candidate)
	}
	info, err := os.Lstat(candidateAbs)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to delete non-directory or link: %s", candidate)
	}
	return os.RemoveAll(candidateAbs)
}

func nameSet(names []string) map[string]struct{} {
	set := make(map[string]struct{}, len(names))
	for _, name := range names {
		if name != "" && filepath.Base(name) == name {
			set[strings.ToLower(name)] = struct{}{}
		}
	}
	return set
}

func isWithin(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel))
}

func samePath(first, second string) bool {
	firstAbs, err := filepath.Abs(first)
	if err != nil {
		return false
	}
	secondAbs, err := filepath.Abs(second)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(filepath.Clean(firstAbs), filepath.Clean(secondAbs))
	return err == nil && rel == "."
}

func containsName(set map[string]struct{}, name string) bool {
	_, ok := set[strings.ToLower(name)]
	return ok
}
