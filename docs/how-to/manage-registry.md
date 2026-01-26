# How-to: Manage Registry Interactions

`skr` allows you to push and pull Agent Skills from OCI registries like GitHub Packages, Docker Hub, or bespoke OCI registries.

## Authentication

Before interacting with a private registry, you must log in.

```bash
skr registry login ghcr.io --username <your-username>
```

You will be prompted for a password (or PAT for GitHub). Credentials are stored securely in your system keyring.

### Using Stdin for CI/CD

For automated environments, you can pipe the password:

```bash
echo $CR_PAT | skr registry login ghcr.io -u <user> --password-stdin
```

## Pushing a Skill

Once built, you can push a skill to a registry.

1.  **Build with the full registry tag:**

    ```bash
    skr build ./my-skill --tag ghcr.io/myuser/my-skill:v1
    ```

    *Note: `skr build` automatically detects your git remote and adds it as a source annotation for GitHub Packages.*

2.  **Push:**

    ```bash
    skr push ghcr.io/myuser/my-skill:v1
    ```

## Pulling a Skill

You can download a skill without installing it (useful for inspection):

```bash
skr pull ghcr.io/myuser/my-skill:v1
```

To install directly (which also pulls if needed):

```bash
skr install ghcr.io/myuser/my-skill:v1
```
