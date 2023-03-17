# mapper
Utility to marshal structs while mapping fields from different locations in structs

## Usage
Add the `mapper` tag to your struct tags. The value should be a json path to the data you want the field to be unmarshaled from, or marshaled into.
## Simple example
The following example would marshal the `externalStructIDontControl` into an `internalStructIDoControl`. 
The `Derp` field will have the `Neato` value, and the `Toggle` field will have the `Question` value.
You can marshal these two structs back and forth easily.
You can then marshal using json and your struct will marshal as normal, using the `derp` and `toggle` field names, instead of the field names from the external struct.
```go
type externalStructIDontControl struct {
	Neato string `json:"neato"`
    Question bool `json:"question"`
}

type internalStructIDoControl struct {
	Derp string `json:"derp" mapper:"neato"`
	Toggle bool `json:"toggle" mapper:"question"`
}
```
## Type Coercion example
The following example would marshal the `externalStructIDontControl` into an `internalStructIDoControl`.
The `SpecificInt` field will have the `AThing` value converted to an int16 instead of an int, 
and the `StringyThingy` field will have the `AnotherThing` value converted from a string into a boolean. If the value of 
`AnotherThing` can't be converted to a boolean, you'll be returned an error, similar to json marshaling.
The `AJsonBlob` field would have the `BigOleNastyThing` struct converted to json in string form. marshaling back to 
`externalStructIDontControl` would result in a struct of type `SomeStructType`
```go
type externalStructIDontControl struct {
	AThing int `json:"a_thing"`
    AnotherThing string `json:"another_thing"`
	BigOleNastyThing SomeStructType `json:"big_ole_nasty_thing"`
}

type internalStructIDoControl struct {
	SpecificInt int16 `json:"specific_int" mapper:"a_thing,coerce"`
	StringyThingy bool `json:"stringy_thingy" mapper:"another_thing,coerce"`
	AJsonBlob string `json:"a_json_blob" mapper:"big_ole_nasty_thing,coerce"`
}
```
# Features
## Does not conflict with json marshaling
You can use json tags as normal and there will be no conflicts.
## Mapping Fields From External Structs / Data
When interfacing with data from applications outside of your control it can be difficult and brittle to keep your own objects in sync. Such as when some incoming data is deeply nested but you only need a few fields from it. marshaling from the incoming data into your own structs would require some code or intermediate structs to extract it and transform it into the shape you want. With mapper you can accomplish this with a struct tag.
## Type Coercion
Mapper can handle type coercion for you by converting the data at the given path to the field type that the struct tag is on. Add `coerce` to the struct tag to enable type coercion for that field.
## JSON Path Support
Mapper supports json path syntax so you can map a struct field to a nested field on another struct.
## Limitations
Only basic types are supported. Converting arrays to arrays and objects to objects is not supported yet. You can convert an object or an array into a json string on a string field however.
## Gotchas
Type coercion is useful and fault tolerant but might not always be what you want and can result in data loss. For example if `field_a` is a float and it's mapped to an `int` field the original float value will be converted to an int, therefore losing the floating precision.
## Benchmarks
Benchmarks were performed using generated data with 12 fields of various types

Marshaling
```text
goos: linux
goarch: amd64
pkg: github.com/catalystsquad/mapper/test
cpu: AMD Ryzen 9 5950X 16-Core Processor            
BenchmarkMapperMarshal
BenchmarkMapperMarshal-32    	  399133	      5142 ns/op
```
Unmarshaling
```text
goos: linux
goarch: amd64
pkg: github.com/catalystsquad/mapper/test
cpu: AMD Ryzen 9 5950X 16-Core Processor            
BenchmarkMapperUnmarshal
BenchmarkMapperUnmarshal-32    	   20544	     65755 ns/op
```