
# LSM Storage Engine (Go)

An **LSM-style key–value storage engine** implemented in Go.

This engine provides **durable, ordered key–value storage** using an append-only
write-ahead log and an in-memory index. It implements the persistence, recovery,
and compaction mechanisms that form the core of modern LSM-based systems.

The design prioritizes **correctness, deterministic recovery, and explicit
durability guarantees**.

---

## Overview

The storage engine maintains state using two primary components:

- an **in-memory index (memtable)** containing the latest key–value mappings
- an **append-only write-ahead log (WAL)** providing durable persistence

All writes flow through the log before being reflected in memory, ensuring that
the on-disk state is always sufficient to recover the system.

---

## Architecture

```

Client Operations
↓
Memtable (in-memory index)
↓
Write-Ahead Log (append-only)
↓
Durable Storage (fsync-backed)

```

---

## Storage Guarantees

The engine provides the following guarantees:

- **Durability**  
  A write is acknowledged only after it has been persisted to disk via `fsync`.

- **Deterministic Recovery**  
  The complete state of the engine can be reconstructed by replaying the log.

- **Atomic Record Application**  
  Records are applied in full or not applied at all during recovery.

- **Monotonic Disk Writes**  
  On-disk data is never modified in place.

---

## Data Model

- Keys and values are arbitrary byte sequences
- The latest value for a key always supersedes earlier entries
- Ordering is derived from insertion and replay order

---

## Log Format

Each record in the write-ahead log is encoded using a fixed,
length-prefixed binary format:

```

[key_length][key][value_length][value]

```

- `key_length`   → 4 bytes (uint32)
- `key`          → raw bytes
- `value_length` → 4 bytes (uint32)
- `value`        → raw bytes

This format enables deterministic parsing and unambiguous record boundaries.

---

## Write Path

1. Serialize the record into binary format
2. Append the record to the write-ahead log
3. Issue `fsync` to ensure durability
4. Update the in-memory index
5. Acknowledge completion

```

SET(key, value)
→ append to log
→ fsync
→ memtable[key] = value

````

---

## Read Path

Reads are served directly from the in-memory index.  
The log is not consulted during normal reads.

---

## Recovery

On startup, the engine reconstructs its state by replaying the write-ahead log
from beginning to end. Each valid record updates the in-memory index.

Recovery is deterministic and requires no auxiliary metadata.

---

## Compaction

As updates accumulate, the log contains obsolete records.

Compaction rewrites the log by serializing the current in-memory state into a new
log file and atomically replacing the old one. This bounds disk usage while
preserving correctness.

---

## Interface

The engine exposes a minimal command interface:

```bash
go run main.go SET key value
go run main.go GET key
````

Each invocation performs a single operation.

---

## Implementation Notes

* Single-process execution model
* Single write-ahead log
* No background tasks
* All persistence is explicit and synchronous

---

## Roadmap

Planned extensions include:

* delete tombstones
* record checksums
* partial-write detection
* concurrent access
* background compaction
* multi-level SSTables

---

