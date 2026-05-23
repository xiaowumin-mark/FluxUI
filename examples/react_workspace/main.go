package main

import (
	"fmt"
	"image/color"
	"sort"
	"strings"
	"time"

	statepkg "github.com/xiaowumin-mark/FluxUI/state"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type task struct {
	ID        string
	Title     string
	ProjectID string
	Done      bool
}

type workspaceSettings struct {
	Name    string
	Compact bool
	Accent  color.NRGBA
}

type uiColors struct {
	Surface    color.NRGBA
	Panel      color.NRGBA
	PanelMuted color.NRGBA
	Text       color.NRGBA
	Muted      color.NRGBA
	Accent     color.NRGBA
	Border     color.NRGBA
	Success    color.NRGBA
}

type workspaceSnapshot struct {
	Settings workspaceSettings
	Colors   uiColors
}

type settingsControls struct {
	Compact    *statepkg.State[bool]
	AccentWarm *statepkg.State[bool]
}

type taskCardProps struct {
	Item     task
	Tasks    *statepkg.State[[]task]
	Snapshot workspaceSnapshot
}

var taskCardPropsContext = ui.ContextKey[taskCardProps]{}

func main() {
	_ = ui.RunElement(App, ui.Title("FluxUI 响应式工作区"), ui.Size(920, 680))
}

func App(ctx *ui.Context) ui.Element {
	compact := ui.UseState(ctx, false)
	accentWarm := ui.UseState(ctx, false)
	settings := workspaceSettings{
		Name:    "FluxUI 响应式工作区",
		Compact: compact.Value(),
		Accent:  accentColor(accentWarm.Value()),
	}
	snapshot := workspaceSnapshot{Settings: settings, Colors: buildColors(settings.Accent)}

	return ui.ContainerElement(
		ui.Style{Background: snapshot.Colors.Surface, Padding: ui.All(18)},
		ui.ColumnElement(
			header(snapshot),
			ui.SpacerElement(0, 12),
			ui.RouterElement(
				ui.RouteElement("/", DashboardPage(snapshot)),
				ui.RouteElement("/projects/:id", ProjectPage(snapshot)),
				ui.RouteElement("/settings", SettingsPage(snapshot, settingsControls{Compact: compact, AccentWarm: accentWarm})),
				ui.RouteElement("/bridge", BridgePage(snapshot)),
			),
		),
	)
}

func SettingsPage(snapshot workspaceSnapshot, controls settingsControls) ui.Component {
	return func(ctx *ui.Context) ui.Element {
		location := ui.UseLocation(ctx)
		navigate := ui.UseNavigate(ctx)

		return pageCard(snapshot,
			"设置",
			"路由工厂将实时设置传递给页面，而 Provider + UseContext 传递每个卡片的数据。",
			routeBadge(snapshot, location),
			ui.SpacerElement(0, 10),
			ui.RowElement(
				settingToggle(snapshot, "紧凑卡片", controls.Compact, "控制带键值任务列表中的间距。"),
				ui.SpacerElement(12, 0),
				settingToggle(snapshot, "暖色强调色", controls.AccentWarm, "改变子组件消费的路由快照。"),
			),
			ui.SpacerElement(0, 12),
			infoBox(snapshot,
				"上下文快照",
				fmt.Sprintf("名称=%s 紧凑=%v 强调色=#%02x%02x%02x", snapshot.Settings.Name, snapshot.Settings.Compact, snapshot.Settings.Accent.R, snapshot.Settings.Accent.G, snapshot.Settings.Accent.B),
			),
			ui.SpacerElement(0, 12),
			actionRow(
				primaryButton(snapshot, "仪表盘", func() { navigate("/", ui.WithNavTransition(ui.TransitionSlideRight)) }),
				secondaryButton(snapshot, "桥接", func() { navigate("/bridge", ui.WithNavTransition(ui.TransitionSlideRight)) }),
			),
		)
	}
}

func DashboardPage(snapshot workspaceSnapshot) ui.Component {
	return func(ctx *ui.Context) ui.Element {
		tasks := ui.UseState(ctx, initialTasks())
		lastEvent := ui.UseState(ctx, "工作区已装载")
		mountNote := ui.UseState(ctx, "挂载等待中")
		navigate := ui.UseNavigate(ctx)
		location := ui.UseLocation(ctx)

		ui.UseMount(ctx, func() func() {
			mountNote.Set("仪表盘已于 " + time.Now().Format("15:04:05") + " 挂载")
			return func() {
				fmt.Println("DashboardPage unmounted")
			}
		})

		doneCount := countDone(tasks.Value())
		totalCount := len(tasks.Value())
		ui.UseEffectWithDeps(ctx, []any{doneCount, totalCount}, func() func() {
			lastEvent.Set(fmt.Sprintf("已于 %s 同步 %d/%d 个任务", time.Now().Format("15:04:05"), doneCount, totalCount))
			return nil
		})

		return ui.Fragment(
			pageCard(snapshot,
				"仪表盘",
				"RunElement + 函数组件 + HookSlot 状态驱动此工作区。",
				routeBadge(snapshot, location),
				ui.SpacerElement(0, 8),
				summaryRow(snapshot, doneCount, totalCount),
				ui.SpacerElement(0, 8),
				ui.TextElement(mountNote.Value(), ui.TextSize(12), ui.TextColor(snapshot.Colors.Muted)),
				ui.TextElement(lastEvent.Value(), ui.TextSize(12), ui.TextColor(snapshot.Colors.Muted)),
				ui.SpacerElement(0, 12),
				actionRow(
					primaryButton(snapshot, "添加任务", func() {
						next := append([]task(nil), tasks.Value()...)
						id := fmt.Sprintf("task-%d", len(next)+1)
						next = append(next, task{ID: id, Title: "审查运行时笔记 " + strings.TrimPrefix(id, "task-"), ProjectID: "docs"})
						tasks.Set(next)
					}),
					secondaryButton(snapshot, "轮换顺序", func() {
						next := append([]task(nil), tasks.Value()...)
						if len(next) > 1 {
							first := next[0]
							next = append(next[1:], first)
							tasks.Set(next)
						}
					}),
					secondaryButton(snapshot, "项目文档", func() {
						navigate("/projects/docs", ui.WithNavTransition(ui.TransitionSlideLeft))
					}),
					secondaryButton(snapshot, "设置", func() {
						navigate("/settings", ui.WithNavTransition(ui.TransitionSlideLeft))
					}),
					secondaryButton(snapshot, "桥接", func() {
						navigate("/bridge", ui.WithNavTransition(ui.TransitionSlideLeft))
					}),
				),
			),
			ui.SpacerElement(0, 12),
			taskList(snapshot, tasks),
		)
	}
}

func ProjectPage(snapshot workspaceSnapshot) ui.Component {
	return func(ctx *ui.Context) ui.Element {
		params := ui.UseParams(ctx)
		location := ui.UseLocation(ctx)
		navigate := ui.UseNavigate(ctx)
		visits := ui.UseState(ctx, 0)
		projectID := params.Get("id")
		if projectID == "" {
			projectID = "未知"
		}

		ui.UseMount(ctx, func() func() {
			visits.Set(visits.Value() + 1)
			return func() {
				fmt.Println("ProjectPage unmounted")
			}
		})

		return pageCard(snapshot,
			"项目详情",
			"RouterElement 在参数变化时通过路由模式保持组件身份。",
			routeBadge(snapshot, location),
			ui.SpacerElement(0, 10),
			infoBox(snapshot, "路由参数", fmt.Sprintf("项目 ID = %s", projectID)),
			ui.SpacerElement(0, 8),
			infoBox(snapshot, "组件状态", fmt.Sprintf("本地挂载计数 = %d", visits.Value())),
			ui.SpacerElement(0, 12),
			actionRow(
				primaryButton(snapshot, "打开设计", func() {
					navigate("/projects/design", ui.WithNavTransition(ui.TransitionFade))
				}),
				secondaryButton(snapshot, "打开文档", func() {
					navigate("/projects/docs", ui.WithNavTransition(ui.TransitionSlideLeft))
				}),
				secondaryButton(snapshot, "仪表盘", func() {
					navigate("/", ui.WithNavTransition(ui.TransitionSlideRight))
				}),
				secondaryButton(snapshot, "设置", func() {
					navigate("/settings", ui.WithNavTransition(ui.TransitionSlideLeft))
				}),
			),
		)
	}
}

func BridgePage(snapshot workspaceSnapshot) ui.Component {
	return func(ctx *ui.Context) ui.Element {
		location := ui.UseLocation(ctx)
		navigate := ui.UseNavigate(ctx)
		progress := ui.UseState(ctx, float32(64))

		ui.UseEffectWithDeps(ctx, []any{progress.Value()}, func() func() {
			fmt.Println("桥接进度", progress.Value())
			return nil
		})

		legacyProgress := ui.FromWidget(ui.ProgressBar(
			progress.Value(),
			ui.ProgressMin(0),
			ui.ProgressMax(100),
			ui.ProgressTrackColor(snapshot.Colors.PanelMuted),
			ui.ProgressFillColor(snapshot.Colors.Accent),
		))

		legacyCard := ui.FromWidget(ui.Card(
			ui.Column(
				ui.Text("嵌入元素树的旧版卡片", ui.TextSize(14)),
				ui.Padding(ui.Insets{Top: 6}, ui.Text("FromWidget 保持旧版组件状态和绘图完整。", ui.TextSize(12), ui.TextColor(snapshot.Colors.Muted))),
			),
			ui.CardBackground(ui.NRGBA(255, 255, 255, 255)),
			ui.CardBorder(snapshot.Colors.Border, 1),
			ui.CardRadius(12),
		))

		return pageCard(snapshot,
			"桥接",
			"FromWidget 允许 React 风格组件在 API 逐步迁移时托管旧版组件。",
			routeBadge(snapshot, location),
			ui.SpacerElement(0, 12),
			legacyCard,
			ui.SpacerElement(0, 14),
			ui.TextElement(fmt.Sprintf("旧版进度条值：%.0f%%", progress.Value()), ui.TextSize(14)),
			ui.SpacerElement(0, 6),
			legacyProgress,
			ui.SpacerElement(0, 12),
			actionRow(
				primaryButton(snapshot, "+12", func() {
					progress.Set(clampProgress(progress.Value() + 12))
				}),
				secondaryButton(snapshot, "-12", func() {
					progress.Set(clampProgress(progress.Value() - 12))
				}),
				secondaryButton(snapshot, "仪表盘", func() { navigate("/", ui.WithNavTransition(ui.TransitionSlideLeft)) }),
				secondaryButton(snapshot, "设置", func() { navigate("/settings", ui.WithNavTransition(ui.TransitionSlideLeft)) }),
			),
		)
	}
}

func taskList(snapshot workspaceSnapshot, tasks *statepkg.State[[]task]) ui.Element {
	items := tasks.Value()
	children := []ui.Element{
		panel(snapshot,
			sectionTitle(snapshot, "带键值的任务卡片"),
			helperText(snapshot, "每个卡片都有本地展开状态。轮换列表以验证状态跟随任务 ID 而非索引。"),
		),
		ui.SpacerElement(0, 8),
	}

	for _, item := range items {
		children = append(children,
			ui.Key(item.ID, ui.Provider(taskCardPropsContext, taskCardProps{Item: item, Tasks: tasks, Snapshot: snapshot}, ui.ComponentElement(TaskCard))),
			ui.SpacerElement(0, 8),
		)
	}

	return ui.Fragment(children...)
}

func TaskCard(ctx *ui.Context) ui.Element {
	props := ui.UseContext(ctx, taskCardPropsContext)
	item := props.Item
	tasks := props.Tasks
	expanded := ui.UseState(ctx, false)
	snapshot := props.Snapshot
	status := "待办"
	statusColor := snapshot.Colors.Muted
	if item.Done {
		status = "已完成"
		statusColor = snapshot.Colors.Success
	}

	children := []ui.Element{
		ui.RowElement(
			ui.CheckboxElement("", item.Done, ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) {
				updateTask(tasks, item.ID, func(t *task) { t.Done = checked })
			})),
			ui.SpacerElement(8, 0),
			ui.ColumnElement(
				ui.TextElement(item.Title, ui.TextSize(15), ui.TextColor(snapshot.Colors.Text)),
				ui.TextElement(fmt.Sprintf("%s - 项目 %s - 本地展开=%v", status, item.ProjectID, expanded.Value()), ui.TextSize(12), ui.TextColor(statusColor)),
			),
			ui.SpacerElement(12, 0),
			secondaryButton(snapshot, "详情", func() {
				expanded.Set(!expanded.Value())
			}),
		),
	}

	if expanded.Value() {
		children = append(children,
			ui.SpacerElement(0, 8),
			ui.ContainerElement(
				ui.Style{Background: snapshot.Colors.PanelMuted, Padding: ui.All(10), Radius: 10},
				ui.TextElement("此本地状态存储在使用任务 ID 作为键值的组件 HookSlot 中。", ui.TextSize(12), ui.TextColor(snapshot.Colors.Muted)),
			),
		)
	}

	return ui.ContainerElement(
		ui.Style{Background: ui.NRGBA(255, 255, 255, 255), Padding: cardPadding(snapshot.Settings.Compact), Radius: 14},
		ui.ColumnElement(children...),
	)
}

func header(snapshot workspaceSnapshot) ui.Element {
	return ui.ContainerElement(
		ui.Style{Background: snapshot.Colors.Panel, Padding: ui.All(14), Radius: 18},
		ui.RowElement(
			ui.ColumnElement(
				ui.TextElement(snapshot.Settings.Name, ui.TextSize(24), ui.TextColor(snapshot.Colors.Text)),
				ui.TextElement("一个完整的 React 风格运行时示例：钩子、上下文、路由、键值、片段和桥接。", ui.TextSize(12), ui.TextColor(snapshot.Colors.Muted)),
			),
		),
	)
}

func pageCard(snapshot workspaceSnapshot, title, subtitle string, children ...ui.Element) ui.Element {
	items := []ui.Element{
		ui.TextElement(title, ui.TextSize(22), ui.TextColor(snapshot.Colors.Text)),
		ui.TextElement(subtitle, ui.TextSize(13), ui.TextColor(snapshot.Colors.Muted)),
	}
	items = append(items, children...)
	return panel(snapshot, items...)
}

func panel(snapshot workspaceSnapshot, children ...ui.Element) ui.Element {
	return ui.ContainerElement(
		ui.Style{Background: snapshot.Colors.Panel, Padding: ui.All(16), Radius: 18},
		ui.ColumnElement(children...),
	)
}

func summaryRow(snapshot workspaceSnapshot, done, total int) ui.Element {
	percent := 0
	if total > 0 {
		percent = done * 100 / total
	}
	return ui.RowElement(
		metricCard(snapshot, "已完成", fmt.Sprintf("%d/%d", done, total)),
		ui.SpacerElement(10, 0),
		metricCard(snapshot, "进度", fmt.Sprintf("%d%%", percent)),
		ui.SpacerElement(10, 0),
		metricCard(snapshot, "运行时", "元素"),
	)
}

func metricCard(snapshot workspaceSnapshot, label, value string) ui.Element {
	return ui.ContainerElement(
		ui.Style{Background: snapshot.Colors.PanelMuted, Padding: ui.All(12), Radius: 14},
		ui.ColumnElement(
			ui.TextElement(label, ui.TextSize(12), ui.TextColor(snapshot.Colors.Muted)),
			ui.TextElement(value, ui.TextSize(20), ui.TextColor(snapshot.Colors.Accent)),
		),
	)
}

func routeBadge(snapshot workspaceSnapshot, location *ui.Location) ui.Element {
	path := ""
	if location != nil {
		path = location.Path
	}
	return infoBox(snapshot, "路由", path)
}

func infoBox(snapshot workspaceSnapshot, label, value string) ui.Element {
	return ui.ContainerElement(
		ui.Style{Background: snapshot.Colors.PanelMuted, Padding: ui.All(10), Radius: 12},
		ui.ColumnElement(
			ui.TextElement(label, ui.TextSize(12), ui.TextColor(snapshot.Colors.Muted)),
			ui.TextElement(value, ui.TextSize(14), ui.TextColor(snapshot.Colors.Text)),
		),
	)
}

func settingToggle(snapshot workspaceSnapshot, label string, value *statepkg.State[bool], helper string) ui.Element {
	return ui.ContainerElement(
		ui.Style{Background: snapshot.Colors.PanelMuted, Padding: ui.All(12), Radius: 14},
		ui.ColumnElement(
			ui.RowElement(
				ui.SwitchElement(value.Value(), ui.SwitchOnChange(func(ctx *ui.Context, checked bool) {
					value.Set(checked)
				})),
				ui.SpacerElement(8, 0),
				ui.TextElement(label, ui.TextSize(15), ui.TextColor(snapshot.Colors.Text)),
			),
			ui.SpacerElement(0, 6),
			ui.TextElement(helper, ui.TextSize(12), ui.TextColor(snapshot.Colors.Muted)),
		),
	)
}

func sectionTitle(snapshot workspaceSnapshot, text string) ui.Element {
	return ui.TextElement(text, ui.TextSize(18), ui.TextColor(snapshot.Colors.Text))
}

func helperText(snapshot workspaceSnapshot, text string) ui.Element {
	return ui.TextElement(text, ui.TextSize(12), ui.TextColor(snapshot.Colors.Muted))
}

func actionRow(children ...ui.Element) ui.Element {
	return ui.RowElement(spaced(children, 8)...)
}

func primaryButton(snapshot workspaceSnapshot, label string, onClick func()) ui.Element {
	return ui.ButtonElement(
		ui.TextElement(label),
		ui.ButtonPadding(ui.Symmetric(7, 12)),
		ui.ButtonRadius(10),
		ui.ButtonBackground(snapshot.Colors.Accent),
		ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
		ui.OnClick(func(ctx *ui.Context) { onClick() }),
	)
}

func secondaryButton(snapshot workspaceSnapshot, label string, onClick func()) ui.Element {
	return ui.ButtonElement(
		ui.TextElement(label),
		ui.ButtonPadding(ui.Symmetric(7, 12)),
		ui.ButtonRadius(10),
		ui.ButtonBackground(snapshot.Colors.PanelMuted),
		ui.ButtonForeground(snapshot.Colors.Text),
		ui.OnClick(func(ctx *ui.Context) { onClick() }),
	)
}

func spaced(children []ui.Element, gap float32) []ui.Element {
	items := make([]ui.Element, 0, len(children)*2)
	for index, child := range children {
		if index > 0 {
			items = append(items, ui.SpacerElement(gap, 0))
		}
		items = append(items, child)
	}
	return items
}

func updateTask(tasks *statepkg.State[[]task], id string, update func(*task)) {
	next := append([]task(nil), tasks.Value()...)
	for index := range next {
		if next[index].ID == id {
			update(&next[index])
			break
		}
	}
	tasks.Set(next)
}

func countDone(items []task) int {
	count := 0
	for _, item := range items {
		if item.Done {
			count++
		}
	}
	return count
}

func initialTasks() []task {
	items := []task{
		{ID: "task-docs", Title: "审核组件文档", ProjectID: "docs", Done: true},
		{ID: "task-router", Title: "验证路由元素钩子", ProjectID: "runtime", Done: true},
		{ID: "task-keys", Title: "检查键值列表标识", ProjectID: "runtime"},
		{ID: "task-bridge", Title: "记录 FromWidget 桥接", ProjectID: "docs"},
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items
}

func cardPadding(compact bool) ui.Insets {
	if compact {
		return ui.Symmetric(8, 10)
	}
	return ui.All(12)
}

func clampProgress(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func accentColor(warm bool) color.NRGBA {
	if warm {
		return ui.NRGBA(217, 119, 6, 255)
	}
	return ui.NRGBA(14, 116, 144, 255)
}

func buildColors(accent color.NRGBA) uiColors {
	return uiColors{
		Surface:    ui.NRGBA(238, 242, 246, 255),
		Panel:      ui.NRGBA(255, 255, 255, 255),
		PanelMuted: ui.NRGBA(241, 245, 249, 255),
		Text:       ui.NRGBA(15, 23, 42, 255),
		Muted:      ui.NRGBA(100, 116, 139, 255),
		Accent:     accent,
		Border:     ui.NRGBA(203, 213, 225, 255),
		Success:    ui.NRGBA(22, 163, 74, 255),
	}
}
