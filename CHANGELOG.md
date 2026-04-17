# Changelog

All notable changes to this project will be documented in this file.

The format is inspired by Keep a Changelog and this project follows Semantic Versioning.

## [Unreleased]

### Added
- Linux loop-device integration tests for write and verify workflows.
- Stronger disk identity support with by-id aliases and normalized serial/model matching.

### Changed
- URL downloads now support retry backoff and resume from partial `.partial` files when remote range requests are supported.
- Byte-sidecar fetch now retries transient network/server failures.

## [0.1.0] - 2026-04-17

### Added
- Initial terminal-native Go CLI scaffold with subcommands.
- End-to-end burn workflow for local or URL ISO sources.
- Checksum validation support with sidecar `.sha256` for URL sources.
- Removable disk discovery and safety confirmation checks.
- Raw write and post-write verification profiles (`quick`, `thorough`, `full`).
- Secure cleanup of temporary downloaded ISO files.
- Colorized progress output, Makefile, and initial unit tests.
