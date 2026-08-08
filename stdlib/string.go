package stdlib

import (
	"math"
	"silver/object"
	"strings"
	"unicode"
	"unicode/utf8"
)

// stringDefinitions contains the functions exported by import("string").
func stringDefinitions(trueValue, falseValue *object.Boolean) []definition {
	return []definition{
		{name: "capitalize", fn: stringCapitalize},
		{name: "chars", fn: stringChars},
		{name: "compare", fn: stringCompare},
		{name: "contains", fn: binaryStringPredicate("contains", strings.Contains, trueValue, falseValue)},
		{name: "count", fn: stringCount},
		{name: "endswith", fn: binaryStringPredicate("endswith", strings.HasSuffix, trueValue, falseValue)},
		{name: "equal_fold", fn: binaryStringPredicate("equal_fold", strings.EqualFold, trueValue, falseValue)},
		{name: "fields", fn: stringFields},
		{name: "find", fn: stringFind},
		{name: "isalnum", fn: stringClassifier("isalnum", isAlphanumeric, trueValue, falseValue)},
		{name: "isalpha", fn: stringClassifier("isalpha", isAlpha, trueValue, falseValue)},
		{name: "isascii", fn: stringClassifier("isascii", isASCII, trueValue, falseValue)},
		{name: "isdecimal", fn: stringClassifier("isdecimal", isDecimal, trueValue, falseValue)},
		{name: "isdigit", fn: stringClassifier("isdigit", isDigitString, trueValue, falseValue)},
		{name: "islower", fn: stringClassifier("islower", isLower, trueValue, falseValue)},
		{name: "isnumeric", fn: stringClassifier("isnumeric", isNumeric, trueValue, falseValue)},
		{name: "isprintable", fn: stringClassifier("isprintable", isPrintable, trueValue, falseValue)},
		{name: "isspace", fn: stringClassifier("isspace", isSpace, trueValue, falseValue)},
		{name: "istitle", fn: stringClassifier("istitle", isTitle, trueValue, falseValue)},
		{name: "isupper", fn: stringClassifier("isupper", isUpper, trueValue, falseValue)},
		{name: "join", fn: stringJoin},
		{name: "lower", fn: unaryStringFunction("lower", strings.ToLower)},
		{name: "lstrip", fn: unaryStringFunction("lstrip", func(value string) string { return strings.TrimLeftFunc(value, unicode.IsSpace) })},
		{name: "removeprefix", fn: stringRemovePrefix},
		{name: "removesuffix", fn: stringRemoveSuffix},
		{name: "repeat", fn: stringRepeat},
		{name: "replace", fn: stringReplace},
		{name: "reverse", fn: unaryStringFunction("reverse", reverseString)},
		{name: "rfind", fn: stringRFind},
		{name: "rstrip", fn: unaryStringFunction("rstrip", func(value string) string { return strings.TrimRightFunc(value, unicode.IsSpace) })},
		{name: "split", fn: stringSplit},
		{name: "splitlines", fn: stringSplitLines},
		{name: "startswith", fn: binaryStringPredicate("startswith", strings.HasPrefix, trueValue, falseValue)},
		{name: "strip", fn: unaryStringFunction("strip", strings.TrimSpace)},
		{name: "swapcase", fn: unaryStringFunction("swapcase", swapCase)},
		{name: "title", fn: unaryStringFunction("title", titleString)},
		{name: "upper", fn: unaryStringFunction("upper", strings.ToUpper)},
	}
}

func requireString(name string, index int, value object.Object) (string, *object.Error) {
	text, ok := value.(*object.String)
	if !ok {
		return "", newError(object.RuntimeErrorKindType, "argument %d to `%s` must be STRING, got %s", index+1, name, value.Type())
	}
	return text.Value, nil
}

func unaryStringFunction(name string, operation func(string) string) object.BuiltinFunction {
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

func binaryStringPredicate(name string, predicate func(string, string) bool, trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		left, err := requireString(name, 0, args[0])
		if err != nil {
			return err
		}
		right, err := requireString(name, 1, args[1])
		if err != nil {
			return err
		}
		if predicate(left, right) {
			return trueValue
		}
		return falseValue
	}
}

func stringClassifier(name string, predicate func(string) bool, trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		value, err := requireString(name, 0, args[0])
		if err != nil {
			return err
		}
		if predicate(value) {
			return trueValue
		}
		return falseValue
	}
}

func stringCapitalize(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString("capitalize", 0, args[0])
	if err != nil {
		return err
	}
	if value == "" {
		return &object.String{Value: ""}
	}
	first, size := utf8.DecodeRuneInString(value)
	return &object.String{Value: string(unicode.ToUpper(first)) + strings.ToLower(value[size:])}
}

func stringChars(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString("chars", 0, args[0])
	if err != nil {
		return err
	}
	elements := make([]object.Object, 0, utf8.RuneCountInString(value))
	for _, character := range value {
		elements = append(elements, &object.String{Value: string(character)})
	}
	return &object.Array{Elements: elements}
}

func stringCompare(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	left, err := requireString("compare", 0, args[0])
	if err != nil {
		return err
	}
	right, err := requireString("compare", 1, args[1])
	if err != nil {
		return err
	}
	return &object.Integer{Value: int64(strings.Compare(left, right))}
}

func stringCount(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	value, err := requireString("count", 0, args[0])
	if err != nil {
		return err
	}
	substring, err := requireString("count", 1, args[1])
	if err != nil {
		return err
	}
	return &object.Integer{Value: int64(strings.Count(value, substring))}
}

func stringFields(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString("fields", 0, args[0])
	if err != nil {
		return err
	}
	return stringArray(strings.Fields(value))
}

func stringFind(args ...object.Object) object.Object {
	return stringIndex("find", strings.Index, args)
}

func stringRFind(args ...object.Object) object.Object {
	return stringIndex("rfind", strings.LastIndex, args)
}

func stringIndex(name string, operation func(string, string) int, args []object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	value, err := requireString(name, 0, args[0])
	if err != nil {
		return err
	}
	substring, err := requireString(name, 1, args[1])
	if err != nil {
		return err
	}
	return &object.Integer{Value: int64(operation(value, substring))}
}

func stringJoin(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	array, err := requireArray("join", args[0])
	if err != nil {
		return err
	}
	separator, stringErr := requireString("join", 1, args[1])
	if stringErr != nil {
		return stringErr
	}
	values := make([]string, len(array.Elements))
	for index, element := range array.Elements {
		text, ok := element.(*object.String)
		if !ok {
			return newError(object.RuntimeErrorKindType, "element %d of argument 1 to `join` must be STRING, got %s", index+1, element.Type())
		}
		values[index] = text.Value
	}
	return &object.String{Value: strings.Join(values, separator)}
}

func stringRemovePrefix(args ...object.Object) object.Object {
	return stringCut("removeprefix", strings.TrimPrefix, args)
}

func stringRemoveSuffix(args ...object.Object) object.Object {
	return stringCut("removesuffix", strings.TrimSuffix, args)
}

func stringCut(name string, operation func(string, string) string, args []object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	value, err := requireString(name, 0, args[0])
	if err != nil {
		return err
	}
	affix, err := requireString(name, 1, args[1])
	if err != nil {
		return err
	}
	return &object.String{Value: operation(value, affix)}
}

func stringRepeat(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	value, err := requireString("repeat", 0, args[0])
	if err != nil {
		return err
	}
	count, ok := args[1].(*object.Integer)
	if !ok {
		return newError(object.RuntimeErrorKindType, "argument 2 to `repeat` must be INTEGER, got %s", args[1].Type())
	}
	if count.Value < 0 {
		return newError(object.RuntimeErrorKindValue, "argument 2 to `repeat` must be nonnegative")
	}
	if value == "" || count.Value == 0 {
		return &object.String{Value: ""}
	}
	const maximumStringBytes = 1_000_000
	if count.Value > int64(math.MaxInt/len(value)) || count.Value > maximumStringBytes/int64(len(value)) {
		return newError(object.RuntimeErrorKindValue, "result of `repeat` exceeds %d bytes", maximumStringBytes)
	}
	return &object.String{Value: strings.Repeat(value, int(count.Value))}
}

func stringReplace(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 3); err != nil {
		return err
	}
	value, err := requireString("replace", 0, args[0])
	if err != nil {
		return err
	}
	oldValue, err := requireString("replace", 1, args[1])
	if err != nil {
		return err
	}
	newValue, err := requireString("replace", 2, args[2])
	if err != nil {
		return err
	}
	return &object.String{Value: strings.ReplaceAll(value, oldValue, newValue)}
}

func stringSplit(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	value, err := requireString("split", 0, args[0])
	if err != nil {
		return err
	}
	separator, err := requireString("split", 1, args[1])
	if err != nil {
		return err
	}
	if separator == "" {
		return newError(object.RuntimeErrorKindValue, "separator argument to `split` must not be empty")
	}
	return stringArray(strings.Split(value, separator))
}

func stringSplitLines(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireString("splitlines", 0, args[0])
	if err != nil {
		return err
	}
	if value == "" {
		return stringArray(nil)
	}
	normalized := strings.ReplaceAll(strings.ReplaceAll(value, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(normalized, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return stringArray(lines)
}

func stringArray(values []string) object.Object {
	elements := make([]object.Object, len(values))
	for index, value := range values {
		elements[index] = &object.String{Value: value}
	}
	return &object.Array{Elements: elements}
}

func reverseString(value string) string {
	runes := []rune(value)
	for left, right := 0, len(runes)-1; left < right; left, right = left+1, right-1 {
		runes[left], runes[right] = runes[right], runes[left]
	}
	return string(runes)
}

func swapCase(value string) string {
	return strings.Map(func(character rune) rune {
		switch {
		case unicode.IsUpper(character):
			return unicode.ToLower(character)
		case unicode.IsLower(character):
			return unicode.ToUpper(character)
		default:
			return character
		}
	}, value)
}

func titleString(value string) string {
	startOfWord := true
	return strings.Map(func(character rune) rune {
		if !unicode.IsLetter(character) {
			startOfWord = true
			return character
		}
		if startOfWord {
			startOfWord = false
			return unicode.ToTitle(character)
		}
		return unicode.ToLower(character)
	}, value)
}

func isAlphanumeric(value string) bool {
	return allRunes(value, func(character rune) bool { return unicode.IsLetter(character) || unicode.IsNumber(character) }, false)
}

func isAlpha(value string) bool {
	return allRunes(value, unicode.IsLetter, false)
}

func isASCII(value string) bool {
	for index := 0; index < len(value); index++ {
		if value[index] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

func isDecimal(value string) bool {
	return allRunes(value, unicode.IsDigit, false)
}

func isDigitString(value string) bool {
	return allRunes(value, unicode.IsDigit, false)
}

func isNumeric(value string) bool {
	return allRunes(value, unicode.IsNumber, false)
}

func isPrintable(value string) bool {
	return allRunes(value, unicode.IsPrint, true)
}

func isSpace(value string) bool {
	return allRunes(value, unicode.IsSpace, false)
}

func allRunes(value string, predicate func(rune) bool, emptyResult bool) bool {
	if value == "" {
		return emptyResult
	}
	for _, character := range value {
		if !predicate(character) {
			return false
		}
	}
	return true
}

func isLower(value string) bool {
	hasCased := false
	for _, character := range value {
		if unicode.IsUpper(character) || unicode.IsTitle(character) {
			return false
		}
		if unicode.IsLower(character) {
			hasCased = true
		}
	}
	return hasCased
}

func isUpper(value string) bool {
	hasCased := false
	for _, character := range value {
		if unicode.IsLower(character) {
			return false
		}
		if unicode.IsUpper(character) || unicode.IsTitle(character) {
			hasCased = true
		}
	}
	return hasCased
}

func isTitle(value string) bool {
	hasCased := false
	startOfWord := true
	for _, character := range value {
		switch {
		case unicode.IsUpper(character) || unicode.IsTitle(character):
			if !startOfWord {
				return false
			}
			hasCased = true
			startOfWord = false
		case unicode.IsLower(character):
			if startOfWord {
				return false
			}
			hasCased = true
		default:
			startOfWord = true
		}
	}
	return hasCased
}
