# How-to: Manage Local System

`skr` maintains a local OCI-compliant store at `~/.local/share/skr/store` (Linux) or equivalent XDG data path.

## Listing Artifacts

See all artifacts currently in your local store:

```bash
skr system list
```

## Inspecting Artifacts

View detailed metadata (manifest, config, annotations) for an artifact:

```bash
skr system inspect <tag-or-digest>
```

## Removing Artifacts

To remove a specific tag (and its reference) from the local store:

```bash
skr system rm <tag>
```

You can remove multiple tags at once:

```bash
skr system rm tag1:v1 tag2:v1
```

## Pruning (Garbage Collection)

Removing a tag with `rm` does not immediately delete the underlying content blobs (layers), just the reference. To free up space by removing unreferenced blobs:

```bash
skr system prune
```
