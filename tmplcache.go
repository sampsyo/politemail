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
	Debug     bool
}

func NewCache(dir string) *TemplateCache {
	c := new(TemplateCache)
	c.Templates = make(map[string]*template.Template)
	c.SearchDir = dir
	c.Extension = ".html"
	c.Debug = false
	return c
}

func (c *TemplateCache) Load(name string) *template.Template {
	path := path.Join(c.SearchDir, name+c.Extension)
	return template.Must(template.ParseFiles(path))
}

func (c *TemplateCache) Lookup(name string) *template.Template {
	if c.Debug {
		return c.Load(name)
	} else {
		tmpl, found := c.Templates[name]
		if found {
			return tmpl
		} else {
			tmpl := c.Load(name)
			c.Templates[name] = tmpl
			return tmpl
		}
	}
}

func (c *TemplateCache) Render(w io.Writer, name string, data interface{}) {
	tmpl := c.Lookup(name)
	tmpl.Execute(w, data)
}
