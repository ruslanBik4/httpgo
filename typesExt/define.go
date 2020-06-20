package typesExt

import (
	"go/types"
)

const (
    TArray types.BasicKind = -1
    TMap    types.BasicKind = -2
    TStruct  types.BasicKind = -3
)

var StringTypeKinds = map[types.BasicKind]string{
	types.Invalid: "Invalid",

	// predeclared types
	types.Bool:          "Bool",
	types.Int:           "Int",
	types.Int8:          "Int8",
	types.Int16:         "Int16",
	types.Int32:         "Int32",
	types.Int64:         "Int64",
	types.Uint:          "Uint",
	types.Uint8:         "Uint8",
	types.Uint16:        "Uint16",
	types.Uint32:        "Uint32",
	types.Uint64:        "Uint64",
	types.Uintptr:       "Uintptr",
	types.Float32:       "Float32",
	types.Float64:       "Float64",
	types.Complex64:     "Complex64",
	types.Complex128:    "Complex128",
	types.String:        "String",
	types.UnsafePointer: "UnsafePointer",

	// types for untyped values
	types.UntypedBool:    "bool",
	types.UntypedInt:     "int",
	types.UntypedRune:    "rune",
	types.UntypedFloat:   "float64",
	types.UntypedComplex: "complex128",
	types.UntypedString:  "string",
	types.UntypedNil:     "nil",
	TArray:               "array",
	TMap:				  "map",
	TStruct: 			  "struct",

}