# Supported Emoji and Replacements

This file documents every entry in `DefaultReplacements()` (`replacements.go`).
All entries are keyed by Unicode codepoint sequence; FE0F (variation selector-16)
variants are listed alongside their bare equivalents.

When the map is passed to `Replace`, `ReplaceFile`, or `ScanDir` (via
`ScanConfig.Replacements`), each sequence is substituted with the text value
shown below. Any emoji **not** present in the map is subsequently stripped by
`Demojify`.

---
## Warning and Alerts

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+26A0 U+FE0F | Warning sign (emoji) | `[WARNING]` |
| U+26A0 | Warning sign | `[WARNING]` |

---

## Status Symbols

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+2705 | White heavy check mark | `[PASS]` |
| U+2705 U+FE0F | White heavy check mark (emoji) | `[PASS]` |
| U+2713 | Check mark | `[PASS]` |
| U+2714 | Heavy check mark | `[PASS]` |
| U+2714 U+FE0F | Heavy check mark (emoji) | `[PASS]` |
| U+274C | Cross mark | `[FAIL]` |
| U+274C U+FE0F | Cross mark (emoji) | `[FAIL]` |
| U+2717 | Ballot X | `[FAIL]` |
| U+2718 | Heavy ballot X | `[FAIL]` |
| U+274E | Negative squared cross mark | `[FAIL]` |
| U+2757 | Heavy exclamation mark | `[ALERT]` |
| U+2757 U+FE0F | Heavy exclamation mark (emoji) | `[ALERT]` |
| U+2755 | White exclamation mark | `[ALERT]` |
| U+2755 U+FE0F | White exclamation mark (emoji) | `[ALERT]` |
| U+203C | Double exclamation mark | `[ALERT]` |
| U+203C U+FE0F | Double exclamation mark (emoji) | `[ALERT]` |
| U+2753 | Question mark | `[?]` |
| U+2754 | White question mark | `[?]` |

---

## Information Symbol

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+2139 | Information source | `[INFO]` |
| U+2139 U+FE0F | Information source (emoji) | `[INFO]` |

---

## Severity Indicators

Colored circles widely used in CI dashboards, status pages, and documentation.

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+1F534 | Red circle | `[ERROR]` |
| U+1F7E0 | Orange circle | `[WARNING]` |
| U+1F7E1 | Yellow circle | `[CAUTION]` |
| U+1F7E2 | Green circle | `[OK]` |
| U+1F535 | Blue circle | `[INFO]` |
| U+26AB | Medium black circle | `[INACTIVE]` |
| U+26AA | Medium white circle | `[INACTIVE]` |

---

## Stop and Prohibition

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+1F6D1 | Stop sign | `[STOP]` |
| U+26D4 | No entry sign | `[NO ENTRY]` |
| U+1F6AB | Prohibited sign | `[PROHIBITED]` |

---

## Favorites and Highlights

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+2B50 | White medium star | `[FEATURED]` |
| U+2605 | Black star | `[FEATURED]` |
| U+2606 | White star | `[FEATURED]` |
| U+1F4A1 | Bulb | `[TIP]` |
| U+1F514 | Bell | `[NOTIFICATION]` |
| U+1F4CC | Pushpin | `[PINNED]` |
| U+1F511 | Key | `[KEY]` |
| U+1F512 | Padlock | `[LOCKED]` |
| U+1F513 | Open padlock | `[UNLOCKED]` |

---

## Cloud and Deployment

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+2601 U+FE0F | Cloud (emoji) | `[CLOUD]` |
| U+2601 | Cloud | `[CLOUD]` |
| U+1F4CA | Bar chart | `[REPORT]` |
| U+1F4C8 | Chart with upwards trend | `[GROWTH]` |
| U+1F4C9 | Chart with downwards trend | `[DECLINE]` |
| U+1F4DA | Books | `[DOCS]` |
| U+1F4D6 | Open book | `[GUIDE]` |
| U+1F4DD | Memo | `[NOTE]` |
| U+1F4C1 | File folder | `[DIR]` |
| U+1F4C2 | Open file folder | `[FOLDER]` |
| U+1F50D | Left-pointing magnifying glass | `[SEARCH]` |
| U+1F50E | Right-pointing magnifying glass | `[SEARCH]` |
| U+1F510 | Closed lock with key | `[SECURITY]` |
| U+2699 | Gear | `[CONFIG]` |
| U+2699 U+FE0F | Gear (emoji) | `[CONFIG]` |
| U+26A1 | High voltage | `[SETTINGS]` |
| U+1F3D7 | Building construction | `[BUILD]` |
| U+1F3AF | Direct hit | `[TARGET]` |
| U+1F3A8 | Artist palette | `[DESIGN]` |
| U+1F4BB | Personal computer | `[CODE]` |
| U+1F5A5 | Desktop computer | `[SERVER]` |
| U+1F310 | Globe with meridians | `[NETWORK]` |
| U+1F30E | Globe showing Americas | `[GLOBAL]` |
| U+1F5FA | World map | `[MAP]` |
| U+1F4CD | Round pushpin | `[MAP]` |
| U+1F449 | White right-pointing backhand index | `[SEE]` |

---

## CI/CD Workflow

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+1F680 | Rocket | `[DEPLOY]` |
| U+1F4E6 | Package | `[PACKAGE]` |
| U+1F389 | Party popper | `[SUCCESS]` |
| U+2728 | Sparkles | `[NEW]` |
| U+1F3C1 | Chequered flag | `[DONE]` |
| U+1F527 | Wrench | `[FIX]` |
| U+1F6E0 | Hammer and wrench | `[TOOLS]` |
| U+1F9F0 | Toolbox | `[TOOLS]` |
| U+267B | Recycling symbol | `[RECYCLE]` |
| U+267B U+FE0F | Recycling symbol (emoji) | `[RECYCLE]` |
| U+1F4BE | Floppy disk | `[SAVE]` |
| U+1F525 | Fire | `[HOT]` |
| U+1F4AF | Hundred points | `[100]` |
| U+1F6A8 | Police car revolving light | `[ALERT]` |
| U+1FA79 | Adhesive bandage | `[PATCH]` |

---

## Status Indicators

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+23F3 | Hourglass with flowing sand | `[PENDING]` |
| U+23F3 U+FE0F | Hourglass with flowing sand (emoji) | `[PENDING]` |
| U+23F1 | Stopwatch | `[TIMER]` |
| U+23F1 U+FE0F | Stopwatch (emoji) | `[TIMER]` |
| U+23F0 | Alarm clock | `[TIMER]` |
| U+1F504 | Counterclockwise arrows | `[REFRESH]` |
| U+231B | Hourglass | `[LOADING]` |
| U+231B U+FE0F | Hourglass (emoji) | `[LOADING]` |
| U+2B06 | Upwards black arrow | `[UP]` |
| U+2B07 | Downwards black arrow | `[DOWN]` |
| U+27A1 | Black rightwards arrow | `[NEXT]` |
| U+2B05 | Leftwards black arrow | `[PREV]` |
| U+1F440 | Eyes | `[SEE]` |
| U+1F4E4 | Outbox tray | `[FROM]` |
| U+2611 U+FE0F | Ballot box with check (emoji) | `[SELECTED]` |

---

## Arrows

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+2192 | Rightwards arrow | `->` |
| U+2190 | Leftwards arrow | `<-` |
| U+2191 | Upwards arrow | `^` |
| U+2193 | Downwards arrow | `v` |
| U+21D2 | Rightwards double arrow | `=>` |
| U+21D0 | Leftwards double arrow | `<=` |
| U+21D1 | Upwards double arrow | `^^` |
| U+21D3 | Downwards double arrow | `vv` |
| U+27A1 U+FE0F | Black rightwards arrow (emoji) | `->` |
| U+2B05 U+FE0F | Leftwards black arrow (emoji) | `<-` |
| U+2B06 U+FE0F | Upwards black arrow (emoji) | `^` |
| U+2B07 U+FE0F | Downwards black arrow (emoji) | `v` |

---

## Math Operators

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+2716 | Heavy multiplication X | `x` |
| U+2716 U+FE0F | Heavy multiplication X (emoji) | `x` |
| U+2795 | Heavy plus sign | `+` |
| U+2796 | Heavy minus sign | `-` |
| U+2797 | Heavy division sign | `/` |
| U+267E | Infinity (permanent paper sign) | `[INFINITY]` |
| U+267E U+FE0F | Infinity (emoji) | `[INFINITY]` |

---

## Geometric Shapes

Used as decorative bullet points in AI-generated and rich-text content.

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+25CF | Black circle | `*` |
| U+25CB | White circle | `o` |
| U+25A0 | Black square | `*` |
| U+25A1 | White square | `[]` |
| U+25B2 | Black up-pointing triangle | `^` |
| U+25B3 | White up-pointing triangle | `^` |
| U+25BC | Black down-pointing triangle | `v` |
| U+25BD | White down-pointing triangle | `v` |
| U+25C6 | Black diamond | `*` |
| U+25C7 | White diamond | `<>` |
| U+25AA | Black small square | `*` |
| U+25AB | White small square | `[]` |

---

## Checkboxes

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+2611 | Ballot box with check | `[x]` |
| U+2612 | Ballot box with X | `[x]` |
| U+2610 | Ballot box | `[ ]` |

---

## Common Dingbats

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+2764 | Heavy black heart | `<3` |
| U+2764 U+FE0F | Heavy black heart (emoji) | `<3` |
| U+2665 | Black heart suit | `<3` |
| U+2665 U+FE0F | Black heart suit (emoji) | `<3` |
| U+2666 | Black diamond suit | `<>` |
| U+2666 U+FE0F | Black diamond suit (emoji) | `<>` |
| U+2022 | Bullet | `*` |
| U+2023 | Triangular bullet | `>` |
| U+25E6 | White bullet | `o` |
| U+2219 | Bullet operator | `*` |

---

## Heart Variants

Colored and decorative hearts — all map to `[HEART]`.

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+1F499 | Blue heart | `[HEART]` |
| U+1F49A | Green heart | `[HEART]` |
| U+1F49B | Yellow heart | `[HEART]` |
| U+1F49C | Purple heart | `[HEART]` |
| U+1F5A4 | Black heart | `[HEART]` |
| U+1F90D | White heart | `[HEART]` |
| U+1F90E | Brown heart | `[HEART]` |
| U+1F9E1 | Orange heart | `[HEART]` |
| U+1F494 | Broken heart | `[HEART]` |
| U+1F495 | Two hearts | `[HEART]` |
| U+1F496 | Sparkling heart | `[HEART]` |
| U+1F497 | Growing heart | `[HEART]` |
| U+1F493 | Beating heart | `[HEART]` |
| U+1F49E | Revolving hearts | `[HEART]` |
| U+1F49D | Heart with ribbon | `[HEART]` |

---

## Project and Issue Tracking

Gitmoji conventions used in commit messages, changelogs, and READMEs.

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+1F41B | Bug | `[BUG]` |
| U+1F41E | Lady beetle | `[BUG]` |
| U+1F4A5 | Collision | `[BREAKING]` |
| U+1F6A7 | Construction sign | `[CONSTRUCTION]` |
| U+1F9EA | Test tube | `[TEST]` |
| U+1F9EB | Petri dish | `[TEST]` |
| U+1F516 | Bookmark | `[RELEASE]` |
| U+1F3F7 | Label | `[TAG]` |
| U+1F3F7 U+FE0F | Label (emoji) | `[TAG]` |
| U+1F9F9 | Broom | `[CLEANUP]` |
| U+1F517 | Link | `[LINK]` |
| U+1F4AC | Speech balloon | `[COMMENT]` |
| U+1F4AD | Thought balloon | `[COMMENT]` |
| U+1F5E3 | Speaking head | `[COMMENT]` |
| U+1F5E3 U+FE0F | Speaking head (emoji) | `[COMMENT]` |
| U+1F4E3 | Megaphone | `[ANNOUNCE]` |
| U+1F4E2 | Loudspeaker | `[ANNOUNCE]` |
| U+1F44D | Thumbs up | `[APPROVED]` |
| U+1F44E | Thumbs down | `[REJECTED]` |
| U+1F9E9 | Puzzle piece | `[PLUGIN]` |
| U+1F3C6 | Trophy | `[AWARD]` |
| U+1F396 | Military medal | `[AWARD]` |
| U+1F396 U+FE0F | Military medal (emoji) | `[AWARD]` |
| U+1F4CB | Clipboard | `[CLIPBOARD]` |
| U+1F5D1 | Wastebasket | `[TRASH]` |
| U+1F5D1 U+FE0F | Wastebasket (emoji) | `[TRASH]` |
| U+1F4CE | Paperclip | `[ATTACHMENT]` |
| U+1F381 | Wrapped gift | `[GIFT]` |
| U+1F48E | Gem stone | `[GEM]` |

---

## Colored Squares

Used in CI dashboards and status tables alongside the colored circle set.

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+1F7E5 | Red square | `[ERROR]` |
| U+1F7E9 | Green square | `[OK]` |
| U+1F7E8 | Yellow square | `[CAUTION]` |
| U+1F7E6 | Blue square | `[INFO]` |
| U+1F7E7 | Orange square | `[WARNING]` |
| U+1F7EA | Purple square | `[INFO]` |
| U+2B1B | Black large square | `[INACTIVE]` |
| U+2B1C | White large square | `[INACTIVE]` |

---

## Media Controls

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+23F8 | Pause button | `[PAUSED]` |
| U+23F8 U+FE0F | Pause button (emoji) | `[PAUSED]` |
| U+23F9 | Stop button | `[STOPPED]` |
| U+23F9 U+FE0F | Stop button (emoji) | `[STOPPED]` |
| U+23FA | Record button | `[RECORDING]` |
| U+23FA U+FE0F | Record button (emoji) | `[RECORDING]` |
| U+23E9 | Fast-forward button | `[NEXT]` |
| U+23EA | Fast-reverse button | `[PREV]` |

---

## Community and Contributors

Common in CONTRIBUTING.md, acknowledgement sections, and bot-generated content.

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+1F691 | Ambulance | `[HOTFIX]` |
| U+1F500 | Twisted rightwards arrows | `[MERGE]` |
| U+1F501 | Clockwise rightwards arrows | `[RETRY]` |
| U+23EB | Black up-pointing double triangle | `[UPGRADE]` |
| U+23EC | Black down-pointing double triangle | `[DOWNGRADE]` |
| U+1F6E1 | Shield | `[PROTECTED]` |
| U+1F6E1 U+FE0F | Shield (emoji) | `[PROTECTED]` |
| U+1F916 | Robot face | `[BOT]` |
| U+1F91D | Handshake | `[CONTRIB]` |
| U+1F464 | Bust in silhouette | `[USER]` |
| U+1F465 | Busts in silhouette | `[USERS]` |
| U+1F64F | Folded hands | `[THANKS]` |
| U+1F4C4 | Page facing up | `[FILE]` |
| U+1F4C3 | Page with curl | `[FILE]` |
| U+1F4E7 | Envelope | `[EMAIL]` |
| U+1F4B0 | Money bag | `[SPONSOR]` |
| U+1F4B5 | Dollar banknote | `[SPONSOR]` |
| U+1F30D | Globe showing Europe/Africa | `[GLOBAL]` |
| U+1F30F | Globe showing Asia/Australia | `[GLOBAL]` |
| U+21A9 | Leftwards arrow with hook | `[BACK]` |
| U+21A9 U+FE0F | Leftwards arrow with hook (emoji) | `[BACK]` |
| U+21AA | Rightwards arrow with hook | `[FORWARD]` |
| U+21AA U+FE0F | Rightwards arrow with hook (emoji) | `[FORWARD]` |
| U+1F507 | Speaker with cancellation stroke | `[MUTE]` |
| U+1F515 | Bell with cancellation stroke | `[MUTE]` |

---

## Platform and Language Indicators

Used in README badges, setup sections, and toolchain documentation.

| Codepoint | Sequence | Replacement |
|-----------|----------|-------------|
| U+1F433 | Spouting whale | `[DOCKER]` |
| U+1F40B | Whale | `[DOCKER]` |
| U+1F427 | Penguin | `[LINUX]` |
| U+1F40D | Snake | `[PYTHON]` |
| U+1F980 | Crab | `[RUST]` |
| U+1F439 | Hamster | `[GO]` |
| U+2219 | Bullet operator | `*` |

---
## Adding Custom Replacements

`DefaultReplacements()` returns a fresh copy of the map on every call. To add
or override entries, modify the returned map before passing it to `Replace`,
`ReplaceFile`, or `ScanConfig.Replacements`:

```go
repl := demojify.DefaultReplacements()
repl["\U0001F600"] = "[GRIN]" // add a custom mapping
repl["\u2705"] = "OK"         // override an existing mapping

cleaned := demojify.Replace(text, repl)
```

Any emoji that has no entry in the map is stripped by `Demojify` after
substitution, so unmapped codepoints never reach output.

---
## Unicode Ranges Removed by Demojify

`Demojify` removes codepoints regardless of the replacements map. The regex
covers:

| Range | Description |
|-------|-------------|
| U+2139 | Information source |
| U+231A–U+231B | Watch, hourglass |
| U+23CF | Eject symbol |
| U+23E9–U+23F3 | Media controls, hourglasses |
| U+23F8–U+23FA | Pause, stop, record |
| U+24C2 | Circled M |
| U+25AA–U+25AB | Small squares |
| U+25B6 | Play button |
| U+25C0 | Reverse button |
| U+25FB–U+25FE | Medium squares |
| U+2600–U+27BF | Miscellaneous symbols, dingbats, arrows |
| U+2934–U+2935 | Curved arrows |
| U+2B05–U+2B07 | Directional arrows |
| U+2B1B–U+2B1C | Large squares |
| U+2B50 | White medium star |
| U+2B55 | Heavy large circle |
| U+3030 | Wavy dash |
| U+303D | Part alternation mark |
| U+3297 | Circled ideograph congratulation |
| U+3299 | Circled ideograph secret |
| U+1F000–U+1FAFF | Mahjong tiles through symbols and pictographs (includes all Emoji 17.0 additions) |
| U+200D | Zero width joiner (used in sequences) |
| U+20E3 | Combining enclosing keycap |
| U+E0020–U+E007F | Tags block (subdivision flags: England, Scotland, Wales) |
| U+FE00–U+FE0F | Variation selectors |