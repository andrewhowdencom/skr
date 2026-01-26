# Reference: CLI Commands

## `skr`

Root command.

### `skr build [path] --tag <tag>`
Build an Agent Skill artifact from a directory.
-   **path**: Path to skill directory (default: `.`)
-   **--tag, -t**: Name and optional tag (e.g., `my-skill:v1`).

### `skr install <ref>`
Install a skill into the current project.
-   **ref**: Tag or digest of the skill (e.g., `ghcr.io/user/skill:v1`).

### `skr list`
List skills installed in the current project or available globally.

### `skr rm <name>`
Remove a skill from the current project configuration.

### `skr sync`
Synchronize the local`.agent/skills` directory with the `.skr.yaml` configuration.

---

## `skr registry`

Manage registry interactions.

### `skr registry login <server>`
Log in to a registry.
-   **server**: Registry address (default: `ghcr.io`).
-   **--username, -u**: Registry username.
-   **--password, -p**: Registry password/token.
-   **--password-stdin**: Read password from stdin.

### `skr registry logout <server>`
Log out from a registry.

### `skr push <ref>`
Push an artifact to a remote registry.

### `skr pull <ref>`
Pull an artifact from a remote registry to the local store.

---

## `skr system`

Manage the local system store.

### `skr system list`
List all artifacts (tags) in the local store.

### `skr system inspect <ref>`
View metadata for a specific artifact.

### `skr system rm <ref>...`
Remove one or more artifact references (tags) from the local store.

### `skr system prune`
Delete unreferenced blobs (garbage collection) to free space.
