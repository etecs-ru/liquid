package liquid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

var emptyBindings = map[string]interface{}{}

// There's a lot more tests in the filters and tags sub-packages.
// This collects a minimal set for testing end-to-end.
var liquidTests = []struct{ in, expected string }{
	{`{{ page.title }}`, "Introduction"},
	{`{% if x %}true{% endif %}`, "true"},
	{`{{ "upper" | upcase }}`, "UPPER"},
	{`{{ array | upcase }}`, "FIRSTSECONDTHIRD"},
	{`{{ interface_array | upcase }}`, "FIRSTSECONDTHIRD"},
	{`{{ empty_array | upcase }}`, ""},
}

var testBindings = map[string]interface{}{
	"x":               123,
	"array":           []string{"first", "second", "third"},
	"interface_array": []interface{}{"first", "second", "third"},
	"empty_array":     []interface{}{},
	"page": map[string]interface{}{
		"title": "Introduction",
	},
}

func TestEngine_ParseAndRenderString(t *testing.T) {
	engine := NewEngine()
	for i, test := range liquidTests {
		testV := test
		t.Run(fmt.Sprint(i+1), func(t *testing.T) {
			out, err := engine.ParseAndRenderString(testV.in, testBindings)
			require.NoErrorf(t, err, testV.in)
			require.Equalf(t, testV.expected, out, testV.in)
		})
	}
}

func TestEngine_ParseAndRenderString_ptr_to_hash(t *testing.T) {
	params := map[string]interface{}{
		"message": &map[string]interface{}{
			"Text":       "hello",
			"jsonNumber": json.Number("123"),
		},
	}
	engine := NewEngine()
	template := "{{ message.Text }} {{message.jsonNumber}}"
	str, err := engine.ParseAndRenderString(template, params)
	require.NoError(t, err)
	require.Equal(t, "hello 123", str)
}

type testStruct struct{ Text string }

func TestEngine_ParseAndRenderString_struct(t *testing.T) {
	params := map[string]interface{}{
		"message": testStruct{
			Text: "hello",
		},
	}
	engine := NewEngine()
	template := "{{ message.Text }}"
	str, err := engine.ParseAndRenderString(template, params)
	require.NoError(t, err)
	require.Equal(t, "hello", str)
}

func TestEngine_ParseAndRender_errors(t *testing.T) {
	_, err := NewEngine().ParseAndRenderString("{{ syntax error }}", emptyBindings)
	require.Error(t, err)
	_, err = NewEngine().ParseAndRenderString("{% if %}", emptyBindings)
	require.Error(t, err)
	_, err = NewEngine().ParseAndRenderString("{% undefined_tag %}", emptyBindings)
	require.Error(t, err)
	_, err = NewEngine().ParseAndRenderString("{% a | undefined_filter %}", emptyBindings)
	require.Error(t, err)
}

func BenchmarkEngine_Parse(b *testing.B) {
	engine := NewEngine()
	buf := new(bytes.Buffer)
	for i := 0; i < 1000; i++ {
		io.WriteString(buf, `if{% if true %}true{% elsif %}elsif{% else %}else{% endif %}`)
		io.WriteString(buf, `loop{% for item in array %}loop{% break %}{% endfor %}`)
		io.WriteString(buf, `case{% case value %}{% when a %}{% when b %{% endcase %}`)
		io.WriteString(buf, `expr{{ a and b }}{{ a add: b }}`)
	}
	s := buf.Bytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.ParseTemplate(s) // nolint: errcheck
	}
}
