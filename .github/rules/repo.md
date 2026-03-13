---
applyTo: "**"
---

# Repository Instructions

## .github Layout

Files present in this repository:

- `.github/copilot-instructions.md` — repo-wide Copilot context
- `.github/repogov-config.json` — repogov line-limit configuration
- `.github/rules/` — scoped Copilot instruction files (`*.md`)
- `.github/workflows/ci.yml` — CI pipeline

Optional GitHub-standard files (add as needed):

- `.github/ISSUE_TEMPLATE/` or `.github/ISSUE_TEMPLATE.md`
- `.github/PULL_REQUEST_TEMPLATE/` or `.github/pull_request_template.md`
- `.github/CODEOWNERS`
- `.github/dependabot.yml`
- `.github/FUNDING.yml`
- `.github/SUPPORT.md`

Standard `.gitlab/` files: `issue_templates/`, `merge_request_templates/`, `CODEOWNERS`, `.gitlab-ci.yml`.

Shared root files: `README.md`, `LICENSE`, `CHANGELOG.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, `AGENTS.md`, `.gitignore`, `.gitattributes`.

## Pull Requests / Merge Requests

- Keep pull requests focused on a single concern
- Use a descriptive title that summarizes the change (imperative mood)
- Reference related issues in the PR description
- Ensure all CI checks pass before requesting review
- Resolve or respond to every review comment before merging
- Update documentation in the same PR that changes behavior

## Commit Standards

Format: `<type>(<scope>): <subject>` -- subject in imperative mood, under 72 characters.

- Separate subject from body with a blank line when detail is needed
- Reference issue or PR numbers in the body when relevant
- Do not include generated, vendor, or binary files in commits
- Do not commit secrets, credentials, or environment-specific values

| Type | Use |
|------|-----|
| `feat:` | New exported symbol or option |
| `fix:` | Bug fix |
| `docs:` | Documentation only |
| `style:` | Formatting (no logic change) |
| `refactor:` | Code restructuring |
| `test:` | Adding/updating tests |
| `chore:` | Maintenance, dependencies |
| `perf:` | Performance improvement |
| `ci:` | CI/CD changes |
