# Template Strings

Template strings combine literal text with delayed Silver expressions. They are useful for queries, log formats,
messages, and other text that should be rendered later or more than once.

## Creating and evaluating templates

Triple backticks create a `TemplateString`, not a `str`. Call `.eval()` to render it:

````silver
let table = "users"
let minimum = 21
let query: TemplateString = ```SELECT * FROM {table} WHERE age >= {minimum}```

let rendered: str = query.eval()
````

Template text is raw and may span physical lines. Newlines and other text between the delimiters are preserved. Unlike
ordinary quoted strings, backslash escape sequences in literal template text are not decoded.

Template strings are also Silver's multi-line string type by evaluating them.
````silver
let multi_line: TemplateString = ```
hello from
multiple lines
```

io.print(multi_line.eval())

>> helo from
>> multiple lines
````

## Interpolations are Silver expressions

Text between a single pair of braces is parsed as an ordinary expression:

````silver
let maps = import("map")
let template = ```sum={1 + 2}; answer={maps.get({"answer": 42}, "answer")}```

template.eval() # "sum=3; answer=42"
````

Calls, operators, indexing, member access, maps, and nested delimiters can all appear inside an interpolation. The
rendered form uses the resulting value's normal Silver inspection text.

A template can even evaluate another template:

````silver
let nested = ```outer {```inner```.eval()}```
nested.eval() # "outer inner"
````

## Delayed evaluation and captured scope

Creating a template does not evaluate its interpolations. The template captures the lexical environment in which it is
declared and reads the current values of those bindings each time `.eval()` runs:

````silver
let value = 1
let template = ```value={value}```

value = 2
template.eval() # "value=2"
````

Templates returned from functions retain that function's scope:

````silver
let make_label = fn(value: str) TemplateString {
    ```captured {value}```
}

let label = make_label("scope")
label.eval() # "captured scope"
````

Every call reevaluates every interpolation, including side effects:

````silver
let count = 0
let next = fn() int {
    count = count + 1
    count
}

let value = ```call {next()}```
value.eval() # "call 1"
value.eval() # "call 2"
````

## Literal braces

Double braces produce literal brace characters in template text:

````silver
let value = "inside"
let template = ```{{literal}} {value}```

template.eval() # "{literal} inside"
````

An unmatched closing brace is a syntax error. Braces belonging to an interpolation's nested Silver expression, such as
a map literal, are tracked normally and do not end the interpolation early.

## Type and callable contract

`TemplateString` is a built-in nominal struct type with one method:

| Method | Description |
| ------ | ----------- |
| `eval() str` | Evaluate every interpolation and return the rendered string. |

The method accepts no arguments and can be stored under the detailed callable type `call() str`:

```silver
let template: TemplateString = ```hello```
let render: call() str = template.eval
```

Use an ordinary quoted string when text should be computed immediately or does not need delayed interpolation.

## Errors

Syntax inside interpolation braces is checked when the source file is parsed, but names, types, calls, and other runtime
behavior are not evaluated until `.eval()`. A runtime failure therefore appears at render time:

````silver
let template = ```value={missing}```

let rendered = try {
    template.eval()
} catch NameError err {
    "could not render: " + err.message
}
````

See [Errors](errors.md) for handling delayed failures and [Functions](functions.md#closures) for the lexical-capture
model shared by closures and templates.

[Language guide](language_guide.md) | [Documentation index](../table_of_contents.md)
