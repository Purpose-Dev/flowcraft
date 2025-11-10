# Flowcraft Public Roadmap

Welcome to the public roadmap! This document outlines the major features and capabilities we are working on.

Flowcraft is being built as a "Dual Mode" platform:

1. A best-in-class **local-first CLI** for developers to get instant feedback.
2. A scalable, enterprise-ready **CI/CD Platform** (server and agents) for your team's production pipelines.

Our goal is to ensure that a pipeline you run locally works exactly the same way when run by the platform.

---

## Core Engine & Local Experience

This is about making `flowcraft run` the most powerful local-first orchestrator available.

* [x] **Concurrency Limiter:** Control how many jobs run in parallel (`parallelism = 8`).
* [x] **Job Retries:** Automatically retry flaky steps (`retry = 3`).
* [ ] **Timeouts:** Kill jobs or steps that run for too long (`timeout = "5m"`)
* [ ] **Conditional Execution (`when`):** Run jobs/steps based on conditions (`when = "env:CI_BRANCH == 'main'"` or
  `when = "failure()"`).
* [ ] **Matrix Builds:** Natively run jobs across a matrix of configurations (`matrix: { node: [18, 20, 22] }`).
* [ ] **Container Runtime (`uses:`):** Run any step inside a container (`uses: "node:22"`).
* [ ] **Container Volumes:** Mount volumes into container steps for caching (`.m2`, `.npm`) or tools (`docker.sock`).
* [ ] **Service Networking:** Automatically network `service = true` jobs (like databases, message brokers) with your
  test containers.

## Observability

You can't optimize what you can't seed. We are building first-class observability.

* [ ] **DAG Visualization:** A new `flowcraft graph` command to export your pipeline as a `graphviz` (DOT) file.
* [ ] **Centralized Local Logging:** A "summary" view for `flowcraft run` (no more 8-way parallel log spam) and a
  `flowcraft logs <job_name>` command to inspect individual logs.
* [ ] **Prometheus Metrics:** Expose an endpoint with metrics (job duration, cache hits, etc.).
* [ ] **OpenTelemetry Tracing:** Generate traces for your runs to visualize bottlenecks in tools like Jaeger.

## Performance & Data Management

Fast pipelines are happy pipelines.

* [ ] **Local Caching:** Smart caching based on `inputs` (file hashes) and `outputs`.
* [ ] **Remote Cache Backend:** Share your cache (S3, GCS) between developers and CI runs.
* [ ] **Cache Policies:** Define cache `scope` (branch vs. global) and `retention_days` to manage costs.
* [ ] **Artifact Management:** Pass artifacts (binaries, `dist` folders) between jobs, even in containers.

## Security & Integration

Enterprise-grade features for security and control.

* [x] **Log Secret Masking:** Automatically scrub secrets from all log output.
* [ ] **HashiCorp Vault Integration:** A `vault:` provider to pull secrets directly from Vault.
* [ ] **Manual Approve Gates:** Pause a pipeline and require human approval (`approve = true`), both in the CLI and the
  UI.

## The Platform (Server & Agents)

* [ ] **`flowcraft-server`:** The central orchestrator with a job queue and API.
* [ ] **`flowcraft-agent`:** The standalone worker binary that executes jobs.
* [ ] **`flowcraft run --remote`:** A CLI flag to send your local pipeline to the server for execution.
* [ ] **VCS Integration:** Natively clone Git repositories as the first step of a platform run.
* [ ] **Agent Tags & Job Placement (`runs_on`):** Smart scheduling. Run `jobs_on = ["macos", "m1"]` on agents with the
  correct tags.
* [ ] **Resource Management (`resources`):** Define a `cpu` and `mem` for jobs. The scheduler will "bin pack" jobs onto
* [ ] **Webhook Triggers:** Start pipelines from GitHub/GitLab `git push` events.
* [ ] **Web UI Dashboard:** A full dashboard to view pipeline history, live logs, agent status, and approve jobs.
  agents.
* [ ] **`flowcraft login`:** A CLI command to securely connect your local CLI to the `flowcraf-server`.

---
