---
name: "builder"
description: "A skill for building and maintaining Agent Skills."
author: "Andrew Howden"
version: "0.1.0"
---

# Builder Skill

This skill provides instructions for creating, building, and maintaining Agent Skills using the `skr` tool.

## Concepts

An **Agent Skill** is a directory containing standardized knowledge and capabilities for an AI agent. It is packaged as an OCI artifact (like a Docker image) and distributed via container registries.

### Directory Structure

A valid skill directory must contain a `SKILL.md` file at its root.

```
my-skill/
├── SKILL.md          (Required) Entry point and instructions.
├── references/       (Optional) Supporting documentation.
├── scripts/          (Optional) Executable tools.
└── assets/           (Optional) Static files.
```

### SKILL.md Format

The `SKILL.md` file is the core of the skill. It MUST strictly follow this format:

1.  **Frontmatter**: YAML block defining metadata.
2.  **Body**: Markdown content defining the skill's capabilities.

```markdown
---
name: "my-skill"
description: "A brief description of the skill."
---

# My Skill

Detailed instructions on how to use this skill...
```

**Frontmatter Rules:**
-   `name`: [Required] 1-64 characters, lowercase alphanumeric and hyphens. Should match the directory name.
-   `description`: [Required] 1-1024 characters.

## Capabilities

### Create a New Skill

To create a new skill, create a directory and a `SKILL.md` file.

1.  Create directory: `mkdir path/to/skill-name`
2.  Create `SKILL.md` with appropriate frontmatter.

### Build a Skill

To package a skill into an artifact:

```bash
skr build path/to/skill-name --tag my-skill:v1
```

### Best Practices

-   **Inlined SOPs**: Include Standard Operating Procedures (SOPs) directly in `SKILL.md` for zero-latency execution.
-   **Atomic Skills**: Keep skills focused on a single domain or capability.
-   **References**: Use the `references/` directory for deep-dive technical documentation to keep `SKILL.md` concise.
