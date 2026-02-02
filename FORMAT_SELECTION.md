# Interactive Format Selection

When using the `--list-formats` flag, ytdl-go now provides an interactive format selector that allows you to:

1. **Navigate formats** using arrow keys (↑/↓) or vim keys (j/k)
2. **Quick selection** by typing digits 0-9 to select formats by itag number
   - Type multiple digits to match exact itags (e.g., type "101" to select itag 101)
   - Press the same digit repeatedly within 500ms to cycle through matches
   - Buffer resets after 1.5 seconds of inactivity
3. **Page navigation** with Page Up/Page Down (or just pgup/pgdown)
4. **Go to first/last** with Home/End (or g/G)
5. **Select a format** by highlighting it and pressing Enter
6. **Go back** by pressing 'b' to cancel selection
7. **Quit** by pressing 'q', 'esc', or 'ctrl+c'

## Usage

```bash
# List formats interactively
./ytdl-go --list-formats <URL>

# Once in the format selector:
# - Use ↑/↓ to highlight different formats
# - Type digits to select by itag (e.g., "101" for itag 101)
# - Press Enter to download the selected format
# - Press 'b' to go back without downloading
# - Press 'q' to quit
```

## Features

- **Visual highlighting**: Selected format is highlighted in cyan
- **Real-time feedback**: Shows the currently selected itag number
- **Multi-digit selection**: Type full itag numbers (e.g., "101") with visual feedback
- **Cycling**: Press the same digit quickly (within 500ms) to cycle through matches
- **Smart timeout**: Buffer resets after 1.5 seconds of inactivity
- **Smooth scrolling**: Supports mouse wheel and keyboard scrolling
- **Format details**: Shows itag, extension, quality, size, audio channels, and video resolution

## Selection Examples

- Type "1" → selects first format with itag starting with 1 (e.g., itag 18)
- Press "1" again quickly (within 500ms) → cycles to next itag starting with 1 (e.g., itag 140)
- Type "1", wait, then "0" → buffer becomes "10", selects first itag starting with 10
- Type "1", "0", "1" → buffer becomes "101", exact match selects itag 101
- Type "2", "2" slowly (pausing between) → buffer becomes "22", selects itag 22
- Wait 1.5 seconds → buffer automatically resets

The interactive selector replaces the previous read-only pager, making it much easier to select and download specific formats without having to remember itag numbers.
