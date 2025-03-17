package internal

import (
	"os"
	"strconv"
	"strings"
)

func GenerateSrc(screenWndRectsQueue []*[]*Rect, fps, frameWidth, frameHeight int) (string, error) {
	useWindowCnt := 0
	for _, rects := range screenWndRectsQueue {
		if len(*rects) > useWindowCnt {
			useWindowCnt = len(*rects)
		}
	}

	var stepCppArray strings.Builder
	stepCppArray.WriteString("{")
	for _, rects := range screenWndRectsQueue {
		stepCppArray.WriteString("{")
		for i, rect := range *rects {
			stepCppArray.WriteString("{" + strconv.Itoa(rect.X) + "," + strconv.Itoa(rect.Y) + "," + strconv.Itoa(rect.W) + "," + strconv.Itoa(rect.H) + "}")
			if i < len(*rects)-1 {
				stepCppArray.WriteString(",")
			}
		}
		stepCppArray.WriteString("}")
		if rects != screenWndRectsQueue[len(screenWndRectsQueue)-1] {
			stepCppArray.WriteString(",")
		}
	}
	stepCppArray.WriteString("};")

	src := `
#undef UNICODE
#undef _UNICODE
#include <windows.h>
#include <cstdio>
#include <functional>
#include <map>

#define WD_USE_WINDOW_CNT ` + strconv.Itoa(useWindowCnt) + `
#define WD_FPS ` + strconv.Itoa(fps) + `
#define WD_TOTAL_FRAME_CNT ` + strconv.Itoa(len(screenWndRectsQueue)) + `
#define WD_FRAME_WIDTH ` + strconv.Itoa(frameWidth) + `
#define WD_FRAME_HEIGHT ` + strconv.Itoa(frameHeight) + `

int steps[WD_TOTAL_FRAME_CNT][WD_USE_WINDOW_CNT][4] = ` + stepCppArray.String() + `

// Adapt steps' frames to the window size
void adaptFramesToWindowSize() {
  int screenWidth = GetSystemMetrics(SM_CXSCREEN);
  int screenHeight = GetSystemMetrics(SM_CYSCREEN);
  int videoWidth = WD_FRAME_WIDTH; // video resolution width
  int videoHeight = WD_FRAME_HEIGHT; // video resolution height
  int baseOffsetX, baseOffsetY, framePixelSize;
  if (static_cast<double>(screenWidth) / screenHeight > static_cast<double>(videoWidth) / videoHeight) {
    baseOffsetX = (screenWidth - static_cast<int>(static_cast<double>(screenHeight) * videoWidth / videoHeight)) / 2;
    baseOffsetY = 0;
    framePixelSize = screenHeight / videoHeight;
  } else {
    baseOffsetX = 0;
    baseOffsetY = (screenHeight - static_cast<int>(static_cast<double>(screenWidth) * videoHeight / videoWidth)) / 2;
    framePixelSize = screenWidth / videoWidth;
  }
  for (int i = 0; i < WD_TOTAL_FRAME_CNT; i++) {
    for (int j = 0; j < WD_USE_WINDOW_CNT; j++) {
      steps[i][j][0] = baseOffsetX + steps[i][j][0] * framePixelSize;
      steps[i][j][1] = baseOffsetY + steps[i][j][1] * framePixelSize;
      steps[i][j][2] = steps[i][j][2] * framePixelSize;
      steps[i][j][3] = steps[i][j][3] * framePixelSize;
    }
  }
}

long long windowUID = 0;
char classnameBuf[40];
HWND windowList[WD_USE_WINDOW_CNT];
HWND createWindow (int x, int y, int width, int height) {
  WNDCLASS wc = {0};
  wc.lpfnWndProc = DefWindowProc;
  wc.hInstance = GetModuleHandle(NULL);
  std::snprintf(classnameBuf, sizeof(classnameBuf), "windance%d", windowUID++);
  wc.lpszClassName = classnameBuf;
  if (!RegisterClass(&wc)) { return nullptr; }
  HWND hwnd = CreateWindow(
    wc.lpszClassName,
    "",
    WS_POPUP | WS_BORDER,
    x, y,
    width, height,
    NULL, NULL,
    wc.hInstance,
    NULL
  );
  if (!hwnd) { return nullptr; }
  return hwnd;
}

void showWindow (HWND hwnd) {
  ShowWindow(hwnd, SW_SHOW);
  UpdateWindow(hwnd);
}

void destroyWindow(HWND hwnd) {
  DestroyWindow(hwnd);
}

void moveWindow(HWND hwnd, int x, int y, int width, int height) {
  MoveWindow(hwnd, x, y, width, height, TRUE);
  PAINTSTRUCT ps;
  HDC hdc = BeginPaint(hwnd, &ps);
  FillRect(hdc, &ps.rcPaint, (HBRUSH) (COLOR_WINDOW + 1));
  EndPaint(hwnd, &ps);
}

// animation
long long getCurrentTimeMicroseconds() {
  LARGE_INTEGER frequency;
  LARGE_INTEGER counter;
  QueryPerformanceFrequency(&frequency);
  QueryPerformanceCounter(&counter);
  return (counter.QuadPart * 1000000) / frequency.QuadPart;
}
long long startTime = getCurrentTimeMicroseconds();
int nextFrame[WD_USE_WINDOW_CNT] = {0};
void CALLBACK TimerProc(HWND hwnd, UINT message, UINT timerId, DWORD dwTime) {
  // here timerId is the index of the window
  int frame = (getCurrentTimeMicroseconds() - startTime) / 1000000.0 * WD_FPS;
  if (frame < nextFrame[timerId]) { return; }
  if (frame >= WD_TOTAL_FRAME_CNT) {
    PostQuitMessage(0);
    return;
  }
  nextFrame[timerId]++;
  moveWindow(hwnd, steps[frame][timerId][0], steps[frame][timerId][1], steps[frame][timerId][2], steps[frame][timerId][3]);
}

int WINAPI WinMain(HINSTANCE hInstance, HINSTANCE hPrevInstance, LPSTR lpCmdLine, int nCmdShow) {
  adaptFramesToWindowSize();

  for (int i = 0; i < WD_USE_WINDOW_CNT; i++) {
    windowList[i] = createWindow(steps[0][i][0], steps[0][i][1], steps[0][i][2], steps[0][i][3]);
    if (windowList[i] == nullptr) { return 1; }
    showWindow(windowList[i]);
    SetTimer(windowList[i], i, 1, (TIMERPROC)TimerProc);
  }

  MSG msg = {0};
  while (GetMessage(&msg, NULL, 0, 0)) {
    TranslateMessage(&msg);
    DispatchMessage(&msg);
  }

  return 0;
}
	`
	return src, nil
}

func SaveSrc(dir, src string) (string, error) {
	srcFile, err := os.CreateTemp(dir, "win-dance-*.cc")
	srcFile.WriteString(src)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()
	return srcFile.Name(), nil
}
