package engine

import (
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
