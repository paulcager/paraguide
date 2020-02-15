package templates

import (
	"html/template"
	"io"

	"github.com/paulcager/paraguide/sites"
)

func Execute(name string, w io.Writer, sites []sites.Site, clubs []sites.Club) bool {
	t, ok := all[name]
	if !ok {
		return false
	}
	if err := t.Execute(w, map[string]interface{}{"sites": sites, "Clubs": clubs}); err != nil {
		// TODO
		panic(err)
	}
	return true
}

var (
	all = make(map[string]*template.Template)
)

func init() {
	all["index"] = template.Must(template.New("index").Parse(`
	This would be index.html
`))
}
