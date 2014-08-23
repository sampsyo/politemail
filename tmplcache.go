package main

import (
	"html/template"
	"io"
	"path"
)

type TemplateCache struct {
	Templates map[string]*template.Template
	SearchDir string
	Extension string
}

func NewCache(dir string) *TemplateCache {
	c := new(TemplateCache)
	c.Templates = make(map[string]*template.Template)
	c.SearchDir = dir
	c.Extension = ".html"
	return c
}

func (c *TemplateCache) Lookup(name string) *template.Template {
	tmpl, found := c.Templates[name]
	if found {
		return tmpl
	} else {
		path := path.Join(c.SearchDir, name+c.Extension)
		tmpl = template.Must(template.ParseFiles(path))
		c.Templates[name] = tmpl
		return tmpl
	}
}

func (c *TemplateCache) Render(w io.Writer, name string, data interface{}) {
	tmpl := c.Lookup(name)
	tmpl.Execute(w, data)
}
