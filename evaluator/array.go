package evaluator

import "silver/object"

// evalArrayIndexExpression returns an IndexError outside the array bounds.
func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	position, err := requireArrayIndex(arrayObject, index)
	if err != nil {
		return err
	}
	return arrayObject.Elements[position]
}

// requireArrayIndex validates the index shared by native array reads and
// assignments. Primitive arrays never use struct method dispatch.
func requireArrayIndex(array *object.Array, index object.Object) (int, *object.Error) {
	integer, ok := index.(*object.Integer)
	if !ok {
		return 0, newError(object.RuntimeErrorKindType, "array index must be INTEGER, got %s", index.Type())
	}
	if integer.Value < 0 || integer.Value >= int64(len(array.Elements)) {
		return 0, newError(object.RuntimeErrorKindIndex, "array index out of range: %d", integer.Value)
	}
	return int(integer.Value), nil
}
