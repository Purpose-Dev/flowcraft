# Flowcraft

[![CI & Dogfood Test](https://github.com/Purpose-Dev/flowcraft/actions/workflows/ci.yml/badge.svg)](https://github.com/Purpose-Dev/flowcraft/actions/workflows/ci.yml)
[![Latest Release](https://img.shields.io/github/v/release/Purpose-Dev/flowcraft)](https://github.com/Purpose-Dev/flowcraft/releases)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25.3%2B-blue.svg)](https://go.dev/)

Flowcraft is a fast, portable, and dependency-aware build orchestrator. It reads a `flow.toml`, builds an execution
graph (DAG), and runs tasks with maximum parallelism.

It's designed to run CI/CD pipelines locally for fast feedback, and then execute the _exact same pipeline_ in
automation, giving you true build parity.

---

## The Problem

Modern CI/CD is powerful, but let's be honest, it's often a source of frustration:

1. **"It works on my machine!"** The classic developer nightmare. A pipeline that only runs on a server is a black box
   that's impossible to debug, leading to "commit-and-pray" development.
2. **YAML Hell:** We've all seen it. A 1000-line, fragile, copy-pasted YAML file that no one understands. It's complex,
   hard to read, and even harder to maintain.
3. **Slow Feedback Loops:** Most CI systems run sequentially, wasting time. You wait 20 minutes for a build that could
   have run in 5, just to see a simple linting error. This wastes time and expensive runners' minutes.

## The Solution (One Pipeline, Anywhere)

Flowcraft solves this by defining the entire pipeline logic in a single `flow.toml` file that runs identically anywhere.

* **Run Locally, Run in CI:** Execute `flowcraft run` on your machine to get instant feedback. Run the *exact same
  command* in GitHub Actions, GitLab, or any CI provider. **True CI/CD parity.**
* **Parallel by Default:** Flowcraft builds a dependency graph (DAG) and runs everything that can be parallelized in
  _parallel_. Stop paying for sequential builds you don't need.
* **Simple & Declarative:** No more YAML hell. A clean, powerful TOML syntax that's easy to read and version-controlled
  with your code.
* **Container-First:** Define your job's runtime with `uses = "node:22"`. Flowcraft runs your steps in isolated
  containers, guaranteeing a clean, portable, and secure build environment.

---

## Core Features

* **Dual Mode:** Run 100% locally with `flowcraft run`, or scale up with `flowcraft-server` and `flowcraft-agent` for a
  full, distributed, enterprise-grade platform.
* **Container Runtime:** `uses = "node:22"` or `uses = "docker:dind"`. All steps are isolated.
* **Service Networking:** Spin up `service = true` jobs (like Postgres) and they're automatically networked with your
  jobs.
* **Smart Caching:** `inputs` and `outputs` directives ensure you never rebuild what isn't broken. Supports local and
  remote (S3, GCS) caches.
* **Matrix Builds:** Test again `matrix: { node: [18, 20, 22] }` without copy-pasting your jobs.
* **Powerful Orchestration:** `retry = 3`, `timeout = "5m"`, `when = "failure()"`... Control your flow like a pro.

---

## Roadmap

Flowcraft is actively developed. We are working on a host of enterprise-grade features to build a complete, end-to-end
CI/CD platform.

You can follow our high-level public milestones, here: [Todo Plan](./todo.md)

---

## Quick Start

### 1. Installation

Download the latest binary for your OS from the [GitHub Releases](https://github.com/Purpose-Dev/flowcraft/releases)
page and place it in your `PATH`.

### 2. Create `flow.toml`

Create a `flow.toml` file in the root of your project.

```toml
# flow.toml

# The 'setup' job has no dependencies
[jobs.setup]
depends_on = []
[[jobs.setup.steps]]
name = "Create artifacts directory"
cmd = "mkdir -p bin"

# The 'build-api' job depends on 'setup'
[jobs.build-api]
depends_on = ["setup"]
[[jobs.build-api.steps]]
name = "Go Build API"
cmd = "go build -o ../bin/api-server"
dir = "api" # Run this command inside the 'api' directory

# The 'build-webapp' job also depends on 'setup'
[jobs.build-webapp]
depends_on = ["setup"]
[[jobs.build-webapp.steps]]
name = "NPM Install & Build"
cmd = "npm install && npm run build"
dir = "webapp"

# The 'deploy' job depends on both builds
[jobs.deploy]
depends_on = ["build-api", "build-webapp"]
[[jobs.deploy.steps]]
name = "Deploy"
cmd = "echo 'Deploying...'"
```

### 3. Run It

From the root of your project, run:

```shell
flowcraft run
```

Flowcraft will parse the file, build the graph, and execute the jobs:

1. Run `setup`.
2. Run `build-api` and `build-webapp` in **parallel**.
3. Run `deploy` (only if both builds succeed).

---

## Commands

### `flowcraft run`

Runs the pipeline defined in the `flow.toml` file.

- `--file` (or `-f`): Specify a different config file (default: `flow.toml`)
- `--remote`: (Coming soon) Execute the pipeline on a remote `flowcraft-server`.

### `flowcraft validate`

Parses the config file and validates the dependency graph. This is a "dry run" command.

Use this in your CI to quickly fail a build if you have a syntax error or a circular dependency.

- `--file` (or `-f`): Specify a different config file (default: `flow.toml`)

### `flowcraft graph`

(Coming soon) Parse the pipeline and output the dependency graph in `DOT` (Graphviz) format.

### `flowcraft logs <job_name>`

(Coming soon) Show the detailed, collected logs for a specific job from the last run.

---

## Configuration Reference

`flow.toml` is designed to be powerful yet simple. Here are the main concepts:

- `[env]` **(Global):** A top level table for global environment variables.
- `[jobs.<job_name>]`: The main build unit.
    - `depends_on = []`: An array of job names this job depends on.
    - `env = {}`: A map of job-specific environment variable.
    - `when = ""`: A condition to run this job (e.g., `"env.CI_BRANCH == 'main'"`).
    - `retry = 3`: (Coming soon) Number of times to retry a failed job.
    - `timeout = "1h"`: (Coming soon) Max duration for the job.
    - `runs_on = []`: (Coming soon) Tags required for an agent to run this job (e.g., `["macos", "m1"]`).
- `[[jobs.<job_name>.steps]]`: An array of steps to run *sequentially*.
    - `name = ""`: A descriptive name for logging.
    - `cmd = ""`: The shell command to execute.
    - `dir = ""`: The working directory to `cd` into before running.
    - `shell = "bash"`: (Coming soon) Specify the shell (`bash`, `pwsh`, `cmd`).
    - `uses = "image:tag"`: (Coming soon) A container image to run this step in.
- `[[jobs.<job_name>.parallel]]`: An array of steps to run *concurrently*.
    - `name = ""`: A descriptive name for logging.
    - `cmd = ""`: The shell command to execute.
    - `dir = ""`: The working directory to `cd` into before running.
    - `shell = "bash"`: (Coming soon) Specify the shell (`bash`, `pwsh`, `cmd`).
    - `uses = "image:tag"`: (Coming soon) A container image to run this step in.
---

## License

Flowcraft is released under the **Apache 2.0 License**. You can find the full license text in the [`LICENSE`](LICENSE)
file included in this repository.

Copyright 2025 Riyane El Qoqui.
