# Material 3 视觉回归

本页记录 Phase G 的本地截图、pixel smoke check、交互状态关键帧和 CI artifact 流程。当前不引入 golden image 比较，目标是先建立可重复产出截图、可防空白渲染、可人工复核的闭环。

## 运行命令

```sh
make visual
```

等价底层命令：

```sh
go test -tags visual ./examples/material3_showcase -run TestMaterial3ShowcaseScreenshots -count=1
```

默认输出目录：

```text
out/material3-screenshots/
```

可通过环境变量覆盖输出目录：

```sh
FLUXUI_VISUAL_OUTPUT=out/custom-material3-screenshots make visual
```

PowerShell 写法：

```powershell
$env:FLUXUI_VISUAL_OUTPUT = "out/custom-material3-screenshots"
go test -tags visual ./examples/material3_showcase -run TestMaterial3ShowcaseScreenshots -count=1
Remove-Item Env:\FLUXUI_VISUAL_OUTPUT
```

相对输出路径会按仓库根目录解析。

该测试使用 Gio headless renderer 直接渲染 `examples/material3_showcase` 和专用 visual fixture，不会打开桌面窗口。普通 `go test ./...` 不会运行该测试，因为入口带有 `visual` build tag。

本地前置条件是当前机器需要可用的 Gio headless GPU backend；若 backend 不可用，显式视觉测试会失败并提示 headless window 创建错误。

## 截图矩阵

当前一次运行会生成 30 张 PNG 和 1 个 `manifest.json`。

Showcase 全屏截图：

| 文件 | 主题 | Viewport |
| --- | --- | --- |
| `showcase-light-desktop.png` | Light | `1440x1000` |
| `showcase-light-narrow.png` | Light | `480x900` |
| `showcase-dark-desktop.png` | Dark | `1440x1000` |
| `showcase-dark-narrow.png` | Dark | `480x900` |

关键组件区域截图，每个区域同时生成 Light 和 Dark：

| 文件模式 | 覆盖内容 |
| --- | --- |
| `region-{theme}-buttons.png` | Button variants、IconButton、FAB |
| `region-{theme}-inputs-cards.png` | Outlined/Filled TextField、Select、Card variants |
| `region-{theme}-selection.png` | Checkbox、Radio、Switch、Slider |
| `region-{theme}-navigation.png` | Tabs、BottomNavigation、NavigationRail、NavigationDrawer |
| `region-{theme}-overlay.png` | Dialog、Popup、Toast、Snackbar |
| `region-{theme}-chips-progress.png` | Chips、SearchBar、Linear/Circular Progress |

交互状态关键帧截图，每个状态同时生成 Light 和 Dark：

| 文件模式 | 覆盖状态 |
| --- | --- |
| `state-{theme}-hover.png` | hover state layer specimen |
| `state-{theme}-pressed-ripple.png` | pressed state layer 和 ripple 目标区域 |
| `state-{theme}-focus.png` | focused TextField 和 focus ring |
| `state-{theme}-selected.png` | Tabs、BottomNavigation、FilterChip selected |
| `state-{theme}-expanded.png` | expanded DropdownMenu / Select trigger |
| `state-{theme}-toast-enter.png` | toast visible frame |
| `state-{theme}-toast-exit.png` | toast expired frame |

`manifest.json` 会记录运行命令、截图 category、输出路径、尺寸和 smoke check 统计值。

## 自动 Smoke Check

截图测试会检查：

- PNG 尺寸必须等于对应 viewport。
- 非透明像素比例必须足够高，避免空白或透明截图。
- 不透明像素比例必须足够高，避免只渲染了局部内容。
- 量化后的颜色数量必须超过最低阈值，避免整张图只有单色背景。
- 亮度范围必须超过最低阈值，避免整张图过于平坦。

这些检查只用于兜底，不替代人工视觉验收，也不做 golden diff。

## CI Artifact

`.github/workflows/ci.yml` 已新增 `visual-regression` job：

- 运行环境：`ubuntu-latest`
- 命令：`make visual`
- Artifact：`material3-screenshots`
- 路径：`out/material3-screenshots`
- 保留时间：14 天

CI 失败时也会尝试上传已有截图，方便排查渲染或 smoke check 问题。

## 人工验收清单

每次调整 MD3 默认样式后，运行截图命令并检查以下项目：

- Light 和 Dark 截图都不是空白、纯色、明显过暗或明显过亮。
- `Material 3 Tokens` 区域中的 primary、secondary container、surface container、error container 文字对比度可读。
- `Dynamic Color` 区域能看到不同 seed 主题差异，暗色 seed 面板没有低对比文字。
- `Type Scale` 区域字号、行高、权重层级清晰，长文本没有被裁剪。
- Button、TextField、Card、Selection Controls 的圆角、边框、填充、禁用态符合 MD3 默认视觉。
- Tabs、BottomNavigation、NavigationRail、NavigationDrawer 的 selected 状态清晰，图标和标签没有重叠。
- Snackbar、Tooltip、Popup、Toast、Dialog 的 overlay 层级正确，没有被背景内容遮挡。
- narrow viewport 下没有异常崩溃、空白截图或整屏重叠；若出现横向裁剪，需要决定是修 showcase 响应式布局还是记录为已知限制。
- 交互状态截图中的 hover、pressed、focus、selected、expanded、toast enter/exit 状态应能被直接识别。
- 截图中没有明显 mojibake、乱码、缺字、文字溢出按钮或文字覆盖相邻组件。

## 当前边界

- 当前只做固定截图和低成本 pixel smoke check。
- 当前不比较 golden image。
- hover/pressed 截图使用 visual fixture 固定关键状态，不替代后续更细粒度的 pointer event 行为测试。
