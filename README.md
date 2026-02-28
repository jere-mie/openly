# openly

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A self-hosted, open-source URL shortener - a lightweight alternative to bit.ly. Single binary, no CGO, SQLite-backed, with built-in click analytics.

## Features

- **Single Binary** - templates, CSS, JS, and migrations are all embedded. One file to deploy.
- **No CGO** - uses [ncruces/go-sqlite3](https://github.com/nicholasgasior/gopher-sqlite3) (WASM-based) so it cross-compiles cleanly.
- **Click Analytics** - tracks clicks, referrers, user agents, and IP addresses.
- **Custom Short Codes** - choose your own short code or let one be generated.
- **Admin Dashboard** - password-protected dashboard to manage links and view stats.
- **Responsive UI** - clean editorial design that works on desktop and mobile.

## Getting Started

### Prerequisites

- Go 1.23.4+

### Setup

1. Clone the repository:
   ```sh
   git clone https://github.com/jere-mie/openly.git
   cd openly
   ```
2. Install dependencies:
   ```sh
   go mod tidy
   ```
3. Create your config:
   ```sh
   cp example.env .env
   ```
   Edit `.env` to set your admin password, port, and other options.

4. Run the application:
   ```sh
   go run .
   ```
5. Visit [http://localhost:8080](http://localhost:8080) (or whatever port you specified in `.env`).

### CLI Commands

```sh
# Run database migrations manually (also runs automatically on startup)
./openly migrate

# Print the current version
./openly version
```

### Building

Build for your current platform:

```sh
go build -o bin/openly .
```

Cross-compile for all supported platforms:

```powershell
# PowerShell
./scripts/build.ps1
```

```sh
# Bash
./scripts/build.sh
```

Binaries are output to the `bin/` directory.

### Releasing

The release scripts use the [GitHub CLI](https://cli.github.com/) to create a GitHub release from the version in `version.txt` and upload all binaries from `bin/`:

```powershell
# PowerShell
./scripts/release.ps1
```

```sh
# Bash
./scripts/release.sh
```

### Development with Air

You can use [Air](https://github.com/air-verse/air) for live reloading during development:

```sh
go install github.com/air-verse/air@latest
air
```

## Configuration

All configuration is done via environment variables (or a `.env` file):

| Variable | Default | Description |
|---|---|---|
| `ADMIN_PASSWORD` | `admin` | Password for the admin dashboard |
| `PORT` | `8080` | Server port |
| `HOST` | `localhost` | Server bind address |
| `DATABASE_PATH` | `openly.db` | Path to SQLite database file |
| `BASE_URL` | `http://localhost:8080` | Public base URL for generated short links |

## License

MIT - see [LICENSE](LICENSE) for details.

## Download a Release Binary

You can download a prebuilt binary directly from GitHub Releases without cloning the repo.

### Linux (amd64)

```sh
curl -Lo openly https://github.com/jere-mie/openly/releases/latest/download/openly_linux_amd64
chmod +x openly
```

### Linux (arm64)

```sh
curl -Lo openly https://github.com/jere-mie/openly/releases/latest/download/openly_linux_arm64
chmod +x openly
```

### macOS (Apple Silicon)

```sh
curl -Lo openly https://github.com/jere-mie/openly/releases/latest/download/openly_darwin_arm64
chmod +x openly
```

### macOS (Intel)

```sh
curl -Lo openly https://github.com/jere-mie/openly/releases/latest/download/openly_darwin_amd64
chmod +x openly
```

### Windows (PowerShell)

```powershell
Invoke-WebRequest -Uri "https://github.com/jere-mie/openly/releases/latest/download/openly_windows_amd64.exe" -OutFile "openly.exe"
```

### Available Binaries

| Platform | Architecture | Filename |
|---|---|---|
| Linux | amd64 | `openly_linux_amd64` |
| Linux | 386 | `openly_linux_386` |
| Linux | arm64 | `openly_linux_arm64` |
| Linux | arm | `openly_linux_arm` |
| macOS | amd64 | `openly_darwin_amd64` |
| macOS | arm64 | `openly_darwin_arm64` |
| Windows | amd64 | `openly_windows_amd64.exe` |
| Windows | 386 | `openly_windows_386.exe` |
| Windows | arm64 | `openly_windows_arm64.exe` |
