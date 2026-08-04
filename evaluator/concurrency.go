package evaluator

import (
	"fmt"
	"silver/ast"
	"silver/object"
)

// evalTaskExpression launches work immediately and returns without observing
// its result. Runtime failures are retained on the handle until collection.
func (e *Evaluator) evalTaskExpression(node *ast.TaskExpression, env *object.Environment) object.Object {
	handle := object.NewTask()
	env.RegisterTask(handle)
	taskEvaluator := e.fork()
	started := make(chan struct{})

	go func() {
		// Handshake before evaluation makes task start order match declaration
		// order without waiting for any task to complete.
		close(started)
		callable := taskEvaluator.Eval(node.Work, env)
		var result object.Object = callable
		if !isError(callable) {
			result = taskEvaluator.applyFunction(callable, nil)
			taskEvaluator.prependCallerFrame(result, node)
			if failure, ok := result.(*object.Error); ok {
				failure.SetOrigin(taskEvaluator.traceFrame(node))
			}
		}
		result = unwrapReturnValue(result)
		if result == nil {
			result = NULL
		}
		handle.Complete(result)
	}()
	<-started

	return handle
}

// evalCollectExpression consumes every handle before waiting, then waits for
// all work even if one task produced a runtime error. Null results are omitted
// from the anonymous result struct.
func (e *Evaluator) evalCollectExpression(node *ast.CollectExpression, env *object.Environment) object.Object {
	handles := make([]*object.Task, len(node.Handles))
	for index, identifier := range node.Handles {
		value, ok := env.Get(identifier.Value)
		if !ok {
			return newError(object.RuntimeErrorKindName, "identifier not found: %s", identifier.Value)
		}
		handle, ok := value.(*object.Task)
		if !ok {
			return newError(object.RuntimeErrorKindType, "%q is not a task handle", identifier.Value)
		}
		handles[index] = handle
	}

	for index, handle := range handles {
		if !handle.MarkCollected() {
			for prior := 0; prior < index; prior++ {
				handles[prior].Await()
			}
			return newError(object.RuntimeErrorKindTask, "task handle %q may only be collected once", node.Handles[index].Value)
		}
	}

	results := make([]object.Object, len(handles))
	var firstFailure object.Object
	for index, handle := range handles {
		results[index] = handle.Await()
		if firstFailure == nil && isError(results[index]) {
			firstFailure = results[index]
		}
	}
	if firstFailure != nil {
		return firstFailure
	}

	fields := make([]string, 0, len(results))
	fieldTypes := make([]*ast.TypeAnnotation, 0, len(results))
	values := make(map[string]object.Object, len(results))
	for index, result := range results {
		if result == NULL {
			continue
		}
		name := node.Handles[index].Value
		fields = append(fields, name)
		fieldTypes = append(fieldTypes, nil)
		values[name] = result
	}
	definition := &object.Struct{
		Name:       "collect",
		Fields:     fields,
		FieldTypes: fieldTypes,
		Env:        env,
	}
	return &object.StructInstance{Struct: definition, Values: values}
}

// finishTasks warns about and joins uncollected work. Joining ensures a scope
// cannot silently abandon a goroutine when it exits.
func (e *Evaluator) finishTasks(env *object.Environment) {
	for _, task := range env.Tasks() {
		if task.Collected() {
			continue
		}
		name := task.Name()
		if name == "" {
			name = "<unnamed>"
		}
		fmt.Fprintf(e.warnings, "WARNING: task handle %q was never collected; waiting for completion\n", name)
		task.MarkCollected()
		task.Await()
	}
}
