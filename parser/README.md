# Silver Parser

The Silver parser is a recursive descent parser using Pratt parsing for operator precedence. Its
core contains the parser state, operator precedence table, and prefix and infix tables used for
recursive descent.

Each file in the parser package implements a specific part of the parser, such as parsing
expressions, statements, or declarations. They are implemented as methods of the parser object,
separated into individual files of each concept for better organization and maintainability.
