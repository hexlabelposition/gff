# Changelog

## Unreleased

### Added

- Single-file HTTP/HTTPS downloads with streaming writes.
- Filename detection from the original URL with a fallback to `download`.
- Redirect support and protection against overwriting existing files.
- Version command.
- Typed HTTP status errors and wrapped errors for underlying failures.
- Downloader API with an injected HTTP client, an optional destination path,
  and results containing the path, byte count, duration, and HTTP status.
- Automated tests for downloads, HTTP 404 responses, redirects, existing files,
  and downloader API results.
- Basic CI for formatting, vet, lint, tests, and builds.
