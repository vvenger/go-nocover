# go-nocover

A CLI tool that filters Go coverage profiles by removing blocks that are not meaningful to test:
blocks annotated with `//nocover:block`, error-logging blocks, and `if err != nil { return }` blocks.

## Installation

```bash
go install nocover@latest
```

## Usage

```bash
go test -cover -coverprofile=coverage.out -coverpkg=./... ./...
# Filter and write to a new file
nocover -coverage=coverage.out -output=coverage.filtered.out
# Code coverage tool
gocover-cobertura < coverage.filtered.out > coverage.xml
diff-cover coverage.xml --compare-branch=origin/dev
```

### Flags

| Flag        | Default               | Description                                                     |
| ----------- | --------------------- | --------------------------------------------------------------- |
| `-coverage` | `.`                   | Path to the input coverage profile                              |
| `-output`   | (overwrite input)     | Path to the output file                                         |
| `-root`     | `.`                   | Project root — directory containing `go.mod` and `nocover.yaml` |
| `-config`   | `<root>/nocover.yaml` | Path to config file                                             |

## Code annotation

Mark a function or block with `//nocover:block` to exclude it from the coverage profile regardless of config:

```go
func (s *Service) runMigrations() error { //nocover:block
    if err := s.db.Migrate(); err != nil {
        return err
    }
    return nil
}
```

The comment can be placed on the same line as the opening `{` or on the line above:

```go
//nocover:block
if condition {
    ...
}
```

## Configuration (nocover.yaml)

Place `nocover.yaml` in the project root next to `go.mod`. If the file is absent, only `//nocover:block` annotations are applied.

```yaml
# How to handle blocks annotated with //nocover:block.
# deleted — remove from the profile (default)
# tested  — keep in the profile with count=1 (marked as covered)
mark-no-cover: deleted

# How to handle blocks excluded by config options (exclude-err-nil, exclude-err-regexp, etc.).
# deleted — remove from the profile
# tested  — keep in the profile with count=1 (marked as covered, default)
mark-options: tested

# Exclude logging statement or `if err != nil` blocks whose body contains only logging calls
# matching one of the given regexps.
exclude-log-regexp:
  - \.Info\(
  - \.Error\(
  - \.logger\.

# Exclude ALL `if err != nil { return ... }` blocks where `return` is a single statement.
exclude-err-nil: false

# Exclude `if err != nil { return ... }` blocks only when the error was produced
# by a call whose source text matches one of the given regexps.
exclude-err-regexp:
  - json\.Marshal\(
  - json\.Unmarshal\(

# Exclude `if err != nil { return ... }` blocks only when the error was produced
# by a method call whose receiver type matches the pattern.
# Uses go/types for matching — the variable name does not matter.
exclude-err-method:
  - (pgx.Rows) Scan
  - (pgx.Rows) Err
  - (pgxpool.Pool) Query
```

### exclude-err-nil example

Excludes any `if err != nil` block where the body is a single `return` statement (with or without values):

```go
// excluded
if err != nil {
    return err
}

// NOT excluded — has else branch
if err != nil {
    return err
} else {
    return nil
}

// NOT excluded — body has extra statements
if err != nil {
    log.Error("failed")  
    return err
}
```

> When used together with `exclude-log-regexp`, log calls inside the block are considered already excluded, so a block like `if err != nil { log.Error(...); return err }` is also excluded.

---

### exclude-log-regexp example

The pattern (a standard Go regexp) is matched against the source text of the log call expression.

**Single log statement in a mixed function** — only the statement itself is excluded:
```go
// matches \.Info\(
func process() {
    log.Info("start")  // excluded
    doWork()
}
```

**If block with only log calls** — the entire `if` block is excluded:
```go
// matches \.Info\(
if debug {
    log.Info("debug mode")  // entire if block excluded
}
```

**Function body with only log calls** — the entire function is excluded:
```go
// matches \.Info\(
func logStart() {
    log.Info("start")  // entire function excluded
}
```

---

### exclude-err-regexp example

The pattern (a standard Go regexp) is matched against the source text of the call expression that assigned `err`. 

**Init statement:**
```go
// matches json\.Marshal\(
if err := json.Marshal(v); err != nil {
    return err
}
```

**Preceding statement:**
```go
// matches json\.Marshal\(
err = json.Marshal(v)
if err != nil {
    return err
}
```

---

## TODO

## Configuration (nocover.yaml)

```yaml
# Exclude `if err != nil { return ... }` blocks only when the error was produced
# by a method call whose receiver type matches the pattern.
# Uses go/types for matching — the variable name does not matter.
exclude-err-method:
  - (pgx.Rows) Scan
  - (pgx.Rows) Err
  - (pgxpool.Pool) Query
```

### exclude-err-method example

Pattern format: `(pkg.Type) Method`. Matching is done via `go/types` — the variable name is irrelevant:

```yaml
exclude-err-method:
  - (pgx.Rows) Scan
```

```go
// r, result, rows — any name works, as long as the type is pgx.Rows
r, _ := pool.Query(ctx, sql)
err = r.Scan(&v)       // matches
if err != nil {
    return err
}

if err := rows.Scan(&v); err != nil {  // also matches
    return err
}
```

Unlike `exclude-err-regexp`, the type is checked statically — `(pgx.Rows) Scan` will not match a `Scan` call on any other type.
