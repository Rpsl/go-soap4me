package structtag

import (
	"errors"
	"reflect"
)

// ErrNotAPointer error when not a pointer
var ErrNotAPointer = errors.New("Not a pointer")

// ErrNotAStruct error when not a struct
var ErrNotAStruct = errors.New("Not a struct")

// ErrTypeNotSupported error when type not supported
var ErrTypeNotSupported = errors.New("Type not supported")

// ErrUndefinedTag error when Tag var is not defined
var ErrUndefinedTag = errors.New("Undefined tag")

// Tag set the main tag
var Tag string

// TagDefault set tag default
var TagDefault string

// TagDisabled used to not process an input
var TagDisabled string

// TagSeparator separe names on environment variables
var TagSeparator string

// Prefix is a string that would be placed at the beginning of the generated tags.
var Prefix string

// ReflectFunc type used to create funcrions to parse struct and tags
type ReflectFunc func(
	field *reflect.StructField,
	value *reflect.Value,
	tag string) (err error)

// ParseMap points to each of the supported types
var ParseMap map[reflect.Kind]ReflectFunc

// Setup maps and variables
func Setup() {
	TagDisabled = "-"
	TagSeparator = "_"

	ParseMap = make(map[reflect.Kind]ReflectFunc)

	ParseMap[reflect.Struct] = ReflectStruct
}

// Reset maps caling setup function
func Reset() {
	Setup()
}

//Parse tags on struct instance
func Parse(s interface{}, superTag string) (err error) {

	if Tag == "" {
		err = ErrUndefinedTag
		return
	}

	st := reflect.TypeOf(s)

	if st.Kind() != reflect.Ptr {
		err = ErrNotAPointer
		return
	}

	refField := st.Elem()
	if refField.Kind() != reflect.Struct {
		err = ErrNotAStruct
		return
	}

	refValue := reflect.ValueOf(s).Elem()
	for i := 0; i < refField.NumField(); i++ {
		field := refField.Field(i)
		value := refValue.Field(i)
		kind := field.Type.Kind()

		if field.PkgPath != "" {
			continue
		}

		t := updateTag(&field, superTag)
		if t == "" {
			continue
		}

		if f, ok := ParseMap[kind]; ok {
			err = f(&field, &value, t)
			if err != nil {
				return
			}
		} else {
			err = ErrTypeNotSupported
			return
		}
	}
	return
}

func updateTag(field *reflect.StructField, superTag string) (ret string) {
	ret = field.Tag.Get(Tag)
	if ret == TagDisabled {
		ret = ""
		return
	}

	if ret == "" {
		ret = field.Name
	}

	if superTag != "" {
		ret = superTag + TagSeparator + ret
	} else {
		if Prefix != "" {
			ret = Prefix + TagSeparator + ret
		}
	}
	return
}

// ReflectStruct is called when the Parse encounters a sub-structure in the current structure and then calls Parsr again to treat the fields of the sub-structure.
func ReflectStruct(field *reflect.StructField, value *reflect.Value, tag string) (err error) {

	err = Parse(value.Addr().Interface(), tag)
	return
}

/*
// ReflectDebug used to debug tags
func ReflectDebug(field *reflect.StructField, value *reflect.Value, tag string) (err error) {
	fmt.Printf("name: %v, value %v, Tag %v, TagDefault %v, type %v\n",
		field.Name,
		value,
		field.Tag.Get(Tag),
		field.Tag.Get(TagDefault),
		field.Type)
	return
}
*/
