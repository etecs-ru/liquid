// Package filters is an internal package that defines the standard Liquid filters.
package filters

import (
	"encoding/json"
	"fmt"
	"html"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/osteele/liquid/values"
	"github.com/osteele/tuesday"
)

// A FilterDictionary holds filters.
type FilterDictionary interface {
	AddFilter(string, interface{})
}

// AddStandardFilters defines the standard Liquid filters.
func AddStandardFilters(fd FilterDictionary) { // nolint: gocyclo
	// value filters
	fd.AddFilter("default", func(value, defaultValue interface{}) interface{} {
		if value == nil || value == false || values.IsEmpty(value) {
			value = defaultValue
		}
		return value
	})

	// array filters
	fd.AddFilter("concat", func(a, b []interface{}) []interface{} {
		result := make([]interface{}, len(a)+len(b))
		copy(result, a)
		return append(result, b...)
	})

	fd.AddFilter("compact", func(a interface{}) interface{} {
		arr, ok := values.IsArray(a)
		if !ok {
			return a
		}

		var result []interface{}
		for _, item := range arr {
			if item != nil {
				result = append(result, item)
			}
		}
		return result
	})
	fd.AddFilter("join", joinFilter)
	fd.AddFilter("map", func(a []map[string]interface{}, key string) (result []interface{}) {
		for _, obj := range a {
			result = append(result, obj[key])
		}
		return result
	})
	fd.AddFilter("reverse", reverseFilter)
	fd.AddFilter("sort", sortFilter)
	// https://shopify.github.io/liquid/ does not demonstrate first and last as filters,
	// but https://help.shopify.com/themes/liquid/filters/array-filters does
	fd.AddFilter("first", func(a []interface{}) interface{} {
		if len(a) == 0 {
			return nil
		}
		return a[0]
	})
	fd.AddFilter("last", func(a []interface{}) interface{} {
		if len(a) == 0 {
			return nil
		}
		return a[len(a)-1]
	})
	fd.AddFilter("uniq", uniqFilter)

	// date filters
	fd.AddFilter("date", func(t time.Time, format func(string) string) (string, error) {
		f := format("%a, %b %d, %y")
		return tuesday.Strftime(f, t)
	})

	// number filters
	fd.AddFilter("abs", stdUnaryMathOperation(math.Abs).Call)
	fd.AddFilter("ceil", func(a values.Number) int64 {
		return int64(math.Ceil(a.AsFloat64()))
	})
	fd.AddFilter("floor", func(a values.Number) int64 {
		return int64(math.Floor(a.AsFloat64()))
	})
	fd.AddFilter("at_least", atLeast)
	fd.AddFilter("at_most", atMost)
	fd.AddFilter("modulo", stdBinaryMathOperation(math.Mod).Call)
	fd.AddFilter("minus", commonNumberOperation{
		Int64: func(a, b int64) int64 {
			return a - b
		},
		Float64: func(a, b float64) float64 {
			return a - b
		},
	}.Call)
	fd.AddFilter("plus", commonNumberOperation{
		Int64: func(a, b int64) int64 {
			return a + b
		},
		Float64: func(a, b float64) float64 {
			return a + b
		},
	}.Call)
	fd.AddFilter("times", commonNumberOperation{
		Int64: func(a, b int64) int64 {
			return a * b
		},
		Float64: func(a, b float64) float64 {
			return a * b
		},
	}.Call)
	fd.AddFilter("divided_by", func(a float64, b values.Number) (interface{}, error) {
		if b.IsFloat {
			return a / b.AsFloat64(), nil
		} else {
			i := b.AsInt64()
			if i == 0 {
				return nil, fmt.Errorf("divided by 0")
			}
			return int64(a) / i, nil
		}
	})
	fd.AddFilter("round", func(n values.Number, places func(int) int) interface{} {
		pl := places(0)
		exp := math.Pow10(pl)
		result := math.Floor(n.AsFloat64()*exp+0.5) / exp

		if n.IsFloat && pl > 0 {
			return result
		} else {
			return int64(result)
		}
	})

	// sequence filters
	fd.AddFilter("size", values.Length)

	// string filters
	fd.AddFilter("append", func(s, suffix string) string {
		return s + suffix
	})
	fd.AddFilter("capitalize", func(s string) string {
		if len(s) == 0 {
			return s
		}
		return strings.ToUpper(s[:1]) + s[1:]
	})
	fd.AddFilter("downcase", strings.ToLower)
	fd.AddFilter("escape", html.EscapeString)
	fd.AddFilter("escape_once", func(s string) string {
		return html.EscapeString(html.UnescapeString(s))
	})
	fd.AddFilter("newline_to_br", func(s string) string {
		return strings.Replace(s, "\n", "<br />", -1)
	})
	fd.AddFilter("prepend", func(s, prefix string) string {
		return prefix + s
	})
	fd.AddFilter("remove", func(s, old string) string {
		return strings.Replace(s, old, "", -1)
	})
	fd.AddFilter("remove_first", func(s, old string) string {
		return strings.Replace(s, old, "", 1)
	})
	fd.AddFilter("replace", func(s, old, new string) string {
		return strings.Replace(s, old, new, -1)
	})
	fd.AddFilter("replace_first", func(s, old, new string) string {
		return strings.Replace(s, old, new, 1)
	})
	fd.AddFilter("sort_natural", sortNaturalFilter)
	fd.AddFilter("slice", func(s string, start int, length func(int) int) string {
		// runes aren't bytes; don't use slice
		n := length(1)
		if start < 0 {
			start = utf8.RuneCountInString(s) + start
		}
		p := regexp.MustCompile(fmt.Sprintf(`^.{%d}(.{0,%d}).*$`, start, n))
		return p.ReplaceAllString(s, "$1")
	})
	fd.AddFilter("split", splitFilter)
	fd.AddFilter("strip_html", func(s string) string {
		// TODO this probably isn't sufficient
		return regexp.MustCompile(`<.*?>`).ReplaceAllString(s, "")
	})
	fd.AddFilter("strip_newlines", func(s string) string {
		return strings.Replace(s, "\n", "", -1)
	})
	fd.AddFilter("strip", strings.TrimSpace)
	fd.AddFilter("lstrip", func(s string) string {
		return strings.TrimLeftFunc(s, unicode.IsSpace)
	})
	fd.AddFilter("rstrip", func(s string) string {
		return strings.TrimRightFunc(s, unicode.IsSpace)
	})
	fd.AddFilter("truncate", func(s string, length func(int) int, ellipsis func(string) string) string {
		n := length(50)
		el := ellipsis("...")
		// runes aren't bytes; don't use slice
		re := regexp.MustCompile(fmt.Sprintf(`^(.{%d})..{%d,}`, n-len(el), len(el)))
		return re.ReplaceAllString(s, `$1`+el)
	})
	fd.AddFilter("truncatewords", func(s string, length func(int) int, ellipsis func(string) string) string {
		el := ellipsis("...")
		n := length(15)
		re := regexp.MustCompile(fmt.Sprintf(`^(?:\s*\S+){%d}`, n))
		m := re.FindString(s)
		if m == "" {
			return s
		}
		return m + el
	})
	fd.AddFilter("upcase", strings.ToUpper)
	fd.AddFilter("url_encode", url.QueryEscape)
	fd.AddFilter("url_decode", url.QueryUnescape)

	// debugging filters
	// inspect is from Jekyll
	fd.AddFilter("inspect", func(value interface{}) string {
		s, err := json.Marshal(value)
		if err != nil {
			return fmt.Sprintf("%#v", value)
		}
		return string(s)
	})
	fd.AddFilter("type", func(value interface{}) string {
		return fmt.Sprintf("%T", value)
	})
}

func joinFilter(a []interface{}, sep func(string) string) interface{} {
	ss := make([]string, 0, len(a))
	s := sep(" ")
	for _, v := range a {
		if v != nil {
			ss = append(ss, fmt.Sprint(v))
		}
	}
	return strings.Join(ss, s)
}

func reverseFilter(a []interface{}) interface{} {
	result := make([]interface{}, len(a))
	for i, x := range a {
		result[len(result)-1-i] = x
	}
	return result
}

var wsre = regexp.MustCompile(`[[:space:]]+`)

func splitFilter(s, sep string) interface{} {
	result := strings.Split(s, sep)
	if sep == " " {
		// Special case for Ruby, therefore Liquid
		result = wsre.Split(s, -1)
	}
	// This matches Ruby / Liquid / Jekyll's observed behavior.
	for len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}
	return result
}

func uniqFilter(a []interface{}) (result []interface{}) {
	seenMap := map[interface{}]bool{}
	seen := func(item interface{}) bool {
		if k := reflect.TypeOf(item).Kind(); k < reflect.Array || k == reflect.Ptr || k == reflect.UnsafePointer {
			if seenMap[item] {
				return true
			}
			seenMap[item] = true
			return false
		}
		// the O(n^2) case:
		for _, other := range result {
			if eqItems(item, other) {
				return true
			}
		}
		return false
	}
	for _, item := range a {
		if !seen(item) {
			result = append(result, item)
		}
	}
	return
}

func eqItems(a, b interface{}) bool {
	if reflect.TypeOf(a).Comparable() && reflect.TypeOf(b).Comparable() {
		return a == b
	}
	return reflect.DeepEqual(a, b)
}

type commonNumberOperation struct {
	Int64   func(int64, int64) int64
	Float64 func(float64, float64) float64
}

func (op commonNumberOperation) Call(lhs, rhs values.Number) interface{} {
	if lhs.IsFloat || rhs.IsFloat {
		return op.Float64(lhs.AsFloat64(), rhs.AsFloat64())
	} else {
		return op.Int64(lhs.AsInt64(), rhs.AsInt64())
	}
}

type stdUnaryMathOperation func(float64) float64

func (op stdUnaryMathOperation) Call(num values.Number) interface{} {
	result := op(num.AsFloat64())
	if num.IsFloat {
		return result
	} else {
		return int64(result)
	}
}

type stdBinaryMathOperation func(float64, float64) float64

func (op stdBinaryMathOperation) Call(lhs, rhs values.Number) interface{} {
	result := op(lhs.AsFloat64(), rhs.AsFloat64())
	if lhs.IsFloat || rhs.IsFloat {
		return result
	} else {
		return int64(result)
	}
}

// equivalent to math.Max
func atLeast(num, comp values.Number) interface{} {
	// both integers
	if !num.IsFloat && !comp.IsFloat {
		if num.AsInt64() > comp.AsInt64() {
			return num.Value
		} else {
			return comp.Value
		}
	}

	fNum := num.AsFloat64()
	fComp := comp.AsFloat64()

	// special cases (from math.Max)
	switch {
	case math.IsInf(fNum, 1) || math.IsInf(fComp, 1):
		return math.Inf(1)
	case math.IsNaN(fNum) || math.IsNaN(fComp):
		return math.NaN()
	case fNum == 0 && fNum == fComp:
		if math.Signbit(fNum) {
			return comp.Value
		}
		return num.Value
	}

	if fNum > fComp {
		return num.Value
	}
	return comp.Value
}

// equivalent to math.Min
func atMost(num, comp values.Number) interface{} {
	// both integers
	if !num.IsFloat && !comp.IsFloat {
		if num.AsInt64() < comp.AsInt64() {
			return num.Value
		} else {
			return comp.Value
		}
	}

	fNum := num.AsFloat64()
	fComp := comp.AsFloat64()

	// special cases (from math.Min)
	switch {
	case math.IsInf(fNum, -1) || math.IsInf(fComp, -1):
		return math.Inf(-1)
	case math.IsNaN(fNum) || math.IsNaN(fComp):
		return math.NaN()
	case fNum == 0 && fNum == fComp:
		if math.Signbit(fNum) {
			return num.Value
		}
		return comp.Value
	}

	if fNum < fComp {
		return num.Value
	}
	return comp.Value
}
