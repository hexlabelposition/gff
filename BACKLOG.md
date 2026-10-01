# GFF — Go Fast Fetch

`gff` (**Go Fast Fetch**) is a CLI download manager written in Go.

The primary goal of the project is to learn Go through practice by gradually building a real CLI application while covering as many important language features and standard library capabilities as possible.

The project should start as a simple single-file downloader and gradually evolve into a full download manager with concurrent downloads, resume support, a download queue, configuration, and advanced diagnostics.

---

## Project Goals

### Primary Goal

Learn Go by building a practical application.

During development, the project should provide hands-on experience with:

- Go project structure;
- packages and modules;
- structs;
- methods;
- interfaces;
- pointers;
- error handling;
- custom error types;
- `defer`;
- `context.Context`;
- goroutines;
- channels;
- `sync.WaitGroup`;
- `sync.Mutex`;
- `sync.Once`;
- worker pools;
- HTTP clients;
- streaming I/O;
- file systems;
- JSON;
- configuration;
- generics;
- logging;
- Unix signals;
- graceful shutdown;
- unit testing;
- integration testing;
- benchmarks;
- profiling;
- build flags;
- cross-compilation.

---

## Basic CLI

Basic interface:

```bash
gff <url>
```

Example:

```bash
gff https://example.com/archive.zip
```

Example output from v0.2.0 onward (v0.1.0 reports completion without a progress bar):

```text
archive.zip
100% |████████████████████████████| 124 MB
Downloaded in 4.2s
```

---

## Development Order and Release Discipline

Implement phases in numerical order. Milestones are cumulative: each release keeps the functionality of earlier releases.

| Release | Phases | Result |
| --- | --- | --- |
| v0.1.0 — Fetch | 0–2 | Project setup, single-file streaming download, downloader API |
| v0.2.0 — Progress | 3 | Progress, size, speed, ETA |
| v0.3.0 — Reliable Fetch | 4–6 | Cancellation, partial files, resume, retries |
| v0.4.0 — Concurrent Fetch | 7–9 | Multiple downloads, worker pool, concurrency limits, race checks |
| v0.5.0 — Download Manager | 10–12 | Queue, persistent metadata, management commands |
| v0.6.0 — Fast Fetch | 13–14 | Parallel chunks and rate limiting |
| v0.7.0 — Production CLI | 15–23 | Configuration, logging, checksums, design review, broader tests, benchmarks, profiling |
| v1.0.0 | 24–28 | Recovery hardening, build metadata, platform builds, release automation, final architecture review |

The **MVP is v0.4.0**. A persistent queue and parallel chunks follow after the MVP.

### Checks Throughout Development

- Add focused unit tests and HTTP integration tests when the corresponding behavior is introduced, starting with v0.1.0. Phases 20–21 expand and review coverage rather than introduce testing for the first time.
- Run formatting, `go vet`, available lint checks, tests, and a build before each release.
- Start race checks when concurrency is introduced in v0.4.0; repeat them when shared state or parallel downloading changes.
- Add flags, help text, and error exit codes alongside the features that need them. Phase 12 introduces management subcommands; Phase 25 adds build metadata to version output.
- Create packages only when they contain useful code. The structure examples are guides, not a requirement to create empty directories. `go.sum` appears when dependencies require it.
- Mark backlog items complete only after implementing and checking them.

### Release Checklist — Repeat for Every Version

- [ ] Verify all phases and acceptance criteria for the target version.
- [ ] Run the applicable checks and manually try the documented CLI examples.
- [ ] Update `README.md` with supported behavior and usage.
- [ ] Record release changes in `CHANGELOG.md`.
- [ ] Set the version reported by `gff version` to the release version; use a development marker between releases.
- [ ] Commit the verified changes and create the matching Git tag, for example `v0.1.0`.
- [ ] Publish release binaries once release automation is available in Phase 27.

---

## Phase 0 — Bootstrap

Target release: **v0.1.0**.

### Project Initialization

- [x] Initialize a Go module.
- [x] Create the basic project structure.
- [x] Add `README.md`.
- [x] Add `.gitignore`.
- [x] Add `LICENSE`.
- [x] Set up formatting with `gofmt`.
- [x] Set up `go vet`.
- [x] Set up `golangci-lint`.
- [x] Add basic CI.

Initial structure (add other packages as features appear):

```text
gff/
├── cmd/
│   └── gff/
│       └── main.go
├── go.mod
├── README.md
├── BACKLOG.md
├── .gitignore
└── LICENSE
```

#### What We Learn

- `go mod`
- packages
- imports
- visibility
- `internal`
- `cmd`
- Go project conventions

---

## Phase 1 — Minimal Downloader

Target release: **v0.1.0**.

The goal is to implement the smallest functional single-file downloader.

### CLI

- [x] Accept a URL as a command-line argument.

```bash
gff https://example.com/file.zip
```

- [x] Check that a URL was provided.
- [x] Validate the URL.
- [x] Return a clear error for invalid URLs.
- [x] Print basic usage/help and return a nonzero exit code on failure.
- [ ] Add basic `gff version` output for release tracking.

### HTTP

- [ ] Perform an HTTP `GET` request.
- [ ] Check the HTTP status code.
- [ ] Support redirects.
- [ ] Read the response body as a stream.

### File I/O

- [ ] Determine the filename from the URL.
- [ ] Create the destination file.
- [ ] Stream the response directly into the file.

Do not load the entire file into memory.

Use streaming:

```text
HTTP response
      ↓
   io.Reader
      ↓
   io.Writer
      ↓
     file
```

### Errors

- [ ] Return errors correctly.
- [ ] Use error wrapping with `%w`.
- [ ] Add custom error types where appropriate.
- [ ] Use `errors.Is` when checking an underlying or sentinel error.
- [ ] Use `errors.As` when inspecting a typed error; defer this item if no real need exists.

#### What We Learn

- `net/http`
- `os`
- `io`
- `io.Copy`
- structs
- functions
- methods
- pointers
- `defer`
- error handling
- error wrapping

---

## Phase 2 — Downloader API

Target release: **v0.1.0**.

Move download logic out of `main`.

Example architecture:

```go
type Downloader struct {
    client *http.Client
}

func (d *Downloader) Download(
    request DownloadRequest,
) (DownloadResult, error)
```

Add `context.Context` to this API in Phase 4 when cancellation is introduced.

### Tasks

- [ ] Create `Downloader`.
- [ ] Pass dependencies through a constructor.
- [ ] Remove the HTTP client from global state.
- [ ] Implement a dedicated `DownloadRequest` type.
- [ ] Implement `DownloadResult`.

Example:

```go
type DownloadRequest struct {
    URL         string
    Destination string
}

type DownloadResult struct {
    Path       string
    Size       int64
    Duration   time.Duration
    StatusCode int
}
```

#### What We Learn

- structs
- constructors
- pointer receivers
- value receivers
- dependency injection
- package boundaries

---

## Phase 3 — Progress Bar

Target release: **v0.2.0**.

Add download progress reporting.

Example:

```text
ubuntu.iso
████████████████░░░░░░░░ 67%
1.8 GB / 2.7 GB
24 MB/s
ETA 38s
```

### Tasks

- [ ] Read `Content-Length`.
- [ ] Track the number of downloaded bytes.
- [ ] Calculate percentage.
- [ ] Calculate download speed.
- [ ] Calculate ETA.
- [ ] Handle unknown or zero content length without dividing by zero or inventing a percentage/ETA.
- [ ] Format file sizes:

```text
1024 B
1.4 KB
34.8 MB
2.1 GB
```

### Writer Wrapper

Create a custom `io.Writer`.

Example:

```go
type ProgressWriter struct {
    Writer  io.Writer
    Total   int64
    Written int64
}
```

#### What We Learn

- interfaces
- `io.Reader`
- `io.Writer`
- interface composition
- custom types
- methods
- time
- formatting

---

## Phase 4 — Context and Cancellation

Target release: **v0.3.0**.

Add the ability to stop a download.

```text
Ctrl+C
```

After cancellation, the application should:

1. cancel the HTTP request;
2. close the file;
3. preserve the incomplete file for resume support in Phase 5;
4. shut down cleanly.

### Tasks

- [ ] Introduce `.part` files, preserve them on cancellation, and rename them only after successful completion.
- [ ] Use `context.Context`.
- [ ] Use `context.WithCancel`.
- [ ] Handle `SIGINT`.
- [ ] Handle `SIGTERM`.
- [ ] Implement graceful shutdown.

#### What We Learn

- context
- cancellation
- `os.Signal`
- `signal.NotifyContext`
- graceful shutdown

---

## Phase 5 — Resume Downloads

Target release: **v0.3.0**.

Add support for resuming incomplete downloads.

```bash
gff https://example.com/file.iso
```

If a partial file exists:

```text
file.iso.part
```

GFF should attempt to resume the download.

### HTTP Range

Use:

```http
Range: bytes=1048576-
```

Inspect the advertised support:

```http
Accept-Ranges: bytes
```

Confirm support using the actual Range response; the header alone is not a guarantee.

### Tasks

- [ ] Reuse the `.part` file lifecycle introduced in Phase 4.
- [ ] Determine the current partial file size.
- [ ] Send a `Range` request.
- [ ] Verify `206 Partial Content` and the matching `Content-Range` offset.
- [ ] Handle `200 OK` by restarting safely rather than appending a full response.
- [ ] Handle `416 Range Not Satisfiable`.
- [ ] Save and check ETag or Last-Modified validators where available so a changed remote file is not appended to an old partial file.
- [ ] Continue writing from the correct position.
- [ ] Rename the `.part` file after successful completion.

#### What We Learn

- HTTP headers
- HTTP Range Requests
- random access files
- `Seek`
- file metadata

---

## Phase 6 — Retry System

Target release: **v0.3.0**.

Add automatic retry support.

```bash
gff --retries 5 https://example.com/file.zip
```

### Backoff

Example:

```text
retry 1 → 1 sec
retry 2 → 2 sec
retry 3 → 4 sec
retry 4 → 8 sec
```

### Tasks

- [ ] Implement a retry policy.
- [ ] Implement exponential backoff.
- [ ] Add a maximum delay.
- [ ] Add jitter.
- [ ] Avoid retrying non-retryable errors.
- [ ] Respect cancellation during requests and backoff waits.
- [ ] Reuse resume support so retries do not duplicate downloaded bytes.
- [ ] Add the `--retries` flag and help text.

#### What We Learn

- algorithms
- time
- error classification
- retry patterns

---

## Phase 7 — Concurrent Downloads

Target release: **v0.4.0**.

Add support for multiple URLs.

```bash
gff \
  https://example.com/a.zip \
  https://example.com/b.zip \
  https://example.com/c.zip
```

### Tasks

- [ ] Run downloads concurrently.
- [ ] Use goroutines.
- [ ] Wait for completion with `sync.WaitGroup`.
- [ ] Limit the number of simultaneous downloads.

Target interface, completed with URL-list file input in Phase 8:

```bash
gff --workers 4 urls.txt
```

#### What We Learn

- goroutines
- concurrency
- `sync.WaitGroup`
- synchronization

---

## Phase 8 — Worker Pool

Target release: **v0.4.0**.

Implement a job queue.

Architecture:

```text
                 ┌── worker 1
jobs channel ────┼── worker 2
                 ├── worker 3
                 └── worker 4
```

### Tasks

- [ ] Create `Job`.
- [ ] Create `Result`.
- [ ] Create a jobs channel.
- [ ] Create a results channel.
- [ ] Implement a worker pool.
- [ ] Add configurable worker count with `--workers`.
- [ ] Accept a text file containing one URL per line, as in `gff --workers 4 urls.txt`.
- [ ] Distinguish URL arguments from the URL-list file and document accepted inputs.

#### What We Learn

- channels
- buffered channels
- goroutines
- worker pools
- `select`
- synchronization

---

## Phase 9 — Thread Safety

Target release: **v0.4.0**.

Review shared state introduced by concurrent downloads and the worker pool. Repeat this review for queue state and parallel chunks.

### Tasks

- [ ] Use `sync.Mutex`.
- [ ] Use `sync.RWMutex` where appropriate.
- [ ] Run:

```bash
go test -race ./...
```

- [ ] Fix detected race conditions.

#### What We Learn

- race conditions
- mutexes
- atomic operations
- race detector

---

## Phase 10 — Download Queue

Target release: **v0.5.0**.

Create the queue model and transitions. Persistence and public management commands are completed in Phases 11–12 of this release.

The following CLI examples describe the completed v0.5.0 behavior.

```bash
gff add https://example.com/a.zip
gff add https://example.com/b.zip

gff list
```

Output:

```text
ID   STATUS       PROGRESS   FILE
1    paused       42%        a.zip
2    queued       0%         b.zip
```

### Statuses

```text
queued
downloading
paused
completed
failed
cancelled
```

### Tasks

- [ ] Create `Download`.
- [ ] Create `DownloadStatus`.
- [ ] Create a queue manager.
- [ ] Add unique IDs.
- [ ] Add timestamps.
- [ ] Add retry count.

#### What We Learn

- custom types
- constants
- state management
- enum-like values with `iota`

---

## Phase 11 — Metadata

Target release: **v0.5.0**.

Persist information about downloads.

Example:

```text
~/.local/share/gff/downloads.json
```

Data:

```json
{
  "id": "abc123",
  "url": "...",
  "file": "ubuntu.iso",
  "size": 4294967296,
  "downloaded": 2147483648,
  "status": "paused"
}
```

### Tasks

- [ ] Serialize state.
- [ ] Load state at startup.
- [ ] Restore the download queue.
- [ ] Handle corrupted state.
- [ ] Save state atomically to avoid partially written metadata.
- [ ] Restore interrupted downloads as paused or queued according to a documented policy.

#### What We Learn

- persistence
- JSON
- filesystem
- serialization

---

## Phase 12 — CLI Commands

Target release: **v0.5.0**.

Expand the CLI with queue management commands backed by persistent metadata.

For v0.5.0, commands run in the foreground: `add` saves a queued job; `resume <id>` starts a saved job; `list` reads persisted state. `pause <id>` and `cancel <id>` update inactive jobs; use Ctrl+C to pause the active foreground process. Control of a download running in another process requires a future daemon/IPC design and is outside this release.

```bash
gff download <url>
gff add <url>
gff list
gff pause <id>
gff resume <id>
gff cancel <id>
gff remove <id>
```

Additional commands:

```bash
gff retry <id>
gff clear
gff version
```

### Tasks

- [ ] Implement subcommands.
- [ ] Prevent conflicting processes from writing the same queue state or destination file.
- [ ] Add flags.
- [ ] Add help output.
- [ ] Preserve the basic version output introduced in Phase 1.
- [ ] Keep `gff <url>` working alongside `gff download <url>`.
- [ ] Document command behavior for queued, paused, completed, and cancelled jobs.
- [ ] Add exit codes.

#### What We Learn

- CLI architecture
- command routing
- flags
- exit codes

---

## Phase 13 — Parallel Chunk Downloading

Target release: **v0.6.0**.

Allow a single file to be downloaded using multiple connections.

Example file:

```text
0 MB ----------------------------- 100 MB
```

Split into:

```text
worker 1: 0-25 MB
worker 2: 25-50 MB
worker 3: 50-75 MB
worker 4: 75-100 MB
```

### CLI

```bash
gff --connections 4 https://example.com/file.iso
```

### Tasks

- [ ] Check Range support.
- [ ] Determine file size.
- [ ] Split the file into chunks.
- [ ] Download chunks concurrently.
- [ ] Write chunks at the correct offsets.
- [ ] Assemble the final file.

#### What We Learn

- concurrency
- Range requests
- mutexes
- concurrent file access
- synchronization

---

## Phase 14 — Rate Limiting

Target release: **v0.6.0**.

Add download speed limits.

```bash
gff --limit 5MB/s https://example.com/file.iso
```

### Tasks

- [ ] Limit read throughput.
- [ ] Add a global rate limit.
- [ ] Add per-download rate limits.
- [ ] Add the `--limit` flag and help text.
- [ ] Repeat race checks for chunk downloading and shared rate-limit state.

#### What We Learn

- rate limiting
- timers
- custom readers

---

## Phase 15 — Configuration

Target release: **v0.7.0**.

Add configuration support.

Example:

```text
~/.config/gff/config.json
```

```json
{
  "workers": 4,
  "connections": 4,
  "retries": 3,
  "downloadDir": "~/Downloads"
}
```

### Tasks

- [ ] Create `Config`.
- [ ] Read JSON configuration.
- [ ] Write JSON configuration.
- [ ] Implement defaults.
- [ ] Support CLI overrides.
- [ ] Read environment variable overrides.
- [ ] Add `gff config` for inspecting or updating configuration.

Priority:

```text
CLI arguments
      ↓
environment variables
      ↓
config file
      ↓
defaults
```

#### What We Learn

- `encoding/json`
- struct tags
- configuration patterns
- environment variables

---

## Phase 16 — Logging

Target release: **v0.7.0**.

Add structured logging.

Levels:

```text
DEBUG
INFO
WARN
ERROR
```

Example:

```text
INFO  download started url=...
WARN  retry attempt=2
ERROR request failed status=500
```

### Tasks

- [ ] Use `log/slog`.
- [ ] Add log levels.
- [ ] Add JSON output.

```bash
gff --log-format json
```

#### What We Learn

- structured logging
- `log/slog`

---

## Phase 17 — Checksums

Target release: **v0.7.0**.

Add integrity verification.

```bash
gff download https://example.com/file.zip \
  --sha256 7d865e...
```

Support:

```text
SHA256
SHA512
```

### Tasks

- [ ] Calculate hashes while downloading.
- [ ] Avoid rereading the file when possible.
- [ ] Compare checksums.

#### What We Learn

- `crypto/sha256`
- `crypto/sha512`
- `hash.Hash`
- `io.MultiWriter`

---

## Phase 18 — Interfaces

Target release: **v0.7.0**.

Extract useful abstractions.

Example:

```go
type Storage interface {
    Save(...)
    Open(...)
    Remove(...)
}
```

```go
type ProgressReporter interface {
    Start(...)
    Update(...)
    Finish(...)
}
```

```go
type DownloadStore interface {
    Save(...)
    Find(...)
    List(...)
}
```

### Goal

Understand in practice:

- when an interface is actually useful;
- when an interface only adds unnecessary complexity;
- why Go generally favors small interfaces.

#### What We Learn

- interfaces
- implicit implementation
- dependency inversion
- interface segregation

---

## Phase 19 — Generics

Target release: **v0.7.0**.

Use generics only where they provide real value.

Example:

```go
type Result[T any] struct {
    Value T
    Err   error
}
```

Or create a generic collection helper.

### Tasks

- [ ] Find a real use case.
- [ ] Implement a generic abstraction if a real use case exists; otherwise document the decision and revisit later.
- [ ] Compare generic and interface-based approaches.

#### What We Learn

- type parameters
- constraints
- `any`
- `comparable`

---

## Phase 20 — Testing

Target release: **v0.7.0**.

Expand and review tests added in earlier phases; cover edge cases and regressions.

### Unit Tests

Add tests for:

- [ ] filename parser;
- [ ] URL validation;
- [ ] byte formatter;
- [ ] progress calculator;
- [ ] chunk splitting;
- [ ] retry logic;
- [ ] configuration.

Use table-driven tests:

```go
tests := []struct {
    name  string
    input int64
    want  string
}{
    // ...
}
```

#### What We Learn

- `testing`
- table-driven tests
- test helpers
- subtests

---

## Phase 21 — HTTP Integration Tests

Target release: **v0.7.0**.

Expand the HTTP integration tests added alongside download, resume, and retry behavior.

Use:

```go
httptest.NewServer
```

Test:

- [ ] normal downloads;
- [ ] `404`;
- [ ] `500`;
- [ ] redirects;
- [ ] Range requests;
- [ ] resume;
- [ ] retry.

#### What We Learn

- `httptest`
- fake HTTP servers
- integration testing

---

## Phase 22 — Benchmarks

Target release: **v0.7.0**.

Add benchmarks.

```bash
go test -bench=. ./...
```

Benchmark:

- progress calculations;
- chunk splitting;
- metadata parsing;
- buffering.

#### What We Learn

- benchmarks
- allocations
- performance measurement

---

## Phase 23 — Profiling

Target release: **v0.7.0**.

Add profiling.

Use:

```text
pprof
```

Investigate:

- CPU usage;
- memory allocations;
- goroutines;
- blocking;
- mutex contention.

#### What We Learn

- `pprof`
- CPU profiling
- memory profiling
- performance optimization

---

## Phase 24 — Graceful Recovery

Target release: **v1.0.0**.

GFF should recover correctly from:

- application crashes;
- `Ctrl+C`;
- network loss;
- system reboots;
- temporary server failures.

After the next startup:

```bash
gff list
```

should show incomplete downloads.

#### What We Learn

- durability
- recovery
- state machines
- fault tolerance

---

## Phase 25 — Build System

Target release: **v1.0.0**.

Add a Makefile or Taskfile. Extend the existing `gff version` output with commit, build time, and Go version.

Example:

```bash
make build
make test
make lint
make bench
make run
```

Version command:

```bash
gff version
```

```text
gff 1.0.0-dev
commit: a18c4f2
built: 2026-10-01
go: go1.x
```

Inject values using:

```text
-ldflags
```

#### What We Learn

- Go linker flags
- build metadata
- reproducible builds

---

## Phase 26 — Cross Compilation

Target release: **v1.0.0**.

Build GFF for:

```text
linux/amd64
linux/arm64
windows/amd64
darwin/amd64
darwin/arm64
```

Example:

```bash
GOOS=linux GOARCH=amd64 go build
```

#### What We Learn

- `GOOS`
- `GOARCH`
- static binaries
- cross-compilation

---

## Phase 27 — CI/CD

Target release: **v1.0.0**.

GitHub Actions pipeline:

```text
push
  ↓
fmt
  ↓
vet
  ↓
lint
  ↓
test
  ↓
race
  ↓
build
```

Extend the basic CI from Phase 0 with the full build matrix and release pipeline.

For tags such as:

```text
v1.0.0
```

create release binaries.

### Tasks

- [ ] Add test workflow.
- [ ] Add lint workflow.
- [ ] Add build matrix.
- [ ] Add release workflow.
- [ ] Generate SHA256 checksums for release binaries.

---

## Phase 28 — Final Architecture

Target release: **v1.0.0**.

Review the accumulated architecture and refactor only where useful. Possible project structure:

```text
gff/
├── cmd/
│   └── gff/
│       └── main.go
│
├── internal/
│   ├── app/
│   ├── cli/
│   ├── config/
│   ├── downloader/
│   ├── queue/
│   ├── worker/
│   ├── progress/
│   ├── retry/
│   ├── storage/
│   ├── metadata/
│   └── logging/
│
├── pkg/
│
├── tests/
│
├── .github/
│   └── workflows/
│
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── BACKLOG.md
└── LICENSE
```

---

## Additional Tasks

After implementing the core version, consider adding the following features.

### Optional — Download Hooks

Add post-download actions.

Example:

```bash
gff download https://example.com/archive.zip \
  --exec "unzip archive.zip"
```

Implement this carefully and treat it as an advanced/optional feature.

#### What We Learn

- `os/exec`
- child processes
- command execution
- process lifecycle

---

### Optional — Observability

Add:

```bash
gff stats
```

Example output:

```text
Active downloads: 3
Queued downloads: 12
Downloaded today: 14.8 GB
Average speed: 32 MB/s
Failed downloads: 2
```

#### What We Learn

- aggregation
- metrics
- state inspection

---

### Network

- [ ] Proxy support.
- [ ] HTTP proxy.
- [ ] SOCKS proxy.
- [ ] Custom headers.
- [ ] User-Agent.
- [ ] Cookies.
- [ ] Basic Auth.
- [ ] Bearer Token.

### Download Management

- [ ] Download priorities.
- [ ] Pause all downloads.
- [ ] Resume all downloads.
- [ ] Retry failed downloads.
- [ ] Duplicate detection.
- [ ] Overwrite policy.

Example:

```bash
gff --overwrite
```

or:

```text
file.zip
file (1).zip
```

### Terminal UI

An interactive TUI can be added later:

```text
┌──────────────────────────────────────────────┐
│ GFF — Go Fast Fetch                         │
├──────────────────────────────────────────────┤
│ ubuntu.iso         ███████████░░  82%       │
│ archlinux.iso      █████░░░░░░░  41%       │
│ project.zip        queued                    │
└──────────────────────────────────────────────┘
```

This should be treated as a separate phase and should not be part of the MVP.

---

## MVP

The first practically useful version of `gff` should be able to:

- [ ] download a file from a URL;
- [ ] automatically determine the filename;
- [ ] display download progress;
- [ ] display download speed;
- [ ] display ETA;
- [ ] preserve incomplete downloads;
- [ ] resume downloads using HTTP Range;
- [ ] handle `Ctrl+C` correctly;
- [ ] process multiple downloads;
- [ ] limit concurrent downloads;
- [ ] retry temporary failures.

Example:

```bash
gff https://example.com/file.iso
```

and:

```bash
gff --workers 4 urls.txt
```

---

## Milestones

### v0.1.0 — Fetch

Scope: **Phases 0–2**.

Simple CLI downloader.

```text
URL
 ↓
HTTP
 ↓
File
```

Features:

- single URL;
- streaming download;
- filename detection;
- basic error handling;
- downloader API;
- basic help, version output, tests, and CI.

### v0.2.0 — Progress

Scope: **Phase 3**.

Add:

- progress bar;
- file size;
- download speed;
- ETA.

### v0.3.0 — Reliable Fetch

Scope: **Phases 4–6**.

Add:

- context;
- graceful shutdown;
- `.part` files;
- resume;
- retries.

### v0.4.0 — Concurrent Fetch

Scope: **Phases 7–9**.

Add:

- multiple files;
- goroutines;
- channels;
- worker pool;
- concurrency limits;
- URL-list file input;
- race checks.

This release completes the MVP.

### v0.5.0 — Download Manager

Scope: **Phases 10–12**.

Add:

- queue;
- download states;
- metadata;
- `list`;
- `pause`;
- `resume`;
- `cancel`.

### v0.6.0 — Fast Fetch

Scope: **Phases 13–14**.

Add:

- parallel chunks;
- multiple HTTP Range connections;
- configurable connection count;
- rate limiting.

At this stage, the name **Go Fast Fetch** starts describing not only the project name but also its functionality.

### v0.7.0 — Production CLI

Scope: **Phases 15–23**. Interface and generics reviews do not require adding abstractions without a useful purpose.

Add:

- configuration;
- structured logging;
- checksum validation;
- expanded unit and integration coverage;
- repeated race checks;
- benchmarks;
- profiling.

### v1.0.0

Scope: **Phases 24–28**.

First stable release.

Requirements:

- stable CLI;
- resume support;
- concurrent downloads;
- chunked downloads;
- download queue;
- retries;
- configuration;
- checksums;
- graceful shutdown;
- persistent state;
- tests;
- CI;
- binaries for major platforms.

---

## Go Learning Checklist

Track these topics as they naturally appear. Optional features such as generics, `panic/recover`, and specific synchronization primitives are not release requirements when no useful application exists.

By `v1.0.0`, review practical experience with the following Go features:

- [ ] Variables
- [ ] Constants
- [ ] Functions
- [ ] Multiple return values
- [ ] Named types
- [ ] Structs
- [ ] Methods
- [ ] Pointers
- [ ] Interfaces
- [ ] Interface composition
- [ ] Errors
- [ ] Error wrapping
- [ ] `errors.Is`
- [ ] `errors.As`
- [ ] `defer`
- [ ] `panic/recover`
- [ ] Slices
- [ ] Maps
- [ ] Generics
- [ ] Goroutines
- [ ] Channels
- [ ] `select`
- [ ] `sync.WaitGroup`
- [ ] `sync.Mutex`
- [ ] `sync.RWMutex`
- [ ] Atomics
- [ ] Context
- [ ] HTTP
- [ ] File I/O
- [ ] JSON
- [ ] Hashing
- [ ] Signals
- [ ] Testing
- [ ] Table-driven tests
- [ ] Integration tests
- [ ] Race detector
- [ ] Benchmarks
- [ ] Profiling
- [ ] Logging
- [ ] Cross-compilation
- [ ] Build metadata
- [ ] CI/CD

---

## Core Project Principle

Do not add a feature solely to demonstrate a particular Go language feature.

Every language feature should appear when it naturally solves a problem in the project.

For example:

```text
goroutines
```

should be introduced not because "we need to learn goroutines", but because GFF starts downloading multiple files concurrently.

```text
channels
```

should appear when a job queue and worker pool become necessary.

```text
context
```

should appear when HTTP requests need to support cancellation.

```text
interfaces
```

should be introduced when multiple implementations exist or when a component needs to be decoupled from a concrete dependency.

```text
generics
```

should only be introduced when a genuine reusable abstraction across types appears.

The goal is for `gff` to remain a well-designed Go project while also serving as a comprehensive hands-on environment for learning the language.
