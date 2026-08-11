# Contributing to Dynamite

Thanks for your interest in contributing to `Dynamite`! This guide covers
everything you need to get started.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Style](#code-style)
- [Commit Conventions](#commit-conventions)
- [Pull Request Process](#pull-request-process)
- [Project Architecture](#project-architecture)
- [Testing](#testing)

---

## Getting Started

### Prerequisites

- **Go 1.26+** &mdash; [go.dev/dl](https://go.dev/dl/)
- **gofmt** &mdash; code formatting
- **AWS credentials** &mdash; for manual testing
- **[mockgen](https://github.com/uber-go/mock)** &mdash; for generating mocks

### Setup

```bash
git clone https://github.com/wolfwfr/dynamite.git
cd dynamite
make build   # compile local binary
make install # install local build with version & commit information
make test    # run tests with race detection
```

---

## Development Workflow

### Branch Strategy

Use feature branches for all changes. Never commit directly to `main`.

```bash
# Create a feature branch
git checkout -b feature/my-feature

# Or for bug fixes
git checkout -b fix/my-fix
```

### Build & Test Commands

| Command      | What it does                                     |
| ------------ | ------------------------------------------------ |
| `make build` | Compile `Dynamite` binary with version info      |
| `make test`  | Generate mocks & run all tests (`go test ./...`) |
| `make fmt`   | Format code with gofmt                           |

Run a single test:

```bash
go test -run TestName ./internal/config/
```

---

## Code Style

- Format with **gofmt**
- Follow [Effective Go](https://go.dev/doc/effective_go) conventions
- Keep functions focused and short
- Use meaningful variable names; avoid single-letter names outside of loops
- Handle errors explicitly; don't ignore them

---

## Commit Conventions

All commits must follow
[Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>
```

### Allowed Types

| Type       | Use for                                  |
| ---------- | ---------------------------------------- |
| `feat`     | New feature                              |
| `fix`      | Bug fix                                  |
| `docs`     | Documentation only                       |
| `chore`    | Maintenance, dependencies                |
| `refactor` | Code restructuring (no behavior change)  |
| `test`     | Adding or updating tests                 |
| `perf`     | Performance improvement                  |
| `ci`       | CI/CD changes                            |
| `build`    | Build system changes                     |
| `style`    | Code style (formatting, no logic change) |
| `revert`   | Reverting a previous commit              |

### Rules

- Description must start with a lowercase letter
- First line must be 72 characters or fewer
- Breaking changes use `!` suffix: `feat!: redesign config format`

### AI contributions

This project aims to be explicit about its application of AI, favouring human
effort where possible and reasonable.

Commits that contain AI-generated code must include one of the following
suffixes in the commit summary:

- `(AI-assist)`
- `(AI-gen)`

For example:

- `perf [dynamodb]: optimised item parsing (AI-assist)`
- `chore [readme]: added contributor section (AI-gen)`

You are kindly requested to include at least the `(AI-assist)` suffix when an
LLM has been used in any capacity to contribute to your change. When an LLM was
used to consult your idea, but has not contributed any changes to your approach
or implementation, you are free to omit the suffix, however.

The `(AI-assist)` suffix is sufficient for cases where an LLM was used to verify
your ideas but you implemented (the vast majority of) your change manually.

The `(AI-gen)` suffix is applied when an LLM was used to generate a considerable
portion (e.g. > 50%) of your commit, even if the idea was your own.

---

## Pull Request Process

1. Fork the `Dynamite` repository
2. **Create a branch** from `main` with a descriptive name, conventially
   prefixed with `feat/`, `fix/`, `chore/`, `docs/`, etc.
3. **Make your changes** with clear, atomic commits
4. **Run checks locally:**
   ```bash
   make fmt
   make test
   ```
5. **Push and open a PR** against `github.com/wolfwfr/dynamite/main`
6. **Fill out the PR form** with a summary, changes, and test plan
7. **Required checks** must pass: `test`
8. **One approving review** is required before merging
9. PRs are **merged** with automatic branch cleanup

### PR Title Guidelines

- Keep under 70 characters
- Use the same conventions as commit messages: `feat: add EC2 instance browser`
- Use the description for details, not the title

---

## Project Architecture

```
cmd/dynamite/main.go         CLI entry point
lib/styles/                  Custom per-rune styling for e.g. syntax highlighting
pkg/
  adapters/dynamodb/         DynamoDB adapter & parsing
  aws/                       AWS SDK v2 factory functions
  common/                    Commonly shared, detailed logic (e.g. regex)
  file/                      Managers & definitions for config & .state files
  logging/                   Common logging-related resources & definitions
  theme/                     Global theme variables, init & update functions
  util/                      Commonly shared, generic utility functions
  ui/
    home.go                  Initial UI entrypoint (Bubble Tea model)
    internal/
        components/          Reusable UI components (e.g. search & table)
        dialogs/             UI dialogs
        messages/            System-wide Bubble Tea message definitions
        views/               Distinct Views (displayed above bottom gutter)
            items/           DynamoDB items view & internals
            tables/          DynamoDB tables view
```

### Key dependencies

| Package                              | Purpose                                   |
| ------------------------------------ | ----------------------------------------- |
| `github.com/charmbracelet/bubbletea` | TUI framework (Elm architecture)          |
| `github.com/charmbracelet/bubbles`   | TUI components (viewport, textinput)      |
| `github.com/charmbracelet/lipgloss`  | Terminal styling                          |
| `github.com/aws/aws-sdk-go-v2`       | AWS SDK v2                                |
| `github.com/urfave/cli/v3`           | CLI framework                             |
| `github.com/atotto/clipboard`        | clipboard integration                     |
| `github.com/uber-go/mock`            | mocking framework and used in `make test` |

---

## Testing

### Running Tests

```bash
# Generate mocks and execute all tests
make test

# Single test
go test -run TestItemSelectionPreviews ./pkg/ui/internal/views/items/

# Verbose output
go test -V -run TestItemSelectionPreviews ./pkg/ui/internal/views/items/
```

### Writing Tests

- Use table-driven tests where appropriate
- Mock AWS API calls using interface-based clients
- Test pagination (first page, continuation, last page)
- Test error handling

---

## License

By contributing to `Dynamite`, you agree that your contributions will be
licensed under the [MIT License](LICENSE).
