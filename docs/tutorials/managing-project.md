# Tutorial: Managing Skills in a Project

This tutorial explains how to use `skr` to manage Agent Skills within a specific project repository. You will learn how to configure `.skr.yaml`, install skills, and keep them synchronized.

## Prerequisites

-   `skr` installed.
-   An existing git repository or project folder (e.g., `my-agent-project`).

## 1. Initialize Configuration

Create a `.skr.yaml` file in the root of your project. This file declares the configuration for your agent and the list of skills it requires.

```yaml
# .skr.yaml
agent:
  type: "custom" # or "google-genai", etc.

skills: []
```

## 2. Install a Skill

Use the `install` command to add a skill to your project. This will:
1.  Fetch the skill (if not already local).
2.  Update `.skr.yaml`.
3.  Sync the skill content to `.agent/skills/`.

Let's install a demo skill (assuming one is available in your registry or local store, e.g., `my-skill:v1`).

```bash
skr install my-skill:v1
```

If successful, check your `.skr.yaml`:

```yaml
agent:
  type: "custom"
skills:
  - "my-skill:v1"
```

And verify the skill files were "hydration" into the project:

```bash
ls -F .agent/skills/
# Output: my-skill/
```

## 3. Synchronizing Skills

If you manually edit `.skr.yaml` (e.g., to add a list of skills from another project), or if you clone this repo on a fresh machine, you need to sync the `.agent/skills` directory to match the config.

Run:

```bash
skr sync
```

This ensures that `.agent/skills` contains exactly what is listed in `.skr.yaml`.

## 4. Version Control Guidelines

When using `skr` in a team or CI/CD environment, following these `.gitignore` best practices is recommended:

-   **Commit**: `.skr.yaml` (This is your source of truth).
-   **Ignore**: `.agent/skills/` (These are generated artifacts, similar to `node_modules`).

Add to your `.gitignore`:

```text
.agent/skills/
```

This way, other developers simply clone the repo and run `skr sync` to get all necessary skills.

## Summary

You have set up a project where Agent Skills are managed declaratively via `.skr.yaml`. This provides a reproducible environment for your agents.
