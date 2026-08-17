# `terminal`

`terminal` provides terminal sizing, styled output, cursor control, alternate-screen buffers, and immediate keyboard
input. Control sequences require an ANSI/VT-compatible terminal; run terminal applications in a real terminal rather
than an IDE Output or Debug Console.

```silver
let terminal = import("terminal")

terminal.alternate_screen()
defer terminal.main_screen()
terminal.cursor_hide()
defer terminal.cursor_show()
terminal.raw_mode()
defer terminal.normal_mode()

terminal.clear()
terminal.cursor_move(2, 1)
terminal.style("Silver", terminal.Style.Bold)
terminal.colored(" terminal", "#7DD3FC")
terminal.flush()

let key = terminal.read_key()
```

Coordinates are zero-based `(x, y)`. Output functions do not append a newline. Operations that need an attached
terminal, such as `raw_mode`, fail with a catchable `RuntimeError` when input or output is redirected. `width` and
`height` use the positive `COLUMNS` and `LINES` environment values when output is not an attached terminal.

| Function | Description |
| --- | --- |
| `width() int` | Return the current width in columns. |
| `height() int` | Return the current height in rows. |
| `flush()` | Flush buffered output when the configured writer supports flushing. |
| `clear()` | Erase the visible screen and move the cursor to `(0, 0)`. |
| `colored(text: str, color: str)` | Print text with a `#RRGGBB` true-color foreground, then restore the foreground. |
| `style(text: str, value: Style)` | Print text with one style, then reset styling. |
| `cursor_move(x: int, y: int)` | Move to an absolute zero-based position. |
| `cursor_hide()` / `cursor_show()` | Hide or show the cursor. |
| `cursor_save() array` | Query the cursor and return its zero-based position as `[x, y]`. |
| `read_key() str` | Read one UTF-8 keypress without waiting for Enter. |
| `raw_mode()` | Switch input to raw mode. Pair it with `defer terminal.normal_mode()`. |
| `normal_mode()` | Restore normal line-buffered input. |
| `clear_line()` | Erase from the cursor through the end of the line. |
| `alternate_screen()` | Switch to the alternate screen buffer. |
| `main_screen()` | Return to the main screen buffer. |

`Style` contains `Normal`, `Bold`, `Underline`, `Italic`, and `Dim`.

`cursor_save()` sends a terminal cursor-position query and waits for its response. It should be called only when stdin
and stdout refer to the same interactive terminal.

[Standard library index](../table_of_contents.md#standard-library)
