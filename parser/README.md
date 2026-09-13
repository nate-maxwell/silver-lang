# Silver Parser

The Silver parser is a recursive descent parser using Pratt parsing for operator precedence. Its
core contains the parser state, operator precedence table, and prefix and infix tables used for
recursive descent.

Each file in the parser package implements a specific part of the parser, such as parsing
expressions, statements, or declarations. They are implemented as methods of the parser object,
separated into individual files of each concept for better organization and maintainability.

`types.go` parses `type Name = ...` declarations as `ast.TypeStatement` nodes. Their right-hand sides are
struct or enum type bodies, or the same annotations accepted after `:`. `type` is a reserved declaration keyword;
the parser accepts it after `.` for the standard library's `core.type(value)` API. Type expressions have their own
AST interface and cannot appear as value expressions.
