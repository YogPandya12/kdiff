# kdiff User Guide

This guide provides detailed instructions on how to use `kdiff` to solve common Kubernetes configuration challenges.

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

## 📖 Command Reference

### 1. `kdiff compare`
**Check the differences between two local YAML files.**

#### 💡 Use Case
You have a `deployment.yaml` and a `deployment-new.yaml`. You want to know exactly what changed before you commit the new version. Standard `diff` shows you raw text, but `kdiff` shows you semantic changes.

#### 🛡️ How it Helps
*   **Visual Clarity**: Instantly spot if you accidentally changed a port number or an environment variable.
*   **CI/CD Integration**: Use JSON output to programmatically check for forbidden changes in your pipeline.
*   **Reporting**: Generate Markdown tables to paste into Pull Request descriptions.

#### Usage
```bash
kdiff compare <file1> <file2> [flags]
```

#### Examples
```bash
# Basic comparison
kdiff compare ./base.yaml ./modified.yaml

# Output as a Table (Great for PR comments)
kdiff compare ./base.yaml ./modified.yaml --output table

# Output as JSON (Great for automated policy checks)
kdiff compare ./base.yaml ./modified.yaml --output json
```

---

### 2. `kdiff validate`
**Ensure your YAML files are valid Kubernetes objects before committing.**

#### 💡 Use Case
You are writing a new manifest from scratch. You want to make sure you didn't misspell `replicas` as `replica` or forget the `selector` field in a Service.

#### 🛡️ How it Helps
*   **Shift Left**: Catch errors on your laptop, not in the CI pipeline or (worse) during deployment.
*   **Schema Compliance**: Ensures your manifests strictly adhere to the official Kubernetes OpenAPI spec.
*   **Strict Mode**: Prevents "garbage" fields that Kubernetes ignores but might confuse other developers.

#### Usage
```bash
kdiff validate <file> [flags]
```

#### Examples
```bash
# Validate a single file
kdiff validate ./deployment.yaml

# Enable strict mode (fails on unknown fields)
kdiff validate --strict ./deployment.yaml
```

---

### 3. `kdiff live`
**Compare a local manifest against the live resource in your cluster.**

#### 💡 Use Case
You are about to apply a change to production. You want to verify that the *only* thing changing is what you expect. You also want to see if someone manually changed the cluster (e.g., scaled it up) without updating the code.

#### 🛡️ How it Helps
*   **Pre-Deployment Safety**: Acts as a "dry run" with better visualization than `kubectl diff`.
*   **Noise Reduction**: Automatically strips out dynamic fields like `status`, `uid`, and `managedFields` so you only see real changes.
*   **Verification**: Confirms that your local file actually matches the resource you think it does.

#### Usage
```bash
kdiff live <file>
```

#### Example
```bash
# Compare local file vs live cluster
kdiff live ./nginx-deployment.yaml
```

---

### 4. `kdiff cluster`
**Compare resources between two different clusters or contexts.**

#### 💡 Use Case
You have a `staging` cluster and a `production` cluster. They *should* be identical, but you suspect they aren't. You want to find out exactly what is different.

#### 🛡️ How it Helps
*   **Environment Parity**: Quickly audit differences between environments to debug "it works in staging but fails in prod" issues.
*   **Migration Verification**: When migrating clusters, verify that the new cluster has the exact same resources as the old one.
*   **Targeted Auditing**: Filter by namespace or labels to check specific applications.

#### Usage
```bash
kdiff cluster <context1> <context2> [flags]
```

#### Examples
```bash
# Compare default namespace between two contexts
kdiff cluster context-staging context-prod

# Compare specific resources
kdiff cluster context-staging context-prod -r Deployment,Service
```

---

### 5. `kdiff drift`
**Check an entire directory of local manifests against the live cluster.**

#### 💡 Use Case
You manage your infrastructure as code (GitOps). You want to know if the actual state of the cluster has "drifted" from what is defined in your git repository (e.g., someone did a `kubectl edit` manually).

#### 🛡️ How it Helps
*   **GitOps Integrity**: Ensures your git repo remains the single source of truth.
*   **Security Audit**: Detects unauthorized manual changes to the cluster.
*   **Batch Processing**: Checks hundreds of files at once, saving you from running `kdiff live` manually for every file.
*   **Audit Trail**: Logs all detected drift events to `drift.log` for historical analysis.

#### Usage
```bash
kdiff drift <directory>
```

#### Example
```bash
# Check for drift
kdiff drift ./deployments/prod
```
