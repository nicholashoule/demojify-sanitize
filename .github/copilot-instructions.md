# Copilot Instructions

See modular instruction files in [instructions/](instructions/) for scoped rules:
[general](instructions/general.instructions.md)
[library](instructions/library.instructions.md)
[testing](instructions/testing.instructions.md)

Long-form project context lives in [README.md](../README.md) and [docs/](../docs/).Emoji prevention policy: [emoji-prevention.md](emoji-prevention.md).
## File Constraints

**`.github` File Limit**: All files in `.github/` directory must not exceed **500 lines**. Keep instructions concise and focused. Use external documentation in `docs/` for detailed explanations.

**Handling 500-Line Limit:**

When approaching or needing to exceed the 500-line limit, follow this priority order:

1. **Refactor First** - Remove redundancy, condense verbose explanations, eliminate duplicate examples
2. **Move to `docs/`** - Relocate detailed explanations, comprehensive examples, and long-form content to `docs/` directory
3. **Link, Don't Repeat** - Reference external documentation instead of duplicating content
4. **Split Only When Necessary** - Create a new file only when content cannot be condensed and represents a distinct, logical separation of concerns (e.g., splitting backend and frontend instructions, not arbitrary pagination)

**When splitting is unavoidable:**
- Create semantically cohesive files (grouped by domain/concern, not arbitrary line counts)
- Use clear, descriptive filenames that indicate scope
- Add cross-references between related files
- Update index files to maintain discoverability

## File Naming Conventions

**Prefer lowercase filenames** in `docs/` and `.github/` directories:
- Use `kebab-case` or `snake_case` for multi-word filenames (e.g., `developer_guide.md`, `api-reference.md`)
- Exception: `*_AUDIT.md` files in `docs/` may use uppercase (e.g., `security_AUDIT.md`)
- Exception: GitHub-mandated filenames must remain uppercase (e.g., `PULL_REQUEST_TEMPLATE.md`)

When creating or renaming files, enforce lowercase unless an explicit exception applies.
