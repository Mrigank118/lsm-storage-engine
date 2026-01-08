
# LSM Storage Engine (Go)

![Go](https://img.shields.io/badge/language-Go-blue)
![Tests](https://img.shields.io/badge/tests-passing-orange)
![Release](https://img.shields.io/github/v/release/Mrigank118/lsm-storage-engine?include_prereleases&label=release)
![License](https://img.shields.io/github/license/Mrigank118/lsm-storage-engine)

This repository implements a single-node **Log-Structured Merge (LSM) storage engine core** in Go.  
The engine is correctness-focused and models the fundamental persistence, recovery, and metadata invariants used by modern key–value storage engines. Performance, concurrency, and distribution are explicitly out of scope.

The design follows a strict rule: **disk is the source of truth; memory is disposable**. All in-memory state is derived and can be reconstructed deterministically from on-disk data structures after a crash or restart.

---

## Features

- Append-only **Write-Ahead Log (WAL)** with fsync-backed durability
- Deterministic crash recovery via WAL replay
- In-memory memtable for fast reads and writes
- Immutable, sorted **SSTables** generated via explicit flush
- Support for **multiple SSTables** with correct read precedence
- Explicit **MANIFEST** for authoritative metadata tracking
- Tombstone-based deletes to prevent key resurrection
- Append-only persistence model (no in-place mutation)
- Invariant-driven test suite validating correctness guarantees

---

## Architecture

All write operations are first recorded in an append-only **Write-Ahead Log (WAL)** and fsync’d before being applied to the in-memory table (memtable). The WAL is the authoritative history of mutations. During recovery, the WAL is replayed sequentially to rebuild the most recent state.

To bound WAL growth and improve read efficiency, the engine supports explicit flushing of the memtable into immutable, sorted on-disk tables (**SSTables**). SSTables are written sequentially, persisted atomically, and never modified in place. Multiple SSTables may coexist, with newer tables strictly shadowing older ones during reads.

Structural metadata is managed explicitly via a **MANIFEST** file. The MANIFEST is an append-only metadata log that records active SSTables and defines their logical ordering. Engine recovery relies exclusively on MANIFEST replay rather than filesystem inspection, ensuring deterministic startup behavior and providing the foundation for future compaction and WAL trimming.

---

## Storage Model

On-disk formats are deliberately minimal and correctness-oriented. WAL records are binary encoded as  
`[key_length][key][value_length][value][checksum]`, enabling detection of partial writes and corruption. Deletes are represented using tombstones (`value_length = 0`) to prevent resurrection during replay.

SSTables store sorted key–value records in a sequential binary layout and are treated as immutable snapshots. All persistence follows an append-only model; no data file is ever modified in place.

Reads consult the memtable first, followed by SSTables in newest-to-oldest order, guaranteeing that the most recent value always takes precedence.

---

## Scope and Non-Goals

This engine intentionally excludes SQL, networking, distributed replication, concurrency control, background compaction, and performance optimizations. These are higher-level concerns that depend on a correct and well-defined storage core and are deliberately deferred.

---

## Testing

Testing is invariant-driven rather than behavior-driven. Validated properties include durable writes, crash recovery, tombstone semantics, SSTable shadowing, MANIFEST replay correctness, and safe handling of orphaned SSTable files.

```bash
go test ./engine -v
````

---

## CLI Usage

A minimal command-line interface is provided as a **diagnostic layer** over the engine API. It exists solely to exercise the storage engine during development and carries no architectural significance.

Run the CLI from the project root:

```bash
go run ./cmd/lsm
```

Supported commands:

* `SET <key> <value>` — insert or update a key
* `GET <key>` — retrieve the current value
* `DEL <key>` — delete a key via tombstone
* `FLUSH` — flush memtable to a new SSTable
* `EXIT` — terminate the session

---

## Release Status

**Current Version:** `v0.2.0`
**Stability:** Pre-release (Alpha)

The engine is correctness-complete at the core level. Internal APIs and on-disk formats may evolve as compaction, WAL trimming, and indexing are introduced.

---
