package embedx

import (
	"fmt"
	"html/template"
	"io/fs"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

// Template is convenient wrapper of html template
type Template struct {
	*template.Template
}

// ParseFS is different from go's standard implement that it can recursively find all
// template files ending with `.tmpl` and load them with a template name of their path.
func (t Template) ParseFS(fsys fs.FS) (*template.Template, error) {
	var filenames []string
	_ = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logx.Errorf("Parse template fs %q: %v", path, err)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".tmpl") {
			filenames = append(filenames, path)
		}
		return nil
	})
	return parseFiles(t.Template, readFileFS(fsys), filenames...)
}

// parseFiles is the helper for the method and function. If the argument
// template is nil, it is created from the first file.
func parseFiles(t *template.Template, readFile func(string) (string, []byte, error), filenames ...string) (*template.Template, error) {
	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("html/template: no files named in call to ParseFiles")
	}
	for _, filename := range filenames {
		name, b, err := readFile(filename)
		if err != nil {
			return nil, err
		}
		s := string(b)
		// First template becomes return value if not already defined,
		// and we use that one for subsequent New calls to associate
		// all the templates together. Also, if this file has the same name
		// as t, this file becomes the contents of t, so
		//  t, err := New(name).Funcs(xxx).ParseFiles(name)
		// works. Otherwise we create a new template associated with t.
		var tmpl *template.Template
		if t == nil {
			t = template.New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		_, err = tmpl.Parse(s)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

func readFileFS(fsys fs.FS) func(string) (string, []byte, error) {
	return func(file string) (name string, b []byte, err error) {
		name = file
		b, err = fs.ReadFile(fsys, file)
		return
	}
}
