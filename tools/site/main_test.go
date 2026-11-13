package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	sound := `<html lang="en"><head><title>x</title><meta name="viewport" content="w"><meta name="description" content="d">` + soundSearch + `</head>` +
		`<body><p>&#x2705;</p><a href="#top">t</a><a href="https://example.com/">e</a><img src="images/a.svg"><a href="images/a.svg#x">f</a></body></html>`
	write("index.html", sound)
	write("images/a.svg", "<svg/>")
	if problems, err := check(dir); err != nil || len(problems) != 0 {
		t.Fatalf("a sound site: %v, %v", problems, err)
	}

	write("index.html", "<html><head></head><body>\u2705<a href=\"http://example.com/\">p</a><img src=\"missing.png\"><a href=\"/docs/\">abs</a></body></html>")
	problems, err := check(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"lacks a <title>", "lacks a lang", "lacks a viewport", "lacks a description", "literal emoji", "plain HTTP link", "not in " + dir + ": missing.png", "absolute path", "lacks a canonical link"} {
		if !strings.Contains(strings.Join(problems, "\n"), want) {
			t.Errorf("problems %q lack %q", problems, want)
		}
	}
	if _, err := check(filepath.Join(dir, "nowhere")); err == nil {
		t.Error("a missing site passed")
	}
}

// soundSearch is the head markup checkSearch accepts for a site whose files
// include images/a.svg.
const soundSearch = `<link rel="canonical" href="https://example.com/p/">` +
	`<meta property="og:url" content="https://example.com/p/">` +
	`<meta property="og:title" content="t"><meta property="og:description" content="a > b">` +
	`<meta property="og:image" content="https://example.com/p/images/a.svg">` +
	`<script type="application/ld+json">{"url": "https://example.com/p/", "image": ["https://example.com/p/images/a.svg"]}</script>`

func TestCheckSearch(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "images"), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"index.html", filepath.Join("images", "a.svg")} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	sitemap := func(locs ...string) string {
		var s strings.Builder
		s.WriteString(`<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
		for _, loc := range locs {
			s.WriteString("<url><loc>" + loc + "</loc></url>")
		}
		s.WriteString("</urlset>")
		return s.String()
	}
	for _, tc := range []struct {
		name, head, sitemap, want string
	}{
		{name: "sound", head: soundSearch},
		{name: "sound with sitemap", head: soundSearch, sitemap: sitemap(" https://example.com/p/ ")},
		{name: "no canonical", head: "", want: "lacks a canonical link"},
		{name: "plain HTTP canonical", head: strings.Replace(soundSearch, `href="https://example.com/p/"`, `href="http://example.com/p/"`, 1), want: "https:// address ending in /"},
		{name: "canonical without slash", head: strings.Replace(soundSearch, `href="https://example.com/p/"`, `href="https://example.com/p"`, 1), want: "https:// address ending in /"},
		{name: "og:url elsewhere", head: strings.Replace(soundSearch, `og:url" content="https://example.com/p/"`, `og:url" content="https://example.com/old/"`, 1), want: "og:url must be the canonical address"},
		{name: "no og:title", head: strings.Replace(soundSearch, `og:title`, `og:other`, 1), want: "lacks og:title"},
		{name: "no og:image", head: strings.Replace(soundSearch, `property="og:image"`, `property="og:other"`, 1), want: "lacks og:image"},
		{name: "og:image off site", head: strings.Replace(soundSearch, `content="https://example.com/p/images/a.svg"`, `content="https://cdn.example.com/a.png"`, 1), want: "og:image must be a file on this site"},
		{name: "og:image missing", head: strings.Replace(soundSearch, `p/images/a.svg">`, `p/images/gone.png">`, 1), want: "not in " + dir + ": images/gone.png (as https://example.com/p/images/gone.png)"},
		{name: "structured data not JSON", head: strings.Replace(soundSearch, `"url":`, `url:`, 1), want: "structured data that is not valid JSON"},
		{name: "structured data names a missing file", head: strings.Replace(soundSearch, `["https://example.com/p/images/a.svg"]`, `["https://example.com/p/images/gone.svg"]`, 1), want: "images/gone.svg"},
		{name: "sitemap off site", head: soundSearch, sitemap: sitemap("https://example.com/old/"), want: "not under the canonical"},
		{name: "sitemap missing file", head: soundSearch, sitemap: sitemap("https://example.com/p/gone.html"), want: "not in " + dir + ": gone.html"},
		{name: "sitemap empty", head: soundSearch, sitemap: sitemap(), want: "lists no addresses"},
		{name: "sitemap wrong namespace", head: soundSearch, sitemap: `<urlset><url><loc>https://example.com/p/</loc></url></urlset>`, want: "not a sitemap"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, "sitemap.xml")
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
			if tc.sitemap != "" {
				if err := os.WriteFile(path, []byte(tc.sitemap), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			problems := strings.Join(checkSearch(dir, "<html><head>"+tc.head+"</head></html>"), "\n")
			if tc.want == "" && problems != "" || !strings.Contains(problems, tc.want) {
				t.Errorf("problems %q, want %q", problems, tc.want)
			}
		})
	}
}

func TestStartTagsAndAttr(t *testing.T) {
	html := `<LINK REL=canonical HREF='https://a/'><linkx href="no"><meta content="a > b" property="og:x"><meta/>`
	links := startTags(html, "link")
	if len(links) != 1 || attr(links[0].text, "href") != "https://a/" || attr(links[0].text, "rel") != "canonical" {
		t.Errorf("startTags(link) = %q", links)
	}
	metas := startTags(html, "meta")
	if len(metas) != 2 || attr(metas[0].text, "content") != "a > b" || attr(metas[0].text, "property") != "og:x" || attr(metas[1].text, "content") != "" {
		t.Errorf("startTags(meta) = %q", metas)
	}
	if got := scriptBodies(`<script>x</script><script type="application/ld+json"> {} </SCRIPT>`, "application/ld+json"); len(got) != 1 || got[0] != " {} " {
		t.Errorf("scriptBodies = %q", got)
	}
}

func TestCheckReference(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "site")
	for _, name := range []string{filepath.Join(dir, "images", "a b.svg"), filepath.Join(dir, "index.html"), filepath.Join(root, "outside.txt")} {
		if err := os.MkdirAll(filepath.Dir(name), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for ref, want := range map[string]string{
		"":                              "",
		"#install":                      "",
		"https://example.com/":          "",
		"mailto:someone@example.com":    "",
		"images/a%20b.svg":              "",
		"images/a%20b.svg#frag":         "",
		"index.html?x=1":                "",
		"http://example.com/":           "plain HTTP",
		"HTTP://example.com/":           "plain HTTP",
		"javascript:alert(1)":           "script link",
		"vbscript:msgbox(1)":            "script link",
		"data:image/svg+xml,%3Csvg/%3E": "data: URL",
		"//example.com/x.js":            "protocol-relative",
		"/docs/":                        "absolute path",
		"../outside.txt":                "outside",
		"images/../../outside.txt":      "outside",
		"images/missing.png":            "not in",
		"images/%zz.svg":                "percent-encoded",
	} {
		got := checkReference(dir, ref)
		if want == "" && got != "" || want != "" && !strings.Contains(got, want) {
			t.Errorf("checkReference(%q) = %q, want %q", ref, got, want)
		}
	}
}

func TestReferences(t *testing.T) {
	html := `<a href="a.html"><img SRC='b.png'><link href=c.css><div data-src="not-me"></div>`
	got := strings.Join(references(html), " ")
	if want := "a.html c.css b.png"; got != want {
		t.Errorf("references = %q, want %q", got, want)
	}
}

func TestSite(t *testing.T) {
	problems, err := check(filepath.Join("..", "..", "site"))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range problems {
		t.Error(p)
	}
}
