# FAQ

## What video sources are supported?

ytdl-go supports publicly accessible YouTube videos, playlists, and YouTube Music URLs. Direct URLs to video/audio files are also supported.

## Does ytdl-go support authentication?

No. ytdl-go only downloads publicly accessible content. Private, age-gated, or member-only videos are not supported.

## Do I need FFmpeg?

FFmpeg is optional but recommended. It enables the audio extraction fallback strategy when YouTube blocks direct audio-only downloads. Without it, some audio downloads may fail.

## How do I select a specific format?

Use `-list-formats` to browse available formats interactively, or use `-itag` to download a specific format by its YouTube itag number. See [Format Selection](format-selection.md).

## Can I download multiple videos at once?

Yes. Pass multiple URLs as arguments and use `-jobs N` to download them in parallel:

```bash
ytdl-go -jobs 4 URL1 URL2 URL3 URL4
```

## Where are downloaded files saved?

By default, files are saved to the current directory with the pattern `{title}.{ext}`. Use `-o` to customize. See [Output Templates](output-templates.md).

## How do I use the web UI?

Run `ytdl-go -web` to start the web server. It provides a browser-based interface for downloading and managing media.

## Does ytdl-go bypass DRM or encryption?

No. ytdl-go does not circumvent DRM, encryption, or any access controls. It only downloads publicly accessible content.

## How do I update ytdl-go?

```bash
go install github.com/lvcoi/ytdl-go@latest
```

## What exit codes does ytdl-go use?

Exit codes are categorized by error type. Use `-json` mode for structured error output with category information. See the [CLI Options](../reference/cli-options.md) reference.
