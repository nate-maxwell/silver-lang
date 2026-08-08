package stdlib

import (
	"io/fs"
	"net/url"
	"os"
	pathpkg "path"
	"path/filepath"
	"silver/ast"
	"silver/object"
	"strings"
	"time"
)

// pathDefinitions exports one module-local Path type and factories for it.
// Path methods are native closures stored in ordinary Silver struct fields;
// this requires no global builtin type or evaluator-level method machinery.
func pathDefinitions(null *object.Null, trueValue, falseValue *object.Boolean) []definition {
	pathType := newPathStructDefinition()
	return []definition{
		{name: "Path", value: pathType},
		{name: "new", fn: pathConstructor(pathType, null, trueValue, falseValue), signature: callSignature([]string{"path"}, []*ast.TypeAnnotation{namedType("str")}, namedType("Path"))},
		{name: "cwd", fn: pathFactory(pathType, null, trueValue, falseValue, "cwd", os.Getwd)},
		{name: "home", fn: pathFactory(pathType, null, trueValue, falseValue, "home", os.UserHomeDir)},
	}
}

type pathResultKind uint8

const (
	pathRawResult pathResultKind = iota
	pathValueResult
	pathArrayResult
	pathWalkResult
)

func newPathStructDefinition() *object.Struct {
	environment := object.NewEnvironment()
	pathType := &object.Struct{Name: "Path", Env: environment}
	environment.Set("Path", pathType)

	pathAnnotation := namedType("Path")
	strAnnotation := namedType("str")
	boolAnnotation := namedType("bool")
	intAnnotation := namedType("int")
	arrayAnnotation := namedType("array")
	mapAnnotation := namedType("map")
	fileAnnotation := namedType("File")

	type field struct {
		name       string
		annotation *ast.TypeAnnotation
	}
	fields := []field{
		{name: "path", annotation: strAnnotation},
		{name: "anchor", annotation: strAnnotation},
		{name: "drive", annotation: strAnnotation},
		{name: "name", annotation: strAnnotation},
		{name: "parts", annotation: arrayAnnotation},
		{name: "root", annotation: strAnnotation},
		{name: "stem", annotation: strAnnotation},
		{name: "suffix", annotation: strAnnotation},
		{name: "suffixes", annotation: arrayAnnotation},
		{name: "parent", annotation: callSignature(nil, nil, pathAnnotation)},
		{name: "parents", annotation: callSignature(nil, nil, arrayAnnotation)},
		{name: "absolute", annotation: callSignature(nil, nil, pathAnnotation)},
		{name: "as_posix", annotation: callSignature(nil, nil, strAnnotation)},
		{name: "as_uri", annotation: callSignature(nil, nil, strAnnotation)},
		{name: "chmod", annotation: callSignature([]string{"mode"}, []*ast.TypeAnnotation{intAnnotation}, nil)},
		{name: "exists", annotation: callSignature(nil, nil, boolAnnotation)},
		{name: "expanduser", annotation: callSignature(nil, nil, pathAnnotation)},
		{name: "glob", annotation: callSignature([]string{"pattern"}, []*ast.TypeAnnotation{strAnnotation}, arrayAnnotation)},
		{name: "hardlink_to", annotation: callSignature([]string{"target"}, []*ast.TypeAnnotation{pathAnnotation}, nil)},
		{name: "is_absolute", annotation: callSignature(nil, nil, boolAnnotation)},
		{name: "is_block_device", annotation: callSignature(nil, nil, boolAnnotation)},
		{name: "is_char_device", annotation: callSignature(nil, nil, boolAnnotation)},
		{name: "is_dir", annotation: callSignature(nil, nil, boolAnnotation)},
		{name: "is_fifo", annotation: callSignature(nil, nil, boolAnnotation)},
		{name: "is_file", annotation: callSignature(nil, nil, boolAnnotation)},
		{name: "is_mount", annotation: callSignature(nil, nil, boolAnnotation)},
		{name: "is_relative_to", annotation: callSignature([]string{"other"}, []*ast.TypeAnnotation{pathAnnotation}, boolAnnotation)},
		{name: "is_socket", annotation: callSignature(nil, nil, boolAnnotation)},
		{name: "is_symlink", annotation: callSignature(nil, nil, boolAnnotation)},
		{name: "iterdir", annotation: callSignature(nil, nil, arrayAnnotation)},
		{name: "joinpath", annotation: callSignature([]string{"part"}, []*ast.TypeAnnotation{strAnnotation}, pathAnnotation)},
		{name: "lstat", annotation: callSignature(nil, nil, mapAnnotation)},
		{name: "match", annotation: callSignature([]string{"pattern"}, []*ast.TypeAnnotation{strAnnotation}, boolAnnotation)},
		{name: "mkdir", annotation: callSignature(nil, nil, nil)},
		{name: "open", annotation: callSignature(nil, nil, fileAnnotation, "FileNotFound", "PermissionDenied")},
		{name: "read_bytes", annotation: callSignature(nil, nil, strAnnotation)},
		{name: "read_text", annotation: callSignature(nil, nil, strAnnotation)},
		{name: "readlink", annotation: callSignature(nil, nil, pathAnnotation)},
		{name: "relative_to", annotation: callSignature([]string{"other"}, []*ast.TypeAnnotation{pathAnnotation}, pathAnnotation)},
		{name: "rename", annotation: callSignature([]string{"target"}, []*ast.TypeAnnotation{pathAnnotation}, pathAnnotation)},
		{name: "replace", annotation: callSignature([]string{"target"}, []*ast.TypeAnnotation{pathAnnotation}, pathAnnotation)},
		{name: "resolve", annotation: callSignature(nil, nil, pathAnnotation)},
		{name: "rglob", annotation: callSignature([]string{"pattern"}, []*ast.TypeAnnotation{strAnnotation}, arrayAnnotation)},
		{name: "rmdir", annotation: callSignature(nil, nil, nil)},
		{name: "samefile", annotation: callSignature([]string{"other"}, []*ast.TypeAnnotation{pathAnnotation}, boolAnnotation)},
		{name: "stat", annotation: callSignature(nil, nil, mapAnnotation)},
		{name: "symlink_to", annotation: callSignature([]string{"target"}, []*ast.TypeAnnotation{pathAnnotation}, nil)},
		{name: "touch", annotation: callSignature(nil, nil, nil)},
		{name: "unlink", annotation: callSignature(nil, nil, nil)},
		{name: "walk", annotation: callSignature(nil, nil, arrayAnnotation)},
		{name: "with_name", annotation: callSignature([]string{"name"}, []*ast.TypeAnnotation{strAnnotation}, pathAnnotation)},
		{name: "with_stem", annotation: callSignature([]string{"stem"}, []*ast.TypeAnnotation{strAnnotation}, pathAnnotation)},
		{name: "with_suffix", annotation: callSignature([]string{"suffix"}, []*ast.TypeAnnotation{strAnnotation}, pathAnnotation)},
		{name: "write_bytes", annotation: callSignature([]string{"contents"}, []*ast.TypeAnnotation{strAnnotation}, intAnnotation)},
		{name: "write_text", annotation: callSignature([]string{"contents"}, []*ast.TypeAnnotation{strAnnotation}, intAnnotation)},
	}

	pathType.Fields = make([]string, len(fields))
	pathType.FieldTypes = make([]*ast.TypeAnnotation, len(fields))
	for index, field := range fields {
		pathType.Fields[index] = field.name
		pathType.FieldTypes[index] = field.annotation
	}
	return pathType
}

func pathConstructor(pathType *object.Struct, null *object.Null, trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		value, err := requireString("new", 0, args[0])
		if err != nil {
			return err
		}
		return newPathValue(pathType, null, trueValue, falseValue, value)
	}
}

func pathFactory(pathType *object.Struct, null *object.Null, trueValue, falseValue *object.Boolean, name string, operation func() (string, error)) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 0); err != nil {
			return err
		}
		value, err := operation()
		if err != nil {
			return pathOperationError(name, "", err)
		}
		return newPathValue(pathType, null, trueValue, falseValue, value)
	}
}

func newPathValue(pathType *object.Struct, null *object.Null, trueValue, falseValue *object.Boolean, value string) *object.StructInstance {
	value = filepath.Clean(value)
	values := map[string]object.Object{
		"path":     &object.String{Value: value},
		"anchor":   invokePathProperty(pathAnchor, value),
		"drive":    &object.String{Value: filepath.VolumeName(value)},
		"name":     &object.String{Value: filepath.Base(value)},
		"parts":    invokePathProperty(pathParts, value),
		"root":     invokePathProperty(pathRoot, value),
		"stem":     &object.String{Value: pathStem(value)},
		"suffix":   &object.String{Value: pathSuffix(value)},
		"suffixes": invokePathProperty(pathSuffixes, value),
	}

	newMethod := func(name string, operation object.BuiltinFunction, resultKind pathResultKind) *object.Builtin {
		return bindPathMethod(pathType, null, trueValue, falseValue, value, name, operation, resultKind)
	}
	values["parent"] = &object.Builtin{Fn: func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 0); err != nil {
			return err
		}
		return newPathValue(pathType, null, trueValue, falseValue, filepath.Dir(value))
	}, Signature: pathFieldSignature(pathType, "parent")}
	values["parents"] = newMethod("parents", pathParents, pathArrayResult)
	values["absolute"] = newMethod("absolute", pathAbsolute, pathValueResult)
	values["as_posix"] = newMethod("as_posix", unaryPathString("as_posix", filepath.ToSlash), pathRawResult)
	values["as_uri"] = newMethod("as_uri", pathAsURI, pathRawResult)
	values["chmod"] = newMethod("chmod", pathChmod(null), pathRawResult)
	values["exists"] = newMethod("exists", pathExists(trueValue, falseValue), pathRawResult)
	values["expanduser"] = newMethod("expanduser", pathExpandUser, pathValueResult)
	values["glob"] = newMethod("glob", pathGlob, pathArrayResult)
	values["hardlink_to"] = newMethod("hardlink_to", pathHardlinkTo(null), pathRawResult)
	values["is_absolute"] = newMethod("is_absolute", unaryPathPredicate("is_absolute", filepath.IsAbs, trueValue, falseValue), pathRawResult)
	values["is_block_device"] = newMethod("is_block_device", pathInfoPredicate("is_block_device", func(info fs.FileInfo) bool {
		return info.Mode()&os.ModeDevice != 0 && info.Mode()&os.ModeCharDevice == 0
	}, trueValue, falseValue), pathRawResult)
	values["is_char_device"] = newMethod("is_char_device", pathModePredicate("is_char_device", os.ModeDevice|os.ModeCharDevice, trueValue, falseValue), pathRawResult)
	values["is_dir"] = newMethod("is_dir", pathInfoPredicate("is_dir", func(info fs.FileInfo) bool { return info.IsDir() }, trueValue, falseValue), pathRawResult)
	values["is_fifo"] = newMethod("is_fifo", pathModePredicate("is_fifo", os.ModeNamedPipe, trueValue, falseValue), pathRawResult)
	values["is_file"] = newMethod("is_file", pathInfoPredicate("is_file", func(info fs.FileInfo) bool { return info.Mode().IsRegular() }, trueValue, falseValue), pathRawResult)
	values["is_mount"] = newMethod("is_mount", pathIsMount(trueValue, falseValue), pathRawResult)
	values["is_relative_to"] = newMethod("is_relative_to", pathIsRelativeTo(trueValue, falseValue), pathRawResult)
	values["is_socket"] = newMethod("is_socket", pathModePredicate("is_socket", os.ModeSocket, trueValue, falseValue), pathRawResult)
	values["is_symlink"] = newMethod("is_symlink", pathIsSymlink(trueValue, falseValue), pathRawResult)
	values["iterdir"] = newMethod("iterdir", pathIterDir, pathArrayResult)
	values["joinpath"] = newMethod("joinpath", pathJoin, pathValueResult)
	values["lstat"] = newMethod("lstat", pathLstat, pathRawResult)
	values["match"] = newMethod("match", pathMatch(trueValue, falseValue), pathRawResult)
	values["mkdir"] = newMethod("mkdir", pathMkdir(null), pathRawResult)
	values["open"] = newMethod("open", pathOpen(null), pathRawResult)
	values["read_bytes"] = newMethod("read_bytes", pathRead("read_bytes"), pathRawResult)
	values["read_text"] = newMethod("read_text", pathRead("read_text"), pathRawResult)
	values["readlink"] = newMethod("readlink", pathReadlink, pathValueResult)
	values["relative_to"] = newMethod("relative_to", pathRelativeTo, pathValueResult)
	values["rename"] = newMethod("rename", pathRename("rename"), pathValueResult)
	values["replace"] = newMethod("replace", pathRename("replace"), pathValueResult)
	values["resolve"] = newMethod("resolve", pathResolve, pathValueResult)
	values["rglob"] = newMethod("rglob", pathRGlob, pathArrayResult)
	values["rmdir"] = newMethod("rmdir", pathRmdir(null), pathRawResult)
	values["samefile"] = newMethod("samefile", pathSameFile(trueValue, falseValue), pathRawResult)
	values["stat"] = newMethod("stat", pathStat, pathRawResult)
	values["symlink_to"] = newMethod("symlink_to", pathSymlinkTo(null), pathRawResult)
	values["touch"] = newMethod("touch", pathTouch(null), pathRawResult)
	values["unlink"] = newMethod("unlink", pathUnlink(null), pathRawResult)
	values["walk"] = newMethod("walk", pathWalk, pathWalkResult)
	values["with_name"] = newMethod("with_name", pathWithName, pathValueResult)
	values["with_stem"] = newMethod("with_stem", pathWithStem, pathValueResult)
	values["with_suffix"] = newMethod("with_suffix", pathWithSuffix, pathValueResult)
	values["write_bytes"] = newMethod("write_bytes", pathWrite("write_bytes"), pathRawResult)
	values["write_text"] = newMethod("write_text", pathWrite("write_text"), pathRawResult)
	return &object.StructInstance{Struct: pathType, Values: values}
}

func invokePathProperty(operation object.BuiltinFunction, value string) object.Object {
	return operation(&object.String{Value: value})
}

func pathFieldSignature(pathType *object.Struct, name string) *ast.TypeAnnotation {
	for index, field := range pathType.Fields {
		if field == name {
			return pathType.FieldTypes[index]
		}
	}
	return nil
}

func bindPathMethod(pathType *object.Struct, null *object.Null, trueValue, falseValue *object.Boolean, value, name string, operation object.BuiltinFunction, resultKind pathResultKind) *object.Builtin {
	return &object.Builtin{
		Signature: pathFieldSignature(pathType, name),
		Fn: func(args ...object.Object) object.Object {
			normalized := make([]object.Object, len(args)+1)
			normalized[0] = &object.String{Value: value}
			for index, argument := range args {
				normalized[index+1] = pathArgument(pathType, argument)
			}
			result := operation(normalized...)
			return convertPathResult(pathType, null, trueValue, falseValue, result, resultKind)
		},
	}
}

func pathArgument(pathType *object.Struct, value object.Object) object.Object {
	instance, ok := value.(*object.StructInstance)
	if !ok || instance.Struct != pathType {
		return value
	}
	path, ok := instance.Get("path")
	if !ok {
		return value
	}
	return path
}

func convertPathResult(pathType *object.Struct, null *object.Null, trueValue, falseValue *object.Boolean, result object.Object, kind pathResultKind) object.Object {
	if _, failed := result.(*object.Error); failed || kind == pathRawResult {
		return result
	}
	newPath := func(value string) object.Object {
		return newPathValue(pathType, null, trueValue, falseValue, value)
	}
	switch kind {
	case pathValueResult:
		value, ok := result.(*object.String)
		if !ok {
			return result
		}
		return newPath(value.Value)
	case pathArrayResult:
		array, ok := result.(*object.Array)
		if !ok {
			return result
		}
		elements := make([]object.Object, len(array.Elements))
		for index, element := range array.Elements {
			value, ok := element.(*object.String)
			if !ok {
				return result
			}
			elements[index] = newPath(value.Value)
		}
		return &object.Array{Elements: elements}
	case pathWalkResult:
		array, ok := result.(*object.Array)
		if !ok {
			return result
		}
		for _, element := range array.Elements {
			row, ok := element.(*object.Array)
			if !ok || len(row.Elements) == 0 {
				continue
			}
			if root, ok := row.Elements[0].(*object.String); ok {
				row.Elements[0] = newPath(root.Value)
			}
		}
		return array
	}
	return result
}

func unaryPathString(name string, operation func(string) string) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		value, err := requireString(name, 0, args[0])
		if err != nil {
			return err
		}
		return &object.String{Value: operation(value)}
	}
}

func unaryPathPredicate(name string, predicate func(string) bool, trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		value, err := requireString(name, 0, args[0])
		if err != nil {
			return err
		}
		return pathBoolean(predicate(value), trueValue, falseValue)
	}
}

func pathBoolean(value bool, trueValue, falseValue *object.Boolean) object.Object {
	if value {
		return trueValue
	}
	return falseValue
}

func pathAnchor(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString("anchor", 0, args[0])
	if err != nil {
		return err
	}
	volume := filepath.VolumeName(value)
	if !filepath.IsAbs(value) {
		return &object.String{Value: ""}
	}
	return &object.String{Value: volume + string(os.PathSeparator)}
}

func pathRoot(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString("root", 0, args[0])
	if err != nil {
		return err
	}
	root := ""
	if filepath.IsAbs(value) {
		root = string(os.PathSeparator)
	}
	return &object.String{Value: root}
}

func pathParents(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString("parents", 0, args[0])
	if err != nil {
		return err
	}
	current := filepath.Clean(value)
	parents := []string{}
	for {
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		parents = append(parents, parent)
		current = parent
		if current == "." {
			break
		}
	}
	return pathStringArray(parents)
}

func pathParts(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString("parts", 0, args[0])
	if err != nil {
		return err
	}
	cleaned := filepath.Clean(value)
	volume := filepath.VolumeName(cleaned)
	remainder := strings.TrimPrefix(cleaned, volume)
	absolute := filepath.IsAbs(cleaned)
	components := strings.FieldsFunc(remainder, func(character rune) bool { return character == '/' || character == '\\' })
	parts := make([]string, 0, len(components)+1)
	if absolute {
		parts = append(parts, volume+string(os.PathSeparator))
	} else if volume != "" {
		parts = append(parts, volume)
	}
	parts = append(parts, components...)
	if len(parts) == 0 && cleaned == "." {
		parts = append(parts, ".")
	}
	return pathStringArray(parts)
}

func pathSuffix(name string) string {
	name = filepath.Base(name)
	index := strings.LastIndex(name, ".")
	if index <= 0 {
		return ""
	}
	return name[index:]
}

func pathStem(value string) string {
	name := filepath.Base(value)
	return strings.TrimSuffix(name, pathSuffix(name))
}

func pathSuffixes(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString("suffixes", 0, args[0])
	if err != nil {
		return err
	}
	name := filepath.Base(value)
	start := 0
	if strings.HasPrefix(name, ".") {
		start = 1
	}
	suffixes := []string{}
	for index := start; index < len(name); {
		dot := strings.IndexByte(name[index:], '.')
		if dot < 0 {
			break
		}
		dot += index
		next := strings.IndexByte(name[dot+1:], '.')
		if next < 0 {
			suffixes = append(suffixes, name[dot:])
			break
		}
		next += dot + 1
		suffixes = append(suffixes, name[dot:next])
		index = next
	}
	return pathStringArray(suffixes)
}

func pathAbsolute(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString("absolute", 0, args[0])
	if err != nil {
		return err
	}
	absolute, goErr := filepath.Abs(value)
	if goErr != nil {
		return pathOperationError("absolute", value, goErr)
	}
	return &object.String{Value: absolute}
}

func pathAsURI(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString("as_uri", 0, args[0])
	if err != nil {
		return err
	}
	if !filepath.IsAbs(value) {
		return newError(object.RuntimeErrorKindValue, "argument to `as_uri` must be an absolute path")
	}
	uriPath := filepath.ToSlash(filepath.Clean(value))
	if filepath.VolumeName(value) != "" && !strings.HasPrefix(uriPath, "/") {
		uriPath = "/" + uriPath
	}
	return &object.String{Value: (&url.URL{Scheme: "file", Path: uriPath}).String()}
}

func pathExpandUser(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString("expanduser", 0, args[0])
	if err != nil {
		return err
	}
	if value != "~" && !strings.HasPrefix(value, "~/") && !strings.HasPrefix(value, `~\`) {
		return &object.String{Value: value}
	}
	home, goErr := os.UserHomeDir()
	if goErr != nil {
		return pathOperationError("expanduser", value, goErr)
	}
	if value == "~" {
		return &object.String{Value: home}
	}
	return &object.String{Value: filepath.Join(home, value[2:])}
}

func pathJoin(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want at least=2", len(args))
	}
	parts := make([]string, len(args))
	for index, argument := range args {
		value, err := requireString("joinpath", index, argument)
		if err != nil {
			return err
		}
		parts[index] = value
	}
	return &object.String{Value: filepath.Join(parts...)}
}

func pathMatch(trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		value, err := requireString("match", 0, args[0])
		if err != nil {
			return err
		}
		pattern, err := requireString("match", 1, args[1])
		if err != nil {
			return err
		}
		candidate := filepath.ToSlash(value)
		if !strings.Contains(pattern, "/") {
			candidate = pathpkg.Base(candidate)
		}
		matched, goErr := pathpkg.Match(pattern, candidate)
		if goErr != nil {
			return pathOperationError("match", value, goErr)
		}
		return pathBoolean(matched, trueValue, falseValue)
	}
}

func pathRelativeTo(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	value, err := requireString("relative_to", 0, args[0])
	if err != nil {
		return err
	}
	base, err := requireString("relative_to", 1, args[1])
	if err != nil {
		return err
	}
	relative, goErr := filepath.Rel(base, value)
	if goErr != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return newError(object.RuntimeErrorKindValue, "%q is not relative to %q", value, base)
	}
	return &object.String{Value: relative}
}

func pathIsRelativeTo(trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		value, err := requireString("is_relative_to", 0, args[0])
		if err != nil {
			return err
		}
		base, err := requireString("is_relative_to", 1, args[1])
		if err != nil {
			return err
		}
		relative, goErr := filepath.Rel(base, value)
		if goErr != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
			return falseValue
		}
		return trueValue
	}
}

func pathResolve(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString("resolve", 0, args[0])
	if err != nil {
		return err
	}
	resolved, goErr := filepath.EvalSymlinks(value)
	if goErr != nil {
		return pathOperationError("resolve", value, goErr)
	}
	resolved, goErr = filepath.Abs(resolved)
	if goErr != nil {
		return pathOperationError("resolve", value, goErr)
	}
	return &object.String{Value: resolved}
}

func pathWithName(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	value, err := requireString("with_name", 0, args[0])
	if err != nil {
		return err
	}
	name, err := requireString("with_name", 1, args[1])
	if err != nil {
		return err
	}
	if name == "" || filepath.Base(name) != name || name == "." || name == ".." {
		return newError(object.RuntimeErrorKindValue, "invalid path name %q", name)
	}
	return &object.String{Value: filepath.Join(filepath.Dir(value), name)}
}

func pathWithStem(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	value, err := requireString("with_stem", 0, args[0])
	if err != nil {
		return err
	}
	stem, err := requireString("with_stem", 1, args[1])
	if err != nil {
		return err
	}
	if stem == "" || filepath.Base(stem) != stem {
		return newError(object.RuntimeErrorKindValue, "invalid path stem %q", stem)
	}
	return &object.String{Value: filepath.Join(filepath.Dir(value), stem+pathSuffix(value))}
}

func pathWithSuffix(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	value, err := requireString("with_suffix", 0, args[0])
	if err != nil {
		return err
	}
	suffix, err := requireString("with_suffix", 1, args[1])
	if err != nil {
		return err
	}
	if suffix != "" && (!strings.HasPrefix(suffix, ".") || strings.ContainsAny(suffix, `/\`)) {
		return newError(object.RuntimeErrorKindValue, "invalid suffix %q", suffix)
	}
	name := filepath.Base(value)
	name = strings.TrimSuffix(name, pathSuffix(name)) + suffix
	return &object.String{Value: filepath.Join(filepath.Dir(value), name)}
}

func pathExists(trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		value, err := requireString("exists", 0, args[0])
		if err != nil {
			return err
		}
		_, goErr := os.Stat(value)
		return pathBoolean(goErr == nil, trueValue, falseValue)
	}
}

func pathInfoPredicate(name string, predicate func(fs.FileInfo) bool, trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		value, err := requireString(name, 0, args[0])
		if err != nil {
			return err
		}
		info, goErr := os.Stat(value)
		return pathBoolean(goErr == nil && predicate(info), trueValue, falseValue)
	}
}

func pathModePredicate(name string, mode fs.FileMode, trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return pathInfoPredicate(name, func(info fs.FileInfo) bool { return info.Mode()&mode == mode }, trueValue, falseValue)
}

func pathIsSymlink(trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		value, err := requireString("is_symlink", 0, args[0])
		if err != nil {
			return err
		}
		info, goErr := os.Lstat(value)
		return pathBoolean(goErr == nil && info.Mode()&os.ModeSymlink != 0, trueValue, falseValue)
	}
}

func pathIsMount(trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		value, err := requireString("is_mount", 0, args[0])
		if err != nil {
			return err
		}
		absolute, goErr := filepath.Abs(value)
		if goErr != nil {
			return falseValue
		}
		parent := filepath.Dir(absolute)
		return pathBoolean(parent == absolute, trueValue, falseValue)
	}
}

func pathGlob(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	base, err := requireString("glob", 0, args[0])
	if err != nil {
		return err
	}
	pattern, err := requireString("glob", 1, args[1])
	if err != nil {
		return err
	}
	matches, goErr := filepath.Glob(filepath.Join(base, filepath.FromSlash(pattern)))
	if goErr != nil {
		return pathOperationError("glob", base, goErr)
	}
	return pathStringArray(matches)
}

func pathRGlob(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	base, err := requireString("rglob", 0, args[0])
	if err != nil {
		return err
	}
	pattern, err := requireString("rglob", 1, args[1])
	if err != nil {
		return err
	}
	matches := []string{}
	goErr := filepath.WalkDir(base, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == base {
			return nil
		}
		relative, relErr := filepath.Rel(base, current)
		if relErr != nil {
			return relErr
		}
		candidate := filepath.ToSlash(relative)
		if !strings.Contains(pattern, "/") {
			candidate = entry.Name()
		}
		matched, matchErr := pathpkg.Match(pattern, candidate)
		if matchErr != nil {
			return matchErr
		}
		if matched {
			matches = append(matches, current)
		}
		return nil
	})
	if goErr != nil {
		return pathOperationError("rglob", base, goErr)
	}
	return pathStringArray(matches)
}

func pathIterDir(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString("iterdir", 0, args[0])
	if err != nil {
		return err
	}
	entries, goErr := os.ReadDir(value)
	if goErr != nil {
		return pathOperationError("iterdir", value, goErr)
	}
	paths := make([]string, len(entries))
	for index, entry := range entries {
		paths[index] = filepath.Join(value, entry.Name())
	}
	return pathStringArray(paths)
}

func pathChmod(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		value, err := requireString("chmod", 0, args[0])
		if err != nil {
			return err
		}
		mode, ok := args[1].(*object.Integer)
		if !ok {
			return newError(object.RuntimeErrorKindType, "argument 2 to `chmod` must be INTEGER, got %s", args[1].Type())
		}
		if mode.Value < 0 {
			return newError(object.RuntimeErrorKindValue, "argument 2 to `chmod` must be nonnegative")
		}
		if goErr := os.Chmod(value, fs.FileMode(mode.Value)); goErr != nil {
			return pathOperationError("chmod", value, goErr)
		}
		return null
	}
}

func pathHardlinkTo(null *object.Null) object.BuiltinFunction {
	return pathLinkOperation("hardlink_to", os.Link, null)
}

func pathSymlinkTo(null *object.Null) object.BuiltinFunction {
	return pathLinkOperation("symlink_to", os.Symlink, null)
}

func pathLinkOperation(name string, operation func(string, string) error, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		link, err := requireString(name, 0, args[0])
		if err != nil {
			return err
		}
		target, err := requireString(name, 1, args[1])
		if err != nil {
			return err
		}
		if goErr := operation(target, link); goErr != nil {
			return pathOperationError(name, link, goErr)
		}
		return null
	}
}

func pathMkdir(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if len(args) < 1 || len(args) > 3 {
			return newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=1..3", len(args))
		}
		value, err := requireString("mkdir", 0, args[0])
		if err != nil {
			return err
		}
		parents, boolErr := optionalBoolean("mkdir", args, 1, false)
		if boolErr != nil {
			return boolErr
		}
		existOK, boolErr := optionalBoolean("mkdir", args, 2, false)
		if boolErr != nil {
			return boolErr
		}
		var goErr error
		if parents {
			if !existOK {
				if _, statErr := os.Stat(value); statErr == nil {
					return newError(object.RuntimeErrorKindValue, "mkdir failed for %q: path already exists", value)
				}
			}
			goErr = os.MkdirAll(value, 0777)
		} else {
			goErr = os.Mkdir(value, 0777)
		}
		if goErr != nil && !(existOK && os.IsExist(goErr)) {
			return pathOperationError("mkdir", value, goErr)
		}
		return null
	}
}

func pathOpen(null *object.Null) object.BuiltinFunction {
	opener := builtinOpen(null)
	return func(args ...object.Object) object.Object { return opener(args...) }
}

func pathRead(name string) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		value, err := requireString(name, 0, args[0])
		if err != nil {
			return err
		}
		contents, goErr := os.ReadFile(value)
		if goErr != nil {
			return pathOperationError(name, value, goErr)
		}
		return &object.String{Value: string(contents)}
	}
}

func pathWrite(name string) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		value, err := requireString(name, 0, args[0])
		if err != nil {
			return err
		}
		contents, err := requireString(name, 1, args[1])
		if err != nil {
			return err
		}
		if goErr := os.WriteFile(value, []byte(contents), 0666); goErr != nil {
			return pathOperationError(name, value, goErr)
		}
		return &object.Integer{Value: int64(len(contents))}
	}
}

func pathReadlink(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString("readlink", 0, args[0])
	if err != nil {
		return err
	}
	target, goErr := os.Readlink(value)
	if goErr != nil {
		return pathOperationError("readlink", value, goErr)
	}
	return &object.String{Value: target}
}

func pathRename(name string) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		value, err := requireString(name, 0, args[0])
		if err != nil {
			return err
		}
		target, err := requireString(name, 1, args[1])
		if err != nil {
			return err
		}
		if goErr := os.Rename(value, target); goErr != nil {
			return pathOperationError(name, value, goErr)
		}
		return &object.String{Value: filepath.Clean(target)}
	}
}

func pathRmdir(null *object.Null) object.BuiltinFunction {
	return pathRemove("rmdir", false, null)
}

func pathUnlink(null *object.Null) object.BuiltinFunction {
	return pathRemove("unlink", true, null)
}

func pathRemove(name string, optionalMissing bool, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		maximum := 1
		if optionalMissing {
			maximum = 2
		}
		if len(args) < 1 || len(args) > maximum {
			return newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=1..%d", len(args), maximum)
		}
		value, err := requireString(name, 0, args[0])
		if err != nil {
			return err
		}
		missingOK, boolErr := optionalBoolean(name, args, 1, false)
		if boolErr != nil {
			return boolErr
		}
		if goErr := os.Remove(value); goErr != nil && !(missingOK && os.IsNotExist(goErr)) {
			return pathOperationError(name, value, goErr)
		}
		return null
	}
}

func pathTouch(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if len(args) < 1 || len(args) > 2 {
			return newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=1..2", len(args))
		}
		value, err := requireString("touch", 0, args[0])
		if err != nil {
			return err
		}
		existOK, boolErr := optionalBoolean("touch", args, 1, true)
		if boolErr != nil {
			return boolErr
		}
		flags := os.O_WRONLY | os.O_CREATE
		if !existOK {
			flags |= os.O_EXCL
		}
		file, goErr := os.OpenFile(value, flags, 0666)
		if goErr != nil {
			return pathOperationError("touch", value, goErr)
		}
		if closeErr := file.Close(); closeErr != nil {
			return pathOperationError("touch", value, closeErr)
		}
		if existOK {
			now := time.Now()
			if goErr = os.Chtimes(value, now, now); goErr != nil {
				return pathOperationError("touch", value, goErr)
			}
		}
		return null
	}
}

func optionalBoolean(name string, args []object.Object, index int, fallback bool) (bool, *object.Error) {
	if index >= len(args) {
		return fallback, nil
	}
	value, ok := args[index].(*object.Boolean)
	if !ok {
		return false, newError(object.RuntimeErrorKindType, "argument %d to `%s` must be BOOLEAN, got %s", index+1, name, args[index].Type())
	}
	return value.Value, nil
}

func pathSameFile(trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		left, err := requireString("samefile", 0, args[0])
		if err != nil {
			return err
		}
		right, err := requireString("samefile", 1, args[1])
		if err != nil {
			return err
		}
		leftInfo, goErr := os.Stat(left)
		if goErr != nil {
			return pathOperationError("samefile", left, goErr)
		}
		rightInfo, goErr := os.Stat(right)
		if goErr != nil {
			return pathOperationError("samefile", right, goErr)
		}
		return pathBoolean(os.SameFile(leftInfo, rightInfo), trueValue, falseValue)
	}
}

func pathStat(args ...object.Object) object.Object {
	return pathFileInfo("stat", os.Stat, args)
}

func pathLstat(args ...object.Object) object.Object {
	return pathFileInfo("lstat", os.Lstat, args)
}

func pathFileInfo(name string, operation func(string) (fs.FileInfo, error), args []object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString(name, 0, args[0])
	if err != nil {
		return err
	}
	info, goErr := operation(value)
	if goErr != nil {
		return pathOperationError(name, value, goErr)
	}
	return pathInfoMap(info)
}

func pathInfoMap(info fs.FileInfo) object.Object {
	values := map[string]object.Object{
		"name":     &object.String{Value: info.Name()},
		"size":     &object.Integer{Value: info.Size()},
		"mode":     &object.Integer{Value: int64(info.Mode())},
		"modified": &object.Integer{Value: info.ModTime().Unix()},
		"is_dir":   &object.Boolean{Value: info.IsDir()},
	}
	pairs := make(map[object.HashKey]object.MapPair, len(values))
	for name, value := range values {
		key := &object.String{Value: name}
		pairs[key.HashKey()] = object.MapPair{Key: key, Value: value}
	}
	return &object.Map{Pairs: pairs}
}

func pathWalk(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	base, err := requireString("walk", 0, args[0])
	if err != nil {
		return err
	}
	rows := []object.Object{}
	goErr := filepath.WalkDir(base, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}
		entries, readErr := os.ReadDir(current)
		if readErr != nil {
			return readErr
		}
		directories := []string{}
		files := []string{}
		for _, child := range entries {
			if child.IsDir() {
				directories = append(directories, child.Name())
			} else {
				files = append(files, child.Name())
			}
		}
		rows = append(rows, &object.Array{Elements: []object.Object{
			&object.String{Value: current},
			pathStringArray(directories),
			pathStringArray(files),
		}})
		return nil
	})
	if goErr != nil {
		return pathOperationError("walk", base, goErr)
	}
	return &object.Array{Elements: rows}
}

func pathStringArray(values []string) *object.Array {
	elements := make([]object.Object, len(values))
	for index, value := range values {
		elements[index] = &object.String{Value: value}
	}
	return &object.Array{Elements: elements}
}

func pathOperationError(operation, value string, err error) object.Object {
	if value == "" {
		return newError(object.RuntimeErrorKindValue, "%s failed: %s", operation, err)
	}
	return newError(object.RuntimeErrorKindValue, "%s failed for %q: %s", operation, value, err)
}
