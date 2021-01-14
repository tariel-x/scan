# scan

Scan is the frontend for SANE built with golang and svelte.

![Scan with SANE test devices](screenshot.jpeg)

Features:

- List connected scanners.
- Configure the selected scanner.
- Scan image from the selected scanner.

## Build & run

Install libsane-dev, e.g. for Ubuntu:

```shell
sudo apt install -y libsane-dev
```

Build and run service:

```bash
go build -o scan ./cmd/scan
LISTEN=0.0.0.0:8085 ./scan
```

## Environment variables

- `DEBUG` enables debug log level and text format for logs.
- `LISTEN` is the interface to listen for.