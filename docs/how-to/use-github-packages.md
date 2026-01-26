# Using GitHub Packages with `skr`

`skr` supports any OCI-compliant registry, including GitHub Packages (GHCR).

## Prerequisites

To push or pull skills from GitHub Packages, you need a Personal Access Token (PAT).

1.  Go to **GitHub Settings** -> **Developer Settings** -> **Personal access tokens** -> **Tokens (classic)**.
2.  Generate a new token.
3.  Select the `write:packages` scope (for pushing) and `read:packages` (for pulling).
4.  Copy the token.

## Login

Use `skr` to log in to `ghcr.io`:

```bash
skr registry login ghcr.io --username <your-github-username>
# Paste your PAT when prompted for password
```

## Pushing a Skill

1.  Build your skill with the full registry tag:

```bash
skr build ./my-skill --tag ghcr.io/<your-username>/my-skill:v1
```

2.  Push it:

```bash
skr push ghcr.io/<your-username>/my-skill:v1
```

## Pulling a Skill

To install a skill from GitHub Packages:

```bash
skr install ghcr.io/<your-username>/my-skill:v1
```

Or just pull it to your local store:

```bash
skr pull ghcr.io/<your-username>/my-skill:v1
```
