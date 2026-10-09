# droppederr

[![CI](https://github.com/Orfeo42/droppederr/actions/workflows/ci.yml/badge.svg)](https://github.com/Orfeo42/droppederr/actions/workflows/ci.yml)

A Go analyzer that reports error-handling sites where the root cause never reaches the logs.

## What it reports

`DROPPED` — an `if err != nil { ... }` block whose body never references `err`, so the cause is discarded before any boundary can log it.

```go
if err != nil { // DROPPED: doWork drops the root error: err unused in return errNotFound
	return errNotFound
}
```

Every finding is always reported; there is no baseline.

## Install

```sh
go install github.com/Orfeo42/droppederr/cmd/droppederr@latest
```

## Run

```sh
droppederr ./...
go vet -vettool=$(which droppederr) ./...
```

## golangci-lint module plugin

Build a custom golangci-lint with this module compiled in, using a `.custom-gcl.yml`:

```yaml
version: v2.14.0
plugins:
  - module: github.com/Orfeo42/droppederr
    import: github.com/Orfeo42/droppederr/plugin
    version: latest
```

Then enable it in `.golangci.yml`:

```yaml
linters:
  enable:
    - droppederr
  settings:
    custom:
      droppederr:
        type: module
        description: reports error-handling sites where the root cause never reaches the logs
        settings:
          exclude-paths: internal/dao/,_mock.go,_test.go
```

The plugin reads syntax only, so it declares `LoadModeSyntax` and costs no type-checking.

## Settings

| setting | default | meaning |
| --- | --- | --- |
| `exclude-paths` | — | Comma-separated path fragments; a file whose slash-normalized path contains one is never reported. |

The standalone binary exposes the same setting as the `-exclude-paths` flag.

## Origin

Split from github.com/Orfeo42/nilfield at commit 01c7880, where it shipped together with the SQLCLASS rule (now github.com/Orfeo42/sqlclass).

## License

MIT, see [LICENSE](LICENSE).
