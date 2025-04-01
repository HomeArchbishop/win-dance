# Win Dance

[en] | [[zh-CN](./README.md)]

Compile a video into a program that plays the video on the desktop using multiple bouncing windows.

![demo.gif](docs/demo.gif)

## Requirements

- Platform: Windows

- `gcc` (specifically `g++` command)

- `ffmpeg`

## Quick start

- Download latest `win-dance.exe` from [Releases](https://github.com/homearchbishop/win-dance/releases)

- Prepare a video file, e.g. `badapple.mp4`

- Open terminal, `cd` to your workspace, run:

```sh
win-dance ./badapple.mp4 -o badapple
```

When done, you get a `badapple.exe` under current directory. Execute it and enjoy!

## Usage

```txt
win-dance INPUT_VIDEO --output OUTPUT_EXE [--workdir=WORK_DIR] \
  [--resolution-vertical=RESOLUTION_VER] [--fps=FPS]

  --output                -o  (required)    Output executable file
  --workdir               -w  (optional)    Directory for temp file storage, \
                                            default to windance-working-directory
  --resolution-vertical   -r  (optional)    Vertical resolution, default to 30
  --fps                   -f  (optional)    Frame rate, default to 12
  --keep-workdir          -k  (optional)    If keep workdir, default to false
```

For example, command as below will compile `video/badapple.mp4` into a executable file `./badapple.exe` with fps=40 and vertical-resolution=120, using `temp` as temporary directory.

```sh
win-dance video/badapple.mp4 -o ./badapple.exe -f 40 -r 120 -w temp
```

## License

[MIT License](./LICENSE)
