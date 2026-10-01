# gff

`gff`(**go fast fetch**) is a CLI download manager written in Go as a practical learning project.
The first release, v0.1.0, focuses on downloading a single file over HTTP or HTTPS.
The current version is `0.1.0`.

## Requirements

- Go 1.27.0 or newer.
- Network access to the URL being downloaded.

## Build and Run

From the project root:

```bash
go build -o bin/gff ./cmd/gff
./bin/gff https://example.com/
./bin/gff version
```

The example saves the HTML page in a file named `download` in the current
working directory. To run without building a binary explicitly:

```bash
go run ./cmd/gff https://example.com/
```

Usage:

```text
gff <url>
gff version
```

## Download Behavior

- Accepts one HTTP or HTTPS URL.
- Follows HTTP redirects and requires a final `200 OK` response.
- Streams response bytes directly into a file without loading the whole response
  into memory.
- Determines the filename from the path of the original URL, even after a redirect.
- Uses `download` when the URL has no filename, such as `https://example.com/`.
- Saves the file in the current working directory and refuses to overwrite an
  existing file.
- Reports the saved filename and number of bytes on success.
- Prints errors to stderr and exits with a nonzero status on failure.

## Current Limitations

Progress reporting, resume support, retries, concurrent downloads, and queue
management are planned in [BACKLOG.md](BACKLOG.md).

A failed or interrupted download can leave an incomplete destination file.
The current version cannot resume it and will refuse to overwrite it on the next
attempt. Remove the incomplete file before trying again.

The CLI does not yet provide a destination flag or configurable request timeouts.

## Development Checks

Install the same linter version used by CI:

```bash
curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b ./bin v2.14.0
```

The installation places the linter in `bin/`, which is excluded from Git.
Install it again when setting up a new checkout.

Run checks from the project root:

```bash
gofmt -l .
go vet ./...
go test ./...
./bin/golangci-lint run ./...
go build ./...
```

`gofmt -l .` lists files that need formatting; an empty output means formatting
is correct. Format changed Go files with `gofmt -w <path>`.

Tests use local HTTP servers and temporary directories. They cover successful
downloads, HTTP 404 responses, redirects, preservation of existing files, and
the downloader API's destination and result fields.

GitHub Actions runs formatting, vet, lint, tests, and a build on pushes and pull
requests.

## Project Structure

- `cmd/gff/`: CLI entry point and CLI integration tests.
- `internal/downloader/`: downloader API, HTTP status error type, and API tests.
- `BACKLOG.md`: development phases and release milestones.
- `CHANGELOG.md`: release changes.

## License

See [LICENSE](LICENSE).
