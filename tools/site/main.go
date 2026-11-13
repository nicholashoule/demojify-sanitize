// Command site checks the website in site/ before it is published, for make
// site and the Pages workflow: the page exists, is bounded in size, declares a
// title, a language, a viewport, and a description, contains no literal emoji,
// every link or source it names relative to itself is a file inside site/, and
// no link uses plain HTTP or runs script. For search engines and link previews
// it also checks that the page names a canonical https:// address, that
// og:url matches it, that og:title, og:description, and og:image are set,
// that every address under the canonical one in a meta tag, the structured
// data, or sitemap.xml is a file inside site/, and that the structured data
// parses. The site is hand-written HTML with no build step, so this is the
// only check it gets.
//
// Usage: go run ./tools/site [DIR], where DIR defaults to site.
package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	demojify "github.com/nicholashoule/demojify-sanitize"
)

// maxPage bounds the page; a hand-written page is tens of kilobytes.
const maxPage = 4 << 20

func main() {
	dir := "site"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	problems, err := check(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "site:", err)
		os.Exit(2)
	}
	for _, p := range problems {
		fmt.Fprintln(os.Stderr, "site:", p)
	}
	if len(problems) > 0 {
		os.Exit(1)
	}
	fmt.Printf("ok   %s: index.html and every file it references\n", dir)
}

// check returns the problems found with the site in dir, or an error when
// the page cannot be read at all.
func check(dir string) ([]string, error) {
	page := filepath.Join(dir, "index.html")
	info, err := os.Stat(page)
	if err != nil {
		return nil, err
	}
	if info.Size() > maxPage {
		return nil, fmt.Errorf("%s is %d bytes, larger than %d", page, info.Size(), maxPage)
	}
	data, err := os.ReadFile(page)
	if err != nil {
		return nil, err
	}
	html := string(data)
	var problems []string
	for _, required := range []struct{ marker, what string }{
		{"<title>", "a <title>"},
		{"<html lang=", "a lang attribute on <html>"},
		{`name="viewport"`, "a viewport meta tag, for phones"},
		{`name="description"`, "a description meta tag"},
	} {
		if !strings.Contains(html, required.marker) {
			problems = append(problems, "index.html lacks "+required.what)
		}
	}
	// Examples on the page write emoji as character references, as docs use escapes.
	if demojify.ContainsEmojiWith(html, demojify.SafeOptions()) {
		problems = append(problems, "index.html contains literal emoji; write them as character references such as &#x2705;")
	}
	for _, ref := range references(html) {
		if problem := checkReference(dir, ref); problem != "" {
			problems = append(problems, problem)
		}
	}
	problems = append(problems, checkSearch(dir, html)...)
	return problems, nil
}

// checkSearch returns the problems with what search engines and link
// previews read from the page in dir: its canonical address, its og: tags,
// its structured data, and sitemap.xml. A page moved to a new address keeps
// stale copies of the old one in all of these, so each address under the
// canonical one must name a file that is published.
func checkSearch(dir, html string) []string {
	var base string
	for _, t := range startTags(html, "link") {
		if strings.EqualFold(attr(t.text, "rel"), "canonical") {
			base = attr(t.text, "href")
		}
	}
	switch {
	case base == "":
		return []string{"index.html lacks a canonical link"}
	case !strings.HasPrefix(base, "https://") || !strings.HasSuffix(base, "/"):
		return []string{"the canonical link must be an https:// address ending in /: " + base}
	}

	var problems []string
	og := map[string]string{}
	for _, t := range startTags(html, "meta") {
		content := attr(t.text, "content")
		if property := attr(t.text, "property"); strings.HasPrefix(property, "og:") {
			og[property] = content
		}
		problems = appendSiteURL(problems, dir, base, content)
	}
	for _, property := range []string{"og:title", "og:description", "og:image"} {
		if og[property] == "" {
			problems = append(problems, "index.html lacks "+property+", which link previews show")
		}
	}
	if og["og:url"] != base {
		problems = append(problems, "og:url must be the canonical address "+base+": "+og["og:url"])
	}
	if image := og["og:image"]; image != "" && !strings.HasPrefix(image, base) {
		problems = append(problems, "og:image must be a file on this site, under "+base+": "+image)
	}

	for _, body := range scriptBodies(html, "application/ld+json") {
		var data any
		if err := json.Unmarshal([]byte(body), &data); err != nil {
			problems = append(problems, "index.html has structured data that is not valid JSON: "+err.Error())
			continue
		}
		for _, u := range jsonStrings(data, base) {
			problems = appendSiteURL(problems, dir, base, u)
		}
	}

	return append(problems, checkSitemap(dir, base)...)
}

// checkSitemap returns the problems with sitemap.xml in dir, if there is one:
// it must parse as a sitemap, list at least one address, and list only
// addresses under base that name files in dir.
func checkSitemap(dir, base string) []string {
	path := filepath.Join(dir, "sitemap.xml")
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return []string{err.Error()}
	}
	if info.Size() > maxPage {
		return []string{fmt.Sprintf("sitemap.xml is %d bytes, larger than %d", info.Size(), maxPage)}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{err.Error()}
	}
	var sitemap struct {
		XMLName xml.Name `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
		Locs    []string `xml:"url>loc"`
	}
	if err := xml.Unmarshal(data, &sitemap); err != nil {
		return []string{"sitemap.xml is not a sitemap: " + err.Error()}
	}
	if len(sitemap.Locs) == 0 {
		return []string{"sitemap.xml lists no addresses"}
	}
	var problems []string
	for _, loc := range sitemap.Locs {
		loc = strings.TrimSpace(loc)
		if !strings.HasPrefix(loc, base) {
			problems = append(problems, "sitemap.xml lists an address that is not under the canonical "+base+": "+loc)
			continue
		}
		problems = appendSiteURL(problems, dir, base, loc)
	}
	return problems
}

// appendSiteURL appends the problem with u, an absolute address, to problems
// when u is under base but names no file in dir. Addresses elsewhere are left
// alone.
func appendSiteURL(problems []string, dir, base, u string) []string {
	rest, ok := strings.CutPrefix(u, base)
	if !ok {
		return problems
	}
	if problem := checkReference(dir, rest); problem != "" {
		return append(problems, problem+" (as "+u+")")
	}
	return problems
}

// jsonStrings returns every string in the decoded JSON value v that starts
// with prefix, in a stable order.
func jsonStrings(v any, prefix string) []string {
	switch v := v.(type) {
	case string:
		if strings.HasPrefix(v, prefix) {
			return []string{v}
		}
	case []any:
		var found []string
		for _, e := range v {
			found = append(found, jsonStrings(e, prefix)...)
		}
		return found
	case map[string]any:
		var found []string
		for _, k := range slices.Sorted(maps.Keys(v)) {
			found = append(found, jsonStrings(v[k], prefix)...)
		}
		return found
	}
	return nil
}

// tag is one start tag: its text from < through >, and the offset in the
// page just past it, where the element's content begins.
type tag struct {
	text string
	end  int
}

// startTags returns each start tag for the element name in html, in any
// case. A > inside a quoted attribute value does not end the tag.
func startTags(html, name string) []tag {
	lower := strings.ToLower(html)
	open := "<" + name
	var tags []tag
	for at := 0; ; {
		i := strings.Index(lower[at:], open)
		if i < 0 {
			return tags
		}
		start := at + i
		at = start + len(open)
		if at < len(html) && !isSpace(html[at]) && html[at] != '>' && html[at] != '/' {
			continue // a longer name, such as <linkx
		}
		end := tagEnd(html, at)
		tags = append(tags, tag{text: html[start:end], end: end})
		at = end
	}
}

// tagEnd returns the offset just past the > that closes the tag whose
// attributes start at html[from], or len(html) when nothing closes it.
func tagEnd(html string, from int) int {
	var quote byte
	for i := from; i < len(html); i++ {
		switch c := html[i]; {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '>':
			return i + 1
		}
	}
	return len(html)
}

// attr returns the value of the attribute name in the start tag t, in any
// case, or "" when t has none. Like references, it counts an attribute only
// after whitespace.
func attr(t, name string) string {
	lower := strings.ToLower(t)
	key := name + "="
	for at := 0; ; {
		i := strings.Index(lower[at:], key)
		if i < 0 {
			return ""
		}
		start := at + i
		at = start + len(key)
		if start > 0 && isSpace(t[start-1]) {
			value, _ := attributeValue(t, at)
			return strings.TrimSpace(value)
		}
	}
}

// scriptBodies returns the content of each <script> element in html whose
// type attribute is typ.
func scriptBodies(html, typ string) []string {
	lower := strings.ToLower(html)
	var bodies []string
	for _, t := range startTags(html, "script") {
		if !strings.EqualFold(attr(t.text, "type"), typ) {
			continue
		}
		end := strings.Index(lower[t.end:], "</script")
		if end < 0 {
			end = len(html) - t.end
		}
		bodies = append(bodies, html[t.end:t.end+end])
	}
	return bodies
}

// checkReference returns the problem with one reference from the page, or
// "". Links to other sites must use HTTPS. The schemes that can carry script
// (javascript:, vbscript:, data:) are refused. Any other URL scheme, such as
// mailto:, is left alone. Everything else names a file, which must exist
// inside dir, since only dir is published.
func checkReference(dir, ref string) string {
	ref = strings.TrimSpace(ref)
	lower := strings.ToLower(ref)
	switch {
	case ref == "", strings.HasPrefix(ref, "#"):
		return ""
	case strings.HasPrefix(lower, "http:"):
		return "plain HTTP link: " + ref
	case strings.HasPrefix(lower, "javascript:"), strings.HasPrefix(lower, "vbscript:"):
		return "script link, which runs code when followed: " + ref
	case strings.HasPrefix(lower, "data:"):
		return "data: URL, which can carry script; make it a file in " + dir + ": " + ref
	case strings.HasPrefix(ref, "//"):
		return "protocol-relative link, which may load over plain HTTP: " + ref
	case hasScheme(ref):
		return ""
	case strings.HasPrefix(ref, "/"):
		return "absolute path, which breaks under a project page's prefix: " + ref
	}
	local, _, _ := strings.Cut(ref, "#")
	local, _, _ = strings.Cut(local, "?")
	decoded, err := url.PathUnescape(local)
	if err != nil {
		return "badly percent-encoded reference: " + ref
	}
	path := filepath.Join(dir, filepath.FromSlash(decoded))
	if rel, err := filepath.Rel(dir, path); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "references a file outside " + dir + ", which is not published: " + ref
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return "references a file that is not in " + dir + ": " + ref
	} else if err != nil {
		return err.Error()
	}
	return ""
}

// hasScheme reports whether ref starts with a URL scheme, such as https: or
// mailto:, before any /, ?, or #.
func hasScheme(ref string) bool {
	for i, r := range ref {
		switch {
		case r == ':':
			return i > 0
		case 'a' <= r && r <= 'z', 'A' <= r && r <= 'Z':
		case i > 0 && ('0' <= r && r <= '9' || r == '+' || r == '-' || r == '.'):
		default:
			return false
		}
	}
	return false
}

// references returns the value of every href and src attribute in html, in
// any case and quoted either way or not at all. An attribute counts only
// after whitespace, so data-src is not src.
func references(html string) []string {
	lower := strings.ToLower(html)
	var refs []string
	for _, name := range []string{"href=", "src="} {
		for at := 0; ; {
			i := strings.Index(lower[at:], name)
			if i < 0 {
				break
			}
			start := at + i + len(name)
			at = start
			if before := at - len(name) - 1; before < 0 || !isSpace(html[before]) {
				continue
			}
			value, next := attributeValue(html, start)
			refs = append(refs, value)
			at = next
		}
	}
	return refs
}

// attributeValue returns the attribute value that starts at html[start],
// quoted or not, and where scanning resumes.
func attributeValue(html string, start int) (string, int) {
	if start >= len(html) {
		return "", start
	}
	if quote := html[start]; quote == '"' || quote == '\'' {
		end := strings.IndexByte(html[start+1:], quote)
		if end < 0 {
			return html[start+1:], len(html)
		}
		return html[start+1 : start+1+end], start + 2 + end
	}
	end := start
	for end < len(html) && !isSpace(html[end]) && html[end] != '>' {
		end++
	}
	return html[start:end], end
}

// isSpace reports whether b is HTML whitespace.
func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f'
}
