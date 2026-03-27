# How to Install a Skill from a Git Repository

You can install `skr` Agent Skills directly from their upstream Git repositories without needing the author to have explicitly constructed and published an OCI container image.

## Prerequisites
- Git installed on your device.
- The Git repository must contain a valid `SKILL.md` at its root.

## Installation

Run the `skr install` command using explicit Git schemas like `git://`, `git+https://`, `git+http://`, or `git+ssh://`, or `git@`. Using the `git+` prefix strictly specifies that the repository should be cloned using Git. 

```bash
skr install git+https://github.com/andrewhowdencom/skr-example-skill
```

### Specifying a Branch, Tag, or Commit

If you want a specific branch, Git commit SHA, or tag, append it to the URL using the `#` fragment:

```bash
skr install git+https://github.com/andrewhowdencom/skr-example-skill#v1.0.0
```

## Under the Hood
When you pass a Git URL, `skr` executes a standard Git clone to a temporary directory. It then builds an OCI image from that directory within the local cache. Future dependency resolution mapping to the same Git commit and repository will seamlessly utilize the local cached OCI image.
