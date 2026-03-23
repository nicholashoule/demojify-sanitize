# Emoji Substitution Audit

**Last reviewed**: 2026-03-23
**Map version**: `DefaultReplacements()` ~230 entries, 20 categories
**Source**: [`replacements.go`](../../replacements.go)

---

## Category Summary

| # | Category | Entries | Token labels |
|---|----------|---------|--------------|
| 1 | Warning and Alerts | 2 | `[WARNING]` |
| 2 | Status Symbols | 18 | `[PASS]`, `[FAIL]`, `[DONE]`, `[PENDING]`, `[SKIP]`, `[RUNNING]`, `[BLOCKED]`, `[ERROR]` |
| 3 | Information | 2 | `[INFO]` |
| 4 | Severity (colored circles) | 8 | `[ERROR]`, `[OK]`, `[CAUTION]`, `[INFO]`, `[WARNING]`, `[INACTIVE]` |
| 5 | Stop and Prohibition | 4 | `[STOP]`, `[BLOCKED]` |
| 6 | Favorites and Highlights | 5 | `[STAR]`, `[FIRE]`, `[HIGHLIGHT]`, `[IMPORTANT]` |
| 7 | Cloud and Deployment | 7 | `[CLOUD]`, `[DEPLOY]`, `[SERVER]`, `[DB]`, `[PACKAGE]`, `[LOCK]`, `[KEY]` |
| 8 | CI/CD Workflow | 14 | `[BUILD]`, `[PASS]`, `[FAIL]`, `[MERGE]`, `[DEPLOY]`, `[TAG]`, `[RELEASE]`, `[ROLLBACK]`, `[PIPELINE]`, `[LINT]`, `[FIX]`, `[CLEANUP]`, `[DOCS]`, `[PERF]` |
| 9 | Status Indicators | 8 | `[NEW]`, `[UPDATED]`, `[DEPRECATED]`, `[REMOVED]`, `[EXPERIMENTAL]`, `[STABLE]`, `[BETA]`, `[DISABLED]` |
| 10 | Arrows | 12 | `->`, `<-`, `=>`, `<=`, `<->`, `<=>`, `[UP]`, `[DOWN]`, `[LEFT]`, `[RIGHT]` |
| 11 | Math Operators | 6 | `!=`, `>=`, `<=`, `~=`, `+-`, `~` |
| 12 | Geometric Shapes | 8 | `[BULLET]`, `[DIAMOND]`, `[BOX]`, `[CIRCLE]`, `[SQUARE]` |
| 13 | Checkboxes | 4 | `[x]`, `[ ]`, `[~]`, `[/]` |
| 14 | Common Dingbats | 10 | `<3`, `<>`, `*`, `>`, `o` |
| 15 | Heart Variants | 15 | `[HEART]` |
| 16 | Project and Issue Tracking | 29 | `[BUG]`, `[BREAKING]`, `[CONSTRUCTION]`, `[TEST]`, `[RELEASE]`, `[TAG]`, `[CLEANUP]`, `[LINK]`, `[COMMENT]`, `[ANNOUNCE]`, `[APPROVED]`, `[REJECTED]`, `[PLUGIN]`, `[AWARD]`, `[CLIPBOARD]`, `[TRASH]`, `[ATTACHMENT]`, `[GIFT]`, `[GEM]` |
| 17 | Colored Squares | 8 | `[ERROR]`, `[OK]`, `[CAUTION]`, `[INFO]`, `[WARNING]`, `[INACTIVE]` |
| 18 | Media Controls | 8 | `[PAUSED]`, `[STOPPED]`, `[RECORDING]`, `[NEXT]`, `[PREV]` |
| 19 | Community and Contributors | 26 | `[HOTFIX]`, `[MERGE]`, `[RETRY]`, `[UPGRADE]`, `[DOWNGRADE]`, `[PROTECTED]`, `[BOT]`, `[CONTRIB]`, `[USER]`, `[USERS]`, `[THANKS]`, `[FILE]`, `[EMAIL]`, `[SPONSOR]`, `[GLOBAL]`, `[BACK]`, `[FORWARD]`, `[MUTE]` |
| 20 | Platform and Language Indicators | 6 | `[DOCKER]`, `[LINUX]`, `[PYTHON]`, `[RUST]`, `[GO]` |

---

## Audit Status by Category

| Category | Substituted | Stripped (if not in map) | Notes |
|----------|-------------|--------------------------|-------|
| Warning and Alerts | yes | — | Full coverage including FE0F variant |
| Status Symbols | yes | — | Covers common CI check/cross/clock symbols |
| Information | yes | — | U+2139 and FE0F variant |
| Severity (colored circles) | yes | — | Full 6-color set (red/orange/yellow/green/purple/blue) |
| Stop and Prohibition | yes | — | Road sign and entry-forbidden emoji |
| Favorites and Highlights | yes | — | Star, fire, lightning, bookmark star |
| Cloud and Deployment | yes | — | Cloud, disk, package, lock/key pairs |
| CI/CD Workflow | yes | — | \~14 entries covering the common gitmoji workflow set |
| Status Indicators | yes | — | New/updated/deprecated flags and semantic-version lifecycle |
| Arrows | yes | — | Basic 4-direction + diagonal; ASCII text tokens |
| Math Operators | yes | — | Inequality and approximation symbols |
| Geometric Shapes | yes | — | Small/medium squares, diamonds, circles |
| Checkboxes | yes | — | Markdown-task-list style tokens |
| Common Dingbats | yes | — | Heart suit, bullets, diamond suit |
| Heart Variants | yes | — | 15 colored/decorative hearts; all → `[HEART]` |
| Project and Issue Tracking | yes | — | gitmoji commit convention symbols |
| Colored Squares | yes | — | Large colored squares as status badges |
| Media Controls | yes | — | Pause/stop/record/next/prev buttons |
| Community and Contributors | yes | — | CONTRIBUTING, bot, handshake, globe, nav hooks |
| Platform and Language Indicators | yes | — | Whale/penguin/snake/crab/hamster |

---

## Known Gaps

| Issue | Detail | Priority |
|-------|--------|----------|
| Code fence passthrough | Emoji inside Markdown fenced code blocks (` ``` `) are not distinguished from prose emoji; the library has no Markdown-aware mode | Low — acceptable by design; use a preprocessor if Markdown awareness is required |
| Skin-tone modifier sequences | Multi-codepoint ZWJ sequences that combine a base emoji with a skin-tone modifier (U+1F3FB–U+1F3FF) are stripped codepoint-by-codepoint; no substitution token is produced | Low — these rarely appear in technical documentation |
| Flag sequences | Regional indicator pair flags and subdivision tag sequences are stripped; no substitution token | Low — country/region flags in docs are rare; out of scope for text sanitization |
| Keycap sequences | Digit + U+FE0F + U+20E3 (e.g., `1` + VS16 + combining keycap) is stripped; no substitution defined | Low — uncommon in technical prose |
| Newer Emoji (17.0+) | Emoji added in Unicode 17.0 that have no established text convention are not in the substitution map | Low — add on demand when usage is observed |

---

## Audit Methodology

1. The source of truth is `DefaultReplacements()` in [`replacements.go`](../../replacements.go).
2. New emoji are added when observed in real-world READMEs, changelogs, or gitmoji convention lists.
3. Each new entry must have a corresponding test case in [`replacements_test.go`](../../replacements_test.go).
4. Token labels follow `[ALLCAPS]` bracket convention. Bare ASCII tokens (e.g., `->`, `<3`) are used where the text equivalent is universally understood.
5. Multiple emoji mapping to the same token (e.g., all heart variants → `[HEART]`) is intentional — the goal is machine-readable text, not round-trip fidelity.

---

## Updating This File

After any addition to `DefaultReplacements()`:

1. Update the entry count in the **Map version** header above.
2. Update the **Category Summary** row for the affected category.
3. Add new token labels to the **Token labels** column.
4. Update the **Known Gaps** table if the addition resolves a gap.
5. Update the **Last reviewed** date.
