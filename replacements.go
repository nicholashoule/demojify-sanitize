package demojify

// DefaultReplacements returns a copy of the built-in emoji-to-text substitution
// map. Callers may pass the returned map directly to [Replace] or [ReplaceFile].
// Because a fresh copy is returned on every call, callers can add, remove, or
// override entries without affecting other callers.
//
// The map covers ~137 codepoint sequences across eleven categories:
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
//
// Variation-selector suffixed sequences (e.g., U+26A0 U+FE0F) are listed
// alongside their bare equivalents so that both forms are substituted.
// Any emoji not present in the map is stripped by [Demojify] after substitution.
func DefaultReplacements() map[string]string {
	return map[string]string{
		// Warning and alerts
		"\u26a0\ufe0f": "WARNING",
		"\u26a0":       "WARNING",

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
		"\U0001f512": "LOCKED",
		"\U0001f513": "UNLOCKED",

		// Cloud and deployment
		"\u2601\ufe0f": "Cloud",
		"\u2601":       "Cloud",
		"\U0001f4ca":   "Report",
		"\U0001f4c8":   "Growth",
		"\U0001f4c9":   "Decline",
		"\U0001f4da":   "Documentation",
		"\U0001f4d6":   "Guide",
		"\U0001f4dd":   "Note",
		"\U0001f4c1":   "Directory",
		"\U0001f4c2":   "Folder",
		"\U0001f50d":   "Search",
		"\U0001f50e":   "Search",
		"\U0001f510":   "Security",
		"\u2699":       "Configuration",
		"\u2699\ufe0f": "Configuration",
		"\u26a1":       "Settings",
		"\U0001f3d7":   "Build",
		"\U0001f3af":   "Target",
		"\U0001f3a8":   "Design",
		"\U0001f4bb":   "Code",
		"\U0001f5a5":   "Server",
		"\U0001f310":   "Network",
		"\U0001f30e":   "Global",
		"\U0001f5fa":   "Map",
		"\U0001f4cd":   "Map",

		// CI/CD workflow
		"\U0001f680":   "[DEPLOY]",  // Rocket
		"\U0001f4e6":   "[PACKAGE]", // Package
		"\U0001f389":   "[SUCCESS]", // Party popper
		"\u2728":       "[NEW]",     // Sparkles
		"\U0001f3c1":   "[DONE]",    // Checkered flag
		"\U0001f527":   "[FIX]",     // Wrench
		"\U0001f6e0":   "[TOOLS]",   // Hammer and wrench
		"\u267b":       "[RECYCLE]", // Recycling symbol
		"\u267b\ufe0f": "[RECYCLE]",
		"\U0001f4be":   "[SAVE]", // Floppy disk
		"\U0001f525":   "[HOT]",  // Fire
		"\U0001f4af":   "[100]",  // Hundred points

		// Status indicators
		"\u23f3":       "Pending",
		"\u23f3\ufe0f": "Pending",
		"\u23f1":       "Timer",
		"\u23f1\ufe0f": "Timer",
		"\u23f0":       "Timer",
		"\U0001f504":   "Refresh",
		"\u231b":       "Loading",
		"\u231b\ufe0f": "Loading",
		"\u2b06":       "Up",
		"\u2b07":       "Down",
		"\u27a1":       "Next",
		"\u2b05":       "Previous",
		"\U0001f440":   "See",
		"\U0001f4e4":   "From",
		"\u2611\ufe0f": "Selected",

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
	}
}
