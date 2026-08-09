package stdlib_test

import (
	"silver/object"
	"strings"
	"testing"
	stdtime "time"
)

const timeImport = `let times = import("time")
`

func TestTimeTypesAndFields(t *testing.T) {
	input := `let value: times.Time = times.parse("2024-01-15 13:14:15", "YYYY-MM-DD HH:mm:ss")
core.type(value) == times.Time`
	testBooleanObject(t, testEval(timeImport+coreImport+input), true)

	fields := map[string]int64{
		"year": 2024, "month": 1, "day": 15, "hour": 13,
		"minute": 14, "second": 15, "nanosecond": 0,
	}
	for field, want := range fields {
		t.Run(field, func(t *testing.T) {
			input := `let value = times.parse("2024-01-15 13:14:15", "YYYY-MM-DD HH:mm:ss")
value.` + field
			testIntegerObject(t, testEval(timeImport+input), want)
		})
	}
	result, ok := testEval(timeImport + `times.parse("2024-01-15", "YYYY-MM-DD").timezone`).(*object.String)
	if !ok || result.Value != "UTC" {
		t.Fatalf("timezone is %T (%v), want UTC", result, result)
	}
}

func TestTimeFormatAndParse(t *testing.T) {
	tests := []struct {
		value  string
		format string
	}{
		{value: "2024-01-15 13:14:15", format: "YYYY-MM-DD HH:mm:ss"},
		{value: "2024-01-15T13:14:15.123", format: "YYYY-MM-DD[T]HH:mm:ss.SSS"},
		{value: "2024-01-15T13:14:15.123456", format: "YYYY-MM-DD[T]HH:mm:ss.SSSSSS"},
		{value: "2024-01-15T13:14:15.123456789+02:30", format: "YYYY-MM-DD[T]HH:mm:ss.SSSSSSSSSZ"},
		{value: "2024-01-15T13:14:15+0230", format: "YYYY-MM-DD[T]HH:mm:ssZZ"},
	}
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			input := `let value = times.parse("` + tt.value + `", "` + tt.format + `")
times.format(value, "` + tt.format + `")`
			result, ok := testEval(timeImport + input).(*object.String)
			if !ok || result.Value != tt.value {
				t.Fatalf("round trip is %T (%v), want %q", result, result, tt.value)
			}
		})
	}
}

func TestTimeDurationAndArithmetic(t *testing.T) {
	input := `let value: times.Duration = times.duration(1.5, "seconds")
core.type(value) == times.Duration`
	testBooleanObject(t, testEval(timeImport+coreImport+input), true)

	integerFields := map[string]int64{"nanoseconds": 1_500_000_000}
	for field, want := range integerFields {
		testIntegerObject(t, testEval(timeImport+`times.duration(1.5, "seconds").`+field), want)
	}
	floatFields := map[string]float64{
		"hours": 1.5 / 3600, "minutes": 1.5 / 60, "seconds": 1.5, "milliseconds": 1500,
	}
	for field, want := range floatFields {
		testFloatObject(t, testEval(timeImport+`times.duration(1.5, "seconds").`+field), want)
	}

	input = `let start = times.parse("2024-01-15 00:00:00", "YYYY-MM-DD HH:mm:ss")
let later = times.add(start, times.duration(90, "minutes"))
times.format(later, "YYYY-MM-DD HH:mm:ss")`
	result, ok := testEval(timeImport + input).(*object.String)
	if !ok || result.Value != "2024-01-15 01:30:00" {
		t.Fatalf("add result is %T (%v), want 2024-01-15 01:30:00", result, result)
	}

	input = `let start = times.parse("2024-01-15 00:00:00", "YYYY-MM-DD HH:mm:ss")
let later = times.add(start, times.duration(90, "minutes"))
times.diff(start, later).seconds`
	testFloatObject(t, testEval(timeImport+input), 5400)
}

func TestTimeComparison(t *testing.T) {
	definitions := `let start = times.parse("2024-01-15T00:00:00Z", "YYYY-MM-DD[T]HH:mm:ssZ")
let same = times.parse("2024-01-15T01:00:00+01:00", "YYYY-MM-DD[T]HH:mm:ssZ")
let later = times.add(start, times.duration(1, "nanosecond"))
`
	for expression, want := range map[string]bool{
		"times.equal(start, same)":   true,
		"times.before(start, later)": true,
		"times.after(later, start)":  true,
		"times.equal(start, later)":  false,
	} {
		testBooleanObject(t, testEval(timeImport+definitions+expression), want)
	}
}

func TestTimeAddPreservesNamedTimezoneRules(t *testing.T) {
	input := `let start = times.Time(2024, 3, 10, 1, 30, 0, 0, "America/Los_Angeles")
let later = times.add(start, times.duration(1, "hour"))
times.format(later, "YYYY-MM-DD HH:mm:ss z")`
	result, ok := testEval(timeImport + input).(*object.String)
	if !ok || result.Value != "2024-03-10 03:30:00 PDT" {
		t.Fatalf("DST add result is %T (%v), want 2024-03-10 03:30:00 PDT", result, result)
	}
}

func TestTimeNowUnixAndSleep(t *testing.T) {
	before := stdtime.Now().Unix()
	result := testEval(timeImport + `times.unix()`)
	after := stdtime.Now().Unix()
	timestamp, ok := result.(*object.Integer)
	if !ok || timestamp.Value < before || timestamp.Value > after {
		t.Fatalf("unix result is %T (%v), want timestamp in [%d, %d]", result, result, before, after)
	}

	now, ok := testEval(timeImport + `times.now()`).(*object.StructInstance)
	if !ok || now.Struct.Name != "Time" {
		t.Fatalf("now result is %T (%v), want Time", now, now)
	}
	if timezone, exists := now.Get("timezone"); !exists || timezone.(*object.String).Value == "" {
		t.Fatal("now returned an empty timezone")
	}
	testNullObject(t, testEval(timeImport+`times.sleep(times.duration(0, "milliseconds"))`))
}

func TestTimeErrors(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: `times.duration("one", "seconds")`, message: "argument 1 to `duration` must be INTEGER or FLOAT, got STRING"},
		{input: `times.duration(1, "fortnights")`, message: `unknown duration unit "fortnights"`},
		{input: `times.format(1, "YYYY")`, message: "argument 1 to `format` must be Time, got INTEGER"},
		{input: `times.add(times.parse("2024-01-15", "YYYY-MM-DD"), 1)`, message: "argument 2 to `add` must be Duration, got INTEGER"},
		{input: `times.parse("not a date", "YYYY-MM-DD")`, message: "could not parse"},
		{input: `times.format(times.Time(2024, 2, 30, 0, 0, 0, 0, "UTC"), "YYYY-MM-DD")`, message: "Time has an invalid date or clock value"},
		{input: `times.format(times.Time(2024, 1, 15, 0, 0, 0, 0, "Nowhere/Unknown"), "YYYY-MM-DD")`, message: `invalid timezone "Nowhere/Unknown"`},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, ok := testEval(timeImport + tt.input).(*object.Error)
			if !ok {
				t.Fatalf("result is %T (%v), want error", result, result)
			}
			if !strings.Contains(result.MessageText(), tt.message) {
				t.Fatalf("error is %q, want it to contain %q", result.MessageText(), tt.message)
			}
		})
	}
}
