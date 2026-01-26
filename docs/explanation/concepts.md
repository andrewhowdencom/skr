# Explanation: Core Concepts

## Agent Skill

An **Agent Skill** is a packaged unit of capability for an AI Agent. It is defined by a `SKILL.md` and may contain supporting files (scripts, prompts, schemas).

`skr` treats skills like container images. They are:
-   **Versioned**: Using tags (e.g., `v1.0.0`, `latest`).
-   **Portable**: Can be pushed to standard OCI registries.
-   **Layered**: Deduplicated storage of content.

## OCI Store

`skr` uses the **Open Container Initiative (OCI)** specifications for storage and distribution.
-   **Manifest**: Describes the skill content (config + layers).
-   **Config**: Metadata about the skill (creation time, author, etc.).
-   **Layers**: The actual file content (tar.gz).

This compatibility allows `skr` to work with existing infrastructure like GitHub Packages (`ghcr.io`), Docker Hub, Harbor, etc.

## The Global vs. Local Scope

-   **System Store**: A global cache of all downloaded/built artifacts on your machine.
-   **Project Scope**: When you run `skr install`, skills are "installed" into your project (referenced in `.skr.yaml` and synced to `.agent/skills`).
