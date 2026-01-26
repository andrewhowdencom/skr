# Skill Registry (skr)

**skr** is a tool designed to manage [Agent Skills](https://agentskills.io/what-are-skills)—a lightweight, open format for extending AI agent capabilities with specialized knowledge and workflows.

## Goal

The primary goal of **skr** is to enable the pushing, pulling, and maintenance of Agent Skills using OCI (Open Container Initiative) Image registries, similar to the workflow used for container images on platforms like GitHub or Docker Hub.

This allows for a standardized, versioned, and distributed ecosystem for sharing AI capabilities.

This allows for a standardized, versioned, and distributed ecosystem for sharing AI capabilities.

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
# Build a skill from the current directory
skr build . -t my-registry.com/my-skill:v1.0.0

# Push a skill to a registry
skr push my-registry.com/my-skill:v1.0.0

# Install a skill
skr install my-registry.com/my-skill:v1.0.0

# Remove an installed skill
skr rm my-registry.com/my-skill

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
1. [Codex Skills](https://developers.openai.com/codex/skills/)
1. [Antigravity Skills](https://antigravity.google/docs/skills)
1. [Agent Skills](https://agentskills.io/what-are-skills)

