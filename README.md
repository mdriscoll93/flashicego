# flashicego

`flashicego` is a terminal-native USB image writer built in Go.

It is designed for scriptable workflows where you can pass either a local ISO path or a remote URL.

Inspiration for this project comes from the venerable, though predominantly GUI-based, Balena Etcher and its ability to burn an image based on a provided URL.

Current release: `v0.1.0`

## Features

- Auto-download ISO from URL
- Resume partial downloads and retry transient network/server failures
- SHA256 checksum verification (sidecar `.sha256` for URLs)
- Confirm target disk identity by serial/model
- Display `/dev/disk/by-id` alias for stronger target confirmation
- Raw image write to removable media
- Post-write verification profiles (`quick`, `thorough`, `full`)
- Secure shredding of temporary downloaded ISOs after successful verify
- Colorized terminal output and progress bars

## Safety Model

`burn` only targets removable disks discovered through Linux sysfs and refuses mounted targets.

By default, burn requires explicit interactive confirmation by typing the target device path.

Identity checks are enforced by requiring at least one of `--serial` or `--model`.

## Install

```bash
go build -o flashicego .
```

## Usage

List candidate removable devices:

```bash
flashicego list-disks
```

Burn local ISO:

```bash
sudo flashicego burn ./ubuntu.iso /dev/sdb --serial ABC123 --verify-profile quick
```

Burn URL and keep downloaded ISO at a chosen path:

```bash
sudo flashicego burn https://example.com/os.iso /dev/sdb --model "SanDisk Ultra" --store-iso ~/Downloads/os.iso
```

Burn URL and allow unsigned image (not recommended):

```bash
sudo flashicego burn https://example.com/os.iso /dev/sdb --serial ABC123 --allow-unsigned
```

Standalone verify:

```bash
sudo flashicego verify ./ubuntu.iso /dev/sdb --profile thorough
```

## Notes

- Linux-focused implementation.
- Raw device writes generally require root privileges.
- Secure deletion on modern filesystems is best effort.

## Testing

Run unit tests:

```bash
make test
```

Run loop-device integration tests (requires root and loop tools):

```bash
sudo make integration-test
```

## Versioning

This project follows Semantic Versioning.

- Current version: `v0.1.0` (see `VERSION`)
- Release notes: `CHANGELOG.md`
