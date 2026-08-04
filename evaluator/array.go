package evaluator

import "silver/object"

// evalArrayIndexExpression returns an IndexError outside the array bounds.
func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	indexValue := index.(*object.Integer).Value
	maximum := int64(len(arrayObject.Elements) - 1)

	if indexValue < 0 || indexValue > maximum {
		return newError(object.RuntimeErrorKindIndex, "array index out of range: %d", indexValue)
	}
	return arrayObject.Elements[indexValue]
}
