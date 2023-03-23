package pkg

import (
	"encoding/json"
	"fmt"
	tyyp "github.com/gurukami/typ/v2"
	"github.com/joomcode/errorx"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"reflect"
	"strconv"
	"strings"
)

const (
	mapperTagName = "mapper"
	jsonTagName   = "json"
	omitEmpty     = "omitempty"
	asString      = "string"
	coerce        = "coerce"
)

type tagInfo struct {
	MapperFieldPath string
	AsString        bool
	Field           reflect.StructField
	JsonFieldName   string
	OmitEmpty       bool
}

type change struct {
	Path  string
	Value interface{}
	Raw   bool
}

func Convert(source, dest interface{}) error {
	sourceBytes, err := Marshal(source)
	if err != nil {
		return err
	}
	return Unmarshal(sourceBytes, dest)
}

func Marshal(v any) ([]byte, error) {
	if isSlice(v) {
		return marshalSlice(v)
	}
	if isStruct(v) {
		return marshalStruct(v)
	}
	return nil, errorx.IllegalArgument.New("unsupported type")
}

func marshalSlice(v any) ([]byte, error) {
	marshalledString := "["
	sliceValue := reflect.ValueOf(v)
	if sliceValue.Kind() == reflect.Ptr {
		sliceValue = sliceValue.Elem()
	}
	for i := 0; i < sliceValue.Len(); i++ {
		structBytes, err := marshalStruct(sliceValue.Index(i).Interface())
		if err != nil {
			return nil, err
		}
		marshalledString += string(structBytes)
		if i+1 < sliceValue.Len() {
			marshalledString += ","
		}
	}
	marshalledString += "]"
	return []byte(marshalledString), nil
}

func marshalStruct(v any) ([]byte, error) {
	// read tags
	tagDatas, err := getTagDatas(v)
	if err != nil {
		return nil, err
	}
	//marshall to json first
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	changes := []change{}
	// process any fields that have the mapper tag, track updates in case there is collision on tags
	if len(tagDatas) > 0 {
		for _, tagData := range tagDatas {
			// get the value from the json marshalled data
			value, err := getValue(jsonBytes, tagData.JsonFieldName, tagData.AsString, tagData.Field)
			if err != nil {
				return nil, err
			}
			if tagData.OmitEmpty && isEmptyValue(value) {
				jsonBytes, err = sjson.DeleteBytes(jsonBytes, tagData.JsonFieldName)
				if err != nil {
					return nil, err
				}
				continue
			}
			_, isBytes := value.([]byte)
			changes = append(changes, change{Path: tagData.MapperFieldPath, Value: value, Raw: isBytes})
		}
		// apply updates
		for _, change := range changes {
			// set the value at the mapped path
			if change.Raw {
				jsonBytes, err = sjson.SetRawBytes(jsonBytes, change.Path, change.Value.([]byte))
				if err != nil {
					return nil, err
				}
			} else {
				jsonBytes, err = sjson.SetBytes(jsonBytes, change.Path, change.Value)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return jsonBytes, nil
}

func Unmarshal(data []byte, v interface{}) error {
	vValue := reflect.ValueOf(v)
	if vValue.IsNil() || vValue.Kind() != reflect.Ptr {
		return errorx.IllegalArgument.New("Cannot Unmarshal to nil or non pointer")
	}
	if isSlice(v) {
		return unmarshalSlice(data, v)
	}
	if isStruct(v) {
		return unmarshalStruct(data, v)
	}
	return errorx.IllegalArgument.New("unsupported type")
}

func unmarshalSlice(data []byte, v interface{}) (err error) {
	sliceObjType := reflect.TypeOf(v).Elem().Elem()
	gjson.GetBytes(data, "@this").ForEach(func(key, value gjson.Result) bool {
		var newObj interface{}
		if sliceObjType.Kind() == reflect.Ptr {
			newObj = reflect.New(sliceObjType.Elem()).Interface()
		} else {
			newObj = reflect.New(sliceObjType).Interface()
		}
		err = unmarshalStruct([]byte(value.String()), newObj)
		if err != nil {
			return false
		}
		if sliceObjType.Kind() == reflect.Struct && reflect.TypeOf(newObj).Kind() == reflect.Ptr {
			appendToSlice(v, reflect.ValueOf(newObj).Elem().Interface())
		} else {
			appendToSlice(v, reflect.ValueOf(newObj).Interface())
		}
		return true
	})
	return
}

func unmarshalStruct(data []byte, v any) error {
	// read tags
	tagDatas, err := getTagDatas(v)
	if err != nil {
		return err
	}

	// process any fields that have the mapper tag, track updates in case there is collision on tags
	if len(tagDatas) > 0 {
		changes := []change{}
		for _, tagData := range tagDatas {
			// get the value using the mapped path
			value, err := getValue(data, tagData.MapperFieldPath, tagData.AsString, tagData.Field)
			if err != nil {
				return err
			}
			if tagData.OmitEmpty && isEmptyValue(value) {
				continue
			}
			_, isBytes := value.([]byte)
			changes = append(changes, change{Path: tagData.JsonFieldName, Value: value, Raw: isBytes})
		}
		// apply updates
		for _, change := range changes {
			// set the value at the mapped path
			if change.Raw {
				data, err = sjson.SetRawBytes(data, change.Path, change.Value.([]byte))
				if err != nil {
					return err
				}
			} else {
				data, err = sjson.SetBytes(data, change.Path, change.Value)
				if err != nil {
					return err
				}
			}
		}
	}
	err = json.Unmarshal(data, v)
	return err
}

func getTagDatas(v any) ([]tagInfo, error) {
	// map the marshal fields
	destType := reflect.TypeOf(v)
	if destType.Kind() == reflect.Ptr {
		// if it's a pointer, get the non pointer type
		destType = destType.Elem()
	}

	// Iterate over all available fields and read the tag values
	tagDatas := []tagInfo{}
	if destType.Kind() == reflect.Struct {
		for i := 0; i < destType.NumField(); i++ {
			field := destType.Field(i)
			if field.Tag.Get(mapperTagName) != "" {
				tagData := getTagInfo(field)
				if tagData.MapperFieldPath != "" || tagData.OmitEmpty == true {
					tagDatas = append(tagDatas, tagData)
				}
			}
		}
	}

	return tagDatas, nil
}

func getValue(data []byte, path string, asString bool, field reflect.StructField) (interface{}, error) {
	var value interface{}
	var err error
	result := gjson.GetBytes(data, path)
	value = result.Raw
	value, err = getCoercedValue(result, field)
	if asString {
		value = fmt.Sprintf("%v", value)
	}
	return value, err
}

func getCoercedValue(result gjson.Result, field reflect.StructField) (interface{}, error) {
	typ := field.Type
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	stringValue := result.String()
	hasDecimal := strings.Contains(result.String(), ".")
	switch typ.Kind() {
	case reflect.String:
		return stringValue, nil
	case reflect.Bool:
		if stringValue == "" {
			return false, nil
		}
		return strconv.ParseBool(stringValue)
	case reflect.Int:
		float := tyyp.Of(stringValue).Float().V()
		return int(float), nil
	case reflect.Int8:
		float := tyyp.Of(stringValue).Float().V()
		return int8(float), nil
	case reflect.Int16:
		float := tyyp.Of(stringValue).Float().V()
		return int16(float), nil
	case reflect.Int32:
		float := tyyp.Of(stringValue).Float().V()
		return int32(float), nil
	case reflect.Int64:
		if hasDecimal {
			float := tyyp.Of(stringValue).Float().V()
			return tyyp.Of(float).Int64().V(), nil
		}
		return tyyp.Of(stringValue).Int64().V(), nil
	case reflect.Uint:
		float := tyyp.Of(stringValue).Float().V()
		return uint(float), nil
	case reflect.Uint8:
		float := tyyp.Of(stringValue).Float().V()
		return uint8(float), nil
	case reflect.Uint16:
		float := tyyp.Of(stringValue).Float().V()
		return uint16(float), nil
	case reflect.Uint32:
		float := tyyp.Of(stringValue).Float().V()
		return uint32(float), nil
	case reflect.Uint64:
		if hasDecimal {
			float := tyyp.Of(stringValue).Float().V()
			return tyyp.Of(float).Int64().V(), nil
		}
		return tyyp.Of(stringValue).Uint64().V(), nil
	case reflect.Float32:
		return tyyp.Of(stringValue).Float32().V(), nil
	case reflect.Float64:
		return tyyp.Of(stringValue).Float().V(), nil
	case reflect.Struct:
		return []byte(result.Raw), nil
	case reflect.Map:
		return []byte(result.Raw), nil
	case reflect.Slice:
		return []byte(result.Raw), nil
	default:
		// return the string value by default
		return stringValue, nil
	}
}

func getTagInfo(field reflect.StructField) tagInfo {
	tagData := tagInfo{
		Field: field,
	}
	mapperTagSplit := strings.Split(field.Tag.Get(mapperTagName), ",")
	for _, tagPart := range mapperTagSplit {
		if tagPart == asString {
			tagData.AsString = true
		} else if tagPart == omitEmpty {
			tagData.OmitEmpty = true
		} else {
			tagData.MapperFieldPath = tagPart
		}
	}
	jsonTagSplit := strings.Split(field.Tag.Get(jsonTagName), ",")
	if len(jsonTagSplit) > 0 {
		tagData.JsonFieldName = jsonTagSplit[0]
	} else {
		tagData.JsonFieldName = field.Name
	}
	if tagData.MapperFieldPath == "" {
		tagData.MapperFieldPath = tagData.JsonFieldName
	}
	return tagData
}

func isSlice(v any) bool {
	_, typ := getValueAndType(v)
	return typ.Kind() == reflect.Slice
}

func isStruct(v any) bool {
	_, typ := getValueAndType(v)
	return typ.Kind() == reflect.Struct
}
func getValueAndType(v any) (value reflect.Value, typ reflect.Type) {
	typ = reflect.TypeOf(v)
	value = reflect.ValueOf(v)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		value = value.Elem()
	}
	return
}

func appendToSlice(arrPtr, toAppend interface{}) {
	valuePtr := reflect.ValueOf(arrPtr)
	value := valuePtr.Elem()
	value.Set(reflect.Append(value, reflect.ValueOf(toAppend)))
}

func isEmptyValue(value interface{}) bool {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	}
	return false
}
