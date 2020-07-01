package tags

import (
	"fmt"
	"io"
	"strconv"

	"github.com/etecs-ru/liquid/render"
)

type counter interface {
	defaultValue() int
	operation(int) int
}

type increment struct{}

func (increment) defaultValue() int {
	return 0
}

func (increment) operation(i int) int {
	return i + 1
}

type decrement struct{}

func (decrement) defaultValue() int {
	return -1
}

func (decrement) operation(i int) int {
	return i - 1
}

func incrementTag(_ string) (func(io.Writer, render.Context) error, error) {
	return counterCompiler(increment{})
}

func decrementTag(_ string) (func(io.Writer, render.Context) error, error) {
	return counterCompiler(decrement{})
}

func counterCompiler(c counter) (func(io.Writer, render.Context) error, error) {
	return func(w io.Writer, ctx render.Context) error {
		return counterTag(w, ctx, c)
	}, nil
}

func counterTag(w io.Writer, ctx render.Context, c counter) error {
	argStr := ctx.TagArgs()

	state := ctx.GetState("counters", func() interface{} {
		return make(map[string]int)
	})

	var count int
	if counts, ok := state.(map[string]int); !ok {
		return fmt.Errorf("counters state is not of type map[string]int")
	} else {
		if count, ok = counts[argStr]; !ok {
			count = c.defaultValue()
		} else {
			count = c.operation(count)
		}

		counts[argStr] = count
	}

	_, err := w.Write([]byte(strconv.Itoa(count)))
	return err
}
