# Skill Registry (skr)

**skr** is a tool designed to manage [Agent Skills](https://agentskills.io/what-are-skills)—a lightweight, open format for extending AI agent capabilities with specialized knowledge and workflows.
 
> 📘 **Documentation**: [https://andrewhowdencom.github.io/skr/](https://andrewhowdencom.github.io/skr/)

## Goal

The primary goal of **skr** is to enable developers to natively install and manage Agent Skills directly from standard Git repositories. Under the hood, `skr` automatically creates OCI (Open Container Initiative) artifacts locally, allowing you to optionally leverage standard OCI registries for faster, pre-built distribution of your skills across environments like GitHub Packages or Docker Hub.

By positioning Git as the primary distribution method and OCI as an enhanced performance layer, `skr` fosters a standardized, versioned, and easily accessible ecosystem for sharing AI capabilities.

## Agent Skills

by default, `skr` uses the [Agent Skill](https://agentskills.io) standard to extend its capabilities.

This repository includes two core skills that also serve as the **canonical examples** for how to build and maintain skills:

*   [**builder**](skills/builder/SKILL.md): Instructions on how to build, test, and maintain Agent Skills.
*   [**skr**](skills/skr/SKILL.md): Instructions on how to use the `skr` CLI tool.

### Examples

Since `skr` is self-hosting, the best examples are the files in this repository:

*   **Skill Structure**: See `skills/builder/SKILL.md`
*   **Workflow**: See `.github/workflows/publish-skills.yaml` for a production-ready GitHub Actions workflow to publish skills.
*   **Configuration**: See `.skr.yaml` for an extensively documented configuration example.
*   **Reproducibility**: See `.skr.lock` as the generated lockfile storing guaranteed immutable tag hashes.

## Installation
 
You can install `skr` by downloading the pre-compiled binary from the [Releases page](https://github.com/andrewhowdencom/skr/releases).
 
### Linux / macOS
 
1. Download the archive for your platform.
2. Extract the binary.
3. Move it to a directory in your `PATH`.
 
```bash
# Example for macOS ARM64 (replace version and platform as needed)
VERSION=v0.0.3
wget https://github.com/andrewhowdencom/skr/releases/download/${VERSION}/skr_Darwin_arm64.tar.gz
tar xvf skr_Darwin_arm64.tar.gz
sudo mv skr /usr/local/bin/
```
 
## GitHub Action

Detailed documentation: [Monorepo Workflow Tutorial](docs/tutorials/monorepo-workflow.md)

Use `skr` in your CI/CD pipelines to automatically build and publish skills:

```yaml
steps:
  - uses: actions/checkout@v4
    with:
      fetch-depth: 0 # Required for change detection

  - name: Publish Skills
    uses: andrewhowdencom/skr@main
    with:
      registry: ghcr.io
      username: ${{ github.actor }}
      password: ${{ secrets.GITHUB_TOKEN }}
      namespace: ${{ github.repository_owner }}
      # Optional: path to skills directory (default: .)
      path: ./skills
      # Optional: git ref for change detection (default: none)
      base: ${{ github.event.before }}
```

## CLI Usage

### Basic Commands

```bash
# Validate a skill (checks structure and syntax)
skr validate .

# Build a skill from the current directory
skr build . -t my-registry.com/my-skill:v1.0.0

# Push a skill to a registry
skr push my-registry.com/my-skill:v1.0.0

# Install a skill from a Git repository (default)
skr install git+https://github.com/andrewhowdencom/skr-example-skill

# Install a skill from a local directory (development)
skr install file://./my-skill

# Install a pre-built skill from an OCI registry (fast path)
skr install oci://ghcr.io/user/skill:v1.0.0

# Remove an installed skill
skr rm git+https://github.com/andrewhowdencom/skr-example-skill

# Inspect a remote skill
skr inspect my-registry.com/my-skill:v1.0.0
```

### Registry Authentication

```bash
# Log in to a registry
skr registry login my-registry.com --username <user> --password <token>

# Log out
skr registry logout my-registry.com
```

## Further Reading

1. [anthropic/skills](https://github.com/anthropics/skills)
2. [The complete guide to building Claude skills](https://resources.anthropic.com/hubfs/The-Complete-Guide-to-Building-Skill-for-Claude.pdf)
3. [Codex Skills](https://developers.openai.com/codex/skills/)
4. [Antigravity Skills](https://antigravity.google/docs/skills)
5. [Gemini CLI Skills](https://geminicli.com/docs/skills/)
6. [Agent Skills](https://agentskills.io/what-are-skills)

