# export-ua-history Specification

## 1. Document Status

- Status: Draft for implementation
- Language: English
- Scope: Initial release
- Last updated: 2026-05-23

## 2. Goal

`export-ua-history` is a Go CLI that exports Google Chrome browsing history as JSON for downstream analysis.

The tool is intentionally narrow in scope:

- Browser support is limited to Google Chrome.
- Output format and most CLI behaviors follow the current `robzolkos/web-recap` README as the compatibility baseline.
- Deviations from `web-recap` are explicitly defined in this document.

## 3. Compatibility Baseline

The baseline reference is the `web-recap` README as observed on 2026-05-22:

- Repository: <https://github.com/robzolkos/web-recap>
- JSON output section: <https://github.com/robzolkos/web-recap#json-output-format>

This project is not a full fork or clone of `web-recap`. It only preserves the parts required by the product requirements for a Chrome-only exporter.

## 4. Non-Goals

The initial release does not support:

- Browsers other than Google Chrome
- Automatic extraction from all browsers
- Custom database path selection
- User-selectable timezone interpretation
- UTC output mode
- The `list` command from `web-recap`
- Any daemon, GUI, or long-running background process

## 5. Runtime and Dependency Policy

### 5.1 Language and runtime

- The implementation must use Go.
- Target Go version should be a current stable release supported by the repository at implementation time.

### 5.2 Dependency policy

- Prefer the Go standard library wherever possible.
- Avoid third-party Go packages unless there is a strong justification.
- SQLite access must use `ncruces/go-sqlite3`.

### 5.3 SQLite library requirement

- The implementation must use `github.com/ncruces/go-sqlite3`.
- The implementation may also use companion packages from the same module family when required for backup or driver integration.
- The implementation must not depend on the platform `sqlite3` CLI executable at runtime.

## 6. Supported Data Source

### 6.1 Chrome user data directory

The default Chrome user data directory depends on the operating system:

- macOS: `~/Library/Application Support/Google/Chrome`
- Linux: `$XDG_CONFIG_HOME/google-chrome` or `~/.config/google-chrome`
- Windows: `%LOCALAPPDATA%\Google\Chrome\User Data`

### 6.2 Profile metadata source

The tool must read profile metadata from the selected platform's Chrome user data directory:

- `<user data dir>/Local State`

The `Local State` file is JSON. The implementation must inspect:

- `profile.info_cache`

Each key under `profile.info_cache` is treated as a profile path, for example:

- `Default`
- `Profile 1`
- `Profile 2`

Each profile entry's human-readable display name is taken from:

- `profile.info_cache.<profile_path>.name`

### 6.3 History database location

For a selected profile, the browsing history database path is:

- `<user data dir>/<profile path>/History`

## 7. Commands and CLI Behavior

### 7.1 Commands

The CLI exposes these commands:

- `export`: export history
- `list-profiles`: list Chrome profiles
- `version`: print version information

The CLI does not expose the `list` command from `web-recap`.

### 7.2 `export` command default behavior

Running the `export` command with no date flags:

```sh
export-ua-history export
```

must export history for the current local calendar day from the `Default` profile and write JSON to stdout.

For compatibility, running the root command without a subcommand:

```sh
export-ua-history
```

must export history for the current local calendar day from the `Default` profile and write JSON to stdout.

### 7.3 Supported options

The `export` command must support:

- `--browser <browser>`
- `--profile <profile name>`
- `--profile-path <profile path>`
- `--date <YYYY-MM-DD>`
- `--start-date <YYYY-MM-DD>`
- `--end-date <YYYY-MM-DD>`
- `--start-time <HH:MM>`
- `--end-time <HH:MM>`
- `--time <0-23>`
- `-o, --output <file>`
- `--help`

The root command without a subcommand must accept the same option set and behave identically to `export`.

### 7.4 Browser option semantics

- `--browser` exists for compatibility with `web-recap`.
- The only accepted value is `chrome`.
- If omitted, the effective browser is `chrome`.
- Any other value must produce an error.

### 7.5 Unsupported options

The program must reject the following `web-recap` options with a clear error:

- `--all-browsers`
- `--db-path`
- `--utc`
- `--tz`

### 7.6 Profile selection semantics

#### `--profile <profile name>`

- Matches the display name from `Local State`.
- Matching is exact and case-sensitive.
- If no profile matches, the command fails.
- If multiple profile paths share the same display name, the command fails with an ambiguity error and instructs the user to use `--profile-path`.

#### `--profile-path <profile path>`

- Matches the path key under `profile.info_cache`.
- Examples: `Default`, `Profile 1`
- Matching is exact and case-sensitive.
- The value is not a filesystem absolute path.

#### Mutual exclusion

- `--profile` and `--profile-path` are mutually exclusive.
- Supplying both must fail.

#### Default profile

- If neither `--profile` nor `--profile-path` is supplied, the selected profile is `Default`.
- If `Default` does not exist, the command fails.

### 7.7 Date and time flag semantics

#### Date modes

The tool supports two date modes:

- Single-day mode: `--date`
- Date-range mode: `--start-date` and `--end-date`

Rules:

- `--date` is mutually exclusive with `--start-date` and `--end-date`.
- `--start-date` and `--end-date` must be supplied together.
- If none of `--date`, `--start-date`, or `--end-date` are supplied, the tool behaves as if `--date <today>` was provided in the local timezone.

#### Time-of-day filters

- `--time <0-23>` is shorthand for one local hour within a single day.
- `--start-time` and `--end-time` define a local time range within a single day.
- `--time` is mutually exclusive with `--start-time` and `--end-time`.
- `--time`, `--start-time`, and `--end-time` may only be used together with `--date`.
- Time filtering is not supported with multi-day date ranges in the initial release.

#### Range interpretation

- All user-supplied dates and times are interpreted in the local system timezone.
- `--date 2026-05-22` means the local interval from `2026-05-22T00:00:00` through the start of `2026-05-23T00:00:00`.
- `--start-date` and `--end-date` form an inclusive date range in local time.
- Queries must use a half-open interval: `start <= timestamp < end`.
- The JSON `end_date` field is a display boundary, not the internal exclusive query boundary.
- For a whole-day export, `end_date` must be rendered as the final second of the requested local interval, for example `2026-05-22T23:59:59+09:00`.
- For a timed range such as `--date 2026-05-22 --start-time 12:00 --end-time 13:00`, the JSON `end_date` must be `2026-05-22T12:59:59+09:00`.

#### Validation

- Dates must use `YYYY-MM-DD`.
- Times must use `HH:MM` in 24-hour format.
- `--time` must be an integer from `0` to `23`.
- `--start-time` must be strictly earlier than `--end-time`.

### 7.8 Output destination

- Default output is stdout.
- `-o, --output <file>` writes the JSON payload to the specified file.
- File creation truncates any existing file.

### 7.9 `list-profiles` command

`list-profiles` prints the available Chrome profiles to stdout.

Each output row must contain:

- profile display name
- profile path

Recommended human-readable format:

```text
<profile name>\t<profile path>
```

Example:

```text
Personal	Default
Work	Profile 1
```

If no profiles are found, the command must fail with a clear error.

### 7.10 `version` command

`version` prints version information to stdout as a single line:

```text
export-ua-history <version>
```

Default development builds may use `dev` as the version string.

## 8. History Extraction Behavior

### 8.1 Database access strategy

Chrome keeps the `History` SQLite database open while the browser is running. The tool must not query the live database file directly for normal operation.

Instead, it must:

1. Resolve the selected profile's `History` database.
2. Create a temporary snapshot database.
3. Query the snapshot.
4. Remove temporary files before exit.

The preferred snapshot method is:

- use `VACUUM INTO` or equivalent safe snapshot logic via `ncruces/go-sqlite3`

This keeps database access in-process and avoids reliance on external executables.

Recommended behavior for the initial release:

1. Attempt to create the temporary snapshot database using `VACUUM INTO`.
2. If that attempt fails because the database is locked, fall back to copying:
   - `History`
   - `History-wal`, if present
   - `History-shm`, if present
3. Query the copied snapshot set instead of the live browser database.

Rationale:

- Chrome often keeps the history database locked while the browser is running.
- Copying the main database together with WAL/SHM sidecar files provides a pragmatic fallback for the initial release when direct snapshot creation cannot proceed.

### 8.2 Query shape

The tool must extract visit-level rows by joining `visits` and `urls`.

Required fields:

- visit timestamp from `visits.visit_time`
- URL from `urls.url`
- title from `urls.title`
- visit count from `urls.visit_count`

Recommended SQL shape:

```sql
SELECT
  v.visit_time,
  u.url,
  COALESCE(u.title, ''),
  u.visit_count
FROM visits AS v
JOIN urls AS u ON u.id = v.url
WHERE v.visit_time >= ? AND v.visit_time < ?
ORDER BY v.visit_time ASC;
```

### 8.3 Timestamp conversion

Chrome stores timestamps as microseconds since `1601-01-01 00:00:00 UTC`.

The implementation must:

1. Convert the local user-requested time range to UTC.
2. Convert UTC bounds into Chrome's microsecond epoch for SQL filtering.
3. Convert each result timestamp back into local time for JSON output.

Output timestamps must include the local offset, for example:

- `2026-05-22T09:15:23+09:00`

### 8.4 Domain extraction

For each entry:

- Parse the URL with Go's `net/url`.
- Set `domain` to `URL.Hostname()`.
- Preserve an empty string if parsing fails.

## 9. Output JSON Format

### 9.1 Top-level structure

The output format follows `web-recap` with one intentional change: omit the `timezone` field.

```json
{
  "browser": "chrome",
  "start_date": "2026-05-22T00:00:00+09:00",
  "end_date": "2026-05-22T23:59:59+09:00",
  "total_entries": 2,
  "entries": [
    {
      "timestamp": "2026-05-22T09:15:23+09:00",
      "url": "https://example.com/page",
      "title": "Example Page Title",
      "visit_count": 3,
      "domain": "example.com",
      "browser": "chrome"
    }
  ]
}
```

### 9.2 Field definitions

Top-level fields:

- `browser`: always `chrome`
- `start_date`: requested report start in local time with numeric offset
- `end_date`: requested report end in local time with numeric offset, rendered as an inclusive display boundary
- `total_entries`: number of exported visit rows
- `entries`: array of exported visit rows

Entry fields:

- `timestamp`: visit time in local time with numeric offset
- `url`: visited URL
- `title`: page title, or empty string if absent
- `visit_count`: URL-level visit count from Chrome's `urls` table
- `domain`: extracted hostname
- `browser`: always `chrome`

### 9.3 Encoding details

- JSON must be UTF-8.
- JSON must be pretty-printed with stable indentation for readability.
- The output must end with a trailing newline.

## 10. Error Handling

All user-facing errors must be printed to stderr and the process must exit non-zero.

Required error cases include:

- unsupported browser value
- mutually exclusive flags used together
- incomplete date-range flags
- invalid date or time format
- invalid `--time` value
- selected profile not found
- ambiguous profile name match
- `Default` profile missing when no profile flag is supplied
- `Local State` file missing or unreadable
- `History` database missing
- snapshot creation failure
- output file creation failure

### 10.1 Permission guidance

If the program cannot access Chrome's configuration directory or files, it must:

- print the exact path that could not be accessed
- print the underlying OS error
- print a short actionable hint, for example that the invoking terminal or parent process may not have sufficient permission to access the browser profile directory

## 11. Exit Codes

- `0`: success
- non-zero: any validation, discovery, access, snapshot, query, or output error

The exact non-zero numeric values do not need to be differentiated in the initial release.

## 12. Recommended Internal Package Structure

The implementation should remain simple. A recommended layout is:

- `cmd/export-ua-history/main.go`
- `internal/browser`
- `internal/browser/platform`
- `internal/browser/chrome`
- `internal/cli`
- `internal/output`

Responsibilities:

- `internal/browser`: shared browser-facing types
- `internal/browser/platform`: platform and environment primitives shared across browsers
- `internal/browser/chrome`: Chrome-specific profile discovery and history extraction
- `internal/cli`: command resolution, flag parsing, and validation
- `internal/output`: JSON serialization

## 13. Testing Strategy

The implementation should include unit tests for:

- date parsing and range calculation
- local-time to Chrome-epoch conversion
- Chrome-epoch to local-time conversion
- profile resolution by name
- profile resolution by path
- ambiguity detection for duplicate profile names
- URL hostname extraction
- JSON serialization shape
- CLI validation rules

Integration-style tests should cover:

- `list-profiles` against fixture `Local State` JSON
- history export against a fixture SQLite database
- snapshot command failure handling

Fixtures should be synthetic and must not include real browsing history.

## 14. Future Extension Points

The initial design should avoid blocking future additions such as:

- Chromium-based browser variants
- machine-readable profile listing mode
- optional JSON schema versioning

## 15. References

- `web-recap` repository: <https://github.com/robzolkos/web-recap>
- `web-recap` README JSON format: <https://github.com/robzolkos/web-recap#json-output-format>
- Chromium user data directory documentation: <https://chromium.googlesource.com/chromium/src/+/d4afc97b7/docs/user_data_dir.md>
- Chromium example `Local State` structure: <https://chromium.googlesource.com/chromium/src.git/+/39.0.2164.0/chrome/test/data/diagnostics/user/Local%20State>
