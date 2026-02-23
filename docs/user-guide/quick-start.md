# Quick Start

## Download a Video

```bash
ytdl-go https://www.youtube.com/watch?v=BaW_jenozKc
```

Downloads the best available quality to the current directory.

## Download Audio Only

```bash
ytdl-go -audio https://www.youtube.com/watch?v=BaW_jenozKc
```

## Browse Available Formats

```bash
ytdl-go -list-formats https://www.youtube.com/watch?v=BaW_jenozKc
```

Use the interactive TUI to browse and select a format:

- `↑/↓` to navigate
- `Enter` to download
- `q` or `Esc` to quit

## Get Video Metadata

```bash
ytdl-go -info https://www.youtube.com/watch?v=BaW_jenozKc
```

Prints metadata as JSON without downloading.

## Download a Playlist

```bash
ytdl-go "https://youtube.com/playlist?list=..."
```

## Launch the Web UI

```bash
ytdl-go -web
```

Opens a browser-based interface for downloading and managing media.

## Common Options

| Goal | Command |
| --- | --- |
| Best video | `ytdl-go URL` |
| Audio only | `ytdl-go -audio URL` |
| Specific quality | `ytdl-go -quality 720p URL` |
| Specific format | `ytdl-go -format mp4 URL` |
| Specific itag | `ytdl-go -itag 251 URL` |
| Custom output path | `ytdl-go -o "downloads/{title}.{ext}" URL` |
| JSON output | `ytdl-go -json URL` |
| Quiet mode | `ytdl-go -quiet URL` |

See the full [CLI Options](../reference/cli-options.md) reference for all flags.
