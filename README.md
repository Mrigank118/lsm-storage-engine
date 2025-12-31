
# LSM Storage Engine (Go)

A **log-structured, append-only key–value storage engine** implemented in Go.

This project implements the **core storage mechanics** of an LSM-style system,
including durable write-ahead logging, deterministic crash recovery, tombstone-
based deletion semantics, and explicit persistence guarantees.

The engine is designed with a strict focus on **correctness, crash safety, and
clear invariants**, mirroring the foundational design principles used in systems
such as LevelDB, RocksDB, and other LSM-based storage engines.

---

## System Overview

The engine maintains state through a strict separation of concerns:

* a **volatile in-memory index (memtable)** holding the latest key state
* a **durable append-only write-ahead log (WAL)** serving as the source of truth

All mutations are persisted to disk **before** being reflected in memory.
At any point, the on-disk log alone is sufficient to reconstruct the complete
state of the system.

---

## Architecture

```
Client Commands
      ↓
In-Memory Index (Memtable)
      ↓
Append-Only Write-Ahead Log
      ↓
Stable Storage (fsync-backed)
```

---

## Core Invariants

The engine enforces the following invariants at all times:

* **Append-Only Persistence**
  On-disk data is never modified in place. All updates are recorded as new log
  entries.

* **Durability Before Acknowledgement**
  A write is acknowledged only after the operating system confirms persistence
  via `fsync`.

* **Deterministic Recovery**
  The engine state can always be reconstructed by replaying the log sequentially
  from the beginning.

* **Memory Is Non-Authoritative**
  The in-memory index is treated as a cache derived from disk and may be safely
  discarded and rebuilt.

---

## Data Model

* Keys and values are arbitrary byte sequences
* Updates to a key supersede all prior values
* Deletions are represented explicitly via tombstone records
* Record ordering is derived from log insertion order

---

## Log Encoding

Each record is serialized using a deterministic, length-prefixed binary layout:

```
[key_length][key][value_length][value][checksum]
```

* `key_length`   → 4 bytes (uint32)
* `key`          → raw bytes
* `value_length` → 4 bytes (uint32)
* `value`        → raw bytes (empty for tombstones)
* `checksum`     → CRC32 over `key || value`

---

## Write Path

The write path strictly enforces durability and ordering:

1. Serialize the record into binary form
2. Append the record to the write-ahead log
3. Issue `fsync` to ensure persistence
4. Apply the mutation to the in-memory index
5. Acknowledge completion

```
SET(key, value)
  → append record
  → fsync
  → memtable[key] = value
```

Delete operations are encoded as tombstone records and follow the same path.

---

## Read Path

Reads are served directly from the in-memory index.

The write-ahead log is never consulted during normal reads, ensuring low read
latency and simple execution semantics.

---

## Crash Recovery

On startup, the engine reconstructs state by sequentially replaying the
write-ahead log:

* records are parsed deterministically
* checksums are validated
* partial or corrupted trailing records are ignored safely
* later records supersede earlier ones
* tombstones remove keys from the reconstructed state

No auxiliary metadata, checkpoints, or snapshots are required.

---

## Compaction

As the log grows, it accumulates obsolete records.

Compaction rewrites the log by serializing the current in-memory state into a new
log file and atomically replacing the old one. This bounds disk usage while
preserving correctness and recovery guarantees.

---

## Interface

The engine exposes a minimal command-driven interface:

```bash
go run ./cmd/lsm SET key value
go run ./cmd/lsm GET key
go run ./cmd/lsm DEL key
```

Each invocation performs a single operation.

---

## Implementation Characteristics

* Single-process execution model
* Single append-only write-ahead log
* Explicit synchronous persistence (`fsync`)
* No background tasks or concurrency assumptions
* All correctness guarantees are enforced synchronously

---

## Roadmap

Planned extensions include:

* background compaction
* concurrent readers and writers
* sorted immutable SSTables
* multi-level LSM structure
* snapshotting and checkpoints
* performance optimizations

---

