package evaluator

import (
	"silver/ast"
	"silver/object"
	"strings"
)

// evalTemplateStringLiteral captures the declaration environment but leaves
// every interpolation untouched until eval is called. A clean evaluator fork
// per invocation makes the same TemplateString safe to evaluate from tasks.
func (e *Evaluator) evalTemplateStringLiteral(node *ast.TemplateStringLiteral, env *object.Environment) object.Object {
	baseEvaluator := e.fork()
	definition, _ := object.BuiltinStructDefinitionByName("TemplateString")

	evaluate := func(args ...object.Object) object.Object {
		if len(args) != 0 {
			return newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=0", len(args))
		}

		invocationEvaluator := baseEvaluator.fork()
		var result strings.Builder
		for _, part := range node.Parts {
			if part.Expression == nil {
				result.WriteString(part.Text)
				continue
			}
			value := invocationEvaluator.Eval(part.Expression, env)
			if isError(value) {
				return value
			}
			if value == nil {
				return newError(object.RuntimeErrorKindRuntime, "template interpolation produced no value")
			}
			result.WriteString(value.Inspect())
		}
		return &object.String{Value: result.String()}
	}

	return &object.StructInstance{
		Struct: definition,
		Values: map[string]object.Object{
			"eval": &object.Builtin{Fn: evaluate, Signature: templateStringEvalSignature()},
		},
	}
}

func templateStringEvalSignature() *ast.TypeAnnotation {
	return &ast.TypeAnnotation{
		Parts:          []string{"call"},
		ParameterNames: []string{},
		ParameterTypes: []*ast.TypeAnnotation{},
		ReturnType:     &ast.TypeAnnotation{Parts: []string{"str"}},
		ErrorTypes:     []*ast.TypeAnnotation{},
	}
}
