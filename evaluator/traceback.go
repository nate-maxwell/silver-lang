package evaluator

import (
	"silver/ast"
	"silver/object"
)

// traceFrame combines a node's lexical position with the currently executing
// Silver function. Keeping frame creation centralized ensures every runtime
// error uses the same source and function naming rules.
func (e *Evaluator) traceFrame(node ast.Node) object.TraceFrame {
	position := node.Position()
	return object.TraceFrame{
		Source:   position.Source,
		Line:     position.Line,
		Column:   position.Column,
		Function: e.currentContext(),
	}
}

// prependCallerFrame records the call/import site when an error already has an
// origin in deeper Silver code. An error created by the call itself has no
// origin yet and will be annotated once by Eval, avoiding duplicate frames.
func (e *Evaluator) prependCallerFrame(result object.Object, node ast.Node) {
	if failure, ok := result.(*object.Error); ok {
		if failure.HasTraceback() {
			failure.PrependFrame(e.traceFrame(node))
		}
	}
}

// currentContext returns the active Silver function name, or <module> while
// evaluating top-level source.
func (e *Evaluator) currentContext() string {
	if len(e.contexts) == 0 {
		return "<module>"
	}
	return e.contexts[len(e.contexts)-1]
}

// pushContext records entry into a function or imported module.
func (e *Evaluator) pushContext(name string) {
	e.contexts = append(e.contexts, name)
}

// popContext records normal or deferred exit from the active context.
func (e *Evaluator) popContext() {
	e.contexts = e.contexts[:len(e.contexts)-1]
}
