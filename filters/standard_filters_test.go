package filters

import (
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/etecs-ru/liquid/v2/expressions"
	"github.com/stretchr/testify/require"
)

var filterTests = []struct {
	in       string
	expected interface{}
}{
	// value filters
	{`undefined | default: 2.99`, 2.99},
	{`nil | default: 2.99`, 2.99},
	{`false | default: 2.99`, 2.99},
	{`"" | default: 2.99`, 2.99},
	{`empty_array | default: 2.99`, 2.99},
	{`empty_map | default: 2.99`, 2.99},
	{`empty_map_slice | default: 2.99`, 2.99},
	{`true | default: 2.99`, true},
	{`"true" | default: 2.99`, "true"},
	{`4.99 | default: 2.99`, 4.99},
	{`fruits | default: 2.99 | join`, "apples oranges peaches plums"},

	// array filters
	{`pages | map: 'category' | join`, "business celebrities lifestyle sports technology"},
	{`pages | map: 'category' | compact | join`, "business celebrities lifestyle sports technology"},
	{`"John, Paul, George, Ringo" | split: ", " | join: " and "`, "John and Paul and George and Ringo"},
	{`",John, Paul, George, Ringo" | split: ", " | join: " and "`, ",John and Paul and George and Ringo"},
	{`"John, Paul, George, Ringo," | split: ", " | join: " and "`, "John and Paul and George and Ringo,"},
	{`animals | sort | join: ", "`, "Sally Snake, giraffe, octopus, zebra"},
	{`sort_prop | sort: "weight" | inspect`, `[{"weight":null},{"weight":1},{"weight":3},{"weight":5}]`},
	{`fruits | reverse | join: ", "`, "plums, peaches, oranges, apples"},
	{`fruits | first`, "apples"},
	{`fruits | last`, "plums"},
	{`empty_array | first`, nil},
	{`empty_array | last`, nil},
	// {`empty_array | uniq: 'a'`, []interface{}(nil)}, // TODO uniq with property is not supported
	// {`empty_array | compact: 'a'`, []interface{}(nil)}, // TODO compact with property is not supported
	{`empty_array | sort: 'a'`, []interface{}{}},
	{`dup_ints | uniq | join`, "1 2 3"},
	{`dup_strings | uniq | join`, "one two three"},
	{`dup_maps | uniq | map: "name" | join`, "m1 m2 m3"},
	{`mixed_case_array | sort_natural | join`, "a B c"},
	{`mixed_case_hash_values | sort_natural: 'key' | map: 'key' | join`, "a B c"},

	{`12 | compact`, 12},

	{`dup_ints | concat: dup_strings | join`, "1 2 1 3 one two one three"},
	{`dup_ints | concat: empty_array | join`, "1 2 1 3"},
	{`empty_array | concat: empty_array | join`, ""},

	{`map_slice_has_nil | compact | join`, `a b`},
	{`map_slice_2 | first`, `b`},
	{`map_slice_2 | last`, `a`},
	{`map_slice_2 | join`, `b a`},
	{`map_slice_objs | map: "key" | join`, `a b`},
	{`map_slice_2 | reverse | join`, `a b`},
	{`map_slice_2 | sort | join`, `a b`},
	{`map_slice_dup | join`, `a a b`},
	{`map_slice_dup | uniq | join`, `a b`},

	// date filters
	{`article.published_at | date`, "Fri, Jul 17, 15"},
	{`article.published_at | date: "%a, %b %d, %y"`, "Fri, Jul 17, 15"},
	{`article.published_at | date: "%Y"`, "2015"},
	{`"2017-02-08 19:00:00 -05:00" | date`, "Wed, Feb 08, 17"},
	{`"2017-05-04 08:00:00 -04:00" | date: "%b %d, %Y"`, "May 04, 2017"},
	{`"2017-02-08 09:00:00" | date: "%H:%M"`, "09:00"},
	{`"2017-02-08 09:00:00" | date: "%-H:%M"`, "9:00"},
	{`"2017-02-08 09:00:00" | date: "%d/%m"`, "08/02"},
	{`"2017-02-08 09:00:00" | date: "%e/%m"`, " 8/02"},
	{`"2017-02-08 09:00:00" | date: "%-d/%-m"`, "8/2"},
	{`"March 14, 2016" | date: "%b %d, %y"`, "Mar 14, 16"},
	{`"2017-07-09" | date: "%d/%m"`, "09/07"},
	{`"2017-07-09" | date: "%e/%m"`, " 9/07"},
	{`"2017-07-09" | date: "%-d/%-m"`, "9/7"},

	// sequence (array or string) filters
	{`"Ground control to Major Tom." | size`, 28},
	{`"apples, oranges, peaches, plums" | split: ", " | size`, 4},

	// string filters
	{`"Take my protein pills and put my helmet on" | replace: "my", "your"`, "Take your protein pills and put your helmet on"},
	{`"Take my protein pills and put my helmet on" | replace_first: "my", "your"`, "Take your protein pills and put my helmet on"},
	{`"/my/fancy/url" | append: ".html"`, "/my/fancy/url.html"},
	{`"website.com" | append: "/index.html"`, "website.com/index.html"},
	{`"title" | capitalize`, "Title"},
	{`"my great title" | capitalize`, "My great title"},
	{`"" | capitalize`, ""},
	{`"Parker Moore" | downcase`, "parker moore"},
	{`"Have you read 'James & the Giant Peach'?" | escape`, "Have you read &#39;James &amp; the Giant Peach&#39;?"},
	{`"1 < 2 & 3" | escape_once`, "1 &lt; 2 &amp; 3"},
	{`string_with_newlines | newline_to_br`, "<br />Hello<br />there<br />"},
	{`"1 &lt; 2 &amp; 3" | escape_once`, "1 &lt; 2 &amp; 3"},
	{`"apples, oranges, and bananas" | prepend: "Some fruit: "`, "Some fruit: apples, oranges, and bananas"},
	{`"I strained to see the train through the rain" | remove: "rain"`, "I sted to see the t through the "},
	{`"I strained to see the train through the rain" | remove_first: "rain"`, "I sted to see the train through the rain"},

	{`"Liquid" | slice: 0`, "L"},
	{`"Liquid" | slice: 2`, "q"},
	{`"Liquid" | slice: 2, 5`, "quid"},
	{`"Liquid" | slice: -3, 2`, "ui"},

	{`"a/b/c" | split: '/' | join: '-'`, "a-b-c"},
	{`"a/b/" | split: '/' | join: '-'`, "a-b"},
	{`"a//c" | split: '/' | join: '-'`, "a--c"},
	{`"a//" | split: '/' | join: '-'`, "a"},
	{`"/b/c" | split: '/' | join: '-'`, "-b-c"},
	{`"/b/" | split: '/' | join: '-'`, "-b"},
	{`"//c" | split: '/' | join: '-'`, "--c"},
	{`"//" | split: '/' | join: '-'`, ""},
	{`"/" | split: '/' | join: '-'`, ""},
	{`"a.b" | split: '.' | join: '-'`, "a-b"},
	{`"a..b" | split: '.' | join: '-'`, "a--b"},
	{"'a.\t.b' | split: '.' | join: '-'", "a-\t-b"},
	{`"a b" | split: ' ' | join: '-'`, "a-b"},
	{`"a  b" | split: ' ' | join: '-'`, "a-b"},
	{"'a \t b' | split: ' ' | join: '-'", "a-b"},

	{`"Have <em>you</em> read <strong>Ulysses</strong>?" | strip_html`, "Have you read Ulysses?"},
	{`string_with_newlines | strip_newlines`, "Hellothere"},

	{`"Ground control to Major Tom." | truncate: 20`, "Ground control to..."},
	{`"Ground control to Major Tom." | truncate: 25, ", and so on"`, "Ground control, and so on"},
	{`"Ground control to Major Tom." | truncate: 20, ""`, "Ground control to Ma"},
	{`"Ground" | truncate: 20`, "Ground"},
	{`"Ground control to Major Tom." | truncatewords: 3`, "Ground control to..."},
	{`"Ground control to Major Tom." | truncatewords: 3, "--"`, "Ground control to--"},
	{`"Ground control to Major Tom." | truncatewords: 3, ""`, "Ground control to"},
	{`"Ground control" | truncatewords: 3, ""`, "Ground control"},
	{`"Ground" | truncatewords: 3, ""`, "Ground"},
	{`"  Ground" | truncatewords: 3, ""`, "  Ground"},
	{`"" | truncatewords: 3, ""`, ""},
	{`"  " | truncatewords: 3, ""`, "  "},

	{`"Parker Moore" | upcase`, "PARKER MOORE"},
	{`"          So much room for activities!          " | strip`, "So much room for activities!"},
	{`"          So much room for activities!          " | lstrip`, "So much room for activities!          "},
	{`"          So much room for activities!          " | rstrip`, "          So much room for activities!"},

	{`"%27Stop%21%27+said+Fred" | url_decode`, "'Stop!' said Fred"},
	{`"john@liquid.com" | url_encode`, "john%40liquid.com"},
	{`"Tetsuro Takara" | url_encode`, "Tetsuro+Takara"},

	// number filters
	{`-17 | abs`, int64(17)},
	{`4 | abs`, int64(4)},
	{`"-19.86" | abs`, 19.86},

	{`1.2 | ceil`, int64(2)},
	{`2.0 | ceil`, int64(2)},
	{`183.357 | ceil`, int64(184)},
	{`"183.357" | ceil`, int64(184)},
	{`"3.5" | ceil`, int64(4)},

	{`1.2 | floor`, int64(1)},
	{`2.0 | floor`, int64(2)},
	{`2.0 | floor`, int64(2)},
	{`183.357 | floor`, int64(183)},
	{`183.357 | floor`, int64(183)},

	{`4 | plus: 2`, int64(6)},
	{`"4" | plus: 2`, int64(6)},
	{`"183.357" | plus: "12"`, 195.357},

	{`4 | minus: 2`, int64(2)},
	{`16 | minus: 4`, int64(12)},
	{`"16" | minus: 4`, int64(12)},
	{`183.357 | minus: 12`, 171.357},
	{`"183.357" | minus: 12`, 171.357},

	{`3 | times: 2`, int64(6)},
	{`24 | times: 7`, int64(168)},
	{`"24" | times: 7`, int64(168)},
	{`183.357 | times: 12`, 2200.284},
	{`"183.357" | times: 12`, 2200.284},

	{`3 | modulo: 2`, int64(1)},
	{`24 | modulo: 7`, int64(3)},
	{`"24" | modulo: 7`, int64(3)},
	// {`183.357 | modulo: 12`, 3.357}, // TODO test suite use inexact
	// {`"183.357" | modulo: 12`, 3.357}, // TODO test suite use inexact

	{`16 | divided_by: 4`, int64(4)},
	{`5 | divided_by: 3`, int64(1)},
	{`20 | divided_by: 7`, int64(2)},
	{`"20" | divided_by: 7`, int64(2)},
	{`20 | divided_by: 7.0`, 2.857142857142857},
	{`"20" | divided_by: 7.0`, 2.857142857142857},
	{`"20" | divided_by: 0.0`, math.Inf(1)},
	// {`20 | divided_by: 's'`, nil}, // TODO test for error

	{`1.2 | round`, int64(1)},
	{`2.7 | round`, int64(3)},
	{`"2.7" | round`, int64(3)},
	{`183.357 | round: 2`, 183.36},
	{`"183.357" | round: 2`, 183.36},

	{`4 | at_least: 5`, int64(5)},
	{`4 | at_least: 3`, int64(4)},
	{`3.14 | at_least: 2`, 3.14},
	{`3.14 | at_least: 5`, int64(5)},

	{`4 | at_most: 5`, int64(4)},
	{`4 | at_most: 3`, int64(3)},
	{`3.14 | at_most: 2`, int64(2)},
	{`3.14 | at_most: 5`, 3.14},

	// Jekyll extensions; added here for convenient testing
	// TODO add this just to the test environment
	{`map | inspect`, `{"a":1}`},
	{`1 | type`, `int`},
	{`"1" | type`, `string`},
}

var filterTestBindings = map[string]interface{}{
	"empty_array":     []interface{}{},
	"empty_map":       map[string]interface{}{},
	"empty_map_slice": yaml.MapSlice{},
	"map": map[string]interface{}{
		"a": 1,
	},
	"map_slice_2":       yaml.MapSlice{{Key: 1, Value: "b"}, {Key: 2, Value: "a"}},
	"map_slice_dup":     yaml.MapSlice{{Key: 1, Value: "a"}, {Key: 2, Value: "a"}, {Key: 3, Value: "b"}},
	"map_slice_has_nil": yaml.MapSlice{{Key: 1, Value: "a"}, {Key: 2, Value: nil}, {Key: 3, Value: "b"}},
	"map_slice_objs": yaml.MapSlice{
		{Key: 1, Value: map[string]interface{}{"key": "a"}},
		{Key: 2, Value: map[string]interface{}{"key": "b"}},
	},
	"mixed_case_array": []string{"c", "a", "B"},
	"mixed_case_hash_values": []map[string]interface{}{
		{"key": "c"},
		{"key": "a"},
		{"key": "B"},
	},
	"sort_prop": []map[string]interface{}{
		{"weight": 1},
		{"weight": 5},
		{"weight": 3},
		{"weight": nil},
	},
	"string_with_newlines": "\nHello\nthere\n",
	"dup_ints":             []int{1, 2, 1, 3},
	"dup_strings":          []string{"one", "two", "one", "three"},

	// for examples from liquid docs
	"animals": []string{"zebra", "octopus", "giraffe", "Sally Snake"},
	"fruits":  []string{"apples", "oranges", "peaches", "plums"},
	"article": map[string]interface{}{
		"published_at": timeMustParse("2015-07-17T15:04:05Z"),
	},
	"page": map[string]interface{}{
		"title": "Introduction",
	},
	"pages": []map[string]interface{}{
		{"name": "page 1", "category": "business"},
		{"name": "page 2", "category": "celebrities"},
		{"name": "page 3"},
		{"name": "page 4", "category": "lifestyle"},
		{"name": "page 5", "category": "sports"},
		{"name": "page 6"},
		{"name": "page 7", "category": "technology"},
	},
}

func TestFilters(t *testing.T) {
	require.NoError(t, os.Setenv("TZ", "America/New_York"))

	var (
		m1 = map[string]interface{}{"name": "m1"}
		m2 = map[string]interface{}{"name": "m2"}
		m3 = map[string]interface{}{"name": "m3"}
	)
	filterTestBindings["dup_maps"] = []interface{}{m1, m2, m1, m3}

	cfg := expressions.NewConfig()
	AddStandardFilters(&cfg)
	context := expressions.NewContext(filterTestBindings, cfg)

	for i, test := range filterTests {
		testV := test
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			actual, err := expressions.EvaluateString(testV.in, context)
			require.NoErrorf(t, err, testV.in)
			require.Equalf(t, testV.expected, actual, testV.in)
		})
	}
}

func timeMustParse(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
