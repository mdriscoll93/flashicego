# Changelog

All notable changes to this project will be documented in this file.

The format is inspired by Keep a Changelog and this project follows Semantic Versioning.

## [0.1.0] - 2026-04-17

### Added
- Initial terminal-native Go CLI scaffold with subcommands.
- End-to-end burn workflow for local or URL ISO sources.
- Checksum validation support with sidecar `.sha256` for URL sources.
- Removable disk discovery and safety confirmation checks.
- Raw write and post-write verification profiles (`quick`, `thorough`, `full`).
- Secure cleanup of temporary downloaded ISO files.
- Colorized progress output, Makefile, and initial unit tests.
