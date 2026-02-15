---
name: git
description: Guidelines for validation, staging, amending, and commit messages.
---

# Git

## Validate Changes
Before committing any changes, ensure that your work does not break the application and meets all quality standards.
Run the `validate` task just prior to commit:

```bash
task validate
```

## Stage Changes
Stage changes by reviewing specific changes applied in files using patch mode:

```bash
git add --patch ./path/to/file
```

## Amend Commits
If you need to update the previous commit (e.g., to squash changes or fix a mistake), append changes to the last commit:

```bash
git commit --amend
```

If you have already pushed, you will need to force push:
```bash
git push --force-with-lease
```

## Write Commit Messages
Commit messages should follow the standard format:

#### Title
Maximally 72 characters, descriptive, using imperative mood (e.g., "Deploy changes automatically" not "Deployed changes").

#### Body
Explain **why** the change was made (justification), not just **what** changed. Use headers for structured context if applicable:

```
Design:
  Describe the design of the change here.

Tradeoffs:
  Explain any tradeoffs (performance, complexity, etc).

Justification:
  Provide the reasoning for this change.
```

#### Co-author
If pair programming or crediting others, add co-authors at the end of the body:

```
Co-authored-by: NAME <NAME@EXAMPLE.COM>
```
