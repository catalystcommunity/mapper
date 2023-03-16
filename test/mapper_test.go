package test

import (
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/catalystsquad/mapper/pkg"
	"github.com/gobuffalo/nulls"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type MapperSuite struct {
	suite.Suite
}

func TestMapperSuite(t *testing.T) {
	suite.Run(t, new(MapperSuite))
}

type nonMappedStruct struct {
	ABool    bool    `json:"a_bool"`
	AString  string  `json:"a_string"`
	AnInt    int     `json:"an_int"`
	AnInt8   int8    `json:"an_int_8"`
	AnInt16  int16   `json:"an_int_16"`
	AnInt32  int32   `json:"an_int_32"`
	AnInt64  int64   `json:"an_int_64"`
	AUint    uint    `json:"a_uint"`
	AUint8   uint8   `json:"a_uint_8"`
	AUint16  uint16  `json:"a_uint_16"`
	AUint32  uint32  `json:"a_uint_32"`
	AUint64  uint64  `json:"a_uint_64"`
	AFloat32 float32 `json:"a_float_32"`
	AFloat64 float64 `json:"a_float_64"`
}

type mappedStruct struct {
	SomeOtherBool    bool    `json:"some_other_bool" mapper:"a_bool"`
	SomeOtherString  string  `json:"some_other_string" mapper:"a_string"`
	SomeOtherInt     int     `json:"some_other_int" mapper:"an_int"`
	SomeOtherInt8    int8    `json:"some_other_int_8" mapper:"an_int_8"`
	SomeOtherInt16   int16   `json:"some_other_int_16" mapper:"an_int_16"`
	SomeOtherInt32   int32   `json:"some_other_int_32" mapper:"an_int_32"`
	SomeOtherInt64   int64   `json:"some_other_int_64" mapper:"an_int_64"`
	SomeOtherUint    uint    `json:"some_other_uint" mapper:"a_uint"`
	SomeOtherUint8   uint8   `json:"some_other_uint_8" mapper:"a_uint_8"`
	SomeOtherUint16  uint16  `json:"some_other_uint_16" mapper:"a_uint_16"`
	SomeOtherUint32  uint32  `json:"some_other_uint_32" mapper:"a_uint_32"`
	SomeOtherUint64  uint64  `json:"some_other_uint_64" mapper:"a_uint_64"`
	SomeOtherFloat32 float32 `json:"some_other_float_32" mapper:"a_float_32"`
	SomeOtherFloat64 float64 `json:"some_other_float_64" mapper:"a_float_64"`
}

// struct tests
func (s *MapperSuite) TestNonMappedToMapped() {
	nonMapped := getRandomNonMappedStruct()
	mapped := mappedStruct{}
	bytes, err := pkg.Marshal(nonMapped)
	require.NoError(s.T(), err)
	err = pkg.Unmarshal(bytes, &mapped)
	require.NoError(s.T(), err)
	s.assertNonMappedStructMappedStructEquality(nonMapped, mapped)
}

func (s *MapperSuite) TestMappedToNonMapped() {
	mapped := getRandomMappedStruct()
	nonMapped := nonMappedStruct{}
	bytes, err := pkg.Marshal(mapped)
	require.NoError(s.T(), err)
	err = pkg.Unmarshal(bytes, &nonMapped)
	require.NoError(s.T(), err)
	s.assertNonMappedStructMappedStructEquality(nonMapped, mapped)
}

func (s *MapperSuite) TestMappedToMapped() {
	mapped1 := getRandomMappedStruct()
	mapped2 := getRandomMappedStruct()
	bytes, err := pkg.Marshal(mapped1)
	require.NoError(s.T(), err)
	err = pkg.Unmarshal(bytes, &mapped2)
	require.NoError(s.T(), err)
	s.assertMappedStructEquality(mapped1, mapped2)
}

func (s *MapperSuite) TestNonMappedToNonMapped() {
	nonMapped1 := getRandomNonMappedStruct()
	nonMapped2 := getRandomNonMappedStruct()
	bytes, err := pkg.Marshal(nonMapped1)
	require.NoError(s.T(), err)
	err = pkg.Unmarshal(bytes, &nonMapped2)
	require.NoError(s.T(), err)
	s.assertNonMappedStructEquality(nonMapped1, nonMapped2)
}

func (s *MapperSuite) TestPointerToNonPointer() {
	nonMapped1 := getRandomNonMappedStruct()
	nonMapped2 := getRandomNonMappedStruct()
	bytes, err := pkg.Marshal(&nonMapped1)
	require.NoError(s.T(), err)
	err = pkg.Unmarshal(bytes, &nonMapped2)
	require.NoError(s.T(), err)
	s.assertNonMappedStructEquality(nonMapped1, nonMapped2)
}

func (s *MapperSuite) TestMarshalStructField() {
	theStruct := struct {
		AnotherStruct struct {
			AString string
			AnInt   int
			ABool   bool
		}
	}{AnotherStruct: struct {
		AString string
		AnInt   int
		ABool   bool
	}{AString: "one", AnInt: 2, ABool: false}}
	bytes, err := pkg.Marshal(theStruct)
	require.NoError(s.T(), err)
	require.Equal(s.T(), string(bytes), `{"AnotherStruct":{"AString":"one","AnInt":2,"ABool":false}}`)
}

func (s *MapperSuite) TestMarshalSliceField() {
	theStruct := struct {
		ASlice []string
	}{
		ASlice: []string{"i", "like", "turtles"},
	}
	bytes, err := pkg.Marshal(theStruct)
	require.NoError(s.T(), err)
	require.Equal(s.T(), string(bytes), `{"ASlice":["i","like","turtles"]}`)
}

func (s *MapperSuite) TestPointerField() {
	type structWithPointerField struct {
		MaybeAString *string `json:"maybe_a_string"`
		MaybeAnInt   *int    `json:"maybe_an_int"`
		ABool        bool    `json:"a_bool"`
	}

	type mappedStructWithPointerField struct {
		SomeOtherStringMaybe *string `json:"some_other_string_maybe" mapper:"maybe_a_string"`
		SomeOtherIntMaybe    *int    `json:"some_other_int_maybe" mapper:"maybe_an_int"`
		SomeOtherBool        bool    `json:"some_other_bool" mapper:"a_bool"`
	}
	sourceString := gofakeit.HackerPhrase()
	sourceInt := gofakeit.Number(1, 100000)
	source := structWithPointerField{
		MaybeAString: &sourceString,
		MaybeAnInt:   &sourceInt,
		ABool:        gofakeit.Bool(),
	}
	dest := mappedStructWithPointerField{}
	err := pkg.Convert(source, &dest)
	require.NoError(s.T(), err)
	require.Equal(s.T(), source.MaybeAString, dest.SomeOtherStringMaybe)
	require.Equal(s.T(), source.MaybeAnInt, dest.SomeOtherIntMaybe)
	require.Equal(s.T(), source.ABool, dest.SomeOtherBool)
}

func (s *MapperSuite) TestPointerFieldToNonPointerField() {
	type structWithPointerField struct {
		MaybeAString *string `json:"maybe_a_string"`
		MaybeAnInt   *int    `json:"maybe_an_int"`
		ABool        bool    `json:"a_bool"`
	}

	type mappedStructWithPointerField struct {
		SomeOtherStringMaybe string `json:"some_other_string_maybe" mapper:"maybe_a_string"`
		SomeOtherIntMaybe    int    `json:"some_other_int_maybe" mapper:"maybe_an_int"`
		SomeOtherBool        bool   `json:"some_other_bool" mapper:"a_bool"`
	}
	sourceString := gofakeit.HackerPhrase()
	sourceInt := gofakeit.Number(1, 100000)
	source := structWithPointerField{
		MaybeAString: &sourceString,
		MaybeAnInt:   &sourceInt,
		ABool:        gofakeit.Bool(),
	}
	dest := mappedStructWithPointerField{}
	err := pkg.Convert(source, &dest)
	require.NoError(s.T(), err)
	require.Equal(s.T(), *source.MaybeAString, dest.SomeOtherStringMaybe)
	require.Equal(s.T(), *source.MaybeAnInt, dest.SomeOtherIntMaybe)
	require.Equal(s.T(), source.ABool, dest.SomeOtherBool)
}

func (s *MapperSuite) TestNonPointerFieldToPointerField() {
	type structWithPointerField struct {
		MaybeAString string `json:"maybe_a_string"`
		MaybeAnInt   int    `json:"maybe_an_int"`
		ABool        bool   `json:"a_bool"`
	}

	type mappedStructWithPointerField struct {
		SomeOtherStringMaybe *string `json:"some_other_string_maybe" mapper:"maybe_a_string"`
		SomeOtherIntMaybe    *int    `json:"some_other_int_maybe" mapper:"maybe_an_int"`
		SomeOtherBool        bool    `json:"some_other_bool" mapper:"a_bool"`
	}
	sourceString := gofakeit.HackerPhrase()
	sourceInt := gofakeit.Number(1, 100000)
	source := structWithPointerField{
		MaybeAString: sourceString,
		MaybeAnInt:   sourceInt,
		ABool:        gofakeit.Bool(),
	}
	dest := mappedStructWithPointerField{}
	err := pkg.Convert(source, &dest)
	require.NoError(s.T(), err)
	require.Equal(s.T(), source.MaybeAString, *dest.SomeOtherStringMaybe)
	require.Equal(s.T(), source.MaybeAnInt, *dest.SomeOtherIntMaybe)
	require.Equal(s.T(), source.ABool, dest.SomeOtherBool)
}

// slice tests
func (s *MapperSuite) TestNonMappedPointerSliceToMappedPointerSlice() {
	num := gofakeit.Number(1, 5)
	nonMappedSlice := getRandomNonMappedStructPointers(num)
	mappedSlice := []*mappedStruct{}
	bytes, err := pkg.Marshal(nonMappedSlice)
	require.NoError(s.T(), err)
	err = pkg.Unmarshal(bytes, &mappedSlice)
	require.NoError(s.T(), err)
	s.assertnonMappedPointerSliceMappedPointerSliceEquality(nonMappedSlice, mappedSlice)
}

func (s *MapperSuite) TestNonMappedPointerSliceToMappedSlice() {
	num := gofakeit.Number(1, 5)
	nonMappedSlice := getRandomNonMappedStructPointers(num)
	mappedSlice := []mappedStruct{}
	bytes, err := pkg.Marshal(nonMappedSlice)
	require.NoError(s.T(), err)
	err = pkg.Unmarshal(bytes, &mappedSlice)
	require.NoError(s.T(), err)
	s.assertnonMappedPointerSliceMappedSliceEquality(nonMappedSlice, mappedSlice)
}

func (s *MapperSuite) TestNonMappedSliceToMappedSlice() {
	num := gofakeit.Number(1, 5)
	nonMappedSlice := getRandomNonMappedStructs(num)
	mappedSlice := []mappedStruct{}
	bytes, err := pkg.Marshal(nonMappedSlice)
	require.NoError(s.T(), err)
	err = pkg.Unmarshal(bytes, &mappedSlice)
	require.NoError(s.T(), err)
	s.assertnonMappedSliceMappedSliceEquality(nonMappedSlice, mappedSlice)
}

func (s *MapperSuite) TestNonMappedSliceToMappedPointerSlice() {
	num := gofakeit.Number(1, 5)
	nonMappedSlice := getRandomNonMappedStructs(num)
	mappedSlice := []*mappedStruct{}
	bytes, err := pkg.Marshal(nonMappedSlice)
	require.NoError(s.T(), err)
	err = pkg.Unmarshal(bytes, &mappedSlice)
	require.NoError(s.T(), err)
	s.assertnonMappedSliceMappedPointerSliceEquality(nonMappedSlice, mappedSlice)
}

func (s *MapperSuite) TestNonMappedSliceToNonMappedSlice() {
	num := gofakeit.Number(1, 5)
	nonMappedSlice1 := getRandomNonMappedStructs(num)
	nonMappedSlice2 := []nonMappedStruct{}
	bytes, err := pkg.Marshal(nonMappedSlice1)
	require.NoError(s.T(), err)
	err = pkg.Unmarshal(bytes, &nonMappedSlice2)
	require.NoError(s.T(), err)
	s.assertnonMappedSliceNonMappedSliceEquality(nonMappedSlice1, nonMappedSlice2)
}

func (s *MapperSuite) TestMappedSliceToMappedSlice() {
	num := gofakeit.Number(1, 5)
	mappedSlice1 := getRandomMappedStructs(num)
	mappedSlice2 := []mappedStruct{}
	bytes, err := pkg.Marshal(mappedSlice1)
	require.NoError(s.T(), err)
	err = pkg.Unmarshal(bytes, &mappedSlice2)
	require.NoError(s.T(), err)
	s.assertMappedSliceEquality(mappedSlice1, mappedSlice2)
}

func (s *MapperSuite) TestConvert() {
	num := gofakeit.Number(1, 5)
	nonMappedSlice := getRandomNonMappedStructPointers(num)
	mappedSlice := []*mappedStruct{}
	err := pkg.Convert(nonMappedSlice, &mappedSlice)
	require.NoError(s.T(), err)
	s.assertnonMappedPointerSliceMappedPointerSliceEquality(nonMappedSlice, mappedSlice)
}

// misc tests
func (s *MapperSuite) TestInvalidArguments() {
	err := pkg.Unmarshal([]byte{}, []string{})
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "Cannot Unmarshal to nil or non pointer")
	var strings *[]string
	err = pkg.Unmarshal([]byte{}, strings)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "Cannot Unmarshal to nil or non pointer")
}

func (s *MapperSuite) TestTypeCoercion() {
	type nestedStruct struct {
		AString string
		AnInt   int
	}

	type baseStruct struct {
		AString      string                 `json:"a_string"`
		ABool        string                 `json:"bool_string"`
		AnInt        int                    `json:"an_int"`
		AnObject     map[string]interface{} `json:"an_object"`
		NestedStruct nestedStruct           `json:"nested_struct"`
	}

	type coercedStruct struct {
		AnInt            int     `json:"an_int" mapper:"a_string"`
		ABool            bool    `json:"a_bool" mapper:"bool_string"`
		AFloat           float64 `json:"a_float" mapper:"an_int"`
		SomeJsonObject   string  `json:"some_bytes" mapper:"an_object"`
		SomeNestedStruct string  `json:"some_struct" mapper:"nested_struct"`
	}
	aBaseStruct := baseStruct{
		AString: "10000",
		ABool:   "true",
		AnInt:   100,
		AnObject: map[string]interface{}{
			"one": 1,
			"two": true,
		},
		NestedStruct: nestedStruct{
			AString: "one",
			AnInt:   1,
		},
	}
	aCoerecedStruct := coercedStruct{}

	bytes, err := pkg.Marshal(aBaseStruct)
	require.NoError(s.T(), err)
	err = pkg.Unmarshal(bytes, &aCoerecedStruct)
	require.NoError(s.T(), err)
	require.Equal(s.T(), aCoerecedStruct.AnInt, 10000)
	require.True(s.T(), aCoerecedStruct.ABool)
	require.Equal(s.T(), 100.0, aCoerecedStruct.AFloat)
	objBytes, err := json.Marshal(aBaseStruct.AnObject)
	require.NoError(s.T(), err)
	require.Equal(s.T(), string(objBytes), aCoerecedStruct.SomeJsonObject)
	structBytes, err := json.Marshal(aBaseStruct.NestedStruct)
	require.NoError(s.T(), err)
	require.Equal(s.T(), string(structBytes), aCoerecedStruct.SomeNestedStruct)
}

func (s *MapperSuite) TestOmitEmptyStruct() {
	type dest struct {
		Id nulls.UUID `mapper:"omitempty"`
	}
	theString := gofakeit.HackerPhrase()
	source := struct {
		AString string
		Id      string `json:",omitempty"`
	}{
		AString: theString,
		Id:      "",
	}
	bytes, err := pkg.Marshal(source)
	require.NoError(s.T(), err)
	require.Equal(s.T(), fmt.Sprintf(`{"AString":"%s"}`, theString), string(bytes))
	theDest := &dest{}
	err = pkg.Unmarshal(bytes, theDest)
	require.NoError(s.T(), err)
}

func (s *MapperSuite) TestOmitEmptySlice() {
	type source struct {
		Id *string `json:"id,omitempty"`
	}
	type dest struct {
		Id nulls.UUID `json:"id,omitempty" mapper:"omitempty"`
	}
	theSource := source{
		Id: nil,
	}
	theDest := dest{}
	err := pkg.Convert(theSource, &theDest)
	require.NoError(s.T(), err)
}

func getRandomNonMappedStructPointers(num int) []*nonMappedStruct {
	structs := []*nonMappedStruct{}
	for i := 0; i < num; i++ {
		theStruct := getRandomNonMappedStruct()
		structs = append(structs, &theStruct)
	}
	return structs
}

func getRandomNonMappedStructs(num int) []nonMappedStruct {
	structs := []nonMappedStruct{}
	for i := 0; i < num; i++ {
		structs = append(structs, getRandomNonMappedStruct())
	}
	return structs
}

func getRandomNonMappedStruct() nonMappedStruct {
	return nonMappedStruct{
		ABool:    gofakeit.Bool(),
		AString:  gofakeit.HackerPhrase(),
		AnInt8:   gofakeit.Int8(),
		AnInt16:  gofakeit.Int16(),
		AnInt32:  gofakeit.Int32(),
		AnInt64:  gofakeit.Int64(),
		AUint8:   gofakeit.Uint8(),
		AUint16:  gofakeit.Uint16(),
		AUint32:  gofakeit.Uint32(),
		AUint64:  gofakeit.Uint64(),
		AFloat32: gofakeit.Float32(),
		AFloat64: gofakeit.Float64(),
	}
}

func getRandomMappedStructs(num int) []mappedStruct {
	structs := []mappedStruct{}
	for i := 0; i < num; i++ {
		structs = append(structs, getRandomMappedStruct())
	}
	return structs
}

func getRandomMappedStruct() mappedStruct {
	return mappedStruct{
		SomeOtherBool:    gofakeit.Bool(),
		SomeOtherString:  gofakeit.HackerPhrase(),
		SomeOtherInt8:    gofakeit.Int8(),
		SomeOtherInt16:   gofakeit.Int16(),
		SomeOtherInt32:   gofakeit.Int32(),
		SomeOtherInt64:   gofakeit.Int64(),
		SomeOtherUint8:   gofakeit.Uint8(),
		SomeOtherUint16:  gofakeit.Uint16(),
		SomeOtherUint32:  gofakeit.Uint32(),
		SomeOtherUint64:  gofakeit.Uint64(),
		SomeOtherFloat32: gofakeit.Float32(),
		SomeOtherFloat64: gofakeit.Float64(),
	}
}

func (s *MapperSuite) assertnonMappedSliceMappedSliceEquality(nonMappedSlice []nonMappedStruct, mappedSlice []mappedStruct) {
	require.Equal(s.T(), len(nonMappedSlice), len(mappedSlice))
	for i, nonMapped := range nonMappedSlice {
		s.assertNonMappedStructMappedStructEquality(nonMapped, mappedSlice[i])
	}
}

func (s *MapperSuite) assertnonMappedPointerSliceMappedPointerSliceEquality(nonMappedSlice []*nonMappedStruct, mappedSlice []*mappedStruct) {
	require.Equal(s.T(), len(nonMappedSlice), len(mappedSlice))
	for i, nonMapped := range nonMappedSlice {
		s.assertNonMappedStructMappedStructEquality(*nonMapped, *mappedSlice[i])
	}
}

func (s *MapperSuite) assertnonMappedPointerSliceMappedSliceEquality(nonMappedSlice []*nonMappedStruct, mappedSlice []mappedStruct) {
	require.Equal(s.T(), len(nonMappedSlice), len(mappedSlice))
	for i, nonMapped := range nonMappedSlice {
		s.assertNonMappedStructMappedStructEquality(*nonMapped, mappedSlice[i])
	}
}

func (s *MapperSuite) assertnonMappedSliceMappedPointerSliceEquality(nonMappedSlice []nonMappedStruct, mappedSlice []*mappedStruct) {
	require.Equal(s.T(), len(nonMappedSlice), len(mappedSlice))
	for i, nonMapped := range nonMappedSlice {
		s.assertNonMappedStructMappedStructEquality(nonMapped, *mappedSlice[i])
	}
}

func (s *MapperSuite) assertnonMappedSliceNonMappedSliceEquality(nonMappedSlice1, nonMappedSlice2 []nonMappedStruct) {
	require.Equal(s.T(), len(nonMappedSlice1), len(nonMappedSlice2))
	for i, nonMapped := range nonMappedSlice1 {
		s.assertNonMappedStructEquality(nonMapped, nonMappedSlice2[i])
	}
}

func (s *MapperSuite) assertMappedSliceEquality(mappedSlice1, mappedSlice2 []mappedStruct) {
	require.Equal(s.T(), len(mappedSlice1), len(mappedSlice2))
	for i, mapped := range mappedSlice1 {
		s.assertMappedStructEquality(mapped, mappedSlice2[i])
	}
}

func (s *MapperSuite) assertNonMappedStructMappedStructEquality(aNonMappedStruct nonMappedStruct, aMappedStruct mappedStruct) {
	require.Equal(s.T(), aMappedStruct.SomeOtherBool, aNonMappedStruct.ABool)
	require.Equal(s.T(), aMappedStruct.SomeOtherString, aNonMappedStruct.AString)
	require.Equal(s.T(), aMappedStruct.SomeOtherInt8, aNonMappedStruct.AnInt8)
	require.Equal(s.T(), aMappedStruct.SomeOtherInt16, aNonMappedStruct.AnInt16)
	require.Equal(s.T(), aMappedStruct.SomeOtherInt32, aNonMappedStruct.AnInt32)
	require.Equal(s.T(), aMappedStruct.SomeOtherInt64, aNonMappedStruct.AnInt64)
	require.Equal(s.T(), aMappedStruct.SomeOtherUint8, aNonMappedStruct.AUint8)
	require.Equal(s.T(), aMappedStruct.SomeOtherUint16, aNonMappedStruct.AUint16)
	require.Equal(s.T(), aMappedStruct.SomeOtherUint32, aNonMappedStruct.AUint32)
	require.Equal(s.T(), aMappedStruct.SomeOtherUint64, aNonMappedStruct.AUint64)
	require.Equal(s.T(), aMappedStruct.SomeOtherFloat32, aNonMappedStruct.AFloat32)
	require.Equal(s.T(), aMappedStruct.SomeOtherFloat64, aNonMappedStruct.AFloat64)
}

func (s *MapperSuite) assertMappedStructEquality(expected, actual mappedStruct) {
	require.Equal(s.T(), expected.SomeOtherBool, actual.SomeOtherBool)
	require.Equal(s.T(), expected.SomeOtherString, actual.SomeOtherString)
	require.Equal(s.T(), expected.SomeOtherInt8, actual.SomeOtherInt8)
	require.Equal(s.T(), expected.SomeOtherInt16, actual.SomeOtherInt16)
	require.Equal(s.T(), expected.SomeOtherInt32, actual.SomeOtherInt32)
	require.Equal(s.T(), expected.SomeOtherInt64, actual.SomeOtherInt64)
	require.Equal(s.T(), expected.SomeOtherUint8, actual.SomeOtherUint8)
	require.Equal(s.T(), expected.SomeOtherUint16, actual.SomeOtherUint16)
	require.Equal(s.T(), expected.SomeOtherUint32, actual.SomeOtherUint32)
	require.Equal(s.T(), expected.SomeOtherUint64, actual.SomeOtherUint64)
	require.Equal(s.T(), expected.SomeOtherFloat32, actual.SomeOtherFloat32)
	require.Equal(s.T(), expected.SomeOtherFloat64, actual.SomeOtherFloat64)
}

func (s *MapperSuite) assertNonMappedStructEquality(expected, actual nonMappedStruct) {
	require.Equal(s.T(), expected.ABool, actual.ABool)
	require.Equal(s.T(), expected.AString, actual.AString)
	require.Equal(s.T(), expected.AnInt8, actual.AnInt8)
	require.Equal(s.T(), expected.AnInt16, actual.AnInt16)
	require.Equal(s.T(), expected.AnInt32, actual.AnInt32)
	require.Equal(s.T(), expected.AnInt64, actual.AnInt64)
	require.Equal(s.T(), expected.AUint8, actual.AUint8)
	require.Equal(s.T(), expected.AUint16, actual.AUint16)
	require.Equal(s.T(), expected.AUint32, actual.AUint32)
	require.Equal(s.T(), expected.AUint64, actual.AUint64)
	require.Equal(s.T(), expected.AFloat32, actual.AFloat32)
	require.Equal(s.T(), expected.AFloat64, actual.AFloat64)
}

func BenchmarkMapperMarshal(b *testing.B) {
	mapped := getRandomMappedStruct()

	for n := 0; n < b.N; n++ {
		pkg.Marshal(mapped)
	}
}

func BenchmarkMapperUnmarshal(b *testing.B) {
	var mapped mappedStruct
	nonMapped := getRandomNonMappedStruct()
	bytes, err := pkg.Marshal(nonMapped)
	if err != nil {
		panic(err)
	}

	for n := 0; n < b.N; n++ {
		err = pkg.Unmarshal(bytes, &mapped)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkMarshalUnmarshal(b *testing.B) {
	var mapped mappedStruct
	nonMapped := getRandomNonMappedStruct()

	for n := 0; n < b.N; n++ {
		bytes, err := pkg.Marshal(nonMapped)
		if err != nil {
			panic(err)
		}
		err = pkg.Unmarshal(bytes, &mapped)
		if err != nil {
			panic(err)
		}
	}
}
