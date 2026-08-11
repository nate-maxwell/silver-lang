package stdlib

import (
	"regexp"
	"silver/ast"
	"silver/object"
)

// regexDefinitions contains the regular-expression functions exported by
// import("regex"). Patterns use Go's RE2 syntax, which guarantees linear-time
// matching and supports both numbered and named capture groups.
func regexDefinitions(null *object.Null) []definition {
	matchType, expressionType := newRegexStructDefinitions()
	return []definition{
		{name: "MatchObject", value: matchType},
		{name: "Expression", value: expressionType},
		{name: "match", fn: regexMatch(matchType, null)},
		{name: "search", fn: regexSearch(matchType, null)},
		{name: "findall", fn: regexFindAll},
		{name: "findlist", fn: regexFindList(matchType, null)},
		{name: "sub", fn: regexSub},
		{name: "subn", fn: regexSubN},
		{name: "split", fn: regexSplit},
		{name: "fullmatch", fn: regexFullMatch(matchType, null)},
		{name: "escape", fn: regexEscape},
		{name: "compile", fn: regexCompile(matchType, expressionType, null)},
	}
}

func newRegexStructDefinitions() (*object.Struct, *object.Struct) {
	environment := object.NewEnvironment()
	matchType := &object.Struct{
		Name: "MatchObject",
		Fields: []string{
			"group", "groups", "groupmap", "start", "end", "span", "string",
		},
		FieldTypes: []*ast.TypeAnnotation{
			nil,
			callSignature(nil, nil, namedType("array")),
			callSignature(nil, nil, namedType("map")),
			nil,
			nil,
			nil,
			namedType("str"),
		},
		Env: environment,
	}

	strType := namedType("str")
	expressionType := &object.Struct{
		Name: "Expression",
		Fields: []string{
			"match", "search", "findall", "findlist", "sub", "subn", "split", "fullmatch", "escape",
		},
		FieldTypes: []*ast.TypeAnnotation{
			nil,
			nil,
			callSignature([]string{"string"}, []*ast.TypeAnnotation{strType}, namedType("array")),
			callSignature([]string{"string"}, []*ast.TypeAnnotation{strType}, namedType("array")),
			callSignature([]string{"replacement", "string"}, []*ast.TypeAnnotation{strType, strType}, strType),
			callSignature([]string{"replacement", "string"}, []*ast.TypeAnnotation{strType, strType}, namedType("array")),
			callSignature([]string{"string"}, []*ast.TypeAnnotation{strType}, namedType("array")),
			nil,
			callSignature([]string{"string"}, []*ast.TypeAnnotation{strType}, strType),
		},
		Env: environment,
	}
	environment.Set("MatchObject", matchType)
	environment.Set("Expression", expressionType)
	return matchType, expressionType
}

func regexMatch(matchType *object.Struct, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		compiled, input, err := regexPatternAndInput("match", args)
		if err != nil {
			return err
		}
		return regexFirstMatch(compiled, input, matchType, null, true)
	}
}

func regexSearch(matchType *object.Struct, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		compiled, input, err := regexPatternAndInput("search", args)
		if err != nil {
			return err
		}
		return regexFirstMatch(compiled, input, matchType, null, false)
	}
}

func regexFullMatch(matchType *object.Struct, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		compiled, input, err := regexPatternAndInput("fullmatch", args)
		if err != nil {
			return err
		}
		return regexFirstMatch(fullMatchRegex(compiled), input, matchType, null, true)
	}
}

func regexFindAll(args ...object.Object) object.Object {
	compiled, input, err := regexPatternAndInput("findall", args)
	if err != nil {
		return err
	}
	matches := compiled.FindAllString(input, -1)
	elements := make([]object.Object, len(matches))
	for index, match := range matches {
		elements[index] = &object.String{Value: match}
	}
	return &object.Array{Elements: elements}
}

func regexFindList(matchType *object.Struct, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		compiled, input, err := regexPatternAndInput("findlist", args)
		if err != nil {
			return err
		}
		matches := compiled.FindAllStringSubmatchIndex(input, -1)
		elements := make([]object.Object, len(matches))
		for index, match := range matches {
			elements[index] = newRegexMatchObject(matchType, compiled, input, match, null)
		}
		return &object.Array{Elements: elements}
	}
}

func regexSub(args ...object.Object) object.Object {
	compiled, replacement, input, err := regexSubArguments("sub", args)
	if err != nil {
		return err
	}
	return &object.String{Value: compiled.ReplaceAllString(input, replacement)}
}

func regexSubN(args ...object.Object) object.Object {
	compiled, replacement, input, err := regexSubArguments("subn", args)
	if err != nil {
		return err
	}
	count := len(compiled.FindAllStringIndex(input, -1))
	return &object.Array{Elements: []object.Object{
		&object.String{Value: compiled.ReplaceAllString(input, replacement)},
		&object.Integer{Value: int64(count)},
	}}
}

func regexSplit(args ...object.Object) object.Object {
	compiled, input, err := regexPatternAndInput("split", args)
	if err != nil {
		return err
	}
	parts := compiled.Split(input, -1)
	elements := make([]object.Object, len(parts))
	for index, part := range parts {
		elements[index] = &object.String{Value: part}
	}
	return &object.Array{Elements: elements}
}

func regexEscape(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString("escape", 0, args[0])
	if err != nil {
		return err
	}
	return &object.String{Value: regexp.QuoteMeta(value)}
}

func regexCompile(matchType, expressionType *object.Struct, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		pattern, err := requireString("compile", 0, args[0])
		if err != nil {
			return err
		}
		compiled, compileErr := compileRegex(pattern, "compile")
		if compileErr != nil {
			return compileErr
		}
		return newRegexExpression(expressionType, matchType, compiled, null)
	}
}

func regexPatternAndInput(name string, args []object.Object) (*regexp.Regexp, string, *object.Error) {
	if err := requireArgumentCount(args, 2); err != nil {
		return nil, "", err
	}
	pattern, err := requireString(name, 0, args[0])
	if err != nil {
		return nil, "", err
	}
	input, err := requireString(name, 1, args[1])
	if err != nil {
		return nil, "", err
	}
	compiled, compileErr := compileRegex(pattern, name)
	if compileErr != nil {
		return nil, "", compileErr
	}
	return compiled, input, nil
}

func regexSubArguments(name string, args []object.Object) (*regexp.Regexp, string, string, *object.Error) {
	if err := requireArgumentCount(args, 3); err != nil {
		return nil, "", "", err
	}
	pattern, err := requireString(name, 0, args[0])
	if err != nil {
		return nil, "", "", err
	}
	replacement, err := requireString(name, 1, args[1])
	if err != nil {
		return nil, "", "", err
	}
	input, err := requireString(name, 2, args[2])
	if err != nil {
		return nil, "", "", err
	}
	compiled, compileErr := compileRegex(pattern, name)
	if compileErr != nil {
		return nil, "", "", compileErr
	}
	return compiled, replacement, input, nil
}

func compileRegex(pattern, name string) (*regexp.Regexp, *object.Error) {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, newError(object.RuntimeErrorKindValue, "invalid pattern passed to `%s`: %s", name, err)
	}
	return compiled, nil
}

func fullMatchRegex(compiled *regexp.Regexp) *regexp.Regexp {
	// Checking the end of an ordinary first match is not sufficient for
	// alternatives such as "a|ab". Anchoring the expression lets the engine
	// try later alternatives that can consume the whole input. A pattern that
	// compiled on its own is also valid inside this non-capturing group.
	return regexp.MustCompile(`\A(?:` + compiled.String() + `)\z`)
}

func regexFirstMatch(compiled *regexp.Regexp, input string, matchType *object.Struct, null *object.Null, atStart bool) object.Object {
	indices := compiled.FindStringSubmatchIndex(input)
	if indices == nil || atStart && indices[0] != 0 {
		return null
	}
	return newRegexMatchObject(matchType, compiled, input, indices, null)
}

func newRegexMatchObject(matchType *object.Struct, compiled *regexp.Regexp, input string, indices []int, null *object.Null) *object.StructInstance {
	// FindStringSubmatchIndex returns a fresh slice. Copying it here also makes
	// the match value independent from its producing operation.
	matchIndices := append([]int(nil), indices...)
	values := map[string]object.Object{
		"string": &object.String{Value: input},
	}
	values["group"] = &object.Builtin{Fn: regexMatchGroup(compiled, input, matchIndices, null)}
	values["groups"] = &object.Builtin{Fn: regexMatchGroups(input, matchIndices, null)}
	values["groupmap"] = &object.Builtin{Fn: regexMatchGroupMap(compiled, input, matchIndices, null)}
	values["start"] = &object.Builtin{Fn: regexMatchPosition("start", compiled, matchIndices, 0)}
	values["end"] = &object.Builtin{Fn: regexMatchPosition("end", compiled, matchIndices, 1)}
	values["span"] = &object.Builtin{Fn: regexMatchSpan(compiled, matchIndices)}
	return &object.StructInstance{Struct: matchType, Values: values}
}

func regexMatchGroup(compiled *regexp.Regexp, input string, indices []int, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		group, err := regexGroupArgument("group", compiled, args)
		if err != nil {
			return err
		}
		start, end := indices[group*2], indices[group*2+1]
		if start < 0 {
			// Optional capture groups that did not participate mirror Python's
			// match API by returning null.
			return null
		}
		return &object.String{Value: input[start:end]}
	}
}

func regexMatchGroups(input string, indices []int, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 0); err != nil {
			return err
		}
		elements := make([]object.Object, len(indices)/2-1)
		for group := 1; group < len(indices)/2; group++ {
			start, end := indices[group*2], indices[group*2+1]
			if start < 0 {
				elements[group-1] = null
			} else {
				elements[group-1] = &object.String{Value: input[start:end]}
			}
		}
		return &object.Array{Elements: elements}
	}
}

func regexMatchGroupMap(compiled *regexp.Regexp, input string, indices []int, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 0); err != nil {
			return err
		}
		pairs := make(map[object.HashKey]object.MapPair)
		for group, name := range compiled.SubexpNames() {
			if group == 0 || name == "" {
				continue
			}
			key := &object.String{Value: name}
			start, end := indices[group*2], indices[group*2+1]
			var value object.Object = null
			if start >= 0 {
				value = &object.String{Value: input[start:end]}
			}
			pairs[key.HashKey()] = object.MapPair{Key: key, Value: value}
		}
		return &object.Map{Pairs: pairs}
	}
}

func regexMatchPosition(name string, compiled *regexp.Regexp, indices []int, offset int) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		group, err := regexGroupArgument(name, compiled, args)
		if err != nil {
			return err
		}
		return &object.Integer{Value: int64(indices[group*2+offset])}
	}
}

func regexMatchSpan(compiled *regexp.Regexp, indices []int) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		group, err := regexGroupArgument("span", compiled, args)
		if err != nil {
			return err
		}
		return &object.Array{Elements: []object.Object{
			&object.Integer{Value: int64(indices[group*2])},
			&object.Integer{Value: int64(indices[group*2+1])},
		}}
	}
}

func regexGroupArgument(method string, compiled *regexp.Regexp, args []object.Object) (int, *object.Error) {
	if len(args) > 1 {
		return 0, newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=0 or 1", len(args))
	}
	if len(args) == 0 {
		return 0, nil
	}
	switch group := args[0].(type) {
	case *object.Integer:
		if group.Value < 0 || group.Value > int64(compiled.NumSubexp()) {
			return 0, newError(object.RuntimeErrorKindIndex, "no such group in `%s`: %d", method, group.Value)
		}
		return int(group.Value), nil
	case *object.String:
		index := compiled.SubexpIndex(group.Value)
		if index < 0 {
			return 0, newError(object.RuntimeErrorKindIndex, "no such group in `%s`: %s", method, group.Value)
		}
		return index, nil
	default:
		return 0, newError(object.RuntimeErrorKindType, "argument 1 to `%s` must be INTEGER or STRING, got %s", method, args[0].Type())
	}
}

func newRegexExpression(expressionType, matchType *object.Struct, compiled *regexp.Regexp, null *object.Null) *object.StructInstance {
	compiledFullMatch := fullMatchRegex(compiled)
	patternAndInput := func(name string, operation func(string) object.Object) object.Object {
		return &object.Builtin{Fn: func(args ...object.Object) object.Object {
			if err := requireArgumentCount(args, 1); err != nil {
				return err
			}
			input, err := requireString(name, 0, args[0])
			if err != nil {
				return err
			}
			return operation(input)
		}}
	}

	values := make(map[string]object.Object, len(expressionType.Fields))
	values["match"] = patternAndInput("match", func(input string) object.Object {
		return regexFirstMatch(compiled, input, matchType, null, true)
	})
	values["search"] = patternAndInput("search", func(input string) object.Object {
		return regexFirstMatch(compiled, input, matchType, null, false)
	})
	values["findall"] = patternAndInput("findall", func(input string) object.Object {
		matches := compiled.FindAllString(input, -1)
		elements := make([]object.Object, len(matches))
		for index, match := range matches {
			elements[index] = &object.String{Value: match}
		}
		return &object.Array{Elements: elements}
	})
	values["findlist"] = patternAndInput("findlist", func(input string) object.Object {
		matches := compiled.FindAllStringSubmatchIndex(input, -1)
		elements := make([]object.Object, len(matches))
		for index, match := range matches {
			elements[index] = newRegexMatchObject(matchType, compiled, input, match, null)
		}
		return &object.Array{Elements: elements}
	})
	values["split"] = patternAndInput("split", func(input string) object.Object {
		parts := compiled.Split(input, -1)
		elements := make([]object.Object, len(parts))
		for index, part := range parts {
			elements[index] = &object.String{Value: part}
		}
		return &object.Array{Elements: elements}
	})
	values["fullmatch"] = patternAndInput("fullmatch", func(input string) object.Object {
		return regexFirstMatch(compiledFullMatch, input, matchType, null, true)
	})
	values["escape"] = patternAndInput("escape", func(input string) object.Object {
		return &object.String{Value: regexp.QuoteMeta(input)}
	})
	values["sub"] = regexExpressionSub(compiled, false)
	values["subn"] = regexExpressionSub(compiled, true)
	return &object.StructInstance{Struct: expressionType, Values: values}
}

func regexExpressionSub(compiled *regexp.Regexp, withCount bool) object.Object {
	name := "sub"
	if withCount {
		name = "subn"
	}
	return &object.Builtin{Fn: func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		replacement, err := requireString(name, 0, args[0])
		if err != nil {
			return err
		}
		input, err := requireString(name, 1, args[1])
		if err != nil {
			return err
		}
		replaced := &object.String{Value: compiled.ReplaceAllString(input, replacement)}
		if !withCount {
			return replaced
		}
		return &object.Array{Elements: []object.Object{
			replaced,
			&object.Integer{Value: int64(len(compiled.FindAllStringIndex(input, -1)))},
		}}
	}}
}
