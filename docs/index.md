# Skill Registry (skr)

**skr** is a tool designed to manage [Agent Skills](https://agentskills.io/what-are-skills)—a lightweight, open format for extending AI agent capabilities with specialized knowledge and workflows.

## Goal

The primary goal of **skr** is to enable the pushing, pulling, and maintenance of Agent Skills using OCI (Open Container Initiative) Image registries, similar to the workflow used for container images on platforms like GitHub or Docker Hub.

This allows for a standardized, versioned, and distributed ecosystem for sharing AI capabilities.

## Usage

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
2. [The complete guide to building Claude skills](https://resources.anthropic.com/hubfs/The-Complete-Guide-to-Building-Skill-for-Claude.pdf)
3. [Codex Skills](https://developers.openai.com/codex/skills/)
4. [Antigravity Skills](https://antigravity.google/docs/skills)
5. [Gemini CLI Skills](https://geminicli.com/docs/skills/)
6. [Agent Skills](https://agentskills.io/what-are-skills)

