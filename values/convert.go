package values

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// A TypeError is an error during type conversion.
type TypeError string

func (e TypeError) Error() string { return string(e) }

func typeErrorf(format string, a ...interface{}) TypeError {
	return TypeError(fmt.Sprintf(format, a...))
}

var timeType = reflect.TypeOf(time.Now())
var numberType = reflect.TypeOf(Number{})

func conversionError(modifier string, value interface{}, typ reflect.Type) error {
	if modifier != "" {
		modifier += " "
	}
	switch ref := value.(type) { // nolint: gocritic
	case reflect.Value:
		value = ref.Interface()
	}
	return typeErrorf("can't convert %s%T(%v) to type %s", modifier, value, value, typ)
}

func convertValueToInt(value interface{}, typ reflect.Type) (int64, error) {
	switch value := value.(type) {
	case bool:
		if value {
			return 1, nil
		}
		return 0, nil
	case string:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, conversionError("", value, typ)
		}
		return v, nil
	case json.Number:
		v, err := strconv.ParseInt(value.String(), 10, 64)
		if err != nil {
			return 0, conversionError("", value, typ)
		}
		return v, nil

	}
	return 0, conversionError("", value, typ)
}

func convertValueToFloat(value interface{}, typ reflect.Type) (float64, error) {
	switch value := value.(type) { // nolint: gocritic
	// case int is handled by rv.Convert(typ) in Convert function
	case string:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, conversionError("", value, typ)
		}
		return v, nil
	case json.Number:
		v, err := strconv.ParseFloat(value.String(), 64)
		if err != nil {
			return 0, conversionError("", value, typ)
		}
		return v, nil
	}
	return 0, conversionError("", value, typ)
}

type Number struct {
	Value   interface{}
	IsFloat bool
}

func (n Number) AsInt64() int64 {
	if n.IsFloat {
		return int64(n.Value.(float64))
	} else {
		return n.Value.(int64)
	}
}

func (n Number) AsFloat64() float64 {
	if n.IsFloat {
		return n.Value.(float64)
	} else {
		return float64(n.Value.(int64))
	}
}

func convertValueToNumber(value interface{}, typ reflect.Type) Number {
	if value == nil {
		return Number{int64(0), false}
	}

	switch x := value.(type) {
	case int:
		return Number{int64(x), false}
	case int16:
		return Number{int64(x), false}
	case int32:
		return Number{int64(x), false}
	case int64:
		return Number{x, false}
	case uint:
		return Number{int64(x), false}
	case uint16:
		return Number{int64(x), false}
	case uint32:
		return Number{int64(x), false}
	case uint64:
		return Number{int64(x), false}
	case float32:
		return Number{float64(x), true}
	case float64:
		return Number{x, true}
	}

	if i, err := convertValueToInt(value, typ); err == nil {
		return Number{i, false}
	}
	if f, err := convertValueToFloat(value, typ); err == nil {
		return Number{f, true}
	}
	return Number{int64(0), false}
}

// Convert value to the type. This is a more aggressive conversion, that will
// recursively create new map and slice values as necessary. It doesn't
// handle circular references.
func Convert(value interface{}, typ reflect.Type) (interface{}, error) { // nolint: gocyclo
	value = ToLiquid(value)
	rv := reflect.ValueOf(value)
	// int.Convert(string) returns "\x01" not "1", so guard against that in the following test
	if typ.Kind() != reflect.String && value != nil && rv.Type().ConvertibleTo(typ) {
		return rv.Convert(typ).Interface(), nil
	}
	if typ == timeType && rv.Kind() == reflect.String {
		return ParseDate(value.(string))
	}
	if typ == numberType {
		return convertValueToNumber(value, typ), nil
	}
	// currently unused:
	// case reflect.PtrTo(r.Type()) == typ:
	// 	return &value, nil
	// }
	switch typ.Kind() {
	case reflect.Bool:
		return !(value == nil || value == false), nil
	case reflect.Uint:
		v, err := convertValueToInt(value, typ)
		return uint(v), err
	case reflect.Uint8:
		v, err := convertValueToInt(value, typ)
		return uint8(v), err
	case reflect.Uint16:
		v, err := convertValueToInt(value, typ)
		return uint16(v), err
	case reflect.Uint32:
		v, err := convertValueToInt(value, typ)
		return uint32(v), err
	case reflect.Uint64:
		v, err := convertValueToInt(value, typ)
		return uint64(v), err
	case reflect.Int:
		v, err := convertValueToInt(value, typ)
		return int(v), err
	case reflect.Int8:
		v, err := convertValueToInt(value, typ)
		return int8(v), err
	case reflect.Int16:
		v, err := convertValueToInt(value, typ)
		return int16(v), err
	case reflect.Int32:
		v, err := convertValueToInt(value, typ)
		return int32(v), err
	case reflect.Int64:
		v, err := convertValueToInt(value, typ)
		return v, err
	case reflect.Float32:
		v, err := convertValueToFloat(value, typ)
		return float32(v), err
	case reflect.Float64:
		v, err := convertValueToFloat(value, typ)
		return v, err
	case reflect.Map:
		et := typ.Elem()
		result := reflect.MakeMap(typ)
		if ms, ok := value.(yaml.MapSlice); ok {
			for _, item := range ms {
				var k, v reflect.Value
				if item.Key == nil {
					k = reflect.Zero(typ.Key())
				} else {
					kc, err := Convert(item.Key, typ.Key())
					if err != nil {
						return nil, err
					}
					k = reflect.ValueOf(kc)
				}
				if item.Value == nil {
					v = reflect.Zero(et)
				} else {
					ec, err := Convert(item.Value, et)
					if err != nil {
						return nil, err
					}
					v = reflect.ValueOf(ec)
				}
				result.SetMapIndex(k, v)
			}
			return result.Interface(), nil
		}
		if rv.Kind() != reflect.Map {
			return nil, conversionError("", value, typ)
		}
		for _, key := range rv.MapKeys() {
			if typ.Key().Kind() == reflect.String {
				key = reflect.ValueOf(fmt.Sprint(key))
			}
			if !key.Type().ConvertibleTo(typ.Key()) {
				return nil, conversionError("map key", key, typ.Key())
			}
			key = key.Convert(typ.Key())
			ev := rv.MapIndex(key)
			if et.Kind() == reflect.String {
				ev = reflect.ValueOf(fmt.Sprint(ev))
			}
			if !ev.Type().ConvertibleTo(et) {
				return nil, conversionError("map element", ev, et)
			}
			result.SetMapIndex(key, ev.Convert(et))
		}
		return result.Interface(), nil
	case reflect.Slice:
		et := typ.Elem()
		if ms, ok := value.(yaml.MapSlice); ok {
			result := reflect.MakeSlice(typ, 0, rv.Len())
			for _, item := range ms {
				if item.Value == nil {
					if et.Kind() >= reflect.Array {
						ev := reflect.Zero(et)
						result = reflect.Append(result, ev.Convert(et))
					}
					continue
				}
				ev := reflect.ValueOf(item.Value)
				if et.Kind() == reflect.String {
					ev = reflect.ValueOf(fmt.Sprint(ev))
				}
				if !ev.Type().ConvertibleTo(et) {
					return nil, conversionError("slice element", ev, et)
				}
				result = reflect.Append(result, ev.Convert(et))
			}
			return result.Interface(), nil
		}
		switch rv.Kind() {
		case reflect.Array, reflect.Slice:
			result := reflect.MakeSlice(typ, 0, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				item, err := Convert(rv.Index(i).Interface(), typ.Elem())
				if err != nil {
					return nil, err
				}
				result = reflect.Append(result, reflect.ValueOf(item))
			}
			return result.Interface(), nil
		case reflect.Map:
			result := reflect.MakeSlice(typ, 0, rv.Len())
			for _, key := range rv.MapKeys() {
				item, err := Convert(rv.MapIndex(key).Interface(), typ.Elem())
				if err != nil {
					return nil, err
				}
				result = reflect.Append(result, reflect.ValueOf(item))
			}
			return result.Interface(), nil
		}
	case reflect.String:
		return convertToString(value), nil
	}
	return nil, conversionError("", value, typ)
}

func convertToString(value interface{}) string {
	switch value := value.(type) {
	case []byte:
		return string(value)
	case fmt.Stringer:
		return value.String()
	default:
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			var b strings.Builder
			for i := 0; i < rv.Len(); i++ {
				b.WriteString(convertToString(rv.Index(i).Interface()))
			}
			return b.String()
		default:
			return fmt.Sprint(value)
		}
	}
}

// MustConvert is like Convert, but panics if conversion fails.
func MustConvert(value interface{}, t reflect.Type) interface{} {
	out, err := Convert(value, t)
	if err != nil {
		panic(err)
	}
	return out
}

// MustConvertItem converts item to conform to the type array's element, else panics.
// Unlike MustConvert, the second argument is a value not a type.
func MustConvertItem(item interface{}, array interface{}) interface{} {
	item, err := Convert(item, reflect.TypeOf(array).Elem())
	if err != nil {
		panic(typeErrorf("can't convert %#v to %s: %s", item, reflect.TypeOf(array).Elem(), err))
	}
	return item
}
