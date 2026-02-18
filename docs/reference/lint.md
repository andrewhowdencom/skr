# Lint

The `lint` command validates an Agent Skill against specification and style guidelines. It provides more comprehensive checks than the `validate` command and supports multiple output formats for integration with CI/CD systems.

## Usage

```bash
skr lint [path] [flags]
```

If `[path]` is not provided, it defaults to the current directory.

## Flags

| Flag | Shorthand | Description | Default |
|------|-----------|-------------|---------|
| `--format` | | Output format (`gnu`, `sarif`, `checkstyle`) | `gnu` |
| `--output` | | Output file path | `stdout` |
| `--fail-on` | | Categories to fail on (`spec`, `style`) | `spec`, `style` |

## Output Formats

### GNU (Default)

The default format is compatible with standard GNU error tools and many editors.

```
path/to/SKILL.md:line:col: category: message
```

### SARIF

The Static Analysis Results Interchange Format (SARIF) is a standard JSON format supported by GitHub Advanced Security and other tools.

### Checkstyle

The Checkstyle XML format is supported by many CI/CD systems, including Jenkins.

## Categories

- **spec**: checks for violations of the Skill specification (e.g., missing fields, invalid naming). These are considered Errors.
- **style**: checks for style guidelines (e.g., description formatting). These are considered Warnings.

## Examples

Lint the current directory and output to stdout:
```bash
skr lint
```

Lint a specific directory and output to a file in SARIF format:
```bash
skr lint ./my-skill --format sarif --output results.sarif
```

Fail only on specification violations, ignoring style warnings:
```bash
skr lint --fail-on spec
```
