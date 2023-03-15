package test

import (
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/catalystsquad/mapper/pkg"
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

func (s *MapperSuite) Test() {
	// aNonMappedStruct represents an object that is not under our control thus we can't change the fields
	aNonMappedStruct := getRandomNonMappedStruct()

	// aMappedStruct represents an object that is under our control, so we have different field names and different json tags but we want
	// to marshall it to this using the external field/json names
	var aMappedStruct mappedStruct

	// test marshalling from aNonMappedStruct to aMappedStruct is equal
	bytes, err := pkg.Marshal(aNonMappedStruct)
	require.NoError(s.T(), err)
	err = pkg.Unmarshal(bytes, &aMappedStruct)
	require.NoError(s.T(), err)
	s.assertNonMappedStructMappedStructEquality(aNonMappedStruct, aMappedStruct)

	// test marshalling from aMappedStruct back to aNonMappedStruct is equal
	bytes, err = nil, nil
	bytes, err = pkg.Marshal(aMappedStruct)
	require.NoError(s.T(), err)
	newNonMappedStruct := &nonMappedStruct{}
	err = pkg.Unmarshal(bytes, newNonMappedStruct)
	require.NoError(s.T(), err)
	s.assertNonMappedStructMappedStructEquality(aNonMappedStruct, aMappedStruct)

	// test marshalling from mapped aMappedStruct to mapped aMappedStruct using json works
	newMappedStruct := mappedStruct{}
	bytes, err = json.Marshal(aMappedStruct)
	require.NoError(s.T(), err)
	err = json.Unmarshal(bytes, &newMappedStruct)
	require.NoError(s.T(), err)
	s.assertMappedStructEquality(aMappedStruct, newMappedStruct)
}

func (s *MapperSuite) TestTypeCoercion() {
	type baseStruct struct {
		AString  string                 `json:"a_string"`
		ABool    string                 `json:"bool_string"`
		AnInt    int                    `json:"an_int"`
		AnObject map[string]interface{} `json:"an_object"`
	}

	type coercedStruct struct {
		AnInt          int     `json:"an_int" mapper:"a_string"`
		ABool          bool    `json:"a_bool" mapper:"bool_string"`
		AFloat         float64 `json:"a_float" mapper:"an_int"`
		SomeJsonObject string  `json:"some_bytes" mapper:"an_object"`
	}

	aBaseStruct := baseStruct{
		AString: "10000",
		ABool:   "true",
		AnInt:   100,
		AnObject: map[string]interface{}{
			"one": 1,
			"two": true,
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

func getRandomMappedStruct() nonMappedStruct {
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
