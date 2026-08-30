# Silver Lexer

A typical 2 position lexer with normal look-ahead logic, save for template strings which uses a third position
char. Template strings start and end with triple backticks.

All other language features can be lexed as either two character tokens, single character tokens, or single
character token, or user-defined identifiers.
