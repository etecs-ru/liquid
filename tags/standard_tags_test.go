package tags

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/etecs-ru/liquid/v2/parser"
	"github.com/etecs-ru/liquid/v2/render"
	"github.com/stretchr/testify/require"
)

var parseErrorTests = []struct{ in, expected string }{
	{"{% undefined_tag %}", "undefined tag"},
	{"{% assign v x y z %}", "syntax error"},
	{"{% if syntax error %}", `unterminated "if" block`},
	// TODO once expression parsing is moved to template parse stage
	// {"{% if syntax error %}{% endif %}", "syntax error"},
	// {"{% for a in ar undefined %}{{ a }} {% endfor %}", "TODO"},
}

var tagTests = []struct{ in, expected string }{
	// variable tags
	{`{% assign av = 1 %}{{ av }}`, "1"},
	{`{% assign av = obj.a %}{{ av }}`, "1"},
	{`{% capture x %}captured{% endcapture %}{{ x }}`, "captured"},
	{`{% capture x y %}captured{% endcapture %}{{ x }}`, "captured"},

	// TODO research whether Liquid requires matching interior tags
	{`{% comment %}{{ a }}{% undefined_tag %}{% endcomment %}`, ""},

	// TODO research whether Liquid requires matching interior tags
	{`pre{% raw %}{{ a }}{% undefined_tag %}{% endraw %}post`, "pre{{ a }}{% undefined_tag %}post"},
	{`pre{% raw %}{% if false %}anyway-{% endraw %}post`, "pre{% if false %}anyway-post"},

	// increment/decrement
	{`{% increment abc %}{% increment abc %}{% increment abc %}`, "012"},
	{`{% decrement abc %}{% decrement abc %}{% decrement abc %}`, "-1-2-3"},
	{`{% increment abc %}{% increment abc %}{% increment abc %}{% decrement abc %}`, "0121"},
	{`{% increment abc %}{% increment abc %}{% increment def %}{% decrement abc %}`, "0100"},
}

var tagErrorTests = []struct{ in, expected string }{
	{`{% assign av = x | undefined_filter %}`, "undefined filter"},
}

// this is also used in the other test files
var tagTestBindings = map[string]interface{}{
	"x": 123,
	"obj": map[string]interface{}{
		"a": 1,
	},
	"animals": []string{"zebra", "octopus", "giraffe", "Sally Snake"},
	"pages": []map[string]interface{}{
		{"category": "business"},
		{"category": "celebrities"},
		{},
		{"category": "lifestyle"},
		{"category": "sports"},
		{},
		{"category": "technology"},
	},
	"sort_prop": []map[string]interface{}{
		{"weight": 1},
		{"weight": 5},
		{"weight": 3},
		{"weight": nil},
	},
	"page": map[string]interface{}{
		"title": "Introduction",
	},
}

func TestStandardTags_parse_errors(t *testing.T) {
	settings := render.NewConfig()
	AddStandardTags(settings)
	for i, test := range parseErrorTests {
		testV := test
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := settings.Compile(testV.in, parser.SourceLoc{})
			require.Nilf(t, root, testV.in)
			require.Errorf(t, err, testV.in)
			require.Containsf(t, err.Error(), testV.expected, testV.in)
		})
	}
}

func TestStandardTags(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(config)
	for i, test := range tagTests {
		testV := test
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := config.Compile(testV.in, parser.SourceLoc{})
			require.NoErrorf(t, err, testV.in)
			buf := new(bytes.Buffer)
			err = render.Render(root, buf, tagTestBindings, config)
			require.NoErrorf(t, err, testV.in)
			require.Equalf(t, testV.expected, buf.String(), testV.in)
		})
	}
}

func TestStandardTags_render_errors(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(config)
	for i, test := range tagErrorTests {
		testV := test
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := config.Compile(testV.in, parser.SourceLoc{})
			require.NoErrorf(t, err, testV.in)
			err = render.Render(root, ioutil.Discard, tagTestBindings, config)
			require.Errorf(t, err, testV.in)
			require.Containsf(t, err.Error(), testV.expected, testV.in)
		})
	}
}
