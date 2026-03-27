# How to Install a Skill from a Local Directory

While developing Agent Skills, or when you have the source locally but haven't published it to a registry, you can install the skill directly from its local directory. `skr` will transparently build the local source into an OCI container image using its local store.

## Prerequisites
- A local directory containing an `skr` project (valid `SKILL.md`).

## Installation

Run `skr install` with the `file://` schema syntax, followed by either the absolute or relative path to the skill directory:

```bash
skr install file://./my-local-skill
```

Or an absolute path:

```bash
skr install file:///home/user/work/my-local-skill
```

## What Happens Under the Hood?
1. **Building**: `skr` builds the directory's contents into an OCI image cache (`.config/agent/skills/store`).
2. **Tagging**: The resulting OCI image is tagged using the skill's name and version (`latest` by default).
3. **Extraction**: The layer is unpacked logically into your project's `.agent/skills/` directory.

Once completed, testing or running the skill locally is no different from using one fetched from an OCI registry.
