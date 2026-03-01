package demojify

// DefaultReplacements returns a copy of the built-in emoji-to-text substitution
// map. Callers may pass the returned map directly to [Replace] or [ReplaceFile].
// Because a fresh copy is returned on every call, callers can add, remove, or
// override entries without affecting other callers.
//
// The map covers ~97 codepoint sequences across nine categories:
//
//   - Warning and alert symbols (U+26A0, U+203C, ...)
//   - Status symbols: pass/fail/alert indicators
//   - Favorites, highlights, and annotations
//   - Cloud, deployment, and technical pictographs
//   - Status indicators: pending, loading, directional
//   - Arrows (U+2192 series) common in documentation
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
		"\u2713":       "[PASS]",
		"\u2714":       "[PASS]", // Heavy check mark
		"\u2714\ufe0f": "[PASS]",
		"\u274c":       "[FAIL]",
		"\u2717":       "[FAIL]",
		"\u2718":       "[FAIL]", // Heavy ballot X
		"\u274e":       "[FAIL]", // Negative squared cross mark
		"\u2757":       "[ALERT]",
		"\u2755":       "[ALERT]", // White exclamation mark
		"\u203c":       "[ALERT]", // Double exclamation
		"\u203c\ufe0f": "[ALERT]",

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

		// Status indicators
		"\u23f3":       "Pending",
		"\u23f1":       "Timer",
		"\u23f0":       "Timer",
		"\U0001f504":   "Refresh",
		"\u231b":       "Loading",
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
