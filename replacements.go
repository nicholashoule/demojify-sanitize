package demojify

// DefaultReplacements returns a copy of the built-in emoji-to-text substitution
// map. Callers may pass the returned map directly to [Replace] or [ReplaceFile].
// Because a fresh copy is returned on every call, callers can add, remove, or
// override entries without affecting other callers.
//
// The map covers ~230 codepoint sequences across eighteen categories:
//
//   - Warning and alert symbols (U+26A0, U+203C, ...)
//   - Status symbols: pass/fail/alert/info indicators
//   - Severity indicators: colored circles for CI dashboards
//   - Favorites, highlights, and annotations
//   - Cloud, deployment, and technical pictographs
//   - CI/CD workflow: deploy, package, release, maintenance
//   - Status indicators: pending, loading, directional
//   - Arrows (U+2192 series) common in documentation
//   - Math operators (U+2716, U+2795–U+2797, U+267E)
//   - Geometric shapes and bullet-point codepoints
//   - Checkboxes (U+2610 series)
//   - Common dingbats: hearts, diamonds, bullets
//   - Heart variants: colored and decorative hearts
//   - Project and issue tracking: gitmoji conventions (bug, wip, breaking, ...)
//   - Colored squares: CI dashboards and status tables
//   - Media controls: pause, stop, record, fast-forward/rewind
//   - Community and contributors: bots, thanks, sponsors, contact
//   - Platform and language indicators: Docker, Linux, Python, Rust, Go
//
// Variation-selector suffixed sequences (e.g., U+26A0 U+FE0F) are listed
// alongside their bare equivalents so that both forms are substituted.
// Any emoji not present in the map is stripped by [Demojify] after substitution.
func DefaultReplacements() map[string]string {
	return map[string]string{
		// Warning and alerts
		"\u26a0\ufe0f": "[WARNING]",
		"\u26a0":       "[WARNING]",

		// Status symbols
		"\u2705":       "[PASS]",
		"\u2705\ufe0f": "[PASS]",
		"\u2713":       "[PASS]",
		"\u2714":       "[PASS]", // Heavy check mark
		"\u2714\ufe0f": "[PASS]",
		"\u274c":       "[FAIL]",
		"\u274c\ufe0f": "[FAIL]",
		"\u2717":       "[FAIL]",
		"\u2718":       "[FAIL]", // Heavy ballot X
		"\u274e":       "[FAIL]", // Negative squared cross mark
		"\u2757":       "[ALERT]",
		"\u2757\ufe0f": "[ALERT]",
		"\u2755":       "[ALERT]", // White exclamation mark
		"\u2755\ufe0f": "[ALERT]",
		"\u203c":       "[ALERT]", // Double exclamation
		"\u203c\ufe0f": "[ALERT]",
		"\u2753":       "[?]", // Question mark
		"\u2754":       "[?]", // White question mark

		// Information symbol (U+2139, also covered by emojiRE)
		"\u2139":       "[INFO]", // Information source
		"\u2139\ufe0f": "[INFO]",

		// Severity indicators -- colored circles (widely used in CI dashboards and docs)
		"\U0001f534": "[ERROR]",    // Red circle
		"\U0001f7e0": "[WARNING]",  // Orange circle
		"\U0001f7e1": "[CAUTION]",  // Yellow circle
		"\U0001f7e2": "[OK]",       // Green circle
		"\U0001f535": "[INFO]",     // Blue circle
		"\u26ab":     "[INACTIVE]", // Medium black circle
		"\u26aa":     "[INACTIVE]", // Medium white circle

		// Stop and prohibition
		"\U0001f6d1": "[STOP]",       // Stop sign
		"\u26d4":     "[NO ENTRY]",   // No entry sign
		"\U0001f6ab": "[PROHIBITED]", // Prohibited sign

		// Favorites and highlights
		"\u2b50":     "[FEATURED]",
		"\u2605":     "[FEATURED]", // Filled star
		"\u2606":     "[FEATURED]", // Hollow star
		"\U0001f4a1": "[TIP]",
		"\U0001f514": "[NOTIFICATION]",
		"\U0001f4cc": "[PINNED]",
		"\U0001f511": "[KEY]",
		"\U0001f512": "[LOCKED]",
		"\U0001f513": "[UNLOCKED]",

		// Cloud and deployment
		"\u2601\ufe0f": "[CLOUD]",
		"\u2601":       "[CLOUD]",
		"\U0001f4ca":   "[REPORT]",
		"\U0001f4c8":   "[GROWTH]",
		"\U0001f4c9":   "[DECLINE]",
		"\U0001f4da":   "[DOCS]",
		"\U0001f4d6":   "[GUIDE]",
		"\U0001f4dd":   "[NOTE]",
		"\U0001f4c1":   "[DIR]",
		"\U0001f4c2":   "[FOLDER]",
		"\U0001f50d":   "[SEARCH]",
		"\U0001f50e":   "[SEARCH]",
		"\U0001f510":   "[SECURITY]",
		"\u2699":       "[CONFIG]",
		"\u2699\ufe0f": "[CONFIG]",
		"\u26a1":       "[SETTINGS]",
		"\U0001f3d7":   "[BUILD]",
		"\U0001f3af":   "[TARGET]",
		"\U0001f3a8":   "[DESIGN]",
		"\U0001f4bb":   "[CODE]",
		"\U0001f5a5":   "[SERVER]",
		"\U0001f310":   "[NETWORK]",
		"\U0001f30e":   "[GLOBAL]",
		"\U0001f5fa":   "[MAP]",
		"\U0001f4cd":   "[MAP]",
		"\U0001f449":   "[SEE]", // White right-pointing backhand index (attention/callout marker)

		// CI/CD workflow
		"\U0001f680":   "[DEPLOY]",  // Rocket
		"\U0001f4e6":   "[PACKAGE]", // Package
		"\U0001f389":   "[SUCCESS]", // Party popper
		"\u2728":       "[NEW]",     // Sparkles
		"\U0001f3c1":   "[DONE]",    // Checkered flag
		"\U0001f527":   "[FIX]",     // Wrench
		"\U0001f6e0":   "[TOOLS]",   // Hammer and wrench
		"\U0001f9f0":   "[TOOLS]",   // Toolbox (alternate for U+1F6E0 hammer-and-wrench)
		"\u267b":       "[RECYCLE]", // Recycling symbol
		"\u267b\ufe0f": "[RECYCLE]", // Recycling symbol
		"\U0001f4be":   "[SAVE]",    // Floppy disk
		"\U0001f525":   "[HOT]",     // Fire
		"\U0001f4af":   "[100]",     // Hundred points
		"\U0001f6a8":   "[ALERT]",   // Police car revolving light (gitmoji :rotating_light: -- lint/CI failures)
		"\U0001fa79":   "[PATCH]",   // Adhesive bandage (gitmoji :adhesive_bandage: -- minor/non-critical fix)

		// Status indicators
		"\u23f3":       "[PENDING]",
		"\u23f3\ufe0f": "[PENDING]",
		"\u23f1":       "[TIMER]",
		"\u23f1\ufe0f": "[TIMER]",
		"\u23f0":       "[TIMER]",
		"\U0001f504":   "[REFRESH]",
		"\u231b":       "[LOADING]",
		"\u231b\ufe0f": "[LOADING]",
		"\u2b06":       "[UP]",
		"\u2b07":       "[DOWN]",
		"\u27a1":       "[NEXT]",
		"\u2b05":       "[PREV]",
		"\U0001f440":   "[SEE]",
		"\U0001f4e4":   "[FROM]",
		"\u2611\ufe0f": "[SELECTED]",

		// Arrows (common in documentation)
		"\u2192":       "->", // Rightwards arrow
		"\u2190":       "<-", // Leftwards arrow
		"\u2191":       "^",  // Upwards arrow
		"\u2193":       "v",  // Downwards arrow
		"\u21d2":       "=>", // Rightwards double arrow
		"\u21d0":       "<=", // Leftwards double arrow
		"\u21d1":       "^^", // Upwards double arrow
		"\u21d3":       "vv", // Downwards double arrow
		"\u27a1\ufe0f": "->", // Black rightwards arrow with FE0F
		"\u2b05\ufe0f": "<-", // Leftwards black arrow with FE0F
		"\u2b06\ufe0f": "^",  // Upwards black arrow with FE0F
		"\u2b07\ufe0f": "v",  // Downwards black arrow with FE0F

		// Math operators (common in documentation)
		"\u2716":       "x", // Heavy multiplication X
		"\u2716\ufe0f": "x",
		"\u2795":       "+",          // Heavy plus sign
		"\u2796":       "-",          // Heavy minus sign
		"\u2797":       "/",          // Heavy division sign
		"\u267e":       "[INFINITY]", // Infinity (permanent paper sign)
		"\u267e\ufe0f": "[INFINITY]",

		// Geometric shapes (bullet points)
		"\u25cf": "*",  // Black circle
		"\u25cb": "o",  // White circle
		"\u25a0": "*",  // Black square
		"\u25a1": "[]", // White square
		"\u25b2": "^",  // Black up-pointing triangle
		"\u25b3": "^",  // White up-pointing triangle
		"\u25bc": "v",  // Black down-pointing triangle
		"\u25bd": "v",  // White down-pointing triangle
		"\u25c6": "*",  // Black diamond
		"\u25c7": "<>", // White diamond
		"\u25aa": "*",  // Black small square
		"\u25ab": "[]", // White small square

		// Checkboxes
		"\u2611": "[x]", // Ballot box with check
		"\u2612": "[x]", // Ballot box with X
		"\u2610": "[ ]", // Ballot box

		// Common dingbats
		"\u2764":       "<3", // Heavy black heart
		"\u2764\ufe0f": "<3",
		"\u2665":       "<3", // Black heart suit
		"\u2665\ufe0f": "<3",
		"\u2666":       "<>", // Black diamond suit
		"\u2666\ufe0f": "<>",
		"\u2022":       "*", // Bullet
		"\u2023":       ">", // Triangular bullet
		"\u25e6":       "o", // White bullet
		"\u2219":       "*", // Bullet operator

		// Heart variants (colored and decorative hearts -- all map to [HEART])
		"\U0001f499": "[HEART]", // Blue heart
		"\U0001f49a": "[HEART]", // Green heart
		"\U0001f49b": "[HEART]", // Yellow heart
		"\U0001f49c": "[HEART]", // Purple heart
		"\U0001f5a4": "[HEART]", // Black heart
		"\U0001f90d": "[HEART]", // White heart
		"\U0001f90e": "[HEART]", // Brown heart
		"\U0001f9e1": "[HEART]", // Orange heart
		"\U0001f494": "[HEART]", // Broken heart
		"\U0001f495": "[HEART]", // Two hearts
		"\U0001f496": "[HEART]", // Sparkling heart
		"\U0001f497": "[HEART]", // Growing heart
		"\U0001f493": "[HEART]", // Beating heart
		"\U0001f49e": "[HEART]", // Revolving hearts
		"\U0001f49d": "[HEART]", // Heart with ribbon

		// Project and issue tracking (gitmoji conventions)
		"\U0001f41b":       "[BUG]",          // Bug
		"\U0001f41e":       "[BUG]",          // Lady beetle (also used as bug)
		"\U0001f4a5":       "[BREAKING]",     // Collision / breaking change
		"\U0001f6a7":       "[CONSTRUCTION]", // Construction / work in progress
		"\U0001f9ea":       "[TEST]",         // Test tube
		"\U0001f9eb":       "[TEST]",         // Petri dish
		"\U0001f516":       "[RELEASE]",      // Bookmark / tagged release
		"\U0001f3f7":       "[TAG]",
		"\U0001f3f7\ufe0f": "[TAG]",     // Label with selector
		"\U0001f9f9":       "[CLEANUP]", // Broom
		"\U0001f517":       "[LINK]",    // Link
		"\U0001f4ac":       "[COMMENT]", // Speech balloon
		"\U0001f4ad":       "[COMMENT]", // Thought balloon
		"\U0001f5e3":       "[COMMENT]",
		"\U0001f5e3\ufe0f": "[COMMENT]",  // Speaking head with selector
		"\U0001f4e3":       "[ANNOUNCE]", // Megaphone
		"\U0001f4e2":       "[ANNOUNCE]", // Loudspeaker
		"\U0001f44d":       "[APPROVED]", // Thumbs up
		"\U0001f44e":       "[REJECTED]", // Thumbs down
		"\U0001f9e9":       "[PLUGIN]",   // Puzzle piece
		"\U0001f3c6":       "[AWARD]",    // Trophy
		"\U0001f396":       "[AWARD]",
		"\U0001f396\ufe0f": "[AWARD]",     // Military medal with selector
		"\U0001f4cb":       "[CLIPBOARD]", // Clipboard
		"\U0001f5d1":       "[TRASH]",
		"\U0001f5d1\ufe0f": "[TRASH]",      // Wastebasket with selector
		"\U0001f4ce":       "[ATTACHMENT]", // Paperclip
		"\U0001f381":       "[GIFT]",       // Wrapped gift
		"\U0001f48e":       "[GEM]",        // Gem stone

		// Colored squares (CI dashboards and status tables)
		"\U0001f7e5": "[ERROR]",    // Red square
		"\U0001f7e9": "[OK]",       // Green square
		"\U0001f7e8": "[CAUTION]",  // Yellow square
		"\U0001f7e6": "[INFO]",     // Blue square
		"\U0001f7e7": "[WARNING]",  // Orange square
		"\U0001f7ea": "[INFO]",     // Purple square
		"\u2b1b":     "[INACTIVE]", // Black large square
		"\u2b1c":     "[INACTIVE]", // White large square

		// Media controls
		"\u23f8":       "[PAUSED]",
		"\u23f8\ufe0f": "[PAUSED]", // Pause button with selector
		"\u23f9":       "[STOPPED]",
		"\u23f9\ufe0f": "[STOPPED]", // Stop button with selector
		"\u23fa":       "[RECORDING]",
		"\u23fa\ufe0f": "[RECORDING]", // Record button with selector
		"\u23e9":       "[NEXT]",      // Fast-forward
		"\u23ea":       "[PREV]",      // Fast-rewind

		// Community and contributors
		"\U0001f691":       "[HOTFIX]",    // Ambulance / critical fix
		"\U0001f500":       "[MERGE]",     // Twisted arrows / merge
		"\U0001f501":       "[RETRY]",     // Clockwise arrows / repeat
		"\u23eb":           "[UPGRADE]",   // Black up-pointing double triangle
		"\u23ec":           "[DOWNGRADE]", // Black down-pointing double triangle
		"\U0001f6e1":       "[PROTECTED]", // Shield
		"\U0001f6e1\ufe0f": "[PROTECTED]", // Shield with selector
		"\U0001f916":       "[BOT]",       // Robot face
		"\U0001f91d":       "[CONTRIB]",   // Handshake
		"\U0001f464":       "[USER]",      // Bust in silhouette
		"\U0001f465":       "[USERS]",     // Busts in silhouette
		"\U0001f64f":       "[THANKS]",    // Folded hands
		"\U0001f4c4":       "[FILE]",      // Page facing up
		"\U0001f4c3":       "[FILE]",      // Page with curl
		"\U0001f4e7":       "[EMAIL]",     // Envelope
		"\U0001f4b0":       "[SPONSOR]",   // Money bag
		"\U0001f4b5":       "[SPONSOR]",   // Dollar banknote
		"\U0001f30d":       "[GLOBAL]",    // Globe Europe/Africa
		"\U0001f30f":       "[GLOBAL]",    // Globe Asia/Australia
		"\u21a9":           "[BACK]",      // Leftwards arrow with hook
		"\u21a9\ufe0f":     "[BACK]",      // With variation selector
		"\u21aa":           "[FORWARD]",   // Rightwards arrow with hook
		"\u21aa\ufe0f":     "[FORWARD]",   // With variation selector
		"\U0001f507":       "[MUTE]",      // Speaker with cancellation stroke
		"\U0001f515":       "[MUTE]",      // Bell with cancellation stroke

		// Platform and language indicators
		"\U0001f433": "[DOCKER]", // Spouting whale
		"\U0001f40b": "[DOCKER]", // Whale
		"\U0001f427": "[LINUX]",  // Penguin
		"\U0001f40d": "[PYTHON]", // Snake
		"\U0001f980": "[RUST]",   // Crab
		"\U0001f439": "[GO]",     // Hamster (Go gopher)
	}
}
