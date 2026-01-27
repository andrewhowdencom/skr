---
name: "skr"
description: "A skill for using the skr CLI tool."
---

# skr Tool Skill

This skill provides instructions for using the `skr` CLI tool to manage Agent Skills.

## Overview

`skr` is the CLI tool for the Agent Skills ecosystem. It handles building, installing, publishing, and managing skills.

## Capabilities

### Build Skills
Build a local skill directory into an artifact.

```bash
skr build [path] --tag <tag>
```

### Install Skills
Install a skill into the current workspace.

```bash
skr install <ref>
```

### Publish Skills
Build and push a skill to a remote registry in one step.

```bash
skr publish [path] --tag <tag>
```

### Batch Publish
Publish multiple skills from a monorepo. This is useful for CI/CD pipelines.

```bash
skr batch publish [path] --registry <host> --namespace <ns> [options]
```

### Manage Registry
Login/Logout and manage remote artifacts.

```bash
skr registry login ghcr.io -u <user> -p <token>
skr registry logout ghcr.io
skr push <ref>
skr pull <ref>
```

### Manage System Store
Inspect and clean up local artifacts.

```bash
skr system list
skr system inspect <ref>
skr system prune
```

## Troubleshooting

-   **Context Errors**: If `skr` fails to find context, ensure you are in a directory with a valid `.skr.yaml` or initialize one.
-   **Registry Auth**: Ensure you are logged in via `skr registry login` before pushing/pulling private skills.
