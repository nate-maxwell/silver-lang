package stdlib

import (
	"fmt"
	"math"
	"silver/ast"
	"silver/object"
	"strconv"
	"strings"
	stdtime "time"
)

// timeDefinitions contains the Go implementations exported by import("time").
// Time and Duration are module-local nominal types, rather than maps, so their
// fields can be used with normal Silver member access and type annotations.
func timeDefinitions(null *object.Null, trueValue, falseValue *object.Boolean) []definition {
	timeType, durationType := newTimeStructDefinitions()
	return []definition{
		{name: "Time", value: timeType},
		{name: "Duration", value: durationType},
		{name: "now", fn: timeNow(timeType), signature: callSignature(nil, nil, namedType("Time"))},
		{name: "unix", fn: timeUnix, signature: callSignature(nil, nil, namedType("int"))},
		{name: "format", fn: timeFormat(timeType), signature: callSignature([]string{"value", "format"}, []*ast.TypeAnnotation{namedType("Time"), namedType("str")}, namedType("str"))},
		{name: "parse", fn: timeParse(timeType), signature: callSignature([]string{"value", "format"}, []*ast.TypeAnnotation{namedType("str"), namedType("str")}, namedType("Time"))},
		// duration accepts both INTEGER and FLOAT values, so it intentionally has
		// no narrower callable signature than the runtime checks below.
		{name: "duration", fn: timeDuration(durationType)},
		{name: "add", fn: timeAdd(timeType, durationType), signature: callSignature([]string{"value", "duration"}, []*ast.TypeAnnotation{namedType("Time"), namedType("Duration")}, namedType("Time"))},
		{name: "diff", fn: timeDiff(timeType, durationType), signature: callSignature([]string{"from", "to"}, []*ast.TypeAnnotation{namedType("Time"), namedType("Time")}, namedType("Duration"))},
		{name: "sleep", fn: timeSleep(durationType, null), signature: callSignature([]string{"duration"}, []*ast.TypeAnnotation{namedType("Duration")}, nil)},
		{name: "before", fn: timeCompare(timeType, trueValue, falseValue, "before", func(left, right stdtime.Time) bool { return left.Before(right) })},
		{name: "after", fn: timeCompare(timeType, trueValue, falseValue, "after", func(left, right stdtime.Time) bool { return left.After(right) })},
		{name: "equal", fn: timeCompare(timeType, trueValue, falseValue, "equal", func(left, right stdtime.Time) bool { return left.Equal(right) })},
	}
}

func newTimeStructDefinitions() (*object.Struct, *object.Struct) {
	environment := object.NewEnvironment()
	timeType := &object.Struct{
		Name:   "Time",
		Fields: []string{"year", "month", "day", "hour", "minute", "second", "nanosecond", "timezone"},
		FieldTypes: []*ast.TypeAnnotation{
			namedType("int"), namedType("int"), namedType("int"), namedType("int"),
			namedType("int"), namedType("int"), namedType("int"), namedType("str"),
		},
		Env: environment,
	}
	durationType := &object.Struct{
		Name:   "Duration",
		Fields: []string{"hours", "minutes", "seconds", "milliseconds", "nanoseconds"},
		FieldTypes: []*ast.TypeAnnotation{
			namedType("float"), namedType("float"), namedType("float"), namedType("float"), namedType("int"),
		},
		Env: environment,
	}
	environment.Set("Time", timeType)
	environment.Set("Duration", durationType)
	return timeType, durationType
}

func timeNow(timeType *object.Struct) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 0); err != nil {
			return err
		}
		return newTimeValue(timeType, stdtime.Now())
	}
}

func timeUnix(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	return &object.Integer{Value: stdtime.Now().Unix()}
}

func timeFormat(timeType *object.Struct) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		value, err := requireTime("format", 0, args[0], timeType)
		if err != nil {
			return err
		}
		format, err := requireString("format", 1, args[1])
		if err != nil {
			return err
		}
		return &object.String{Value: formatTime(value, format)}
	}
}

func timeParse(timeType *object.Struct) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		value, err := requireString("parse", 0, args[0])
		if err != nil {
			return err
		}
		format, err := requireString("parse", 1, args[1])
		if err != nil {
			return err
		}
		layout, layoutErr := timeLayout(format)
		if layoutErr != nil {
			return newError(object.RuntimeErrorKindValue, "invalid time format %q: %s", format, layoutErr)
		}
		parsed, parseErr := stdtime.Parse(layout, value)
		if parseErr != nil {
			return newError(object.RuntimeErrorKindValue, "could not parse %q with format %q: %s", value, format, parseErr)
		}
		return newTimeValue(timeType, parsed)
	}
}

func timeDuration(durationType *object.Struct) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		unit, err := requireString("duration", 1, args[1])
		if err != nil {
			return err
		}
		scale, ok := durationUnit(unit)
		if !ok {
			return newError(object.RuntimeErrorKindValue, "unknown duration unit %q", unit)
		}
		nanoseconds, numberErr := durationNanoseconds(args[0], scale)
		if numberErr != nil {
			return numberErr
		}
		return newDurationValue(durationType, stdtime.Duration(nanoseconds))
	}
}

func durationUnit(unit string) (int64, bool) {
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "nanosecond", "nanoseconds", "ns":
		return int64(stdtime.Nanosecond), true
	case "microsecond", "microseconds", "us":
		return int64(stdtime.Microsecond), true
	case "millisecond", "milliseconds", "ms":
		return int64(stdtime.Millisecond), true
	case "second", "seconds", "s":
		return int64(stdtime.Second), true
	case "minute", "minutes", "m":
		return int64(stdtime.Minute), true
	case "hour", "hours", "h":
		return int64(stdtime.Hour), true
	case "day", "days", "d":
		return int64(24 * stdtime.Hour), true
	default:
		return 0, false
	}
}

func durationNanoseconds(value object.Object, scale int64) (int64, *object.Error) {
	switch value := value.(type) {
	case *object.Integer:
		if value.Value > 0 && value.Value > math.MaxInt64/scale || value.Value < 0 && value.Value < math.MinInt64/scale {
			return 0, newError(object.RuntimeErrorKindValue, "duration is out of range")
		}
		return value.Value * scale, nil
	case *object.Float:
		nanoseconds := value.Value * float64(scale)
		if math.IsNaN(nanoseconds) || math.IsInf(nanoseconds, 0) || nanoseconds >= float64(math.MaxInt64) || nanoseconds < float64(math.MinInt64) {
			return 0, newError(object.RuntimeErrorKindValue, "duration is out of range")
		}
		return int64(math.Round(nanoseconds)), nil
	default:
		return 0, newError(object.RuntimeErrorKindType, "argument 1 to `duration` must be INTEGER or FLOAT, got %s", value.Type())
	}
}

func timeAdd(timeType, durationType *object.Struct) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		value, err := requireTime("add", 0, args[0], timeType)
		if err != nil {
			return err
		}
		duration, err := requireDuration("add", 1, args[1], durationType)
		if err != nil {
			return err
		}
		return newTimeValue(timeType, value.Add(duration))
	}
}

func timeDiff(timeType, durationType *object.Struct) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		from, err := requireTime("diff", 0, args[0], timeType)
		if err != nil {
			return err
		}
		to, err := requireTime("diff", 1, args[1], timeType)
		if err != nil {
			return err
		}
		return newDurationValue(durationType, to.Sub(from))
	}
}

func timeSleep(durationType *object.Struct, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		duration, err := requireDuration("sleep", 0, args[0], durationType)
		if err != nil {
			return err
		}
		stdtime.Sleep(duration)
		return null
	}
}

func timeCompare(timeType *object.Struct, trueValue, falseValue *object.Boolean, name string, compare func(stdtime.Time, stdtime.Time) bool) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		left, err := requireTime(name, 0, args[0], timeType)
		if err != nil {
			return err
		}
		right, err := requireTime(name, 1, args[1], timeType)
		if err != nil {
			return err
		}
		if compare(left, right) {
			return trueValue
		}
		return falseValue
	}
}

func newTimeValue(timeType *object.Struct, value stdtime.Time) *object.StructInstance {
	zone, offset := value.Zone()
	if zone == "" {
		zone = formatTimezoneOffset(offset, true)
	}
	return &object.StructInstance{Struct: timeType, Values: map[string]object.Object{
		"year":         &object.Integer{Value: int64(value.Year())},
		"month":        &object.Integer{Value: int64(value.Month())},
		"day":          &object.Integer{Value: int64(value.Day())},
		"hour":         &object.Integer{Value: int64(value.Hour())},
		"minute":       &object.Integer{Value: int64(value.Minute())},
		"second":       &object.Integer{Value: int64(value.Second())},
		"nanosecond":   &object.Integer{Value: int64(value.Nanosecond())},
		"timezone":     &object.String{Value: zone},
		"_timezone":    &object.String{Value: zone},
		"_location":    &object.String{Value: value.Location().String()},
		"_zone_offset": &object.Integer{Value: int64(offset)},
	}}
}

func newDurationValue(durationType *object.Struct, value stdtime.Duration) *object.StructInstance {
	nanoseconds := int64(value)
	return &object.StructInstance{Struct: durationType, Values: map[string]object.Object{
		"hours":        &object.Float{Value: float64(value) / float64(stdtime.Hour)},
		"minutes":      &object.Float{Value: float64(value) / float64(stdtime.Minute)},
		"seconds":      &object.Float{Value: float64(value) / float64(stdtime.Second)},
		"milliseconds": &object.Float{Value: float64(value) / float64(stdtime.Millisecond)},
		"nanoseconds":  &object.Integer{Value: nanoseconds},
	}}
}

func requireTime(name string, index int, value object.Object, timeType *object.Struct) (stdtime.Time, *object.Error) {
	instance, ok := value.(*object.StructInstance)
	if !ok || instance.Struct != timeType {
		return stdtime.Time{}, newError(object.RuntimeErrorKindType, "argument %d to `%s` must be Time, got %s", index+1, name, value.Type())
	}

	field := func(fieldName string) (int, *object.Error) {
		fieldValue, exists := instance.Get(fieldName)
		integer, valid := fieldValue.(*object.Integer)
		if !exists || !valid {
			return 0, newError(object.RuntimeErrorKindValue, "Time field %q is invalid", fieldName)
		}
		if integer.Value < math.MinInt32 || integer.Value > math.MaxInt32 {
			return 0, newError(object.RuntimeErrorKindValue, "Time field %q is out of range", fieldName)
		}
		return int(integer.Value), nil
	}
	year, err := field("year")
	if err != nil {
		return stdtime.Time{}, err
	}
	month, err := field("month")
	if err != nil {
		return stdtime.Time{}, err
	}
	day, err := field("day")
	if err != nil {
		return stdtime.Time{}, err
	}
	hour, err := field("hour")
	if err != nil {
		return stdtime.Time{}, err
	}
	minute, err := field("minute")
	if err != nil {
		return stdtime.Time{}, err
	}
	second, err := field("second")
	if err != nil {
		return stdtime.Time{}, err
	}
	nanosecond, err := field("nanosecond")
	if err != nil {
		return stdtime.Time{}, err
	}
	timezoneValue, exists := instance.Get("timezone")
	timezone, valid := timezoneValue.(*object.String)
	if !exists || !valid || timezone.Value == "" {
		return stdtime.Time{}, newError(object.RuntimeErrorKindValue, "Time field %q is invalid", "timezone")
	}
	location, locationErr := timeLocation(instance, timezone.Value)
	if locationErr != nil {
		return stdtime.Time{}, newError(object.RuntimeErrorKindValue, "invalid timezone %q", timezone.Value)
	}
	if month < 1 || month > 12 || day < 1 || day > 31 || hour < 0 || hour > 23 || minute < 0 || minute > 59 || second < 0 || second > 59 || nanosecond < 0 || nanosecond >= int(stdtime.Second) {
		return stdtime.Time{}, newError(object.RuntimeErrorKindValue, "Time has an invalid date or clock value")
	}
	result := stdtime.Date(year, stdtime.Month(month), day, hour, minute, second, nanosecond, location)
	if result.Year() != year || int(result.Month()) != month || result.Day() != day || result.Hour() != hour || result.Minute() != minute || result.Second() != second || result.Nanosecond() != nanosecond {
		return stdtime.Time{}, newError(object.RuntimeErrorKindValue, "Time has an invalid date or clock value")
	}
	return result, nil
}

func timeLocation(instance *object.StructInstance, timezone string) (*stdtime.Location, error) {
	storedTimezone, timezoneOK := instance.Get("_timezone")
	storedName, nameOK := storedTimezone.(*object.String)
	if timezoneOK && nameOK && storedName.Value == timezone {
		storedLocation, locationOK := instance.Get("_location")
		locationName, valid := storedLocation.(*object.String)
		if locationOK && valid && locationName.Value != "" {
			if location, err := stdtime.LoadLocation(locationName.Value); err == nil {
				return location, nil
			}
		}
		storedOffset, offsetOK := instance.Get("_zone_offset")
		if offset, valid := storedOffset.(*object.Integer); offsetOK && valid {
			return stdtime.FixedZone(timezone, int(offset.Value)), nil
		}
	}
	if location, err := stdtime.LoadLocation(timezone); err == nil {
		return location, nil
	}
	if offset, ok := parseTimezoneOffset(timezone); ok {
		return stdtime.FixedZone(timezone, offset), nil
	}
	return nil, fmt.Errorf("unknown timezone")
}

func requireDuration(name string, index int, value object.Object, durationType *object.Struct) (stdtime.Duration, *object.Error) {
	instance, ok := value.(*object.StructInstance)
	if !ok || instance.Struct != durationType {
		return 0, newError(object.RuntimeErrorKindType, "argument %d to `%s` must be Duration, got %s", index+1, name, value.Type())
	}
	value, exists := instance.Get("nanoseconds")
	nanoseconds, ok := value.(*object.Integer)
	if !exists || !ok {
		return 0, newError(object.RuntimeErrorKindValue, "Duration field %q is invalid", "nanoseconds")
	}
	return stdtime.Duration(nanoseconds.Value), nil
}

type timeFormatToken struct {
	token  string
	layout string
}

var timeFormatTokens = []timeFormatToken{
	{token: "SSSSSSSSS", layout: "000000000"},
	{token: "SSSSSS", layout: "000000"},
	{token: "YYYY", layout: "2006"},
	{token: "SSS", layout: "000"},
	{token: "ZZ", layout: "Z0700"},
	{token: "YY", layout: "06"},
	{token: "MM", layout: "01"},
	{token: "DD", layout: "02"},
	{token: "HH", layout: "15"},
	{token: "mm", layout: "04"},
	{token: "ss", layout: "05"},
	{token: "Z", layout: "Z07:00"},
	{token: "z", layout: "MST"},
}

func timeLayout(format string) (string, error) {
	var layout strings.Builder
	for index := 0; index < len(format); {
		if format[index] == '[' {
			end := strings.IndexByte(format[index+1:], ']')
			if end < 0 {
				return "", fmt.Errorf("unclosed '['")
			}
			end += index + 1
			layout.WriteString(format[index+1 : end])
			index = end + 1
			continue
		}
		matched := false
		for _, candidate := range timeFormatTokens {
			if strings.HasPrefix(format[index:], candidate.token) {
				layout.WriteString(candidate.layout)
				index += len(candidate.token)
				matched = true
				break
			}
		}
		if !matched {
			layout.WriteByte(format[index])
			index++
		}
	}
	return layout.String(), nil
}

func formatTime(value stdtime.Time, format string) string {
	zone, offset := value.Zone()
	values := map[string]string{
		"YYYY":      fmt.Sprintf("%04d", value.Year()),
		"YY":        fmt.Sprintf("%02d", value.Year()%100),
		"MM":        fmt.Sprintf("%02d", value.Month()),
		"DD":        fmt.Sprintf("%02d", value.Day()),
		"HH":        fmt.Sprintf("%02d", value.Hour()),
		"mm":        fmt.Sprintf("%02d", value.Minute()),
		"ss":        fmt.Sprintf("%02d", value.Second()),
		"SSS":       fmt.Sprintf("%03d", value.Nanosecond()/1_000_000),
		"SSSSSS":    fmt.Sprintf("%06d", value.Nanosecond()/1_000),
		"SSSSSSSSS": fmt.Sprintf("%09d", value.Nanosecond()),
		"Z":         formatTimezoneOffset(offset, true),
		"ZZ":        formatTimezoneOffset(offset, false),
		"z":         zone,
	}
	var result strings.Builder
	for index := 0; index < len(format); {
		if format[index] == '[' {
			if relativeEnd := strings.IndexByte(format[index+1:], ']'); relativeEnd >= 0 {
				end := index + 1 + relativeEnd
				result.WriteString(format[index+1 : end])
				index = end + 1
				continue
			}
		}
		matched := false
		for _, candidate := range timeFormatTokens {
			if strings.HasPrefix(format[index:], candidate.token) {
				result.WriteString(values[candidate.token])
				index += len(candidate.token)
				matched = true
				break
			}
		}
		if !matched {
			result.WriteByte(format[index])
			index++
		}
	}
	return result.String()
}

func formatTimezoneOffset(offset int, colon bool) string {
	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}
	hours := offset / 3600
	minutes := offset % 3600 / 60
	if colon {
		return fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
	}
	return fmt.Sprintf("%s%02d%02d", sign, hours, minutes)
}

func parseTimezoneOffset(value string) (int, bool) {
	if value == "Z" {
		return 0, true
	}
	if len(value) != 5 && len(value) != 6 || value[0] != '+' && value[0] != '-' {
		return 0, false
	}
	digits := value[1:]
	if len(value) == 6 {
		if value[3] != ':' {
			return 0, false
		}
		digits = value[1:3] + value[4:]
	}
	hours, errHours := strconv.Atoi(digits[:2])
	minutes, errMinutes := strconv.Atoi(digits[2:])
	if errHours != nil || errMinutes != nil || hours > 23 || minutes > 59 {
		return 0, false
	}
	offset := hours*3600 + minutes*60
	if value[0] == '-' {
		offset = -offset
	}
	return offset, true
}
