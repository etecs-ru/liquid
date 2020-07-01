// Package tags is an internal package that defines the standard Liquid tags.
package tags

import (
	"io"
	"strings"

	"github.com/etecs-ru/liquid/expressions"
	"github.com/etecs-ru/liquid/render"
)

// AddStandardTags defines the standard Liquid tags.
func AddStandardTags(c render.Config) {
	c.AddTag("assign", assignTag)
	c.AddTag("include", includeTag)
	c.AddTag("increment", incrementTag)
	c.AddTag("decrement", decrementTag)

	// blocks
	// The parser only recognize the comment and raw tags if they've been defined,
	// but it ignores any syntax specified here.
	c.AddTag("break", breakTag)
	c.AddTag("continue", continueTag)
	c.AddTag("cycle", cycleTag)
	c.AddBlock("capture").Compiler(captureTagCompiler)
	c.AddBlock("case").Clause("when").Clause("else").Compiler(caseTagCompiler)
	c.AddBlock("comment")
	c.AddBlock("for").Compiler(loopTagCompiler)
	c.AddBlock("if").Clause("else").Clause("elsif").Compiler(ifTagCompiler(true))
	c.AddBlock("raw")
	c.AddBlock("tablerow").Compiler(loopTagCompiler)
	c.AddBlock("unless").Clause("else").Clause("elsif").Compiler(ifTagCompiler(false))
}

func assignTag(source string) (func(io.Writer, render.Context) error, error) {
	stmt, err := expressions.ParseStatement(expressions.AssignStatementSelector, source)
	if err != nil {
		return nil, err
	}
	return func(w io.Writer, ctx render.Context) error {
		value, err := ctx.Evaluate(stmt.Assignment.ValueFn)
		if err != nil {
			return err
		}
		_ = value
		ctx.Set(stmt.Assignment.Variable, value)
		return nil
	}, nil
}

func captureTagCompiler(node render.BlockNode) (func(io.Writer, render.Context) error, error) {
	// only the first "word" gets used as the variable name; i.e., {% capture x y z %} results in a variable named 'x'.
	varname := strings.Fields(node.Args)[0]
	return func(w io.Writer, ctx render.Context) error {
		s, err := ctx.InnerString()
		if err != nil {
			return err
		}
		ctx.Set(varname, s)
		return nil
	}, nil
}
