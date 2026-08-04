# Stack Foundry

**One Principal. AI-First Delivery. Squad-Level Output.**

- Live site: <https://stackfoundry.ai>
- Legacy domain: <https://stackfoundry.co.uk> (redirect target: `.ai`)

StackFoundry is a UK principal engineering consultancy focused on shipping production systems with **AI-first execution and senior-level rigor**.

## Architecture

This site is a self-contained Go system, optimized for speed and low operational overhead.

- **Go (1.24+):** Core app logic and API handlers.
- **Templ:** Type-safe server-rendered UI components.
- **HTMX + Tailwind:** Lightweight interactivity and styling.
- **AWS Lambda (ARM64) + API Gateway:** Scale-to-zero compute.
- **Cloudflare:** DNS, TLS edge, caching, and redirect control.
- **Pulumi (Go):** Unified infrastructure-as-code for AWS and Cloudflare.

## Infrastructure (Pulumi)

Pulumi fully replaces the old CDK setup.

```bash
infra/
├── main.go             # Pulumi entrypoint
├── aws_stack.go        # Lambda, API Gateway, ACM, budget
├── cloudflare_stack.go # DNS records + zone settings
├── Pulumi.yaml
└── Pulumi.prod.yaml
```

The stack deploys:
- `stackfoundry.ai` as the primary domain
- DNS validation for ACM certs via Cloudflare
- API Gateway custom domain mapping (phase 2)

## Prerequisites

- [Go](https://go.dev/) (1.24+)
- [pnpm](https://pnpm.io/)
- [templ](https://templ.guide/) CLI
- [Pulumi](https://www.pulumi.com/docs/iac/download-install/)
- [AWS CLI](https://aws.amazon.com/cli/) configured
- [xc](https://github.com/joerdav/xc) (task runner)

## Tasks

This project uses [xc](https://github.com/joerdav/xc) task blocks in this README.

### build

Compiles the production binary for Lambda into `dist/bootstrap`.

```bash
pnpm install
pnpm build:css
templ generate
rm -rf dist
mkdir -p dist
GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -ldflags="-s -w" -o dist/bootstrap .
```

### test

Runs all Go tests across app and infrastructure.

```bash
go test -v ./...
cd infra && go test -v ./...
```

### dev

Requires: build

Starts local development (Tailwind watcher + Templ watcher + Go app).

```bash
pnpm watch:css &
templ generate --watch --proxy="http://localhost:8080" --cmd="env TEMPL_DEV_MODE=false go run ."
```

### diff

Requires: build

Pulumi preview for the currently selected stack.

```bash
cd infra && pulumi preview
```

### diff-ai

Requires: build

Pulumi preview for prod stack.

```bash
cd infra && pulumi stack select prod && pulumi preview
```

### deploy

Requires: test, build

Deploys with custom domain enabled (normal steady-state deploy path).

```bash
cd infra && pulumi stack select prod && pulumi config set stackfoundry:enableCustomDomain true && pulumi up --stack prod
```

### deploy-ai

Requires: test, build

Alias for primary `.ai` deployment.

```bash
cd infra && pulumi stack select prod && pulumi config set stackfoundry:enableCustomDomain true && pulumi up --stack prod
```

### deploy-phase1

Requires: test, build

First-pass deploy for certificate/bootstrap (`enableCustomDomain=false`).

```bash
cd infra && pulumi stack select prod && pulumi config set stackfoundry:enableCustomDomain false && pulumi up --stack prod
```

### deploy-phase2

Requires: test, build

Second-pass deploy after ACM is issued (`enableCustomDomain=true`).

```bash
cd infra && pulumi stack select prod && pulumi config set stackfoundry:enableCustomDomain true && pulumi up --stack prod
```

### clean

Removes build artifacts and generated files.

```bash
rm -rf dist
rm -f public/css/output.css
rm -fr node_modules
rm -f components/*_templ.go
```

## Deploy with Pulumi (Two-Phase)

Use this flow for first-time or cert-related domain deploys.

### Preflight config

```bash
cd infra
pulumi stack select prod
pulumi config set aws:region eu-west-2
pulumi config set cloudflareZoneID 28170f6ed3495c645ea55b5e49be7d96
# only if needed:
# pulumi config set --secret cloudflare:apiToken <TOKEN>
```

### Phase 1: Bootstrap cert and base infra

```bash
xc test
xc build
cd infra
pulumi config set stackfoundry:enableCustomDomain false
pulumi up --stack prod
```

### Phase 2: Enable custom domain mapping

Run once ACM certificate status is `ISSUED`.

```bash
cd infra
pulumi config set stackfoundry:enableCustomDomain true
pulumi up --stack prod
```

## Go-Live Verification Checklist

1. In AWS ACM, confirm cert for `stackfoundry.ai`/`*.stackfoundry.ai` is `ISSUED`.
2. In API Gateway, confirm custom domain and API mapping exist.
3. In Cloudflare DNS, confirm:
   - ACM validation CNAME is **DNS only** (not proxied).
   - Root and `www` CNAMEs point to API Gateway target and are proxied.
4. Verify HTTPS:
   - `https://stackfoundry.ai` loads successfully.
   - `https://www.stackfoundry.ai` resolves correctly.
5. Smoke test contact form and application routes.

## Rollback and Safety

- If custom domain mapping fails, set `stackfoundry:enableCustomDomain=false` and run `pulumi up --stack prod`.
- Keep Cloudflare DNS as source of truth during migration to avoid registrar-related downtime.
- Use `pulumi preview` before each deploy window.

## `.co.uk` Redirect Plan (Cloudflare DNS retained)

Recommended fast path:

1. Keep registrar as-is for now, keep DNS on Cloudflare.
2. Add Cloudflare Redirect Rule:
   - Match: `stackfoundry.co.uk/*` and `www.stackfoundry.co.uk/*`
   - Action: `301` to `https://stackfoundry.ai/$1`
3. Keep any required non-web records (for example MX/email) untouched.
4. Validate:
   - `curl -I https://stackfoundry.co.uk` returns `301`.
   - Target resolves to `.ai` correctly.

## Domain Ownership Clarification

- **CloudFront is not a domain registrar.**
- Registrar options:
  - Keep current registrar + Cloudflare DNS (recommended now)
  - Transfer registration to Route53 Registrar later if desired
- Do registrar transfer only after `.ai` is stable and `.co.uk` redirect has baked in production.

