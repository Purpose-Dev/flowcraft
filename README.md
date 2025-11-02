# Flowcraft

[![CI & Dogfood Test](https://github.com/Purpose-Dev/flowcraft/actions/workflows/ci.yml/badge.svg)](https://github.com/Purpose-Dev/flowcraft/actions/workflows/ci.yml)
[![Latest Release](https://img.shields.io/github/v/release/Purpose-Dev/flowcraft)](https://github.com/Purpose-Dev/flowcraft/releases)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25.3%2B-blue.svg)](https://go.dev/)

Flowcraft is a fast, portable, and dependency-aware build orchestrator designed to run CI/CD pipelines locally and in
automation. It reads a simple `flow.toml` file, builds an execution graph (DAG), and runs tasks with maximum
parallelism.

---

## The Problem

Modern CI/CD pipelines (like GitHub Actions) are powerful but suffer from three main problems:

1. **Local vs CI Disparity:** A pipeline that only runs on a server is hard to debug. Developers waste time pushing
   small changes to fix syntax or pathing errors.
2. **Complexity and Performance:** YAML-based CI configuration can become complex and unreadable. They often execute
   tasks sequentially, even when parallelism is possible, leading to slow builds.
3. **High Costs:** Slow, sequential pipelines or repeated debugging pushes don't just waste time, they consume expensive
   CI/CD runner minutes (e.g., GitHub-hosted runners), leading to higher bills.

## The Solution

Flowcraft solves this by defining the *entire* pipeline logic in a single `flow.toml` file that can be run anywhere.

* **Run Locally, Run in CI:** Execute `flowcraft run` on your machine or in a GitHub Actions Workflow. The execution
  order, parallelism, and error handling are identical.
* **Parallel by Default:** Flowcraft builds a dependency graph (DAG) from your jobs, automatically parallelizing tasks
  that don't depend on each other.
* **Simple & Declarative:** The TOML configuration is clean, easy to read, and version-controlled with your code.

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

```shell
flowcraft run --file ci/prod.toml
```

### `flowcraft validate`

Parses the config file and validates the dependency graph. This is a "dry run" command.

Use this in your CI to quickly fail a build if you have a syntax error or a circular dependency (e.g., Job A depends on
B, and B depends on A).

- `--file` (or `-f`): Specify a different config file (default: `flow.toml`)

```shell
flowcraft validate
```

If successful, it will exit with code 0. If it fails, it will print the error (e.g., "circular dependency detected") and
exit with code 1.

---

## Configuration Reference

### **Top-Level:** `[jobs]`

The `flow.toml` file is built around a single top-level table: `[jobs]`. Each job is a TOML table entry within `[jobs]`.

```toml
[jobs.my-first-job]
# ... config for this job ...

[jobs.my-second-job]
# ... config for this job ...
```

### **Job Definition**

Each job can have the following keys:

* `depends_on = ["job-a", "job-b"]` An array of job names that must complete successfully before this job can start.
* `steps = [...]` An array of Step objects. These are executed sequentially. If any step fails, the job stops, and no
  parallel steps are run.
* `parallel = [...]` An array of Step objects. These are executed concurrently after all steps have completed
  successfully. If any of these parallel steps fail, the job is marked as failed.

### **Step Definition**

A `Step` object is a TOML inline table and has three keys:

* `name = "My Step"`: A descriptive name used for logging.
* `cmd = "echo 'hello'"`: The shell command to execute.
* `dir = "path/to/dir"`: (Optional) The working directory to `cd` into before running the command.

### Example (`steps`):

```toml
[[jobs.my-job.steps]]
name = "Install"
cmd = "npm install"
dir = "webapp"

[[jobs.my-job.steps]]
name = "Build"
cmd = "npm run build"
dir = "webapp"
```

### Example (`parallel`):

```toml
[[jobs.my-job.parallel]]
name = "Run unit tests"
cmd = "go test ./..."
dir = "api"

[[jobs.my-job.parallel]]
name = "Run linter"
cmd = "golangci-lint run"
dir = "api"
```

---

## License

Flowcraft is released under the **Apache 2.0 License**. You can find the full license text in the [`LICENSE`](LICENSE)
file included in this repository.

Copyright 2025 Riyane El Qoqui.
