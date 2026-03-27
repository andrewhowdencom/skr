# Reference: CLI Commands

## `skr`

Root command.

### `skr init`
Initialize a new `skr` project in the current directory.
-   **--agent, -a**: Agent to configure for (e.g. `antigravity`, `gemini`, `claude`, `standard`). (default: `antigravity`).

### `skr build [path] --tag <tag>`
Build an Agent Skill artifact from a directory.
-   **path**: Path to skill directory (default: `.`)
-   **--tag, -t**: Name and optional tag (e.g., `my-skill:v1`).

### `skr install <ref>`
Install a skill into the current project.
-   **ref**: The source reference for the skill. Supported schemas are:
    -   `oci://<repo>:<tag>` or `<repo>:<tag>`: Pulls an OCI artifact (defaults to `latest` if tag is omitted).
    -   `git://<repo>#<refspec>`, `git+https://<repo>#<refspec>`, `git+ssh://<repo>#<refspec>`, `git@<repo>#<refspec>`: Clones a git repository, checks out `<refspec>` (defaults to HEAD if omitted), builds it, and installs it from the local cache. Note: Use the `git+` prefix to distinguish Git over HTTPS from other types of connections.
    -   `file://<path>`: Builds the skill from a local directory.

### `skr validate [path]`
Validate an Agent Skill's structure and metadata.
-   **path**: Path to skill directory (default: `.`).
-   **--fix**: Automatically fix issues where possible.

### `skr lint [path]`
Lint an Agent Skill against specification and style guidelines.
-   **path**: Path to skill directory (default: `.`).
-   **--format**: Output format (`gnu`, `sarif`, `checkstyle`).
-   **--output**: Output file path.
-   **--fail-on**: Categories to fail on (`spec`, `style`).
-   **--fix**: Automatically fix issues where possible.

### `skr list`
List skills installed in the current project or available globally.

### `skr rm <name>`
Remove a skill from the current project configuration.

### `skr sync`
Synchronize the local`.agent/skills` directory with the `.skr.yaml` configuration.

### `skr update [skill]`
Update installed skills to their latest hashes, bypassing the lockfile to fetch the newest manifest.
-   **skill**: The base reference of the skill to update (e.g., `ghcr.io/user/skill`).
-   **--all**: Update all skills listed in `.skr.yaml`.

### `skr publish [path] --tag <tag>`
Build a skill from a directory and immediately push it to a registry.
-   **path**: Path to skill directory (default: `.`)
-   **--tag, -t**: Registry reference (e.g., `ghcr.io/user/skill:v1`).

### `skr batch publish [path]`
Publish multiple skills from a monorepo structure.
-   **path**: Root directory containing skills (default: `.`)
-   **--registry**: Registry host (required).
-   **--namespace**: Registry namespace (required).
-   **--base**: Git reference for change detection (optional, e.g., `origin/main`).


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

---

## `skr http`

HTTP Server commands.

### `skr http serve`
Start a local HTTP server to browse and manage skills.
-   **--port, -p**: Port to listen on (default: `8080`).
-   **--oci-endpoint**: Remote OCI Registry endpoint to proxy (e.g. `https://registry-1.docker.io`).
-   **--auth-file**: Path to a YAML file containing credentials.
-   **--trace-provider**: Trace provider to use (`stdout`, `otlp`, `none`).
-   **--trace-endpoint**: Endpoint for the OTLP trace provider.

