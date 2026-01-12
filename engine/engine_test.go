package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetAndGet(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "store.log")

	db, err := New(logPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Set("config.max_connections", "100"); err != nil {
		t.Fatal(err)
	}

	val, err := db.Get("config.max_connections")
	if err != nil {
		t.Fatal(err)
	}

	if val != "100" {
		t.Fatalf("expected 100, got %s", val)
	}
}

func TestReplayAfterRestart(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "store.log")

	db, _ := New(logPath)
	db.Set("feature.flag.enabled", "true")

	// simulate restart
	db2, err := New(logPath)
	if err != nil {
		t.Fatal(err)
	}

	val, err := db2.Get("feature.flag.enabled")
	if err != nil {
		t.Fatal(err)
	}

	if val != "true" {
		t.Fatalf("expected true, got %s", val)
	}
}

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "store.log")

	db, _ := New(logPath)
	db.Set("cache.ttl", "60")
	db.Delete("cache.ttl")

	// restart
	db2, _ := New(logPath)

	if _, err := db2.Get("cache.ttl"); err == nil {
		t.Fatal("expected key to be deleted")
	}
}

func TestCompaction(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "store.log")

	db, _ := New(logPath)
	db.Set("service.timeout", "30")
	db.Set("service.timeout", "45")
	db.Set("service.retries", "3")
	db.Delete("service.timeout")

	if err := db.Compact(); err != nil {
		t.Fatal(err)
	}

	// restart after compaction
	db2, _ := New(logPath)

	if _, err := db2.Get("service.timeout"); err == nil {
		t.Fatal("expected key to be deleted")
	}

	val, err := db2.Get("service.retries")
	if err != nil || val != "3" {
		t.Fatal("compaction lost valid key")
	}
}

func TestMultipleSSTablesReadOrder(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "wal.log")

	e, _ := New(logPath)

	_ = e.Set("k", "v1")
	_ = e.Flush()

	_ = e.Set("k", "v2")
	_ = e.Flush()

	// new engine
	e2, _ := New(logPath)

	val, err := e2.Get("k")
	if err != nil {
		t.Fatal(err)
	}
	if val != "v2" {
		t.Fatalf("expected v2, got %s", val)
	}
}

func TestManifestPersistence(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "wal.log")

	db, _ := New(logPath)

	db.Set("a", "1")
	db.Flush()

	db.Set("b", "2")
	db.Flush()

	// restart
	db2, err := New(logPath)
	if err != nil {
		t.Fatal(err)
	}

	val, err := db2.Get("a")
	if err != nil || val != "1" {
		t.Fatalf("expected a=1, got %s", val)
	}

	val, err = db2.Get("b")
	if err != nil || val != "2" {
		t.Fatalf("expected b=2, got %s", val)
	}
}

func TestOrphanSSTableIgnored(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "wal.log")

	db, _ := New(logPath)

	db.Set("x", "1")
	db.Flush()

	// Manually create fake SSTable
	fake := filepath.Join(dir, "sstable-999999.db")
	if err := os.WriteFile(fake, []byte("junk"), 0644); err != nil {
		t.Fatal(err)
	}

	// restart
	db2, _ := New(logPath)

	val, err := db2.Get("x")
	if err != nil || val != "1" {
		t.Fatal("valid SSTable not read correctly")
	}

	if _, err := db2.Get("junk"); err == nil {
		t.Fatal("orphan SSTable should be ignored")
	}
}
