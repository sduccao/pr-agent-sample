# AI Code Review Benchmark: Alibaba Open Code Review (OCR) vs. PR-Agent

This repository provides a hands-on benchmark comparing **Alibaba Open Code Review (`alibaba/open-code-review`)** and **PR-Agent (`The-PR-Agent/pr-agent`)** on an idiomatic Go backend service.

---

## 🎯 Benchmark Objectives

This benchmark evaluates how both AI code review tools handle complex, real-world Go pull requests across 3 specific technical dimensions:

1. **Context Isolation (Scenario A):** Ability to analyze deep multi-file call stacks (handler → service → repository → DTO) without token exhaustion, detecting subtle concurrency defects (goroutine leaks, data races).
2. **Precision vs. Noise Ratio (Scenario B):** Capability to ignore minor stylistic linter noise (missing comments, suboptimal slice allocation) and zero-in strictly on critical security vulnerabilities (SQL injection) and resource leaks (`http.Response.Body` unclosed).
3. **Line-Position Drift Accuracy (Scenario C):** Ability to maintain pinpoint accuracy for inline code review comments when file header imports, package comments, or struct definitions shift line numbers significantly.

---

## 🚀 Quick Start Guide

### Step 1: Push Repository & Set GitHub Secrets

1. Push this repository to GitHub:
   ```bash
   git remote add origin https://github.com/sduccao/pr-agent-sample.git
   git push -u origin main
   ```

2. Add your LLM provider keys in **GitHub Repository Settings → Secrets and variables → Actions**:
   - `GEMINI_API_KEY`: Recommended. Works for both **Alibaba Open Code Review** (via Google OpenAI compatibility endpoint with `gemini-3.6-flash`) and **The-PR-Agent** (`The-PR-Agent/pr-agent@main` using `gemini-3.6-flash`).
   - Alternatively, you can use `OPENAI_API_KEY` (for Alibaba OCR) and `OPENAI_KEY` (for PR-Agent).

---

### Step 2: Reproduce the 3 Benchmark Pull Requests

The repository has 3 pre-committed feature branches. Open PRs for each scenario:

#### 🧪 Scenario A: Context Isolation & Deep Concurrency Test
```bash
git checkout feature/scenario-a-concurrency
git push -u origin feature/scenario-a-concurrency
```
*Create Pull Request on GitHub:* `feature/scenario-a-concurrency` ➔ `main`

#### 🧪 Scenario B: Noise vs. Critical Defect Precision Test
```bash
git checkout feature/scenario-b-precision
git push -u origin feature/scenario-b-precision
```
*Create Pull Request on GitHub:* `feature/scenario-b-precision` ➔ `main`

#### 🧪 Scenario C: Line Position Shift & Differential Test
```bash
git checkout feature/scenario-c-line-drift
git push -u origin feature/scenario-c-line-drift
```
*Create Pull Request on GitHub:* `feature/scenario-c-line-drift` ➔ `main`

---

## 📊 Evaluation Matrix

| Evaluation Criteria | Alibaba Open Code Review (`alibaba/open-code-review`) | The-PR-Agent (`The-PR-Agent/pr-agent`) |
| :--- | :--- | :--- |
| **Action Repo** | [`alibaba/open-code-review`](https://github.com/alibaba/open-code-review) | [`The-PR-Agent/pr-agent`](https://github.com/The-PR-Agent/pr-agent) |
| **Core Philosophy** | **Precision-First:** Focuses on detecting high-confidence defects and resource leaks while avoiding noisy linter comments. | **Comprehensive PR Assistant:** Provides holistic PR summaries, style suggestions (`/improve`), and auto-generated unit tests (`/test`). |
| **Scenario A: Concurrency & Goroutine Leak** | Smart bundling isolates call graphs across 9 files; pinpoints line in `batch_processor.go` missing context cancellation / mutex lock on map update. | Provides strong architectural overview across touched files; may offer broad recommendations on concurrency patterns. |
| **Scenario B: Precision vs. Noise** | Filters out `make([]T, 0)` & missing comment noise. Places inline warning directly on `fmt.Sprintf` SQL Injection and missing `defer resp.Body.Close()`. | Identifies SQL injection & unclosed body while also generating stylistic improvement recommendations for slice allocation and comments. |
| **Scenario C: Line Placement Accuracy** | Dedicated line positioning & reflection parser ensures comment attaches to exact modified line (~line 150+). | Standard diff parser calculates line offset relative to git diff block. |
| **Token Efficiency & Cost** | Optimized call-graph chunking to reduce context token usage on multi-file PRs. | Full PR diff sent per command (`/review`, `/improve`); higher token usage but richer textual explanation. |
| **Interactive Utility** | Focused inline code review comments on pull request diffs. | Interactive slash commands (`/review`, `/improve`, `/describe`, `/test`, `/ask`). |

---

## 📁 Repository Structure

```
.
├── .github/
│   └── workflows/
│       └── ai-code-review-benchmark.yml   # Parallel CI runner for OCR & The-PR-Agent
├── cmd/
│   └── server/
│       └── main.go                        # HTTP server entrypoint
├── internal/
│   ├── config/                            # Application configuration
│   ├── handler/                           # HTTP controllers/handlers
│   ├── model/                             # Data models & DTOs
│   ├── repository/                        # Database persistence (SQLx + SQLite)
│   └── service/                           # Business logic & concurrency management
├── BENCHMARK.md                           # Benchmark testing guide & scoring matrix
└── go.mod
```
