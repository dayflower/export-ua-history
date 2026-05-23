# export-ua-history

`export-ua-history` is a Go CLI that exports Google Chrome browsing history as JSON for downstream analysis.

## Requirements

- Google Chrome
- Go if you want to install or build from source
- Permission for your terminal or parent process to access Chrome profile data

If file access permissions are not available, the command may fail when reading `Local State` or `History`.

## Installation

### Install with Homebrew

```sh
brew tap dayflower/tap
brew install export-ua-history
```

### Install from GitHub Releases

Download the archive for your operating system and architecture from the
[GitHub Releases](https://github.com/dayflower/export-ua-history/releases) page.

For macOS and Linux, extract the `.tar.gz` archive and move the binary somewhere
on your `PATH`:

```sh
tar -xzf export-ua-history_<version>_<os>_<arch>.tar.gz
sudo mv export-ua-history /usr/local/bin/
export-ua-history version
```

For Windows, extract the `.zip` archive and place `export-ua-history.exe` in a
directory where you want to run it.

### Install with `go install`

```sh
go install github.com/dayflower/export-ua-history/cmd/export-ua-history@latest
```

### Build from source

```sh
git clone https://github.com/dayflower/export-ua-history.git
cd export-ua-history
go build -o export-ua-history ./cmd/export-ua-history
```

### Run from a source checkout

```sh
go run ./cmd/export-ua-history help
```

## Usage

### Export today's history from the default profile

```sh
export-ua-history
```

Equivalent explicit form:

```sh
export-ua-history export
```

### Export a specific date

```sh
export-ua-history export --date 2026-05-22
```

### Export a specific profile to a file

```sh
export-ua-history export --profile "Personal" -o history.json
```

### List available Chrome profiles

```sh
export-ua-history list-profiles
```

### Print the current version

```sh
export-ua-history version
```

### Show help

```sh
export-ua-history help
export-ua-history help export
```

### Additional examples

Export a single hour from a day:

```sh
export-ua-history export --date 2026-05-22 --time 9
```

Export a local time range within one day:

```sh
export-ua-history export --date 2026-05-22 --start-time 12:00 --end-time 13:00
```

## Output

The tool writes pretty-printed UTF-8 JSON with a trailing newline.

The top-level payload includes:

- `browser`
- `start_date`
- `end_date`
- `total_entries`
- `entries`

Each entry includes:

- `timestamp`
- `url`
- `title`
- `visit_count`
- `domain`
- `browser`

## Limitations

- Google Chrome only
- No support for `--all-browsers`, `--db-path`, `--utc`, or `--tz`
- No support for browsers other than Chrome in the initial release

When Chrome keeps the live history database locked, the tool first tries to create a SQLite snapshot and then falls back to copying the database together with WAL / SHM sidecar files.

## Acknowledgements

Many thanks to [robzolkos/web-recap](https://github.com/robzolkos/web-recap) for the original inspiration behind this project.
Its README-level UX and JSON export shape were especially helpful reference points while designing this tool.
This repository is not a fork, and the implementation was written from scratch while adapting the idea to my own preferences and scope.

## License

MIT. See [LICENSE](LICENSE).
