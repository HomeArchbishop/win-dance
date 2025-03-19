# Win Dance - 我家的窗口会后空翻！

[[en](./README-en.md)] | [zh-CN]

将一段视频编译成一个程序：通过在桌面上弹出大量窗口来播放视频。

![demo.gif](docs/demo.gif)

## 要求

- Windows 平台

- `gcc` (`g++`)

- `ffmpeg`

## 快速开始

- 从 [Releases](https://github.com/homearchbishop/win-dance/releases) 下载最新的 `win-dance.exe`

- 准备一个视频文件，例如 `badapple.mp4`

- 打开终端，`cd` 到你的工作目录，运行：

```sh
win-dance ./badapple.mp4 -o badapple
```

完成后，你会在当前目录下得到一个 `badapple.exe` 文件。运行它，观看 Bad Apple！

## 用法

```txt
win-dance INPUT_VIDEO --output OUTPUT_EXE [--workdir=WORK_DIR] \
  [--resolution-vertical=RESOLUTION_VER] [--fps=FPS] [--keep-workdir]

  --output                -o  (必需)        输出可执行文件
  --workdir               -w  (可选)        临时文件存储目录，默认为 windance-working-directory
  --resolution-vertical   -r  (可选)        垂直分辨率，默认为 30
  --fps                   -f  (可选)        帧率，默认为 12
  --keep-workdir          -k  (可选)        保留临时文件存储目录，默认为 false
```

例如，下面的命令将 `video/badapple.mp4` 编译成一个名为 `./badapple.exe` 的可执行文件，帧率为 40，垂直分辨率为 120，并使用 `temp` 作为临时目录。

```sh
win-dance video/badapple.mp4 -o ./badapple.exe -f 40 -r 120 -w temp
```
