# Interactive Format Selection

When using the `--list-formats` flag, ytdl-go now provides an interactive format selector that allows you to:

1. **Navigate formats** using arrow keys (↑/↓) or vim keys (j/k)
2. **Quick jump** by pressing digits 1-9 to jump to formats with itags starting with that digit
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
# - Press Enter to download the selected format
# - Press 'b' to go back without downloading
# - Press 'q' to quit
```

## Features

- **Visual highlighting**: Selected format is highlighted in cyan
- **Real-time feedback**: Shows the currently selected itag number
- **Quick selection**: Press digits 1-9 to quickly jump to formats
- **Smooth scrolling**: Supports mouse wheel and keyboard scrolling
- **Format details**: Shows itag, extension, quality, size, audio channels, and video resolution

The interactive selector replaces the previous read-only pager, making it much easier to select and download specific formats without having to remember itag numbers.
