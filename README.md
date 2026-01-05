# Prospect

A desktop application for working with binary protobuf files.

## Description

Prospect is a tool that allows you to open, parse, and view binary Protocol Buffer files. It provides a tree view of the protobuf message structure and allows exporting the parsed data to JSON format.

## Requirements

- Go 1.21+
- C compiler (gcc) for Windows (required for Fyne)
- protoc (Protocol Buffer compiler)

## Installation

```bash
go mod tidy
```

## Building

```bash
$env:CGO_ENABLED=1; go build -o prospect.exe ./cmd/prospect
```

## Running

```bash
$env:CGO_ENABLED=1; go run ./cmd/prospect
```

## Installing protoc

On Windows, you can install protoc via:

```bash
scoop install protobuf
```

---

This application was created with vibe code.
