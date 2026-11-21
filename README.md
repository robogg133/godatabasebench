# godatabasebench

A reproducible Go benchmark comparing four database access paths over the **same**
PostgreSQL instance and the **same** dataset:

| Variant | Stack |
| --- | --- |
| `stdlib` | `database/sql` + `github.com/jackc/pgx/v5/stdlib` |
| `pgxpool` | `github.com/jackc/pgx/v5/pgxpool` (native pool) |
| `ent` | `entgo.io/ent` ORM on top of pgx |
| `sqlc` | `sqlc` generated, type-safe code on top of pgxpool |

The goal is to answer, with numbers instead of opinions:

- Pool/driver overhead: `pgx` stdlib vs native `pgxpool`.
- ORM/query-layer overhead: `ent` vs `sqlc`.
- Behaviour under increasing concurrency and result-set sizes.
- Where the CPU, memory, and lock hotspots actually are (via `pprof`).

## Scope and fairness

All variants run against one PostgreSQL container with one schema and one seeded
dataset. Queries are kept semantically equivalent, connection pool settings are
identical, and timings are collected with Go's `testing` package plus `benchstat`.

See [`PLAN.md`](PLAN.md) for the full methodology, fairness rules, and roadmap.

## Benchmark scenarios

Reading:

- `PointReadByID` — single row by primary key
- `PointReadByEmail` — single row by unique index
- `ListPostsByUser` — filtered + paginated list
- `ListPostsWithAuthor` — join
- `AggregationCountComments` — grouped count
- `SearchPostsByTag` — many-to-many join

Writing:

- `InsertUser`
- `InsertPostWithTags` (transactional multi-statement)
- `UpdateUser`
- `DeletePost`

Transactions / batches:

- `TxCreatePostAndComment`
- `BatchInsertUsers` (N = 100 / 1000)

Each scenario is exposed as `BenchmarkXxx(b *testing.B)` in every variant with
the **same name**, so results line up directly in `benchstat`.

## Metrics

Per scenario (from `go test -bench`):

- `ns/op` — average latency per operation
- `B/op`, `allocs/op` — memory pressure (`-benchmem`)
- custom `b.ReportMetric` values such as `rows/op` and pool wait stats

Resource-level (outside the benchmark loop):

- CPU / heap / goroutine / block / mutex profiles (`pprof`)
- execution traces (`go tool trace`)
- `pgxpool.Stat()` per scenario
- PostgreSQL `pg_stat_statements` and `pg_stat_database`
- container CPU/mem via `docker stats`

## Requirements

- Go (version pinned in `go.mod`; tested with Go 1.23+)
- Docker + Docker Compose (for PostgreSQL)
- `benchstat` (optional, for comparison): `go install golang.org/x/perf/cmd/benchstat@latest`
- `sqlc` and `ent` code generators (only needed when regenerating; generated code is committed)

## Benchmark ranking (PointReadByID, ns/op / allocs)
> CPU: Intel(R) Core(TM) i5-5200U CPU @ 2.20GHz

1. pgxpool — ≈120 µs, 11 allocs
2. sqlc — ≈122 µs, 11 allocs
3. stdlib — ≈125 µs, 25 allocs
4. ent — ≈333 µs, 114 allocs

## Quick start

```sh
# 1. Start PostgreSQL
make up

# 2. Create schema + seed deterministic data
make seed

# 3. Run all benchmarks
make bench

# 4. Run with profiles
make profile
```

Results are written under `results/` (human-readable tables) and `profiles/`
(gitignored, pprof artifacts).

## Running tests

```sh
GOTOOLCHAIN=go1.25.0 go test ./... -run '^$' -bench . -benchmem -count=10
```

Ent and scenarios can be run individually:

```sh
GOTOOLCHAIN=go1.25.0 go test ./variants/pgxpool/... -run '^$' -bench . -benchmem
```

## Common commands

```sh
make up          # start Postgres container
make down        # stop and remove container
make seed        # (re)seed the database
make gen         # regenerate ent + sqlc code
make bench       # run all variants
make bench-stdlib
make bench-pgxpool
make bench-ent
make bench-sqlc
make compare     # benchstat across variants
make profile     # run with CPU + heap profiles
make pprof       # serve last profile on :8081
make clean       # remove profiles/results
```

## Profiling with pprof

The optional pprof HTTP server is enabled with `PPROF=1`:

```sh
PPROF=1 go test ./variants/pgxpool/... -run '^$' \
  -bench 'BenchmarkPointReadByID' -benchmem \
  -cpuprofile cpu.prof -memprofile mem.prof

go tool pprof -http=:8081 cpu.prof
```

Block and mutex profiles require explicit sampling rates (set in the variant
setup when `PPROF=1`); collect them with `-blockprofile` / `-mutexprofile`.

For concurrency-focused investigations also capture a trace:

```sh
go test ./variants/pgxpool/... -run '^$' \
  -bench 'BenchmarkPointReadByID' -trace trace.out
go tool trace trace.out
```

## Repository layout

```
.
├── PLAN.md
├── README.md
├── go.mod
├── docker-compose.yml
├── Makefile
├── schema/        # init.sql
├── seed/          # deterministic data generator
├── internal/
│   ├── model/     # shared structs
│   ├── queries/    # shared SQL / sqlc input
│   └── bench/     # pool setup, timers, metrics helpers
├── variants/
│   ├── stdlib/    # database/sql + pgx stdlib
│   ├── pgxpool/   # pgxpool
│   ├── ent/       # ent ORM
│   └── sqlc/      # sqlc generated code
├── profiles/      # gitignored pprof artifacts
└── results/       # committed comparison reports
```

## Methodology notes

- Each scenario is run in isolation (one process per variant when possible) to
  reduce GC and cache interference.
- Order of variants is randomized across the full suite.
- The first run of each scenario is treated as warm-up and discarded.
- `GOMAXPROCS` and `GOMEMLIMIT` are fixed and documented per run.
- Conclusions require `benchstat` p-values below the configured threshold, not a
  single run.
- Concurrency is varied via `b.RunParallel` and explicit worker counts
  (1 / 8 / 64 / 256).
- Result-set size is varied via sub-benchmarks (small / medium / large).

## Configuration

Environment variables (all optional, with sensible defaults):

| Var | Description | Default |
| --- | --- | --- |
| `DATABASE_URL` | Postgres DSN | `postgres://bench:bench@localhost:5432/bench?sslmode=disable` |
| `MAX_CONNS` | pool max connections | `16` |
| `MIN_CONNS` | pool min connections | `4` |
| `PPROF` | enable pprof HTTP server (`1`) | unset |
| `PPROF_ADDR` | pprof listen address | `:6060` |

## Reproducibility

- Pin Go, PostgreSQL, and library versions (see `go.mod` and
  `docker-compose.yml`).
- Seed is deterministic (fixed RNG seed + fixed cardinalities) — document the
  resulting DB size in `results/`.
- Report host CPU, memory, and kernel in `results/` alongside numbers.
- Profiles that support a conclusion should be saved under `profiles/` with a
  `<variant>-<scenario>-<timestamp>.prof` name.

## Contributing

1. Keep scenarios semantically equivalent across variants; any change to a query
   in one variant must be mirrored in the others.
2. Regenerate `ent` / `sqlc` code with `make gen` and commit the output.
3. Do not commit files under `profiles/`.
4. Include `benchstat` output in pull requests that claim a performance change.

## License

MIT — see `LICENSE`. (To be added.)
