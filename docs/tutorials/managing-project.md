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

Let's install a demo skill directly from a Git repository. `skr` uses Git as the primary pathway for distributing skills.

```bash
skr install git+https://github.com/andrewhowdencom/skr-example-skill
```

If successful, check your `.skr.yaml`:

```yaml
agent:
  type: "custom"
skills:
  - "git+https://github.com/andrewhowdencom/skr-example-skill"
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

This ensures that `.agent/skills` contains exactly what is listed in `.skr.yaml`. If you have a `.skr.lock` file, it will reliably resolve tags to the exact image digests that were originally installed, creating deterministic developer environments.

## 4. Version Control Guidelines

When using `skr` in a team or CI/CD environment, following these `.gitignore` best practices is recommended:

-   **Commit**: `.skr.yaml` and `.skr.lock` (These are your source of truth and determine reproducible versions).
-   **Ignore**: `.agent/skills/` (These are generated artifacts, similar to `node_modules`).

Add to your `.gitignore`:

```text
.agent/skills/
```

This way, other developers simply clone the repo and run `skr sync` to get all necessary skills.

## Summary

You have set up a project where Agent Skills are managed declaratively via `.skr.yaml`. This provides a reproducible environment for your agents.
