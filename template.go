package liquid

import (
	"bytes"

	"github.com/etecs-ru/liquid/parser"
	"github.com/etecs-ru/liquid/render"
)

// A Template is a compiled Liquid template. It knows how to evaluate itself within a variable binding environment, to create a rendered byte slice.
//
// Use Engine.ParseTemplate to create a template.
type Template struct {
	root render.Node
	cfg  *render.Config
}

func newTemplate(cfg *render.Config, source []byte, path string, line int) (*Template, SourceError) {
	loc := parser.SourceLoc{Pathname: path, LineNo: line}
	root, err := cfg.Compile(string(source), loc)
	if err != nil {
		return nil, err
	}
	return &Template{root, cfg}, nil
}

// Render executes the template with the specified variable bindings.
func (t *Template) Render(vars Bindings) ([]byte, SourceError) {
	return t.RenderWithState(vars, map[string]interface{}{})
}

func (t *Template) RenderWithState(vars, state Bindings) ([]byte, SourceError) {
	buf := new(bytes.Buffer)
	err := render.RenderWithState(t.root, buf, vars, state, *t.cfg)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderString is a convenience wrapper for Render, that has string input and output.
func (t *Template) RenderString(b Bindings) (string, SourceError) {
	return t.RenderStringWithState(b, map[string]interface{}{})
}

func (t *Template) RenderStringWithState(b, state Bindings) (string, SourceError) {
	bs, err := t.RenderWithState(b, state)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func (t *Template) FindVariables() (map[string]interface{}, SourceError) {
	return render.FindVariables(t.root, *t.cfg)
}
