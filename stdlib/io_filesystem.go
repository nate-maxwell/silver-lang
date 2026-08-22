package stdlib

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"silver/ast"
	"silver/object"
	"time"
)

// filesystemDefinitions exposes the OS boundary used by Silver-authored
// filesystem libraries. Lexical path behavior intentionally does not live
// here; these functions only perform host queries or mutations.
func filesystemDefinitions(null *object.Null, trueValue, falseValue *object.Boolean) []definition {
	str := namedType("str")
	integer := namedType("int")
	boolean := namedType("bool")
	windowsValue := falseValue
	if runtime.GOOS == "windows" {
		windowsValue = trueValue
	}
	return []definition{
		{name: "_path_separator", value: &object.String{Value: string(os.PathSeparator)}},
		{name: "_is_windows", value: windowsValue},
		{name: "_cwd", fn: fsCwd, signature: callSignature(nil, nil, str, "IOError")},
		{name: "_home", fn: fsHome, signature: callSignature(nil, nil, str, "IOError")},
		{name: "_stat", fn: fsStat(false, trueValue, falseValue), signature: callSignature([]string{"path"}, []*ast.TypeAnnotation{str}, namedType("map"), "IOError")},
		{name: "_lstat", fn: fsStat(true, trueValue, falseValue), signature: callSignature([]string{"path"}, []*ast.TypeAnnotation{str}, namedType("map"), "IOError")},
		{name: "_read_dir", fn: fsReadDir, signature: callSignature([]string{"path"}, []*ast.TypeAnnotation{str}, namedType("array"), "IOError")},
		{name: "_read_file", fn: fsReadFile, signature: callSignature([]string{"path"}, []*ast.TypeAnnotation{str}, str, "IOError")},
		{name: "_write_file", fn: fsWriteFile, signature: callSignature([]string{"path", "contents"}, []*ast.TypeAnnotation{str, str}, integer, "IOError")},
		{name: "_chmod", fn: fsChmod(null), signature: callSignature([]string{"path", "mode"}, []*ast.TypeAnnotation{str, integer}, nil, "IOError")},
		{name: "_mkdir", fn: fsMkdir(null), signature: callSignature([]string{"path", "parents", "exist_ok"}, []*ast.TypeAnnotation{str, boolean, boolean}, nil, "IOError")},
		{name: "_remove", fn: fsRemove(null), signature: callSignature([]string{"path", "missing_ok"}, []*ast.TypeAnnotation{str, boolean}, nil, "IOError")},
		{name: "_rename", fn: fsRename(null), signature: callSignature([]string{"source", "target"}, []*ast.TypeAnnotation{str, str}, nil, "IOError")},
		{name: "_hardlink", fn: fsLink(false, null), signature: callSignature([]string{"target", "link"}, []*ast.TypeAnnotation{str, str}, nil, "IOError")},
		{name: "_symlink", fn: fsLink(true, null), signature: callSignature([]string{"target", "link"}, []*ast.TypeAnnotation{str, str}, nil, "IOError")},
		{name: "_readlink", fn: fsReadlink, signature: callSignature([]string{"path"}, []*ast.TypeAnnotation{str}, str, "IOError")},
		{name: "_realpath", fn: fsRealpath, signature: callSignature([]string{"path"}, []*ast.TypeAnnotation{str}, str, "IOError")},
		{name: "_same_file", fn: fsSameFile(trueValue, falseValue), signature: callSignature([]string{"left", "right"}, []*ast.TypeAnnotation{str, str}, boolean, "IOError")},
		{name: "_touch", fn: fsTouch(null), signature: callSignature([]string{"path", "exist_ok"}, []*ast.TypeAnnotation{str, boolean}, nil, "IOError")},
	}
}

func fsCwd(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	value, err := os.Getwd()
	if err != nil {
		return ioErrorValue("IOError", err)
	}
	return &object.String{Value: value}
}

func fsHome(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	value, err := os.UserHomeDir()
	if err != nil {
		return ioErrorValue("IOError", err)
	}
	return &object.String{Value: value}
}

func fsStat(symlink bool, trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		path, err := requireString("stat", 0, args[0])
		if err != nil {
			return err
		}
		var info fs.FileInfo
		var statErr error
		if symlink {
			info, statErr = os.Lstat(path)
		} else {
			info, statErr = os.Stat(path)
		}
		if statErr != nil {
			return ioErrorValue("IOError", statErr)
		}
		mode := info.Mode()
		booleanValue := func(value bool) object.Object {
			if value {
				return trueValue
			}
			return falseValue
		}
		return objectMap(map[string]object.Object{
			"name":            &object.String{Value: info.Name()},
			"size":            &object.Integer{Value: info.Size()},
			"mode":            &object.Integer{Value: int64(mode)},
			"modified":        &object.Integer{Value: info.ModTime().Unix()},
			"is_dir":          booleanValue(info.IsDir()),
			"is_file":         booleanValue(mode.IsRegular()),
			"is_block_device": booleanValue(mode&os.ModeDevice != 0 && mode&os.ModeCharDevice == 0),
			"is_char_device":  booleanValue(mode&os.ModeDevice != 0 && mode&os.ModeCharDevice != 0),
			"is_fifo":         booleanValue(mode&os.ModeNamedPipe != 0),
			"is_socket":       booleanValue(mode&os.ModeSocket != 0),
			"is_symlink":      booleanValue(mode&os.ModeSymlink != 0),
		})
	}
}

func objectMap(values map[string]object.Object) *object.Map {
	pairs := make(map[object.HashKey]object.MapPair, len(values))
	for name, value := range values {
		key := &object.String{Value: name}
		pairs[key.HashKey()] = object.MapPair{Key: key, Value: value}
	}
	return &object.Map{Pairs: pairs}
}

func fsReadDir(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	path, err := requireString("read_dir", 0, args[0])
	if err != nil {
		return err
	}
	entries, readErr := os.ReadDir(path)
	if readErr != nil {
		return ioErrorValue("IOError", readErr)
	}
	elements := make([]object.Object, len(entries))
	for index, entry := range entries {
		elements[index] = &object.String{Value: entry.Name()}
	}
	return &object.Array{Elements: elements}
}

func fsReadFile(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	path, err := requireString("read_file", 0, args[0])
	if err != nil {
		return err
	}
	contents, readErr := os.ReadFile(path)
	if readErr != nil {
		return ioErrorValue("IOError", readErr)
	}
	return &object.String{Value: string(contents)}
}

func fsWriteFile(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	path, err := requireString("write_file", 0, args[0])
	if err != nil {
		return err
	}
	contents, err := requireString("write_file", 1, args[1])
	if err != nil {
		return err
	}
	if writeErr := os.WriteFile(path, []byte(contents), 0666); writeErr != nil {
		return ioErrorValue("IOError", writeErr)
	}
	return &object.Integer{Value: int64(len(contents))}
}

func fsChmod(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		path, err := requireString("chmod", 0, args[0])
		if err != nil {
			return err
		}
		mode, ok := args[1].(*object.Integer)
		if !ok {
			return newError(object.RuntimeErrorKindType, "argument to `chmod` must be INTEGER, got %s", args[1].Type())
		}
		if chmodErr := os.Chmod(path, fs.FileMode(mode.Value)); chmodErr != nil {
			return ioErrorValue("IOError", chmodErr)
		}
		return null
	}
}

func fsMkdir(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 3); err != nil {
			return err
		}
		path, err := requireString("_mkdir", 0, args[0])
		if err != nil {
			return err
		}
		parents, err := fsRequireBoolean("_mkdir", 1, args[1])
		if err != nil {
			return err
		}
		existOK, err := fsRequireBoolean("_mkdir", 2, args[2])
		if err != nil {
			return err
		}
		if parents {
			if !existOK {
				if _, err := os.Stat(path); err == nil {
					return ioErrorMessage("path already exists")
				}
			}
			if err := os.MkdirAll(path, 0777); err != nil {
				return ioErrorValue("IOError", err)
			}
			return null
		}
		if err := os.Mkdir(path, 0777); err != nil && !(existOK && os.IsExist(err)) {
			return ioErrorValue("IOError", err)
		}
		return null
	}
}

func fsRemove(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		path, err := requireString("_remove", 0, args[0])
		if err != nil {
			return err
		}
		missingOK, err := fsRequireBoolean("_remove", 1, args[1])
		if err != nil {
			return err
		}
		if err := os.Remove(path); err != nil && !(missingOK && os.IsNotExist(err)) {
			return ioErrorValue("IOError", err)
		}
		return null
	}
}

func fsRename(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		source, typeErr := requireString("_rename", 0, args[0])
		if typeErr != nil {
			return typeErr
		}
		target, typeErr := requireString("_rename", 1, args[1])
		if typeErr != nil {
			return typeErr
		}
		if err := os.Rename(source, target); err != nil {
			return ioErrorValue("IOError", err)
		}
		return null
	}
}

func fsLink(symlink bool, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		target, typeErr := requireString("_link", 0, args[0])
		if typeErr != nil {
			return typeErr
		}
		link, typeErr := requireString("_link", 1, args[1])
		if typeErr != nil {
			return typeErr
		}
		var err error
		if symlink {
			err = os.Symlink(target, link)
		} else {
			err = os.Link(target, link)
		}
		if err != nil {
			return ioErrorValue("IOError", err)
		}
		return null
	}
}

func fsReadlink(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	path, typeErr := requireString("_readlink", 0, args[0])
	if typeErr != nil {
		return typeErr
	}
	target, err := os.Readlink(path)
	if err != nil {
		return ioErrorValue("IOError", err)
	}
	return &object.String{Value: target}
}

func fsRealpath(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	path, typeErr := requireString("_realpath", 0, args[0])
	if typeErr != nil {
		return typeErr
	}
	value, err := filepath.EvalSymlinks(path)
	if err != nil {
		return ioErrorValue("IOError", err)
	}
	return &object.String{Value: value}
}

func fsSameFile(trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		leftPath, typeErr := requireString("_same_file", 0, args[0])
		if typeErr != nil {
			return typeErr
		}
		rightPath, typeErr := requireString("_same_file", 1, args[1])
		if typeErr != nil {
			return typeErr
		}
		left, err := os.Stat(leftPath)
		if err != nil {
			return ioErrorValue("IOError", err)
		}
		right, err := os.Stat(rightPath)
		if err != nil {
			return ioErrorValue("IOError", err)
		}
		if os.SameFile(left, right) {
			return trueValue
		}
		return falseValue
	}
}

func fsTouch(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		path, typeErr := requireString("_touch", 0, args[0])
		if typeErr != nil {
			return typeErr
		}
		existOK, typeErr := fsRequireBoolean("_touch", 1, args[1])
		if typeErr != nil {
			return typeErr
		}
		flags := os.O_WRONLY | os.O_CREATE
		if !existOK {
			flags |= os.O_EXCL
		}
		file, osErr := os.OpenFile(path, flags, 0666)
		if osErr != nil {
			return ioErrorValue("IOError", osErr)
		}
		if osErr = file.Close(); osErr != nil {
			return ioErrorValue("IOError", osErr)
		}
		if existOK {
			now := time.Now()
			if osErr = os.Chtimes(path, now, now); osErr != nil {
				return ioErrorValue("IOError", osErr)
			}
		}
		return null
	}
}

func fsRequireBoolean(name string, index int, value object.Object) (bool, *object.Error) {
	boolean, ok := value.(*object.Boolean)
	if !ok {
		return false, newError(object.RuntimeErrorKindType, "argument %d to `%s` must be BOOLEAN, got %s", index+1, name, value.Type())
	}
	return boolean.Value, nil
}
