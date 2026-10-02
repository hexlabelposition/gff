# Changelog

## Unreleased

Progress reporting is planned for v0.2.0.

### Added

- Expected response size in download results.
- Track written bytes during downloads with a progress writer.
- Calculate download percentage when the expected size is positive.
- Calculate average download speed in bytes per second.
- Estimate remaining download time from average speed.
- Format byte sizes using binary units such as KiB, MiB, and GiB.
- Keep percentage and ETA undefined for unknown or zero response sizes.
- Report progress updates through an optional download callback.
- Display downloaded bytes, percentage, average speed, and ETA in the CLI.
- Omit percentage and ETA when the response size is unknown or zero.
- Display a progress bar for downloads with a known positive size.
- Update terminal progress on a single line.
- Test download progress with known, unknown, and zero response sizes.

## [0.1.0] - 2026-10-02

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
