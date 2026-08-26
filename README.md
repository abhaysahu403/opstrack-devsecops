# OpsTrack — Incident & Service Management API

> **Status: Complete.** This repository contains the full Go application,
> database layer, Docker/Compose setup, test suite, GitHub Actions
> CI/CD pipeline, security scanning (Gosec, Trivy, Snyk, SonarQube), the HTML
> report dashboard, and GitHub Pages publishing.
>
> **Terraform and GCP deployment are intentionally not included** in this
> build. The pipeline builds, scans, and (optionally) pushes the Docker image
> to GHCR (`ghcr.io`); wiring that image to a specific cloud target — GCP
> Compute Engine, Cloud Run, or anywhere else — is a follow-on step. See
> [Deployment](#16-deployment-not-included) below for what would need to be added.

OpsTrack is a small, realistic Go REST API used as the vehicle for a larger
DevSecOps demonstration. It lets an operations team track **services** and
the **incidents** raised against them.

## 1. Project Overview

OpsTrack is intentionally moderate in scope: enough real functionality
(REST API, PostgreSQL persistence, validation, state transitions, tests) to
exercise every stage of an enterprise CI/CD pipeline, without becoming a
large application in its own right.

## 2. Architecture

```text
Client
  │
  ▼
net/http (Go 1.22 ServeMux, method + path patterns)
  │
  ▼
internal/handlers   -- HTTP transport (decode/encode JSON, status codes)
  │
  ▼
internal/service     -- business logic, validation, state machine
  │
  ▼
internal/repository  -- persistence interfaces
  │
  ├── postgres.go     -- PostgreSQL implementation (pgx)
  └── memory.go        -- in-memory implementation (used by unit tests)
  │
  ▼
PostgreSQL
```

`internal/config` loads configuration from environment variables.
`internal/database` owns the connection pool and a minimal built-in
migration runner (applies `migrations/*.sql` in order, tracked in a
`schema_migrations` table).

## 3. Application Features

**Services** — `id, name, description, owner, environment, status,
created_at, updated_at`. Status: `ACTIVE | DEGRADED | DOWN | MAINTENANCE`.

**Incidents** — `id, service_id, title, description, severity, status,
assigned_to, created_at, updated_at, resolved_at`. Severity: `LOW | MEDIUM |
HIGH | CRITICAL`. Status: `OPEN | INVESTIGATING | MITIGATED | RESOLVED`,
enforced by an explicit state machine in `internal/models` (e.g. a resolved
incident cannot silently move back to `OPEN`).

Filtering is supported via query parameters (`status`, `severity`,
`service_id`, `environment`). `/health`, `/ready` (checks DB connectivity),
and `/` (API info) are also implemented.

## 4. Technology Stack

* Go 1.22 (standard library `net/http`, method/path-aware `ServeMux`)
* PostgreSQL via [pgx/v5](https://github.com/jackc/pgx)
* Docker / Docker Compose
* GitHub Actions (reusable workflows), golangci-lint
* SonarQube, Trivy (filesystem + image), Snyk, Gosec
* GitHub Pages (report dashboard), GHCR (Docker image registry)

## 5. Local Setup

```bash
git clone <repository-url>
cd opstrack-devsecops
cp .env.example .env
go mod tidy      # first time only, populates go.sum
go build ./...
go test ./...
```

To run against a local PostgreSQL instance without Docker, start Postgres
yourself, set the `DB_*` variables (see `.env.example`), then:

```bash
go run ./cmd/server
```

## 6. Docker Setup

```bash
docker compose up --build
```

This starts PostgreSQL and the OpsTrack API together. The API applies
migrations automatically on startup. Once running:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/ready
curl http://localhost:8080/
```

## 7. API Documentation

Base path: `/api/v1`

### Services

```text
POST   /api/v1/services
GET    /api/v1/services            ?status=ACTIVE&environment=production
GET    /api/v1/services/{id}
PUT    /api/v1/services/{id}
DELETE /api/v1/services/{id}
```

Create example:

```bash
curl -X POST http://localhost:8080/api/v1/services \
  -H "Content-Type: application/json" \
  -d '{"name":"Payment Service","owner":"payments-team","environment":"production","status":"ACTIVE"}'
```

### Incidents

```text
POST   /api/v1/incidents
GET    /api/v1/incidents           ?status=OPEN&severity=CRITICAL&service_id=1
GET    /api/v1/incidents/{id}
PUT    /api/v1/incidents/{id}
DELETE /api/v1/incidents/{id}
```

Create example:

```bash
curl -X POST http://localhost:8080/api/v1/incidents \
  -H "Content-Type: application/json" \
  -d '{"service_id":1,"title":"Payment API latency","severity":"HIGH"}'
```

Response codes: `201` created, `200` ok, `204` deleted, `400` validation
error, `404` not found, `500` unexpected server error.

*Not included*: Swagger/OpenAPI docs at `/docs`. The API surface is small
enough that this README + the handler tests serve as the reference; adding
`swaggo/swag` annotations later is straightforward if it's needed.

## 8. Testing

```bash
go test ./...                              # unit tests (in-memory repos, no DB needed)
go test ./... -coverprofile=coverage.out   # with coverage
go tool cover -func=coverage.out
go test -tags=integration ./tests/...      # integration tests, requires real Postgres
```

Unit tests cover: service/incident validation, the incident status state
machine, service-layer business logic (via in-memory repositories), and
HTTP handlers (via `httptest`). The integration test in `tests/` is gated
behind the `integration` build tag so `go test ./...` stays fast and
dependency-free — it's meant to be run by CI against a Postgres service
container, or locally against `docker compose up postgres`.

## 9. CI/CD Architecture

```text
Developer
   │
   ▼
Feature Branch (feature/*, bugfix/*, hotfix/*)
   │
   ▼
Pull Request  ──────────────►  pr-validation.yml
   │                              ├─ reusable/test.yml      (fmt, vet, build, unit+integration tests, lint, coverage)
   │                              ├─ reusable/security.yml  (gosec, trivy fs, snyk, sonarqube)
   │                              ├─ reusable/docker.yml    (docker build, trivy image scan)
   │                              └─ assemble-reports        (dashboard + ZIP + Actions Summary)
   ▼
Review + required checks pass
   │
   ▼
Merge to main  ─────────────►  main.yml
                                  ├─ reusable/test.yml
                                  ├─ reusable/security.yml
                                  ├─ reusable/docker.yml (+ push image to GHCR)
                                  ├─ assemble-reports
                                  └─ publish-reports.yml → GitHub Pages
```

The three expensive stages (test, security, docker) are independent reusable
workflows and run **in parallel** as separate jobs on both PRs and `main`.
`assemble-reports` runs last (`needs: [test, security, docker]`, `if: always()`)
so a report is always produced even if one stage fails, then generates the
dashboard, zips it, writes the GitHub Actions Summary, and uploads it all as
a workflow artifact.

## 10. Security Scanning

| Tool | What it scans | Where | Requires |
|---|---|---|---|
| **Gosec** | Go source, common security anti-patterns | `reusable/security.yml` | nothing — always runs |
| **Trivy (filesystem)** | Repo files + Go module dependencies (CVEs) | `reusable/security.yml` | nothing — always runs |
| **Trivy (image)** | The built Docker image's OS packages + libs | `reusable/docker.yml` | nothing — always runs |
| **Snyk** | Go module dependency vulnerabilities | `reusable/security.yml` | `SNYK_TOKEN` secret |
| **SonarQube** | Code quality, duplication, coverage, quality gate | `reusable/security.yml` | `SONAR_TOKEN` + `SONAR_HOST_URL` secrets |

Gosec and Trivy always run because they need no external service — they're
genuinely free. Snyk and SonarQube talk to a real external
account/server, so if their secrets aren't set the corresponding stage is
skipped and reported as `NOT CONFIGURED` (not `PASS` — the workflow never
fabricates a result for a tool that didn't actually run).

**Severity gating** is deliberately conservative by default so a stray
low/medium finding in a third-party dependency can't block the prototype.
Both Trivy steps accept a `trivy-fail-on-critical-high` input
(`reusable/security.yml`, `reusable/docker.yml`) that, when `true`, fails
the job on `CRITICAL`/`HIGH` findings. Flip it in `pr-validation.yml` /
`main.yml` (or pass it via `workflow_dispatch` on `security-scan.yml`) once
you're ready to enforce it. Gosec always reports any finding as `FAIL` in
its own report — it doesn't currently gate the whole job on that (see
`reusable/security.yml`'s `gosec` job if you want to change that).

## 11. GitHub Actions

```text
.github/workflows/
├── pr-validation.yml     # top-level: runs on every PR into main
├── main.yml               # top-level: runs on every push to main
├── security-scan.yml       # top-level: manual + nightly (cron) security scan
├── build-image.yml          # top-level: manual docker build/scan entry point
├── publish-reports.yml       # reusable: deploys an already-assembled reports/ dir to GitHub Pages
└── reusable/
    ├── test.yml               # workflow_call: fmt/vet/build/lint/unit+integration tests/coverage
    ├── security.yml            # workflow_call: gosec/trivy-fs/snyk/sonarqube
    └── docker.yml               # workflow_call: docker build/trivy image scan/(optional) push
```

All workflows use least-privilege `permissions:` blocks (mostly
`contents: read`; `pages: write` / `id-token: write` only on the Pages
deploy job, `packages: write` only where GHCR is pushed to). None of the
workflows use `continue-on-error: true` to force a green pipeline — failures
propagate, except for the explicitly-documented "not configured" cases
above.

## 12. Artifacts

Every workflow run uploads its reports as downloadable **GitHub Actions
artifacts** (Actions tab → the workflow run → Artifacts), kept separate per
stage during the run and then combined:

* `reports-test`, `reports-gosec`, `reports-trivy-fs`, `reports-trivy-image`,
  `reports-snyk`, `reports-sonar` — per-stage, uploaded by the reusable
  workflows themselves.
* `opstrack-ci-reports` (on `main`) / `opstrack-ci-reports-pr-<number>` (on
  PRs) — the combined `reports/` directory, including the generated
  dashboard and the downloadable `opstrack-ci-reports.zip`.

## 13. GitHub Pages

On every push to `main`, `main.yml`'s `assemble-reports` job builds the
dashboard and uploads it as an artifact, then `publish-reports.yml` deploys
it to GitHub Pages via `actions/upload-pages-artifact` +
`actions/deploy-pages`. PR runs do **not** publish to Pages, so the public
dashboard only ever reflects what's on `main`.

**One-time setup required:** repo **Settings → Pages → Source: "GitHub
Actions"**. After the first successful `main.yml` run, the dashboard is
live at:

```text
https://<your-org-or-user>.github.io/<repo-name>/
```

## 14. Deployment (not included)

This build intentionally stops at "built, scanned, and (optionally) pushed
to GHCR." There is no Terraform, no GCP Compute Engine provisioning, and no
`deploy.yml`. `reusable/docker.yml` accepts a `push-image` input; when
`true` (the default on `main.yml`), it logs into `ghcr.io` with the
built-in `GITHUB_TOKEN` and pushes `:latest` and `:<git-sha>` tags.

To take this further:

1. Point `reusable/docker.yml`'s push step at a different registry (Artifact
   Registry, ECR, Docker Hub, …) instead of/in addition to GHCR.
2. Add a `deploy.yml` that pulls the pushed image onto your target and runs
   it — the app is a single stateless Docker container reading
   configuration from the `DB_*`/`SERVER_PORT`/`APP_ENV` environment
   variables (see `.env.example`), so it doesn't assume any particular
   host.
3. Provision infrastructure however you prefer (Terraform, a cloud
   console, Ansible, …) — nothing in this repo assumes Terraform
   specifically.

## 15. Terraform (not included)

Not part of this build — see [Deployment](#14-deployment-not-included)
above.

## 16. GitHub Secrets / Variables

| Name | Required? | Used by | Purpose |
|---|---|---|---|
| `SONAR_TOKEN` | Optional — security stage skips gracefully without it | `reusable/security.yml` | Authenticates to your SonarQube server |
| `SONAR_HOST_URL` | Optional — required together with `SONAR_TOKEN` | `reusable/security.yml` | Base URL of your SonarQube server |
| `SNYK_TOKEN` | Optional — security stage skips gracefully without it | `reusable/security.yml` | Authenticates to Snyk |
| `GITHUB_TOKEN` | Automatic — provided by GitHub Actions, no setup needed | `reusable/docker.yml`, `publish-reports.yml` | Pushes images to GHCR; deploys GitHub Pages |

Nothing else is required to run the pipeline end-to-end — Gosec and Trivy
need no credentials. GCP/registry secrets (`GCP_PROJECT_ID`,
`GCP_ARTIFACT_REGISTRY`, service-account keys, etc.) are **not** used by
this build since GCP deployment isn't included; add them yourself if/when
you build a `deploy.yml` on top of this.

To configure a secret: repo **Settings → Secrets and variables → Actions →
New repository secret**.

## 17. Branch Protection

Recommended rule for `main` (**Settings → Branches → Add branch ruleset /
protection rule**):

* Require a pull request before merging (at least 1 approval).
* Require status checks to pass before merging — select the
  `PR Validation` jobs (`Test`, `Security & Quality`, `Docker Build & Scan`,
  `Assemble Reports & Summary`) once they've run at least once on a PR so
  GitHub can list them.
* Require branches to be up to date before merging.
* Do not allow force pushes or deletions of `main`.

Workflows named `feature/*`, `bugfix/*`, and `hotfix/*` should always be
branched from `main` and merged back via PR — never pushed to directly.

## 18. Troubleshooting

* **`go build` fails with missing go.sum entries** — run `go mod tidy` once
  with network access; `go.sum` is intentionally left to be generated
  against the real module proxy rather than hand-written.
* **API can't connect to Postgres** — check `DB_HOST`/`DB_PORT` match your
  Compose/local setup, and that Postgres is actually accepting connections
  (`docker compose ps`).
* **Migrations don't rerun after editing a `.sql` file** — the runner
  tracks applied files by name in `schema_migrations`; add a new file
  instead of editing an already-applied one.
* **Snyk/SonarQube stage shows `NOT CONFIGURED`** — this is expected until
  you add `SNYK_TOKEN` / `SONAR_TOKEN` + `SONAR_HOST_URL` as repository
  secrets (see [GitHub Secrets](#16-github-secrets--variables)). It is not
  a pipeline failure.
* **GitHub Pages 404s after the first `main.yml` run** — confirm
  **Settings → Pages → Source** is set to "GitHub Actions" (it defaults to
  "Deploy from a branch", which this workflow doesn't use).
* **`reusable/docker.yml` push step fails with a permissions error** —
  confirm the repository's **Settings → Actions → General → Workflow
  permissions** allows "Read and write permissions," or that the
  `packages: write` permission on `main.yml` hasn't been overridden by an
  org-level policy.

## 19. Demo Instructions

See [`docs/demo-guide.md`](docs/demo-guide.md) for a full 10–15 minute
walkthrough script.

---

## Project Structure

```text
opstrack-devsecops/
├── cmd/server/main.go        Application entrypoint
├── internal/
│   ├── config/                Environment-based configuration
│   ├── database/               Connection pool + migration runner
│   ├── handlers/                HTTP transport layer
│   ├── middleware/               Logging, recovery, JSON content-type
│   ├── models/                    Domain types + incident state machine
│   ├── repository/                 Postgres + in-memory implementations
│   ├── service/                     Business logic layer
│   └── validation/                   Input validation
├── migrations/                 SQL migrations (schema + optional seed data)
├── tests/                        Integration tests (build-tag gated)
├── scripts/                       Report-generation + local dev helper scripts
├── docs/                          demo-guide.md
├── reports/                       Generated CI/security reports (gitignored, .gitkeep only)
├── .github/workflows/             GitHub Actions (pr-validation, main, security-scan,
│                                    build-image, publish-reports, reusable/*)
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod / go.sum
├── sonar-project.properties
├── .env.example
├── .gitignore
├── LICENSE
└── README.md
```
