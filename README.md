# Prospect

A desktop application for working with binary protobuf files.

## Requirements

- Go 1.21+
- C compiler (gcc) for Windows (required for Fyne)
- protoc (Protocol Buffer compiler)

## Building

```bash
$env:CGO_ENABLED=1; go build -o prospect.exe ./cmd/prospect
```

## Running

```bash
$env:CGO_ENABLED=1; go run ./cmd/prospect
```


---

This application was created with vibe code.
