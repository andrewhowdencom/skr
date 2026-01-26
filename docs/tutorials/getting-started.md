# Getting Started with `skr`

This tutorial will guide you through creating, building, and running your first Agent Skill with `skr`.

## Prerequisites

-   `git` installed
-   `skr` installed (see README)
-   Access to a GitHub repository (options)

## 1. Create a Skill

Create a new directory for your skill:

```bash
mkdir my-skill
cd my-skill
```

Create a `SKILL.md` file. This is the definition of your skill.

```markdown
---
name: my-skill
description: A friendly greeting skill
version: 0.1.0
---

# My Skill

This skill allows agents to say hello.
```

## 2. Build the Skill

Package your skill into an OCI artifact.

```bash
skr build . --tag my-skill:v1
```

You should see output indicating success:
```
Successfully built artifact for skill 'my-skill'
Tagged as: my-skill:v1
```

## 3. Inspect the Artifact

Verify the artifact is in your local system store.

```bash
skr system inspect my-skill:v1
```

You'll see metadata including the size, digest, and creation time.

## 4. Install the Skill

Install the skill to your agent configuration.

```bash
skr install my-skill:v1
```

This updates your `.skr.yaml` and synchronizes the skill to `.agent/skills`.

## Next Steps

-   [Publish your skill to a registry](../how-to/manage-registry.md)
-   [Explore system management](../how-to/manage-system.md)
