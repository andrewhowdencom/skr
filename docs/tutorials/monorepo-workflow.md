# Tutorial: Automating Skill Updates in a Monorepo

This guide explains how to manage a repository containing multiple skills (a "monorepo") and automate their deployment using the official `skr` GitHub Action.

## Repository Structure

We recommend a simple directory structure where each skill resides in its own folder under `skills/`:

```text
my-skills-repo/
├── .github/
│   └── workflows/
│       └── release.yaml
├── skills/
│   ├── git/
│   │   └── SKILL.md
│   ├── docker/
│   │   └── SKILL.md
│   └── text-utils/
│       └── SKILL.md
└── README.md
```

## Automation Strategy

To efficiently manage updates, we use `skr batch publish`. This command:
1.  **Detects changes**: Checks which `skills/*` directories have changed since the last commit.
2.  **Builds**: Creates OCI artifacts for modified skills.
3.  **Pushes**: Uploads them to the registry.

## GitHub Action Workflow

We provide a Docker-based GitHub Action that simplifies this process significantly. You don't need to write complex scripts or matrix strategies.

### `.github/workflows/deploy-skills.yaml`

```yaml
name: Deploy Skills

on:
  push:
    branches: [ "main" ]
    paths:
      - 'skills/**'

env:
  REGISTRY: ghcr.io
  NAMESPACE: ${{ github.repository_owner }}

jobs:
  publish:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0 # Required for change detection

    - name: Publish Skills
      uses: andrewhowdencom/skr@main
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
        namespace: ${{ env.NAMESPACE }}
        base: ${{ github.event.before }}
        path: ./skills
```

## How It Works

1.  **Change Detection**: The action (via `skr batch publish`) compares the current commit against `base` (the previous commit `github.event.before`) to find modified skills.
2.  **Build & Push**: It automatically builds and pushes artifacts for changed skills to the registry.

## Manual Publishing

You can also run this locally using the CLI:

```bash
skr batch publish ./skills \
  --registry ghcr.io \
  --namespace myuser \
  --base origin/main
```
