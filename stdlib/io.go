package stdlib

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"silver/ast"
	"silver/object"
	"strings"
	"sync"
)

// ioDefinitions builds I/O functions around the evaluator's configured
// writer. Capturing the writer keeps the package independent of os.Stdout and
// lets callers redirect program output.
func ioDefinitions(out io.Writer, null *object.Null) []definition {
	return []definition{
		{name: "open", fn: builtinOpen(null), signature: openSignature()},
		{name: "print", fn: builtinPrint(out, null)},
	}
}

// nativeFile owns an open OS handle shared by the closures stored in a Silver
// File struct. Its lock serializes reads, writes, and close across tasks.
type nativeFile struct {
	mu     sync.Mutex
	file   *os.File
	closed bool
	null   *object.Null
}

// builtinOpen opens an existing file for reading and writing and returns a
// File struct whose native call fields close over the handle.
func builtinOpen(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		path, ok := args[0].(*object.String)
		if !ok {
			return newError(object.RuntimeErrorKindType, "argument to `open` must be STRING, got %s", args[0].Type())
		}

		handle, err := os.OpenFile(path.Value, os.O_RDWR, 0)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return ioErrorValue("FileNotFound", err)
			}
			return ioErrorValue("PermissionDenied", err)
		}

		state := &nativeFile{file: handle, null: null}
		definition, _ := object.BuiltinStructDefinitionByName("File")
		return &object.StructInstance{
			Struct: definition,
			Values: map[string]object.Object{
				"path":  path,
				"read":  &object.Builtin{Fn: state.read, Signature: fileReadSignature()},
				"write": &object.Builtin{Fn: state.write, Signature: fileWriteSignature()},
				"close": &object.Builtin{Fn: state.close, Signature: fileCloseSignature()},
			},
		}
	}
}

func (file *nativeFile) read(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	file.mu.Lock()
	defer file.mu.Unlock()
	if file.closed {
		return ioErrorMessage("file is closed")
	}
	if _, err := file.file.Seek(0, io.SeekStart); err != nil {
		return ioErrorValue("IOError", err)
	}
	contents, err := io.ReadAll(file.file)
	if err != nil {
		return ioErrorValue("IOError", err)
	}
	return &object.String{Value: string(contents)}
}

func (file *nativeFile) write(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	contents, ok := args[0].(*object.String)
	if !ok {
		return newError(object.RuntimeErrorKindType, "argument to `File.write` must be STRING, got %s", args[0].Type())
	}
	file.mu.Lock()
	defer file.mu.Unlock()
	if file.closed {
		return ioErrorMessage("file is closed")
	}
	if err := file.file.Truncate(0); err != nil {
		return ioErrorValue("IOError", err)
	}
	if _, err := file.file.Seek(0, io.SeekStart); err != nil {
		return ioErrorValue("IOError", err)
	}
	written, err := io.WriteString(file.file, contents.Value)
	if err != nil {
		return ioErrorValue("IOError", err)
	}
	if written != len(contents.Value) {
		return ioErrorValue("IOError", io.ErrShortWrite)
	}
	return file.null
}

func (file *nativeFile) close(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	file.mu.Lock()
	defer file.mu.Unlock()
	if file.closed {
		return ioErrorMessage("file is closed")
	}
	file.closed = true
	if err := file.file.Close(); err != nil {
		return ioErrorValue("IOError", err)
	}
	return file.null
}

func ioErrorValue(name string, err error) *object.StructInstance {
	return ioError(name, err.Error())
}

func ioErrorMessage(message string) *object.StructInstance {
	return ioError("IOError", message)
}

func ioError(name, message string) *object.StructInstance {
	definition, _ := object.BuiltinStructDefinitionByName(name)
	return &object.StructInstance{
		Struct: definition,
		Values: map[string]object.Object{"message": &object.String{Value: message}},
	}
}

func namedType(name string) *ast.TypeAnnotation {
	return &ast.TypeAnnotation{Parts: []string{name}}
}

func callSignature(parameterNames []string, parameterTypes []*ast.TypeAnnotation, returnType *ast.TypeAnnotation, errorNames ...string) *ast.TypeAnnotation {
	callErrors := make([]*ast.TypeAnnotation, len(errorNames))
	for index, name := range errorNames {
		callErrors[index] = namedType(name)
	}
	if parameterNames == nil {
		parameterNames = []string{}
	}
	if parameterTypes == nil {
		parameterTypes = []*ast.TypeAnnotation{}
	}
	return &ast.TypeAnnotation{
		Parts:          []string{"call"},
		ParameterNames: parameterNames,
		ParameterTypes: parameterTypes,
		ReturnType:     returnType,
		ErrorTypes:     callErrors,
	}
}

func openSignature() *ast.TypeAnnotation {
	return callSignature([]string{"path"}, []*ast.TypeAnnotation{namedType("str")}, namedType("File"), "FileNotFound", "PermissionDenied")
}

func fileReadSignature() *ast.TypeAnnotation {
	return callSignature(nil, nil, namedType("str"), "IOError")
}

func fileWriteSignature() *ast.TypeAnnotation {
	return callSignature([]string{"contents"}, []*ast.TypeAnnotation{namedType("str")}, nil, "IOError")
}

func fileCloseSignature() *ast.TypeAnnotation {
	return callSignature(nil, nil, nil, "IOError")
}

// builtinPrint creates a print function bound to out. Arguments are separated
// by spaces and followed by one newline. The function returns null because it
// exists for its side effect rather than for a value.
func builtinPrint(out io.Writer, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		parts := make([]string, len(args))
		for i, argument := range args {
			parts[i] = argument.Inspect()
		}
		fmt.Fprintln(out, strings.Join(parts, " "))
		return null
	}
}
