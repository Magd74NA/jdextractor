package jdextract

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Store provides filesystem CRUD for JSON-backed entities stored as
// {BasePath}/{slug}/meta.json.
type Store[T any] struct {
	BasePath string           // e.g. app.Paths.Jobs or app.Paths.Contacts
	PostRead func(*T)         // optional fixup after unmarshal (e.g. nil-slice init)
	SetDir   func(*T, string) // populates Dir field after reading from disk
}

// ValidID rejects IDs that could escape the base directory.
func ValidID(id string) bool {
	return id != "" && id != "." && !strings.Contains(id, "/") && !strings.Contains(id, "\\")
}

// ReadMeta reads and unmarshals meta.json from the given subdirectory.
func (s *Store[T]) ReadMeta(dir string) (*T, error) {
	path := filepath.Join(s.BasePath, dir, "meta.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entity T
	if err := json.Unmarshal(data, &entity); err != nil {
		return nil, err
	}
	if s.PostRead != nil {
		s.PostRead(&entity)
	}
	return &entity, nil
}

// WriteMeta marshals and writes meta.json to the given subdirectory using
// write-to-temp + rename for atomic updates. This prevents corrupted JSON
// if the process is interrupted mid-write.
func (s *Store[T]) WriteMeta(dir string, entity *T) error {
	data, err := json.MarshalIndent(entity, "", "\t")
	if err != nil {
		return err
	}
	destDir := filepath.Join(s.BasePath, dir)
	tmp, err := os.CreateTemp(destDir, "meta-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, filepath.Join(destDir, "meta.json"))
}

// List reads all subdirectories, parses each meta.json, and returns the
// slice with Dir set on each entity. Corrupt entries are skipped with a
// stderr warning. The returned slice is unsorted — callers apply their
// own sort.
func (s *Store[T]) List() ([]T, error) {
	entries, err := os.ReadDir(s.BasePath)
	if err != nil {
		return nil, fmt.Errorf("read directory %s: %w", s.BasePath, err)
	}
	var result []T
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		entity, err := s.ReadMeta(e.Name())
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", e.Name(), err)
			continue
		}
		s.SetDir(entity, e.Name())
		result = append(result, *entity)
	}
	return result, nil
}

// Get reads a single entity by exact directory name after validating the ID.
func (s *Store[T]) Get(id string) (*T, error) {
	if !ValidID(id) {
		return nil, fmt.Errorf("invalid id %q", id)
	}
	entity, err := s.ReadMeta(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("not found: %q", id)
		}
		return nil, fmt.Errorf("read meta.json: %w", err)
	}
	s.SetDir(entity, id)
	return entity, nil
}

// Delete removes a directory and all its contents after validating the ID.
func (s *Store[T]) Delete(id string) error {
	if !ValidID(id) {
		return fmt.Errorf("invalid id %q", id)
	}
	path := filepath.Join(s.BasePath, id)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("not found: %q", id)
	}
	return os.RemoveAll(path)
}

// FindByPrefix finds a directory by prefix, returning an error if there are
// zero or multiple matches.
func (s *Store[T]) FindByPrefix(prefix string) (string, error) {
	entries, err := os.ReadDir(s.BasePath)
	if err != nil {
		return "", fmt.Errorf("read directory: %w", err)
	}
	var matches []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			matches = append(matches, e.Name())
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no directory matches prefix %q", prefix)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("prefix %q is ambiguous: matches %s", prefix, strings.Join(matches, ", "))
	}
}

// MkDir creates a new entity directory under BasePath. If the directory
// already exists, it appends "col" as a collision suffix and retries once.
// Returns the directory name (not full path).
func (s *Store[T]) MkDir(slug string) (string, error) {
	dirPath := filepath.Join(s.BasePath, slug)
	err := os.Mkdir(dirPath, 0755)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			slug = slug + "col"
			dirPath = filepath.Join(s.BasePath, slug)
			if err := os.Mkdir(dirPath, 0755); err != nil {
				return "", err
			}
			return slug, nil
		}
		return "", err
	}
	return slug, nil
}
