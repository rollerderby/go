package json

import (
	"fmt"
	"sort"
	"strconv"
)

type ValueType uint8

const (
	StringValue ValueType = iota
	NumberValue
	ObjectValue
	ArrayValue
	TrueValue
	FalseValue
	NullValue
)

func (vt ValueType) String() string {
	switch vt {
	case StringValue:
		return "String"
	case NumberValue:
		return "Number"
	case ObjectValue:
		return "Object"
	case ArrayValue:
		return "Array"
	case TrueValue:
		return "True"
	case FalseValue:
		return "False"
	case NullValue:
		return "Null"
	}
	return fmt.Sprintf("Unknown: %v", uint8(vt))
}

type Value interface {
	Type() ValueType
	json(indent bool, prefix string) string
	JSON(indent bool) string
}

type String struct{ val string }
type Number struct{ val string }
type _True bool
type _False bool
type _Null bool
type Array []Value
type Object map[string]Value

var True _True
var False _False
var Null _Null

func (v _True) Type() ValueType  { return TrueValue }
func (v _False) Type() ValueType { return FalseValue }
func (v _Null) Type() ValueType  { return NullValue }

func NewString(val string) *String { return &String{val: val} }
func NewArray() Array              { return nil }
func NewObject() Object            { return make(Object) }
func NewNumber(val int64) *Number  { num := &Number{}; num.SetInt64(val); return num }

func (v *String) Type() ValueType   { return StringValue }
func (v *String) Set(val string)    { v.val = val }
func (v *String) Get() string       { return v.val }
func (v *String) String() string    { return v.val }
func (v *String) AsNumber() *Number { n := Number(*v); return &n }

func (v *Number) Type() ValueType { return NumberValue }

func (v *Number) SetUint8(val uint8) {
	v.val = strconv.FormatUint(uint64(val), 10)
}
func (v *Number) GetUint8() (uint8, error) {
	val, err := strconv.ParseUint(v.val, 10, 8)
	return uint8(val), err
}
func (v *Number) SetUint16(val uint16) {
	v.val = strconv.FormatUint(uint64(val), 10)
}
func (v *Number) GetUint16() (uint16, error) {
	val, err := strconv.ParseUint(v.val, 10, 16)
	return uint16(val), err
}
func (v *Number) SetUint32(val uint32) {
	v.val = strconv.FormatUint(uint64(val), 10)
}
func (v *Number) GetUint32() (uint32, error) {
	val, err := strconv.ParseUint(v.val, 10, 32)
	return uint32(val), err
}
func (v *Number) SetUint64(val uint64) {
	v.val = strconv.FormatUint(val, 10)
}
func (v *Number) GetUint64() (uint64, error) {
	return strconv.ParseUint(v.val, 10, 64)
}

func (v *Number) SetInt8(val int8) {
	v.val = strconv.FormatInt(int64(val), 10)
}
func (v *Number) GetInt8() (int8, error) {
	val, err := strconv.ParseInt(v.val, 10, 8)
	return int8(val), err
}
func (v *Number) SetInt16(val int16) {
	v.val = strconv.FormatInt(int64(val), 10)
}
func (v *Number) GetInt16() (int16, error) {
	val, err := strconv.ParseInt(v.val, 10, 16)
	return int16(val), err
}
func (v *Number) SetInt32(val int32) {
	v.val = strconv.FormatInt(int64(val), 10)
}
func (v *Number) GetInt32() (int32, error) {
	val, err := strconv.ParseInt(v.val, 10, 32)
	return int32(val), err
}
func (v *Number) SetInt64(val int64) {
	v.val = strconv.FormatInt(val, 10)
}
func (v *Number) GetInt64() (int64, error) {
	return strconv.ParseInt(v.val, 10, 64)
}
func (v *Number) String() string { return v.val }

func (v *Number) SetFloat64(val float64)       { v.val = strconv.FormatFloat(val, 'g', -1, 64) }
func (v *Number) GetFloat64() (float64, error) { return strconv.ParseFloat(v.val, 64) }

func (v Array) Type() ValueType  { return ArrayValue }
func (v Object) Type() ValueType { return ObjectValue }

func formatJSONString(v string) string {
	buf := GetBuffer()

	buf.WriteRune('"')
	for _, r := range v {
		switch r {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '/':
			buf.WriteString(`\/`)
		case '\b':
			buf.WriteString(`\b`)
		case '\f':
			buf.WriteString(`\f`)
		case '\n':
			buf.WriteString(`\n`)
		case '\t':
			buf.WriteString(`\t`)
		case '\r':
			buf.WriteString(`\r`)
		default:
			buf.WriteRune(r)
		}
	}
	buf.WriteRune('"')

	ret := buf.String()
	buf.Return()
	return ret
}

func (v *String) JSON(indent bool) string { return v.json(indent, "") }
func (v *Number) JSON(indent bool) string { return v.json(indent, "") }
func (v _True) JSON(indent bool) string   { return v.json(indent, "") }
func (v _False) JSON(indent bool) string  { return v.json(indent, "") }
func (v _Null) JSON(indent bool) string   { return v.json(indent, "") }
func (v Array) JSON(indent bool) string   { return v.json(indent, "") }
func (v Object) JSON(indent bool) string  { return v.json(indent, "") }

func (v *String) json(indent bool, prefix string) string { return formatJSONString(v.val) }
func (v *Number) json(indent bool, prefix string) string { return v.val }
func (v _True) json(indent bool, prefix string) string   { return "true" }
func (v _False) json(indent bool, prefix string) string  { return "false" }
func (v _Null) json(indent bool, prefix string) string   { return "null" }

func (v Array) json(indent bool, prefix string) string {
	if len(v) == 0 {
		return "[]"
	}

	buf := GetBuffer()
	var childPrefix string
	if indent {
		childPrefix = prefix + "  "
		buf.WriteString("[\n")
	} else {
		buf.WriteString("[")
	}

	first := true
	for _, v2 := range v {
		if !first {
			if indent {
				buf.WriteString(",\n")
			} else {
				buf.WriteString(", ")
			}
		} else {
			first = false
		}
		if indent {
			buf.WriteString(childPrefix)
		}
		buf.WriteString(v2.json(indent, childPrefix))
	}
	if indent {
		buf.WriteString("\n")
		buf.WriteString(prefix)
	}
	buf.WriteRune(']')

	ret := buf.String()
	buf.Return()
	return ret
}

func (v Object) json(indent bool, prefix string) string {
	if len(v) == 0 {
		return "{}"
	}

	buf := GetBuffer()

	keys := make([]string, len(v))
	idx := 0
	for key := range v {
		keys[idx] = key
		idx++
	}
	sort.Strings(keys)

	childPrefix := ""
	if indent {
		childPrefix = prefix + "  "
		buf.WriteString("{\n")
	} else {
		buf.WriteString("{")
	}

	first := true
	for _, key := range keys {
		v2 := v[key]

		if !first {
			if indent {
				buf.WriteString(",\n")
			} else {
				buf.WriteString(", ")
			}
		} else {
			first = false
		}
		if indent {
			buf.WriteString(childPrefix)
		}
		buf.WriteString(formatJSONString(key))
		buf.WriteString(": ")
		buf.WriteString(v2.json(indent, childPrefix))
	}
	if indent {
		buf.WriteString("\n")
		buf.WriteString(prefix)
	}
	buf.WriteRune('}')

	ret := buf.String()
	buf.Return()
	return ret
}
