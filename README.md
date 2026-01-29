# Stack Foundry

**Lean Architecture. High-Throughput Systems. Applied Intelligence.**

- Live site: <https://stackfoundry.co.uk>

This is the official homepage for **StackFoundry Ltd**, a UK-based software consulting company specializing in cutting through the chaos with rigorous coding.

## The Architecture

We reject unnecessary complexity. This site is a **Self-Contained System (SCS)** built to demonstrate the power of modern, server-driven architectures.

- **Go (1.23+):** The core logic. Fast, typed, and compiled to a single static binary.
- **Templ:** Type-safe HTML templating. No runtime parsing errors.
- **htmx:** Progressive enhancement for dynamic UI interactions with minimal client-side JavaScript.
- **Tailwind CSS:** Utility-first styling, embedded directly into the binary.
- **AWS Lambda (ARM64):** Deployed as a single function with zero idle costs.

## Getting Started

### Prerequisites

- [Go](https://go.dev/) (1.24+)
- [pnpm](https://pnpm.io/) (for Tailwind)
- [xc](https://github.com/joerdav/xc) (Task runner)
- [AWS CDK v2](https://docs.aws.amazon.com/cdk/v2/guide/home.html) and Node.js (for infra IaC)
- [AWS CLI](https://aws.amazon.com/cli/) configured with appropriate credentials

### Development

1. Install dependencies: `pnpm install` & `go mod download`
2. Install tools: `go install github.com/a-h/templ/cmd/templ@latest`
3. Run the suite: `xc dev`

## Tasks

This project uses [xc](https://github.com/joerdav/xc) to manage tasks.

### build

Compiles the production binary. It minifies CSS, generates templates, and builds a static Go binary optimized for AWS Lambda (Linux ARM64) into a 'dist' folder.

```bash
pnpm install
pnpm build:css
templ generate

# Create a clean distribution folder
rm -rf dist
mkdir -p dist

# Build the binary directly into 'dist/bootstrap'
GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -ldflags="-s -w" -o dist/bootstrap .
```

### test

Runs all Go tests across the application and infrastructure.

```bash
go test -v ./...
cd infra && go test -v ./...
```

### dev

Requires: build

Starts the development environment. Watches Tailwind and Templ files for changes, and runs the Go server with hot-reload.

```bash
pnpm watch:css &

templ generate --watch --proxy="http://localhost:8080" --cmd="go run ."
```

### diff

Requires: build

Run a CDK plan/diff to preview infrastructure changes.

```bash
cd infra && cdk diff
```

### deploy

Requires: test, build

Deploy infrastructure via CDK.

```bash
cd infra && export NODE_NO_WARNINGS=1 && cdk deploy
```

### clean

Removes build artifacts and generated files.

```bash
rm -rf dist
rm -f public/css/output.css
rm -fr node_modules
rm -f components/*_templ.go
```
