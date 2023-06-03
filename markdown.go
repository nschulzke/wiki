package main

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"html/template"
	"regexp"
)

var replaceLinks = regexp.MustCompile(`\[\[([a-zA-Z0-9 ]+)\]\]`)

func mdToHTML(md []byte) template.HTML {
	md = replaceLinks.ReplaceAllFunc(md, func(s []byte) []byte {
		title := string(s[2 : len(s)-2])
		pageExists := pageExists(title)
		if !pageExists {
			return []byte(`<a class="danger" href="/edit/` + title + `">` + title + `</a>`)
		} else {
			return []byte(`<a class="success" href="/view/` + title + `">` + title + `</a>`)
		}
	})

	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return template.HTML(markdown.Render(doc, renderer))
}
