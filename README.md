# kdiff: The Kubernetes Configuration Diff & Validation Tool

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.24-blue.svg)
![Build Status](https://github.com/YogPandya12/kdiff/actions/workflows/ci.yml/badge.svg)

**kdiff** is a powerful, developer-friendly CLI tool designed to solve the "YAML Hell" in Kubernetes. It simplifies configuration management by providing real-time validation, semantic comparison, and drift detection for your Kubernetes manifests.

---

## 🛑 The Problem: Why kdiff?

Managing Kubernetes configurations across multiple environments (Development, Staging, Production) is a known pain point for developers:

*   **Configuration Drift**: Small, unnoticed changes between environments lead to "it works on my machine" but fails in production.
*   **YAML Complexity**: Manually comparing 1000+ line YAML files to find a single changed environment variable is tedious and error-prone.
*   **Deployment Failures**: Invalid syntax or missing required fields (like `apiVersion` or `kind`) often crash pipelines only *after* code is pushed.
*   **Lack of Visualization**: Standard `diff` tools don't understand Kubernetes structure—they just see text.

**kdiff** was built to address these niche but critical challenges that standard tools like `kubectl diff` or `helm` don't fully solve for the average developer.

---

## 🚀 The Solution

**kdiff** is an "Enhanced Configuration Management Tool" that acts as your safety net. It understands Kubernetes schemas and structure, allowing you to:

1.  **Visualize Differences**: See exactly what changed between two manifests (e.g., `replicas: 2` vs `replicas: 3`) in a color-coded, human-readable format.
2.  **Validate Instantly**: Catch syntax errors, missing fields, and schema violations *before* you apply them to your cluster.
3.  **Prevent Drift**: Ensure your Development and Production environments stay in sync by highlighting unintended discrepancies.

---

## ✨ Features

*   **🔍 Semantic Comparison**: Compares YAML files line-by-line but understands the structure.
    *   *Supports JSON and Table output formats for CI/CD integration.*
*   **✅ Strict Validation**: Validates your manifests against official Kubernetes OpenAPI schemas.
    *   *Detects missing required fields, invalid types, and unknown fields.*
*   **🎨 Rich Visualization**: Color-coded terminal output makes it easy to spot additions (green) and deletions (red).
*   **⚡ Fast & Lightweight**: Built in Go for instant execution.

---

## 🛠️ Installation

### Prerequisites
*   Go 1.24 or higher

### Build from Source
```bash
# Clone the repository
git clone https://github.com/YogPandya12/kdiff.git
cd kdiff

# Build the binary
go build -o bin/kdiff ./cmd/kdiff

# (Optional) Add to your PATH
export PATH=$PATH:$(pwd)/bin
```

---

## 📖 Usage

### 1. Compare Configurations
Check the differences between your local development file and production configuration.

```bash
# Basic comparison
kdiff compare ./tests/data/diff_base.yaml ./tests/data/diff_modified.yaml

# Output as a Table (Great for reports)
kdiff compare ./tests/data/diff_base.yaml ./tests/data/diff_modified.yaml --output table

# Output as JSON (Great for CI/CD parsing)
kdiff compare ./tests/data/diff_base.yaml ./tests/data/diff_modified.yaml --output json
```

**Output Example:**
```text
COMMON  | 1 | apiVersion: v1
COMMON  | 2 | kind: ConfigMap
REMOVED | 5 |   key: value1
ADDED   | 5 |   key: value1-modified
```

### 2. Validate Manifests
Ensure your YAML files are valid Kubernetes objects before committing.

```bash
# Validate a single file
kdiff validate ./tests/data/valid_deployment.yaml

# Validate multiple files
kdiff validate ./tests/data/valid_deployment.yaml ./tests/data/invalid_pod.yaml

# Enable strict mode (fails on unknown fields)
kdiff validate --strict ./tests/data/valid_deployment.yaml
```

**Output Example:**
```text
✅ ./tests/data/valid_deployment.yaml is valid against its Kubernetes schema.
❌ ./tests/data/invalid_missing_kind.yaml has validation issues:
  Error: missing 'kind' field (Path: root)
```

---

## 🛣️ Roadmap

We are building a complete ecosystem for Kubernetes Configuration Management:

*   [x] **Phase 1: CLI Tool** (Current) - Local validation and comparison.
*   [ ] **Phase 2: Web Dashboard** - A graphical interface to visualize diffs and manage configs via a browser.
*   [ ] **Phase 3: Kubernetes Operator** - Automated in-cluster drift detection and synchronization.

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1.  Fork the Project
2.  Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3.  Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4.  Push to the Branch (`git push origin feature/AmazingFeature`)
5.  Open a Pull Request
