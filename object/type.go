package object

// TypeDefinition is a first-class primitive type value. Definitions are
// singletons so ordinary Silver identity equality can compare type() results
// directly with names such as int, str, and array.
type TypeDefinition struct {
	Name        string
	RuntimeType ObjectType
}

// Type returns the runtime tag shared by primitive type values.
func (t *TypeDefinition) Type() ObjectType { return TYPE_OBJ }

// Inspect returns the source-level spelling of the represented type.
func (t *TypeDefinition) Inspect() string { return t.Name }

var (
	intType    = &TypeDefinition{Name: "int", RuntimeType: INTEGER_OBJ}
	floatType  = &TypeDefinition{Name: "float", RuntimeType: FLOAT_OBJ}
	boolType   = &TypeDefinition{Name: "bool", RuntimeType: BOOLEAN_OBJ}
	stringType = &TypeDefinition{Name: "str", RuntimeType: STRING_OBJ}
	nullType   = &TypeDefinition{Name: "null", RuntimeType: NULL_OBJ}
	callType   = &TypeDefinition{Name: "call", RuntimeType: FUNCTION_OBJ}
	arrayType  = &TypeDefinition{Name: "array", RuntimeType: ARRAY_OBJ}
	mapType    = &TypeDefinition{Name: "map", RuntimeType: MAP_OBJ}
	moduleType = &TypeDefinition{Name: "module", RuntimeType: MODULE_OBJ}
	taskType   = &TypeDefinition{Name: "task", RuntimeType: TASK_OBJ}
)

var primitiveTypeDefinitions = map[ObjectType]*TypeDefinition{
	INTEGER_OBJ:  intType,
	FLOAT_OBJ:    floatType,
	BOOLEAN_OBJ:  boolType,
	STRING_OBJ:   stringType,
	NULL_OBJ:     nullType,
	FUNCTION_OBJ: callType,
	BUILTIN_OBJ:  callType,
	ARRAY_OBJ:    arrayType,
	MAP_OBJ:      mapType,
	MODULE_OBJ:   moduleType,
	TASK_OBJ:     taskType,
}

var namedTypeDefinitions = map[string]*TypeDefinition{
	"int":    intType,
	"float":  floatType,
	"bool":   boolType,
	"str":    stringType,
	"null":   nullType,
	"call":   callType,
	"array":  arrayType,
	"map":    mapType,
	"module": moduleType,
}

// TypeDefinitionByName resolves a source-level primitive type name to its
// singleton value.
func TypeDefinitionByName(name string) (*TypeDefinition, bool) {
	definition, ok := namedTypeDefinitions[name]
	return definition, ok
}

// RuntimeTypeName returns the Silver spelling for a primitive runtime tag.
func RuntimeTypeName(objectType ObjectType) (string, bool) {
	if objectType == TYPE_OBJ {
		return "type", true
	}
	definition, ok := primitiveTypeDefinitions[objectType]
	if !ok {
		return "", false
	}
	return definition.Name, true
}

// TypeOf returns the first-class type of value. Nominal definitions are
// idempotent: struct and enum definitions return themselves, and their values
// return the exact definition object from which they were created.
func TypeOf(value Object) Object {
	switch value := value.(type) {
	case *TypeDefinition:
		return value
	case *Struct:
		return value
	case *StructInstance:
		return value.Struct
	case *Enum:
		return value
	case *EnumValue:
		return value.Enum
	case *Function, *BoundMethod, *Builtin:
		return primitiveTypeDefinitions[FUNCTION_OBJ]
	}
	return primitiveTypeDefinitions[value.Type()]
}
