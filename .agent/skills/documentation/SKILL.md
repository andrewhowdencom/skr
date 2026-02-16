---
name: documentation
description: A skill for writing and maintaining technical documentation using the Diátaxis framework.
---

# Documentation Skill

This skill provides instructions for creating, maintaining, and structuring technical documentation within this repository. It enforces a clear separation between **Developer Facing** documentation and **User Facing** documentation, following the [Diátaxis framework](https://diataxis.fr).

## Core Principles

1.  **Documentation is Code**: Treat documentation with the same rigor as code. Review it, test it (verify links/formatting), and version control it.
2.  **User-Centric**: Structure documentation based on *user needs* (learning, solving problems, understanding concepts, looking up information).
3.  **Single Source of Truth**: Avoid duplication. Link to existing documentation rather than repeating it.

## Documentation Structure

### 1. Developer Facing (Root)

These files reside in the repository root and target contributors, developers, and AI agents.

| File | Purpose | Audience |
| :--- | :--- | :--- |
| `README.md` | **Project Entry Point**. High-level overview, quick start, installation, and links to user documentation. | Everyone |
| `ARCHITECTURE.md` | **System Design**. Explains core technologies, interfaces, data flow, and architectural decisions. | Developers, Architects |
| `DEVELOPMENT.md` | **Contribution Guide**. Instructions for setting up the dev environment, running tests, and submitting PRs. | Contributors |
| `AGENTS.md` | **Agent Context**. Specific instructions or context for AI agents (e.g., skill locations, specialized workflows). | AI Agents |

### 2. User Facing (`docs/`) - The Diátaxis Framework

User-facing documentation must be located in the `docs/` directory and structured according to the four Diátaxis quadrants. This structure ensures compatibility with `mkdocs`.

#### 🎓 Tutorials (`docs/tutorials/`)
*   **Goal**: Learning-oriented.
*   **Format**: Step-by-step lessons.
*   **Content**: "Let's build X together." Focus on acquiring new skills.
*   **Style**: Instructional, hand-holding, no gaps.

#### 👐 How-to Guides (`docs/how-to/`)
*   **Goal**: Problem-oriented.
*   **Format**: Steps to complete a specific task.
*   **Content**: "How do I do Y?" Focus on practical application.
*   **Style**: Action-oriented, concise, assumes some basic knowledge.

#### 📖 Reference (`docs/reference/`)
*   **Goal**: Information-oriented.
*   **Format**: Technical descriptions, APIs, CLI commands.
*   **Content**: "What is Z?" Focus on accuracy and completeness.
*   **Style**: Descriptive, structured, dry.

#### 💡 Explanation (`docs/explanation/`)
*   **Goal**: Understanding-oriented.
*   **Format**: Essays, background, concepts.
*   **Content**: "Why did we choose W?" Focus on context and clarity.
*   **Style**: Discursive, explanatory.

## Workflow

### When to Update Documentation

*   **New Feature**:
    *   Add a **Tutorial** if it's a major new capability.
    *   Add **How-to Guides** for common use cases.
    *   Update **Reference** with new API/CLI details.
    *   Add **Explanation** if it involves complex new concepts.
*   **Refactor/Architectural Change**:
    *   Update `ARCHITECTURE.md`.
*   **Build/Test Change**:
    *   Update `DEVELOPMENT.md`.
*   **Bug Fix**:
    *   Update **How-to** or **Reference** if the behavior was ambiguous or incorrect.

### Definition of Done

Documentation is **not** an optional afterthought. A task is considered complete only when:

1.  **Code is Documented**: Relevant developer docs (`README`, `ARCHITECTURE`, `DEVELOPMENT`) are updated.
2.  **User Docs are Updated**: Diátaxis quadrants (`tutorials`, `how-to`, `reference`, `explanation`) are updated if user-facing behavior changed.
3.  **Local Verification**: Documentation builds locally (if applicable) and links are valid.

**Review Requirement**: Pull Requests must include documentation updates. If no documentation was needed, the PR description should explicitly state why.

### Formatting
*   Use standard Markdown.
*   Use relative links for internal navigation (e.g., `[Link](../how-to/guide.md)`).
*   Ensure compatibility with `mkdocs` (e.g., start with an `index.md` in `docs/` and subdirectories).

### Validation
*   For `SKILL.md` files (which are technically documentation + metadata), use `skr validate [path]` to check for structural correctness.
