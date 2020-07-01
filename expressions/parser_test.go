package expressions

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var parseTests = []struct {
	in     string
	expect interface{}
}{
	{`true`, true},
	{`false`, false},
	{`nil`, nil},
	{`2`, 2},
	{`"s"`, "s"},
	{`a`, 1},
	{`obj.prop`, 2},
	{`a | add: b`, 3},
	{`1 == 1`, true},
	{`1 != 1`, false},
	{`true and true`, true},
}

var parseErrorTests = []struct{ in, expected string }{
	{"a syntax error", "syntax error"},
	{`%assign a`, "syntax error"},
	{`%assign a 3`, "syntax error"},
	{`%cycle 'a' 'b'`, "syntax error"},
	{`%loop a in in`, "syntax error"},
	{`%when a b`, "syntax error"},
}

// Since the parser returns funcs, there's no easy way to test them except evaluation
func TestParse(t *testing.T) {
	cfg := NewConfig()
	cfg.AddFilter("add", func(a, b int) int { return a + b })
	ctx := NewContext(map[string]interface{}{
		"a":   1,
		"b":   2,
		"obj": map[string]int{"prop": 2},
	}, cfg)
	for i, test := range parseTests {
		testV := test
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			expr, err := Parse(testV.in)
			require.NoError(t, err, testV.in)
			_ = expr
			value, err := expr.Evaluate(ctx)
			require.NoError(t, err, testV.in)
			require.Equal(t, testV.expect, value, testV.in)
		})
	}
}

func TestParse_errors(t *testing.T) {
	for i, test := range parseErrorTests {
		testV := test
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			expr, err := Parse(testV.in)
			require.Nilf(t, expr, testV.in)
			require.Errorf(t, err, testV.in, testV.in)
			require.Containsf(t, err.Error(), testV.expected, testV.in)
		})
	}
}
