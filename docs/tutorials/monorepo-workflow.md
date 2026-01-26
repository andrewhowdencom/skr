# Tutorial: Automating Skill Updates in a Monorepo

This guide explains how to manage a repository containing multiple skills (a "monorepo") and automate their deployment using GitHub Actions.

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

To efficiently manage updates, we want to:
1.  **Detect changes**: Only build skills that have been modified in the current push/PR.
2.  **Build & Push**: Build the artifact and push it to a registry (e.g., GitHub Container Registry).
3.  **Versioning**: Use the version defined in `SKILL.md` or git SHAs.

## GitHub Action Workflow

Here is a complete workflow that uses `dorny/paths-filter` to detect changes and a matrix strategy to build them.

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
  detect-changes:
    runs-on: ubuntu-latest
    outputs:
      skills: ${{ steps.filter.outputs.changes }}
    steps:
      - uses: actions/checkout@v4
      
      # Detect which folders under skills/ have changed
      - uses: dorny/paths-filter@v3
        id: filter
        with:
          filters: |
            git: skills/git/**
            docker: skills/docker/**
            text-utils: skills/text-utils/**
            # Note: For dynamic discovery, you might need a custom script step 
            # instead of hardcoding filters if you have many skills.

  build-and-push:
    needs: detect-changes
    if: needs.detect-changes.outputs.skills != '[]'
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    strategy:
      matrix:
        skill: ${{ fromJSON(needs.detect-changes.outputs.skills) }}
        
    steps:
    - uses: actions/checkout@v4
    
    # Install skr using the official action
    - uses: andrewhowdencom/skr@main
        
    - name: Log in to the Container registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
        
    - name: Build and Push
      run: |
        SKILL_NAME=${{ matrix.skill }}
        SKILL_PATH="skills/${SKILL_NAME}"
        
        # Read version from SKILL.md (optional, requires yq)
        # VERSION=$(yq '.version' ${SKILL_PATH}/SKILL.md)
        
        # Or use Git SHA for immutable tags
        VERSION="sha-$(git rev-parse --short HEAD)"
        
        TAG="${{ env.REGISTRY }}/${{ env.NAMESPACE }}/${SKILL_NAME}:${VERSION}"
        LATEST="${{ env.REGISTRY }}/${{ env.NAMESPACE }}/${SKILL_NAME}:latest"
        
        echo "Building $TAG..."
        
        # Login to skr registry (wraps docker/login or uses keyring)
        # For CI, we can use the --password-stdin method if needed, 
        # or rely on docker-credential-helpers file if skr supports it.
        # Since 'skr registry login' sets up the keyring, in CI we might need a 
        # simpler auth method or just reuse the docker config if `oras` supports it.
        # Currently skr uses its own keyring.
        
        echo "${{ secrets.GITHUB_TOKEN }}" | skr registry login ${{ env.REGISTRY }} -u ${{ github.actor }} --password-stdin
        
        skr build "${SKILL_PATH}" --tag "$TAG"
        skr registry tag "$TAG" "$LATEST" # Requires 'skr tag' command
        
        skr push "$TAG"
        skr push "$LATEST"
```

> **Note**: This workflow assumes you list your skills in the `paths-filter` step. For a fully dynamic approach (scanning the directory), you would replace the `detect-changes` job with a script that outputs a JSON array of changed directory names.

## Dynamic Discovery Script (Alternative)

If you have many skills and don't want to update the YAML for every new one:

```bash
# Get list of changed files
CHANGED_FILES=$(git diff --name-only ${{ github.event.before }} ${{ github.sha }})

# Extract unique skill directories
SKILLS=$(echo "$CHANGED_FILES" | grep '^skills/' | cut -d/ -f2 | sort -u | jq -R -s -c 'split("\n")[:-1]')

echo "skills=$SKILLS" >> $GITHUB_OUTPUT
```

## Summary

1.  Structure your repo with isolated skill folders.
2.  Use a GitHub Action to detect changes.
3.  Use `matrix` strategy to parallelize builds.
4.  Push tagged artifacts to GHCR.
