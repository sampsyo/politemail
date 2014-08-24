package tmplpool

import (
	"html/template"
	"io"
	"path"
)

type Pool struct {
	Templates map[string]*template.Template
	SearchDir string
	Extension string
	Debug     bool
	Common    []string
	BaseDef   string
}

func New(dir string) *Pool {
	c := new(Pool)
	c.Templates = make(map[string]*template.Template)
	c.SearchDir = dir
	c.Extension = ".html"
	c.Debug = false
	return c
}

func (c *Pool) makeFilename(tmplname string) string {
	return path.Join(c.SearchDir, tmplname+c.Extension)
}

func (c *Pool) Load(name string) *template.Template {
	count := 1
	if c.Common != nil {
		count += len(c.Common)
	}
	filenames := make([]string, count)
	if c.Common != nil {
		for i := 0; i < len(c.Common); i++ {
			filenames[i] = c.makeFilename(c.Common[i])
		}
	}
	filenames[count-1] = c.makeFilename(name)
	return template.Must(template.ParseFiles(filenames...))
}

func (c *Pool) Lookup(name string) *template.Template {
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

func (c *Pool) Render(w io.Writer, name string, data interface{}) {
	tmpl := c.Lookup(name)
	if c.BaseDef == "" {
		tmpl.Execute(w, data)
	} else {
		tmpl.ExecuteTemplate(w, c.BaseDef, data)
	}
}
