
# LSM Storage Engine (Go)

A minimal **Log-Structured Merge (LSM)–style key–value storage engine** implemented in Go.

This project implements the **core storage mechanics** behind LSM-based systems,
focusing on **correctness, durability, and recoverability**.
The implementation is intentionally compact and explicit to make storage-engine
invariants clear and enforceable.

---

## Project Overview

This repository implements the **fundamental building blocks of an LSM-style
storage engine**, including:

* append-only persistence
* in-memory indexing
* crash recovery via log replay
* log compaction
* durability guarantees using `fsync`

The project is designed as a **storage-engine core**, not a full database system.

---

## Scope and Intent

This implementation focuses strictly on the **storage layer**.

It does not attempt to provide:

* SQL or query planning
* relational schemas
* transactions
* concurrency control
* distributed operation

These concerns belong to higher layers and are deliberately excluded to keep the
storage engine small, analyzable, and correct.

---

## High-Level Architecture

```
┌───────────────┐
│   Client API  │   SET / GET commands
└───────┬───────┘
        │
┌───────▼───────┐
│ In-Memory     │   Memtable (latest key → value)
│ Index         │
└───────┬───────┘
        │
┌───────▼───────┐
│ Append-Only   │   Binary log on disk
│ Log File     │
└───────┬───────┘
        │
┌───────▼───────┐
│ Persistent    │   Disk (fsync-backed)
│ Storage       │
└───────────────┘
```

---

## Design Principles

### Append-Only Storage

All writes are append-only.
Data is never modified in place.

This design simplifies crash recovery, avoids partial-write corruption, and
provides a complete history of updates until compaction.

---

### Explicit Separation of Disk and Memory

* **Disk** stores durable history
* **Memory** stores the current state

The in-memory index represents the authoritative view of the latest key–value
pairs, while the log serves as the durable source of truth.

---

### Crash Safety via Deterministic Replay

On startup, the engine reconstructs its state by sequentially replaying the log.
No additional metadata is required to restore correctness.

---

### Durability as a First-Class Invariant

Writes are acknowledged **only after `fsync` completes**.

Once a write returns successfully, it is guaranteed to survive process crashes
and power loss.

---

## Storage Model

### Binary Record Format

Each record in the log uses a fixed, length-prefixed binary format:

```
[key_length][key][value_length][value]
```

* `key_length`   → 4 bytes (uint32)
* `key`          → raw bytes
* `value_length` → 4 bytes (uint32)
* `value`        → raw bytes

This format allows deterministic parsing, efficient replay, and unambiguous
record boundaries.

---

## Write Path

1. Open log file in append mode
2. Serialize record into binary format
3. Append record to disk
4. Call `fsync` to guarantee durability
5. Update in-memory index
6. Acknowledge write

```
SET(key, value)
  → append record
  → fsync
  → index[key] = value
```

---

## Read Path

Reads are served directly from the in-memory index.

The log is not consulted during reads, keeping the read path simple and fast.

---

## Recovery Path

On startup:

1. Open the log file
2. Sequentially read records until EOF
3. Apply each record to the in-memory index
4. Resume normal operation

This process restores the exact state present at the last successful write.

---

## Compaction

As updates accumulate, the log contains obsolete records.

Compaction rewrites the log by retaining only the most recent value for each key.

### Compaction Steps

1. Create a new temporary log
2. Write the current in-memory state to it
3. Atomically replace the old log

This reduces disk usage without compromising correctness.

---

## Command Interface

The engine exposes a minimal command interface:

```bash
go run main.go SET key value
go run main.go GET key
```

Each invocation executes a single command.

---

## Design Boundaries and Future Updates

The current implementation deliberately limits scope to preserve correctness and
clarity in the storage core.

* concurrency control
* transactions
* delete tombstones
* record checksums
* partial-write validation
* background compaction
* multi-level SSTables


---

