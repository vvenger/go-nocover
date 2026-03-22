# go-nocover

A CLI tool that filters Go coverage profiles by removing blocks that are not meaningful to test:
blocks annotated with `//nocover:block`, error-logging blocks, and `if err != nil { return }` blocks.

## Installation

```bash
go install nocover@latest
```

## Usage

```bash
# Filter and write to a new file
nocover -coverage=coverage.out -output=coverage.filtered.out

# Filter in place (overwrite the input file)
nocover -coverage=coverage.out

# Project is not in the current directory
nocover -coverage=coverage.out -root=/path/to/project
```

### Flags

| Flag        | Default           | Description                                                     |
| ----------- | ----------------- | --------------------------------------------------------------- |
| `-coverage` | `.`               | Path to the input coverage profile                              |
| `-output`   | (overwrite input) | Path to the output file                                         |
| `-root`     | `.`               | Project root — directory containing `go.mod` and `nocover.yaml` |

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
# Exclude `if err != nil` blocks whose body contains only logging calls
# matching one of the given regexps.
exclude-log-regexp:
  - \.Info\(
  - \.Error\(
  - \.logger\.

# Exclude ALL `if err != nil { return ... }` blocks.
# When true, exclude-err-regexp and exclude-err-method are ignored.
exclude-errnil: false

# Exclude `if err != nil { return ... }` blocks only when the error was produced
# by a call whose source text matches one of the given regexps.
# Ignored when exclude-errnil: true.
exclude-err-regexp:
  - json\.Marshal\(
  - json\.Unmarshal\(

# Exclude `if err != nil { return ... }` blocks only when the error was produced
# by a method call whose receiver type matches the pattern.
# Uses go/types for matching — the variable name does not matter.
# Ignored when exclude-errnil: true.
exclude-err-method:
  - (pgx.Rows) Scan
  - (pgx.Rows) Err
  - (pgxpool.Pool) Query
```

### Priority rules

| Condition                                      | Behaviour                                                                                                     |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `nocover.yaml` absent                          | Only `//nocover:block` annotations apply                                                                      |
| `exclude-log-regexp` not set                   | Log blocks are not filtered                                                                                   |
| `exclude-errnil: true`                         | All `if err != nil { return }` blocks are excluded; `exclude-err-regexp` and `exclude-err-method` are ignored |
| `exclude-errnil: false` + `exclude-err-regexp` | Excludes blocks where the call text matches a pattern                                                         |
| `exclude-errnil: false` + `exclude-err-method` | Excludes blocks where the receiver type and method match (via `go/types`)                                     |

### How exclude-err-regexp works

The pattern (a standard Go regexp) is matched against the source text of the call expression that assigned `err`. Two cases are supported:

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

### How exclude-err-method works

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
