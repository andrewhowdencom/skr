# How-to: Managing Skill Dependencies with Lockfiles

When developing AI agent skills recursively or maintaining dependent systems, the unpredictability of remote versions like `latest` or floating tag versions can lead to difficult tracking of regression errors.

`skr` guarantees reproducibility by keeping a `.skr.lock` file alongside your declarative `.skr.yaml`. 

This guide demonstrates how to properly manage versioning with `skr`.

## Inspecting the Lockfile

When you use the command `skr install <repository-url>:<tag>`, `skr` automatically creates a corresponding mapped digest in the `.skr.lock` file (usually an SHA-256 identifier containing the explicit manifest of that repository state).

An example `.skr.lock` might look like this:
```yaml
skills:
  ghcr.io/andrewhowdencom/skills.git:latest: sha256:4a001a1dbba4e55e0ef3...
  ghcr.io/andrewhowdencom/skills.go:1.0: sha256:d8c6b75ffcb3a233ecbc...
```

Whenever the above skills are fetched using `skr sync`, the process strictly routes through `sha256:...` internally to avoid "floating tag" drift.

## Updating a Locked Skill

If you have downloaded a skill utilizing a floating tag equivalent like `:latest` or `:1.0` (which generally maps to `1.0.X` under semver conventions), the locked digest won't update itself automatically when a remote author updates the tag. 

To upgrade to the newest manifestation of a floating tag, simply use the `update` command:

```bash
# Update a specific skill by matching its base repository string
skr update ghcr.io/andrewhowdencom/skills.git

# Or update all skills in your configuration simultaneously
skr update --all
```

This bypasses the local lock, fetches the freshest manifest from the remote environment matching your `.skr.yaml` tags, and recalculates a new deterministic lock digest inside `.skr.lock`.

## Checking in the Lockfile

It is highly recommended that you check **both** the `.skr.yaml` configuration and the `.skr.lock` cache tracking logs into your source version control system like Git.
```bash
git add .skr.yaml .skr.lock
git commit -m "chore: hydrate skill locks"
```
This forces identical execution patterns for every remote or teammate environment.
