# Architecture overview

`skr` is designed to streamline the usage and distribution of Agent Skills via OCI artifact packaging conventions.

## 1. Skill Sources (`pkg/source`)

The `pkg/source` package isolates how Agent Skills are requested, fetched, and transformed. Instead of assuming every skill exists in a remote OCI registry, `skr` supports a generic `Fetcher` abstraction logic:

- **Reference Parsing**: A unified parser `Reference.go` examines paths/URIs and divides them into schemas (`file`, `git`, `oci`), origin paths, and specifications (e.g. Git refspec or OCI tag).
- **Git Fetcher (`source/git.go`)**: Delegates cloning of standard HTTP/Git repositories to the underlying `git` command-line tools. The cloned repositories are verified and converted internally into an OCI image inside the `store` module. 
- **File Fetcher (`source/file.go`)**: Bypasses external network calls entirely, targeting local directories and constructing local OCI images directly from disk.
- **OCI Fetcher (`source/oci.go`)**: Functions as the primary pass-through. Uses standard ORAS (`registry.Pull`) logic to resolve artifacts from external container registries natively.

Because every source resolves dynamically into the local `pkg/store` (acting as a local container registry buffer), tasks like validation, installation, and inspection down the pipeline need only to concern themselves with standardized OCI layers and digests.

## 2. Dependency Resolution (`pkg/resolution`)

Dependencies specified formally in `SKILL.md` interact with the generic `pkg/source` module. Before attempting to link a reference internally using OCI conventions, the `Resolver` runs it through the `Fetcher` logic. This ensures transitive dependencies expressed as absolute GitHub repositories or relative file paths are fetched, built contextually, and stored alongside their consuming root application logically.

## 3. Storage (`pkg/store`)

The internal datastore heavily employs `oras.land/oras-go` wrappers to expose a `xdg`-derived memory and disk abstraction of an OCI-compliant cache directory. `st.Build()` constructs canonical container manifests dynamically avoiding dependency on standard Docker daemons.
