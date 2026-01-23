# StackFoundry

**Lean Architecture. High-Throughput Systems. Applied Intelligence.**

This is the official homepage for **StackFoundry Ltd**, a UK-based software consulting company specializing in optimizing chaos through rigorous engineering.

## The Architecture

We reject unnecessary complexity. This site is a **Self-Contained System (SCS)** built to demonstrate the power of modern, server-driven architectures.

* **Go (1.23+):** The core logic. Fast, typed, and compiled to a single static binary.
* **Templ:** Type-safe HTML templating. No runtime parsing errors.
* **Datastar:** Real-time UI updates via Server-Sent Events (SSE). No heavy client-side hydration or React virtual DOM.
* **Tailwind CSS:** Utility-first styling, embedded directly into the binary.
* **AWS Lambda (ARM64):** Deployed as a single function with zero idle costs.

## Getting Started

### Prerequisites

* [Go](https://go.dev/) (1.22+)
* [pnpm](https://pnpm.io/) (for Tailwind)
* [xc](https://github.com/joerdav/xc) (Task runner)

### Development

1. Install dependencies: `pnpm install` & `go mod download`
2. Install tools: `go install github.com/a-h/templ/cmd/templ@latest`
3. Run the suite: `xc dev`

## Tasks

This project uses [xc](https://github.com/joerdav/xc) to manage tasks.

### dev

Starts the development environment. Watches Tailwind and Templ files for changes, and runs the Go server with hot-reload.

```bash
# 1. Start Tailwind in watch mode (background)
pnpm watch:css &

# 2. Start Templ generation in watch mode
# --proxy: Reloads browser when Go server restarts
# --cmd: Re-runs the binary when Go files change
templ generate --watch --proxy="http://localhost:8080" --cmd="go run main.go"
```

### build

Compiles the production binary. It minifies CSS, generates templates, and builds a static Go binary optimized for AWS Lambda (Linux ARM64).

```bash
echo "🔥 Forging assets..."
pnpm build:css

echo "🔨 Generating templates..."
templ generate

echo "📦 Compiling binary (bootstrap)..."
# AWS Lambda requires the binary to be named 'bootstrap'
GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -ldflags="-s -w" -o bootstrap main.go

echo "🤐 Zipping for AWS..."
# -j ignores directory paths (junk paths) ensuring bootstrap is at the root of the zip
zip -j function.zip bootstrap

echo "✅ Ready to deploy: ./function.zip"
```

### clean

Removes build artifacts and generated files.

```bash
rm -f bootstrap
rm -f public/css/output.css
rm -f components/*_templ.go
echo "🧹 Workshop cleaned."
```
