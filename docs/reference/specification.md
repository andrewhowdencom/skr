# Agent Skills Specification

This document outlines the directory structure and file requirements for an Agent Skill, compatible with `skr`.

## Directory Structure

A valid Agent Skill directory must contain a `SKILL.md` file at its root.

```
my-skill/
├── SKILL.md          (Required)
├── references/       (Optional, for technical details)
├── scripts/          (Optional, for executable tools)
└── assets/           (Optional, for static files)
```

## SKILL.md

The `SKILL.md` file is the entry point. It must contain YAML frontmatter.

### Frontmatter

```yaml
---
name: "my-skill"
description: "A brief description of the skill."
---
```

- **name**: [Required] 1-64 characters, lowercase alphanumeric and hyphens. Should match the directory name.
- **description**: [Required] 1-1024 characters.

### Body

The body of the markdown file should contain the instructions and capabilities provided by the skill.
