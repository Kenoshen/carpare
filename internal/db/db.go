// Package db implements a tiny file-backed NoSQL document store.
// Each collection is a directory under the base path, and each document
// is a JSON file named "<id>.json" within it.
package db

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Store struct {
	baseDir string
	mu      sync.RWMutex
}

func Open(baseDir string) (*Store, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("db: creating base dir: %w", err)
	}
	return &Store{baseDir: baseDir}, nil
}

// NewID returns a random, filename-safe document ID.
func NewID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func (s *Store) collectionDir(collection string) string {
	return filepath.Join(s.baseDir, collection)
}

func (s *Store) docPath(collection, id string) string {
	return filepath.Join(s.collectionDir(collection), id+".json")
}

// Save writes v as the document identified by id within collection,
// creating the collection directory if needed. It overwrites any
// existing document with the same id.
func (s *Store) Save(collection, id string, v any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := s.collectionDir(collection)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("db: creating collection %q: %w", collection, err)
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("db: marshaling document: %w", err)
	}

	// Write to a temp file and rename so a crash mid-write never leaves
	// a partially-written document behind.
	final := s.docPath(collection, id)
	tmp := final + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("db: writing document: %w", err)
	}
	if err := os.Rename(tmp, final); err != nil {
		return fmt.Errorf("db: finalizing document: %w", err)
	}
	return nil
}

// Get reads the document identified by id within collection into v.
// It returns fs.ErrNotExist (via errors.Is) if the document is missing.
func (s *Store) Get(collection, id string, v any) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.docPath(collection, id))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// Delete removes the document identified by id within collection.
func (s *Store) Delete(collection, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := os.Remove(s.docPath(collection, id))
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}

// All loads every document in collection, in unspecified order, by
// unmarshaling each into a fresh T and invoking fn with its id.
func All[T any](s *Store, collection string, fn func(id string, doc T) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.collectionDir(collection))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("db: listing collection %q: %w", collection, err)
	}

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".json") {
			continue
		}
		id := strings.TrimSuffix(name, ".json")

		data, err := os.ReadFile(filepath.Join(s.collectionDir(collection), name))
		if err != nil {
			return fmt.Errorf("db: reading document %q: %w", id, err)
		}

		var doc T
		if err := json.Unmarshal(data, &doc); err != nil {
			return fmt.Errorf("db: unmarshaling document %q: %w", id, err)
		}
		if err := fn(id, doc); err != nil {
			return err
		}
	}
	return nil
}
