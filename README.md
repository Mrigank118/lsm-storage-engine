
# LSM Storage Engine (Go)

A **log-structured, append-only key–value storage engine** implemented in Go.

This project implements the **core storage engine primitives** used in LSM-based
systems: durable write-ahead logging, deterministic crash recovery,
checksum-validated records, tombstone-based deletion, and crash-safe log
compaction.

The design prioritizes **correctness, durability, and explicit invariants**.

---

## Core Invariants

* **Append-Only Persistence**
  On-disk data is never modified in place.

* **Durability Before Acknowledgement**
  Writes are acknowledged only after `fsync`.

* **Deterministic Recovery**
  Engine state is reconstructed solely by replaying the log.

* **Memory Is Non-Authoritative**
  The in-memory index is a derived cache and may be rebuilt at any time.

* **Crash-Safe Compaction**
  Compaction rewrites state into a new log and replaces the old log atomically.

---

## Data Model

* Arbitrary byte keys and values
* Latest record for a key supersedes all prior entries
* Deletions are represented via **tombstone records**
* Ordering is derived from log insertion order

---

## Log Format

Each record is encoded as:

```
[key_length][key][value_length][value][checksum]
```

* `key_length`   → uint32
* `value_length` → uint32 (0 indicates tombstone)
* `checksum`     → CRC32 over `key || value`

This enables deterministic parsing and detection of partial or corrupted writes.

---

## Operations

### Write Path

```
SET / DEL
→ append record
→ fsync
→ update memtable
```

### Read Path

Reads are served directly from the in-memory index.

### Recovery

On startup, the log is replayed sequentially.
Corrupted or partial trailing records are safely ignored.

### Compaction

Compaction rewrites the current in-memory state into a new log and atomically
replaces the existing log, bounding disk usage and recovery time.

---

## Command Interface

```bash
go run ./cmd/lsm SET <key> <value>
go run ./cmd/lsm GET <key>
go run ./cmd/lsm DEL <key>
go run ./cmd/lsm COMPACT
```

---

## Testing

Correctness-focused tests validate durability, crash recovery, tombstones, and
compaction safety.

Run tests from the project root:

```bash
go test ./engine
```

---

## Versioning

Current version:

```
v0.1.0
```

This release represents a **correctness-complete single-node storage engine
core**.

---


