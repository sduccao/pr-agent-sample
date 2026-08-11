# Go AI Code Review Benchmark: Alibaba OCR vs CodiumAI PR-Agent

[![Go Report Card](https://goreportcard.com/badge/github.com/benchmark/go-ai-review-benchmark)](https://goreportcard.com/report/github.com/benchmark/go-ai-review-benchmark)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

This repository is a hands-on benchmark suite designed to compare the functional, architectural, and behavioral performance of **Alibaba Open Code Review (OCR)** vs **CodiumAI PR-Agent** when reviewing Go codebase Pull Requests.

## 🚀 Scenario Overview

- **`main` branch:** Clean, production-ready baseline using idiomatic Go (`net/http` + `sqlx` + SQLite).
- **`feature/scenario-a-concurrency`:** Multi-file refactor featuring a **Go Data Race** and **Goroutine Leak**.
- **`feature/scenario-b-precision`:** Noise vs Defect contrast featuring **SQL Injection** + **Unclosed HTTP Response Body** amidst Go linter style noise.
- **`feature/scenario-c-line-drift`:** Large top-of-file import/comment shift with a deep **Nil Pointer Dereference**.

See [BENCHMARK.md](file:///Users/duccao/workspace/pr-agent-sample/BENCHMARK.md) for full instructions on executing the benchmark and evaluating results.

## 🛠️ Requirements

- Go 1.22+
- GitHub Actions with `OPENAI_API_KEY` and `OPENAI_KEY` configured.
