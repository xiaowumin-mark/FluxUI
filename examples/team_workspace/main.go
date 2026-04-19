package main

import (
	"fmt"
	"image/color"
	"sort"
	"strings"
	"time"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

const (
	pageWorkspace = "workspace"
	pageDelivery  = "delivery"
	pageSettings  = "settings"

	tabBoard    = "board"
	tabFocus    = "focus"
	tabActivity = "activity"
)

type task struct {
	ID       int
	Title    string
	Owner    string
	Priority string
	Status   string
	Progress float32
	Blocked  bool
	Due      string
	Notes    string
}

type taskSummary struct {
	Total           int
	Todo            int
	Active          int
	Review          int
	Done            int
	Blocked         int
	AverageProgress float32
	PlannedHours    float32
}

type ownerLoad struct {
	Name        string
	Count       int
	ReviewCount int
	AvgProgress float32
}

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)

		boot := ui.State[bool](ctx)
		tasks := ui.State[[]task](ctx)
		selectedID := ui.State[int](ctx)
		page := ui.State[string](ctx)
		tab := ui.State[string](ctx)
		search := ui.State[string](ctx)
		statusFilter := ui.State[string](ctx)
		sortMode := ui.State[string](ctx)
		blockedOnly := ui.State[bool](ctx)
		showCreateDialog := ui.State[bool](ctx)
		showResetDialog := ui.State[bool](ctx)
		toastMessage := ui.State[string](ctx)
		activityLog := ui.State[[]string](ctx)
		newTitle := ui.State[string](ctx)
		newOwner := ui.State[string](ctx)
		newPriority := ui.State[string](ctx)
		newBlocked := ui.State[bool](ctx)
		autoToast := ui.State[bool](ctx)
		denseMode := ui.State[bool](ctx)
		reviewAlerts := ui.State[bool](ctx)
		weeklyGoal := ui.State[float32](ctx)

		if !boot.Value() {
			seed := sampleTasks()
			tasks.Set(seed)
			selectedID.Set(seed[0].ID)
			page.Set(pageWorkspace)
			tab.Set(tabBoard)
			statusFilter.Set("all")
			sortMode.Set("priority")
			newPriority.Set("medium")
			autoToast.Set(true)
			denseMode.Set(false)
			reviewAlerts.Set(true)
			weeklyGoal.Set(72)
			activityLog.Set([]string{
				stamped("工作台示例已加载"),
				stamped("打开看板，调整筛选并查看任务详情"),
			})
			boot.Set(true)
		}

		allTasks := tasks.Value()
		if len(allTasks) == 0 {
			allTasks = sampleTasks()
			tasks.Set(allTasks)
		}
		if len(allTasks) > 0 && !taskExists(allTasks, selectedID.Value()) {
			selectedID.Set(allTasks[0].ID)
		}

		currentTask, hasCurrentTask := findTask(allTasks, selectedID.Value())
		filteredTasks := filterTasks(allTasks, search.Value(), statusFilter.Value(), blockedOnly.Value(), sortMode.Value())
		summary := summarizeTasks(allTasks)
		owners := buildOwnerLoads(allTasks)
		compactLayout := ctx.MaxConstraints().X < 1100

		notify := func(message string) {
			activityLog.Set(appendActivity(activityLog.Value(), message))
			if autoToast.Value() {
				toastMessage.Set(message)
			} else {
				toastMessage.Set("")
			}
		}

		resetTaskForm := func() {
			newTitle.Set("")
			newOwner.Set("")
			newPriority.Set("medium")
			newBlocked.Set(false)
		}

		createTask := func() {
			title := strings.TrimSpace(newTitle.Value())
			owner := strings.TrimSpace(newOwner.Value())
			if title == "" {
				notify("创建任务前必须填写任务标题")
				return
			}
			if owner == "" {
				owner = "未分配"
			}

			next := task{
				ID:       nextTaskID(allTasks),
				Title:    title,
				Owner:    owner,
				Priority: normalizePriority(newPriority.Value()),
				Status:   "todo",
				Progress: 0,
				Blocked:  newBlocked.Value(),
				Due:      "本周",
				Notes:    "该任务从工作台弹窗创建。",
			}

			updated := append([]task{next}, cloneTasks(allTasks)...)
			tasks.Set(updated)
			selectedID.Set(next.ID)
			showCreateDialog.Set(false)
			resetTaskForm()
			notify("已创建任务：" + next.Title)
		}

		resetDemo := func() {
			seed := sampleTasks()
			tasks.Set(seed)
			selectedID.Set(seed[0].ID)
			search.Set("")
			statusFilter.Set("all")
			sortMode.Set("priority")
			blockedOnly.Set(false)
			showResetDialog.Set(false)
			resetTaskForm()
			notify("已将工作台重置为示例数据")
		}

		advanceTaskStatus := func(taskID int) {
			before, ok := findTask(allTasks, taskID)
			if !ok {
				return
			}

			updated := updateTask(allTasks, taskID, func(item *task) {
				switch item.Status {
				case "todo":
					item.Status = "active"
					if item.Progress < 0.25 {
						item.Progress = 0.25
					}
				case "active":
					item.Status = "review"
					if item.Progress < 0.8 {
						item.Progress = 0.8
					}
				case "review":
					item.Status = "done"
					item.Progress = 1
					item.Blocked = false
				}
			})
			tasks.Set(updated)

			after, _ := findTask(updated, taskID)
			switch after.Status {
			case "review":
				if reviewAlerts.Value() {
					notify("已移动到评审：" + after.Title)
				} else {
					notify("状态已推进：" + after.Title)
				}
			case "done":
				notify("任务已完成：" + after.Title)
			default:
				if before.Status != after.Status {
					notify("状态已推进：" + after.Title)
				}
			}
		}

		markTaskDone := func(taskID int) {
			updated := updateTask(allTasks, taskID, func(item *task) {
				item.Status = "done"
				item.Progress = 1
				item.Blocked = false
			})
			tasks.Set(updated)
			if after, ok := findTask(updated, taskID); ok {
				notify("任务已完成：" + after.Title)
			}
		}

		toggleBlockedState := func(taskID int) {
			updated := updateTask(allTasks, taskID, func(item *task) {
				item.Blocked = !item.Blocked
			})
			tasks.Set(updated)
			if after, ok := findTask(updated, taskID); ok {
				if after.Blocked {
					notify("任务已阻塞：" + after.Title)
				} else {
					notify("任务已解除阻塞：" + after.Title)
				}
			}
		}

		boostProgress := func(taskID int) {
			updated := updateTask(allTasks, taskID, func(item *task) {
				item.Progress += 0.1
				if item.Progress > 1 {
					item.Progress = 1
				}
				if item.Progress >= 1 {
					item.Status = "done"
					item.Blocked = false
				} else if item.Progress >= 0.8 && item.Status == "active" {
					item.Status = "review"
				}
			})
			tasks.Set(updated)
			if after, ok := findTask(updated, taskID); ok {
				notify(fmt.Sprintf("进度微调：%s -> %.0f%%", after.Title, after.Progress*100))
			}
		}

		statsCards := []ui.Widget{
			metricCard(th, "进行中工作", fmt.Sprintf("%d", summary.Todo+summary.Active+summary.Review), "未完成任务数", th.Primary, float32(summary.Todo+summary.Active+summary.Review)/maxFloat32(1, float32(summary.Total))),
			metricCard(th, "评审队列", fmt.Sprintf("%d", summary.Review), "等待评审的任务", warnColor(), float32(summary.Review)/maxFloat32(1, float32(summary.Total))),
			metricCard(th, "已完成", fmt.Sprintf("%d", summary.Done), "完成任务数", successColor(), float32(summary.Done)/maxFloat32(1, float32(summary.Total))),
			metricCard(th, "阻塞", fmt.Sprintf("%d", summary.Blocked), "存在阻塞项", dangerColor(), float32(summary.Blocked)/maxFloat32(1, float32(summary.Total))),
		}

		filterPanel := ui.Card(
			ui.Column(
				sectionHeader("看板筛选", "进入详情前先搜索并筛选任务列表。"),
				ui.Padding(
					ui.Insets{Top: 12},
					ui.TextField(
						search.Value(),
						ui.InputPlaceholder("搜索标题、负责人、优先级或状态"),
						ui.InputOnChange(func(ctx *ui.Context, value string) {
							search.Set(value)
						}),
					),
				),
				ui.Padding(
					ui.Insets{Top: 12},
					ui.Select(
						statusFilter.Value(),
						[]ui.SelectOptionItem[string]{
							{Label: "全部状态", Value: "all"},
							{Label: "待办", Value: "todo"},
							{Label: "进行中", Value: "active"},
							{Label: "评审中", Value: "review"},
							{Label: "已完成", Value: "done"},
						},
						ui.SelectPlaceholder[string]("状态"),
						ui.SelectOnChange[string](func(ctx *ui.Context, value string) {
							statusFilter.Set(value)
						}),
					),
				),
				ui.Padding(
					ui.Insets{Top: 12},
					ui.Checkbox(
						"仅显示阻塞任务",
						blockedOnly.Value(),
						ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) {
							blockedOnly.Set(checked)
						}),
					),
				),
				ui.Padding(
					ui.Insets{Top: 12},
					ui.Text("排序方式", ui.TextSize(13), ui.TextColor(infoColor())),
				),
				ui.Padding(
					ui.Insets{Top: 6},
					ui.RadioGroup(
						sortMode.Value(),
						[]ui.RadioItem{
							{Label: "优先级优先", Value: "priority"},
							{Label: "状态流转", Value: "status"},
							{Label: "负责人名称", Value: "owner"},
							{Label: "进度", Value: "progress"},
						},
						ui.RadioGroupOnChange(func(ctx *ui.Context, value string) {
							sortMode.Set(value)
						}),
					),
				),
			),
			ui.CardBorder(th.SurfaceMuted, 1),
		)

		workspaceTools := ui.Card(
			ui.Column(
				sectionHeader("工作台工具", "无需离开看板即可执行常用操作。"),
				ui.Padding(
					ui.Insets{Top: 12},
					ui.FillWidth(
						ui.Button(
							ui.Text("新建任务"),
							ui.ButtonBackground(th.Primary),
							ui.ButtonForeground(th.TextOnPrimary),
							ui.OnClick(func(ctx *ui.Context) {
								showCreateDialog.Set(true)
							}),
						),
					),
				),
				ui.Padding(
					ui.Insets{Top: 10},
					ui.FillWidth(
						ui.Button(
							ui.Text("恢复示例数据"),
							ui.ButtonBackground(withAlpha(dangerColor(), 230)),
							ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
							ui.OnClick(func(ctx *ui.Context) {
								showResetDialog.Set(true)
							}),
						),
					),
				),
				ui.Padding(
					ui.Insets{Top: 12},
					ui.Text(
						fmt.Sprintf("可见任务：%d / %d", len(filteredTasks), len(allTasks)),
						ui.TextSize(13),
						ui.TextColor(infoColor()),
					),
				),
			),
			ui.CardBorder(th.SurfaceMuted, 1),
		)

		leftPanelContent := ui.Column(
			filterPanel,
			ui.Padding(ui.Insets{Top: 12}, workspaceTools),
		)
		leftPanel := ui.ScrollView(
			leftPanelContent,
			ui.ScrollVertical(true),
		)

		boardTabs := ui.Tabs(
			tab.Value(),
			[]ui.TabItem{
				{Key: tabBoard, Label: "看板"},
				{Key: tabFocus, Label: "聚焦"},
				{Key: tabActivity, Label: "动态"},
			},
			ui.TabsScrollable(true),
			ui.TabsOnChange(func(ctx *ui.Context, key string) {
				tab.Set(key)
			}),
		)

		boardCard := ui.Card(
			ui.Column(
				sectionHeader("任务队列", "选择一行查看详情并推进流程。"),
				ui.Padding(
					ui.Insets{Top: 12},
					func() ui.Widget {
						if len(filteredTasks) == 0 {
							return ui.Container(
								ui.Style{
									Background: withAlpha(th.SurfaceMuted, 30),
									Padding:    ui.All(18),
									Radius:     10,
								},
								ui.Text("没有符合当前筛选条件的任务。", ui.TextColor(infoColor())),
							)
						}

						return ui.FixedHeight(
							listHeight(compactLayout),
							ui.ListView(
								len(filteredTasks),
								func(ctx *ui.Context, index int) ui.Widget {
									item := filteredTasks[index]
									return taskCard(
										th,
										item,
										item.ID == selectedID.Value(),
										denseMode.Value(),
										func(ctx *ui.Context) {
											selectedID.Set(item.ID)
										},
									)
								},
								ui.ListItemSpacing(8),
							),
						)
					}(),
				),
			),
			ui.CardBorder(th.SurfaceMuted, 1),
		)

		focusCard := ui.Card(
			ui.Column(
				sectionHeader("当前聚焦", "选中任务的紧凑详情视图。"),
				ui.Padding(
					ui.Insets{Top: 12},
					func() ui.Widget {
						if !hasCurrentTask {
							return ui.Text("请选择一个任务查看聚焦视图。", ui.TextColor(infoColor()))
						}
						return ui.Column(
							ui.Text(currentTask.Title, ui.TextSize(20)),
							ui.Padding(
								ui.Insets{Top: 8},
								ui.Row(
									statusBadge(th, currentTask.Status),
									ui.Padding(ui.Insets{Left: 8}, priorityBadge(currentTask.Priority)),
									func() ui.Widget {
										if currentTask.Blocked {
											return ui.Padding(ui.Insets{Left: 8}, blockedBadge())
										}
										return ui.Spacer(0, 0)
									}(),
								),
							),
							ui.Padding(
								ui.Insets{Top: 12},
								ui.Text(currentTask.Notes, ui.TextColor(infoColor())),
							),
							ui.Padding(
								ui.Insets{Top: 12},
								ui.Text(fmt.Sprintf("负责人：%s", currentTask.Owner), ui.TextSize(13)),
							),
							ui.Padding(
								ui.Insets{Top: 6},
								ui.Text(fmt.Sprintf("截止：%s", currentTask.Due), ui.TextSize(13)),
							),
							ui.Padding(
								ui.Insets{Top: 12},
								ui.ProgressBar(
									currentTask.Progress*100,
									ui.ProgressMin(0),
									ui.ProgressMax(100),
									ui.ProgressTrackColor(withAlpha(th.SurfaceMuted, 60)),
									ui.ProgressFillColor(priorityColor(currentTask.Priority)),
								),
							),
							ui.Padding(
								ui.Insets{Top: 8},
								ui.Text(fmt.Sprintf("完成度 %.0f%%", currentTask.Progress*100), ui.TextSize(12), ui.TextColor(infoColor())),
							),
						)
					}(),
				),
			),
			ui.CardBorder(th.SurfaceMuted, 1),
		)

		activityCard := ui.Card(
			ui.Column(
				sectionHeader("动态流", "最近变更和导航事件会保留，便于调试。"),
				ui.Padding(
					ui.Insets{Top: 12},
					buildActivityView(activityLog.Value(), th),
				),
			),
			ui.CardBorder(th.SurfaceMuted, 1),
		)

		centerContent := boardCard
		switch tab.Value() {
		case tabFocus:
			centerContent = focusCard
		case tabActivity:
			centerContent = activityCard
		}

		centerPanelContent := ui.Column(
			stackCards(compactLayout, 12, statsCards...),
			ui.Padding(ui.Insets{Top: 12}, boardTabs),
			ui.Padding(ui.Insets{Top: 12}, centerContent),
		)
		centerPanel := ui.ScrollView(
			centerPanelContent,
			ui.ScrollVertical(true),
		)

		detailPanelContent := ui.Column(
			ui.Card(
				ui.Column(
					sectionHeader("任务详情", "使用右侧面板推进状态或解除阻塞。"),
					ui.Padding(
						ui.Insets{Top: 12},
						func() ui.Widget {
							if !hasCurrentTask {
								return ui.Text("尚未选择任务。", ui.TextColor(infoColor()))
							}

							return ui.Column(
								ui.Text(currentTask.Title, ui.TextSize(18)),
								ui.Padding(
									ui.Insets{Top: 10},
									ui.CircularProgress(
										currentTask.Progress*100,
										ui.ProgressMin(0),
										ui.ProgressMax(100),
										ui.ProgressSize(88),
										ui.ProgressFillColor(priorityColor(currentTask.Priority)),
										ui.ProgressTrackColor(withAlpha(th.SurfaceMuted, 70)),
									),
								),
								ui.Padding(
									ui.Insets{Top: 12},
									ui.Text(fmt.Sprintf("负责人：%s", currentTask.Owner), ui.TextSize(13)),
								),
								ui.Padding(
									ui.Insets{Top: 6},
									ui.Text(fmt.Sprintf("优先级：%s", priorityLabel(currentTask.Priority)), ui.TextSize(13)),
								),
								ui.Padding(
									ui.Insets{Top: 6},
									ui.Text(fmt.Sprintf("状态：%s", statusLabel(currentTask.Status)), ui.TextSize(13)),
								),
								ui.Padding(
									ui.Insets{Top: 6},
									ui.Text(fmt.Sprintf("截止：%s", currentTask.Due), ui.TextSize(13)),
								),
								ui.Padding(
									ui.Insets{Top: 12},
									ui.Text(currentTask.Notes, ui.TextColor(infoColor())),
								),
								ui.Padding(
									ui.Insets{Top: 12},
									ui.FillWidth(
										ui.Button(
											ui.Text("推进状态"),
											ui.ButtonBackground(th.Primary),
											ui.ButtonForeground(th.TextOnPrimary),
											ui.OnClick(func(ctx *ui.Context) {
												advanceTaskStatus(currentTask.ID)
											}),
										),
									),
								),
								ui.Padding(
									ui.Insets{Top: 8},
									ui.FillWidth(
										ui.Button(
											ui.Text("标记完成"),
											ui.ButtonBackground(successColor()),
											ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
											ui.OnClick(func(ctx *ui.Context) {
												markTaskDone(currentTask.ID)
											}),
										),
									),
								),
								ui.Padding(
									ui.Insets{Top: 8},
									ui.FillWidth(
										ui.Button(
											ui.Text(blockActionLabel(currentTask.Blocked)),
											ui.ButtonBackground(withAlpha(warnColor(), 230)),
											ui.ButtonForeground(ui.NRGBA(25, 30, 38, 255)),
											ui.OnClick(func(ctx *ui.Context) {
												toggleBlockedState(currentTask.ID)
											}),
										),
									),
								),
								ui.Padding(
									ui.Insets{Top: 8},
									ui.FillWidth(
										ui.Button(
											ui.Text("进度 +10%"),
											ui.OnClick(func(ctx *ui.Context) {
												boostProgress(currentTask.ID)
											}),
										),
									),
								),
							)
						}(),
					),
				),
				ui.CardBorder(th.SurfaceMuted, 1),
			),
			ui.Padding(
				ui.Insets{Top: 12},
				ui.Card(
					ui.Column(
						sectionHeader("实时设置", "这些控制项会影响示例应用行为。"),
						ui.Padding(
							ui.Insets{Top: 12},
							ui.Switch(
								autoToast.Value(),
								ui.SwitchOnChange(func(ctx *ui.Context, checked bool) {
									autoToast.Set(checked)
								}),
							),
						),
						ui.Padding(
							ui.Insets{Top: 6},
							ui.Text("自动提示", ui.TextSize(13)),
						),
						ui.Padding(
							ui.Insets{Top: 12},
							ui.Switch(
								denseMode.Value(),
								ui.SwitchOnChange(func(ctx *ui.Context, checked bool) {
									denseMode.Set(checked)
								}),
							),
						),
						ui.Padding(
							ui.Insets{Top: 6},
							ui.Text("紧凑看板行", ui.TextSize(13)),
						),
					),
					ui.CardBorder(th.SurfaceMuted, 1),
				),
			),
		)
		detailPanel := ui.ScrollView(
			detailPanelContent,
			ui.ScrollVertical(true),
		)

		workspaceBody := func() ui.Widget {
			if compactLayout {
				return ui.ScrollView(
					ui.Column(
						leftPanelContent,
						ui.Padding(ui.Insets{Top: 12}, centerPanelContent),
						ui.Padding(ui.Insets{Top: 12}, detailPanelContent),
					),
					ui.ScrollVertical(true),
				)
			}

			return ui.Row(
				ui.FixedWidth(300, leftPanel),
				ui.Padding(
					ui.Insets{Left: 12, Right: 12},
					ui.Expanded(centerPanel),
				),
				ui.FixedWidth(320, detailPanel),
			)
		}()

		deliveryBody := ui.ScrollView(
			ui.Column(
				sectionHeader("交付视图", "聚焦交付的页面，包含负责人容量、评审压力和周目标跟踪。"),
				ui.Padding(
					ui.Insets{Top: 12},
					stackCards(
						compactLayout,
						12,
						metricCard(th, "每周计划", fmt.Sprintf("%.0fh", summary.PlannedHours), fmt.Sprintf("目标 %.0f 小时", weeklyGoal.Value()), th.Primary, summary.PlannedHours/maxFloat32(1, weeklyGoal.Value())),
						metricCard(th, "平均进度", fmt.Sprintf("%.0f%%", summary.AverageProgress*100), "全部任务", successColor(), summary.AverageProgress),
						metricCard(th, "评审压力", fmt.Sprintf("%d", summary.Review), "等待评审的任务", warnColor(), float32(summary.Review)/maxFloat32(1, float32(summary.Total))),
					),
				),
				ui.Padding(
					ui.Insets{Top: 12},
					ui.Card(
						ui.Column(
							sectionHeader("负责人容量", "每张卡片汇总负责人工作量和平均完成度。"),
							ui.Padding(
								ui.Insets{Top: 12},
								func() ui.Widget {
									if len(owners) == 0 {
										return ui.Text("暂无负责人容量数据。", ui.TextColor(infoColor()))
									}

									cards := make([]ui.Widget, 0, len(owners))
									for _, owner := range owners {
										cards = append(cards, ownerCard(th, owner))
									}
									return stackCards(compactLayout, 12, cards...)
								}(),
							),
						),
						ui.CardBorder(th.SurfaceMuted, 1),
					),
				),
				ui.Padding(
					ui.Insets{Top: 12},
					ui.Card(
						ui.Column(
							sectionHeader("状态泳道", "用于检查任务是流动推进还是堆积。"),
							ui.Padding(
								ui.Insets{Top: 12},
								buildStatusLane(th, "待办", summary.Todo, summary.Total, neutralColor()),
							),
							ui.Padding(
								ui.Insets{Top: 10},
								buildStatusLane(th, "进行中", summary.Active, summary.Total, th.Primary),
							),
							ui.Padding(
								ui.Insets{Top: 10},
								buildStatusLane(th, "评审中", summary.Review, summary.Total, warnColor()),
							),
							ui.Padding(
								ui.Insets{Top: 10},
								buildStatusLane(th, "已完成", summary.Done, summary.Total, successColor()),
							),
						),
						ui.CardBorder(th.SurfaceMuted, 1),
					),
				),
				ui.Padding(
					ui.Insets{Top: 12},
					ui.Card(
						ui.Column(
							sectionHeader("评审与阻塞队列", "更聚焦风险任务的精简列表。"),
							ui.Padding(
								ui.Insets{Top: 12},
								buildRiskQueue(th, allTasks),
							),
						),
						ui.CardBorder(th.SurfaceMuted, 1),
					),
				),
			),
			ui.ScrollVertical(true),
		)

		settingsBody := ui.ScrollView(
			ui.Column(
				sectionHeader("设置", "调整示例行为，并在需要时重置工作台。"),
				ui.Padding(
					ui.Insets{Top: 12},
					ui.Card(
						ui.Column(
							ui.Text("交付偏好", ui.TextSize(16)),
							ui.Padding(
								ui.Insets{Top: 12},
								ui.Row(
									ui.Switch(
										autoToast.Value(),
										ui.SwitchOnChange(func(ctx *ui.Context, checked bool) {
											autoToast.Set(checked)
											notify("自动提示设置已变更")
										}),
									),
									ui.Padding(ui.Insets{Left: 10, Top: 4}, ui.Text("操作后自动提示")),
								),
							),
							ui.Padding(
								ui.Insets{Top: 12},
								ui.Row(
									ui.Switch(
										denseMode.Value(),
										ui.SwitchOnChange(func(ctx *ui.Context, checked bool) {
											denseMode.Set(checked)
											notify("看板密度已调整")
										}),
									),
									ui.Padding(ui.Insets{Left: 10, Top: 4}, ui.Text("紧凑看板行")),
								),
							),
							ui.Padding(
								ui.Insets{Top: 12},
								ui.Checkbox(
									"启用评审提醒",
									reviewAlerts.Value(),
									ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) {
										reviewAlerts.Set(checked)
										notify("评审提醒设置已变更")
									}),
								),
							),
							ui.Padding(
								ui.Insets{Top: 14},
								ui.Text(fmt.Sprintf("每周容量目标：%.0f 小时", weeklyGoal.Value()), ui.TextSize(13), ui.TextColor(infoColor())),
							),
							ui.Padding(
								ui.Insets{Top: 8},
								ui.Slider(
									weeklyGoal.Value(),
									ui.SliderMin(24),
									ui.SliderMax(120),
									ui.SliderStep(4),
									ui.SliderOnChange(func(ctx *ui.Context, value float32) {
										weeklyGoal.Set(value)
									}),
								),
							),
						),
						ui.CardBorder(th.SurfaceMuted, 1),
					),
				),
				ui.Padding(
					ui.Insets{Top: 12},
					ui.Card(
						ui.Column(
							ui.Text("当前快照", ui.TextSize(16)),
							ui.Padding(
								ui.Insets{Top: 10},
								ui.Text(fmt.Sprintf("任务：总计 %d / 已完成 %d / 阻塞 %d", summary.Total, summary.Done, summary.Blocked)),
							),
							ui.Padding(
								ui.Insets{Top: 6},
								ui.Text(fmt.Sprintf("当前页面：%s", pageLabel(page.Value()))),
							),
							ui.Padding(
								ui.Insets{Top: 6},
								ui.Text(fmt.Sprintf("当前选中任务：%d", selectedID.Value())),
							),
							ui.Padding(
								ui.Insets{Top: 12},
								ui.FillWidth(
									ui.Button(
										ui.Text("重置示例工作台"),
										ui.ButtonBackground(withAlpha(dangerColor(), 230)),
										ui.ButtonForeground(ui.NRGBA(255, 255, 255, 255)),
										ui.OnClick(func(ctx *ui.Context) {
											showResetDialog.Set(true)
										}),
									),
								),
							),
						),
						ui.CardBorder(th.SurfaceMuted, 1),
					),
				),
			),
			ui.ScrollVertical(true),
		)

		body := workspaceBody
		switch page.Value() {
		case pageDelivery:
			body = deliveryBody
		case pageSettings:
			body = settingsBody
		}

		appBar := ui.AppBar(
			ui.Column(
				ui.Text("FluxUI 团队工作台", ui.TextSize(17)),
				ui.Text("基于看板、详情和交付视图构建的完整示例应用。", ui.TextSize(11), ui.TextColor(infoColor())),
			),
			ui.AppBarActions(
				ui.Button(
					ui.Text("新建任务"),
					ui.ButtonPadding(ui.Symmetric(6, 10)),
					ui.OnClick(func(ctx *ui.Context) {
						showCreateDialog.Set(true)
					}),
				),
				ui.Button(
					ui.Text("重置"),
					ui.ButtonPadding(ui.Symmetric(6, 10)),
					ui.OnClick(func(ctx *ui.Context) {
						showResetDialog.Set(true)
					}),
				),
			),
		)

		bottomNav := ui.BottomNavigation(
			page.Value(),
			[]ui.NavItem{
				{Key: pageWorkspace, Label: "工作台", Icon: ui.Text("W", ui.TextSize(12))},
				{Key: pageDelivery, Label: "交付", Icon: ui.Text("D", ui.TextSize(12))},
				{Key: pageSettings, Label: "设置", Icon: ui.Text("S", ui.TextSize(12))},
			},
			ui.BottomNavAlignmentOf(ui.BottomNavAlignSpaceEvenly),
			ui.BottomNavOnChange(func(ctx *ui.Context, key string) {
				page.Set(key)
				notify("已切换页面：" + pageLabel(key))
			}),
		)

		layers := []ui.Widget{
			ui.Column(
				appBar,
				ui.Expanded(
					ui.Padding(
						ui.Insets{Left: 14, Right: 14, Top: 12, Bottom: 12},
						body,
					),
				),
				bottomNav,
			),
		}

		if showCreateDialog.Value() {
			layers = append(layers, ui.Dialog(
				showCreateDialog.Value(),
				ui.Column(
					ui.TextField(
						newTitle.Value(),
						ui.InputPlaceholder("任务标题"),
						ui.InputOnChange(func(ctx *ui.Context, value string) {
							newTitle.Set(value)
						}),
					),
					ui.Padding(
						ui.Insets{Top: 10},
						ui.TextField(
							newOwner.Value(),
							ui.InputPlaceholder("负责人"),
							ui.InputOnChange(func(ctx *ui.Context, value string) {
								newOwner.Set(value)
							}),
						),
					),
					ui.Padding(
						ui.Insets{Top: 10},
						ui.Select(
							newPriority.Value(),
							[]ui.SelectOptionItem[string]{
								{Label: "低优先级", Value: "low"},
								{Label: "中优先级", Value: "medium"},
								{Label: "高优先级", Value: "high"},
							},
							ui.SelectPlaceholder[string]("优先级"),
							ui.SelectOnChange[string](func(ctx *ui.Context, value string) {
								newPriority.Set(value)
							}),
						),
					),
					ui.Padding(
						ui.Insets{Top: 10},
						ui.Checkbox(
							"初始为阻塞",
							newBlocked.Value(),
							ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) {
								newBlocked.Set(checked)
							}),
						),
					),
				),
				ui.DialogTitle("创建任务"),
				ui.DialogWidth(360),
				ui.DialogOnOpenChange(func(ctx *ui.Context, open bool) {
					showCreateDialog.Set(open)
				}),
				ui.DialogOnCancel(func(ctx *ui.Context) {
					showCreateDialog.Set(false)
				}),
				ui.DialogOnConfirm(func(ctx *ui.Context) {
					createTask()
				}),
			))
		}

		if showResetDialog.Value() {
			layers = append(layers, ui.Dialog(
				showResetDialog.Value(),
				ui.Text("恢复此演示工作台的示例数据、筛选条件和选中状态。"),
				ui.DialogTitle("重置工作台"),
				ui.DialogWidth(340),
				ui.DialogOnOpenChange(func(ctx *ui.Context, open bool) {
					showResetDialog.Set(open)
				}),
				ui.DialogOnCancel(func(ctx *ui.Context) {
					showResetDialog.Set(false)
				}),
				ui.DialogOnConfirm(func(ctx *ui.Context) {
					resetDemo()
				}),
			))
		}

		if toastMessage.Value() != "" {
			layers = append(layers, ui.Toast(
				toastMessage.Value(),
				ui.ToastTypeOf(ui.ToastSuccess),
				ui.ToastPositionOf(ui.ToastBottom),
				ui.ToastDuration(2200*time.Millisecond),
				ui.ToastOnClose(func(ctx *ui.Context) {
					toastMessage.Set("")
				}),
			))
		}

		return ui.Container(
			ui.Style{Background: th.Surface},
			ui.Stack(layers...),
		)
	}, ui.Title("FluxUI 团队工作台"), ui.Size(1440, 920))
}

func sampleTasks() []task {
	return []task{
		{ID: 101, Title: "优化看板行密度", Owner: "Ava", Priority: "high", Status: "active", Progress: 0.45, Due: "今天", Notes: "调整列表间距，让紧凑模式在高密度数据下仍然清晰可读。"},
		{ID: 102, Title: "整理对话框生命周期文档", Owner: "Noah", Priority: "medium", Status: "review", Progress: 0.84, Due: "明天", Notes: "补充参考示例，说明打开状态、遮罩关闭和确认回调。"},
		{ID: 103, Title: "上线滑块颜色覆盖能力", Owner: "Mia", Priority: "high", Status: "done", Progress: 1, Due: "已完成", Notes: "主题路径已支持显式覆盖，并在缺省时回退到主题色。"},
		{ID: 104, Title: "重构顶栏尺寸逻辑", Owner: "Iris", Priority: "medium", Status: "todo", Progress: 0.1, Due: "本周", Notes: "压测长操作文案和窄宽度场景，确保中间标题仍可读。"},
		{ID: 105, Title: "审查开关边界状态", Owner: "Liam", Priority: "low", Status: "active", Progress: 0.3, Blocked: true, Due: "周五", Notes: "验证禁用、选中和主题驱动状态，同时保持 API 行为可预测。"},
		{ID: 106, Title: "准备发布说明", Owner: "Ava", Priority: "medium", Status: "review", Progress: 0.72, Due: "周五", Notes: "汇总 UX 修复、示例升级和 API 行为变更，用于下一次发布。"},
		{ID: 107, Title: "扩展交付看板示例", Owner: "Mia", Priority: "high", Status: "active", Progress: 0.58, Due: "下周", Notes: "使用更贴近真实场景的应用流程，而不是平铺组件展示。"},
	}
}

func cloneTasks(items []task) []task {
	out := make([]task, len(items))
	copy(out, items)
	return out
}

func nextTaskID(items []task) int {
	maxID := 0
	for _, item := range items {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return maxID + 1
}

func taskExists(items []task, id int) bool {
	_, ok := findTask(items, id)
	return ok
}

func findTask(items []task, id int) (task, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return task{}, false
}

func updateTask(items []task, id int, mutate func(item *task)) []task {
	out := cloneTasks(items)
	for index := range out {
		if out[index].ID != id {
			continue
		}
		if mutate != nil {
			mutate(&out[index])
		}
		return out
	}
	return out
}

func filterTasks(items []task, search, status string, blockedOnly bool, sortMode string) []task {
	search = strings.TrimSpace(strings.ToLower(search))
	status = strings.TrimSpace(strings.ToLower(status))

	filtered := make([]task, 0, len(items))
	for _, item := range items {
		if blockedOnly && !item.Blocked {
			continue
		}
		if status != "" && status != "all" && item.Status != status {
			continue
		}
		if search != "" {
			haystack := strings.ToLower(item.Title + " " + item.Owner + " " + item.Priority + " " + item.Status + " " + item.Notes)
			if !strings.Contains(haystack, search) {
				continue
			}
		}
		filtered = append(filtered, item)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		left := filtered[i]
		right := filtered[j]
		switch sortMode {
		case "owner":
			if left.Owner != right.Owner {
				return left.Owner < right.Owner
			}
		case "progress":
			if left.Progress != right.Progress {
				return left.Progress > right.Progress
			}
		case "status":
			if statusRank(left.Status) != statusRank(right.Status) {
				return statusRank(left.Status) < statusRank(right.Status)
			}
		default:
			if priorityRank(left.Priority) != priorityRank(right.Priority) {
				return priorityRank(left.Priority) < priorityRank(right.Priority)
			}
		}
		if left.Blocked != right.Blocked {
			return left.Blocked
		}
		return left.ID < right.ID
	})

	return filtered
}

func summarizeTasks(items []task) taskSummary {
	var sum taskSummary
	sum.Total = len(items)
	for _, item := range items {
		switch item.Status {
		case "todo":
			sum.Todo++
			sum.PlannedHours += 5
		case "active":
			sum.Active++
			sum.PlannedHours += 8
		case "review":
			sum.Review++
			sum.PlannedHours += 3
		case "done":
			sum.Done++
		}
		if item.Blocked {
			sum.Blocked++
			sum.PlannedHours += 2
		}
		sum.AverageProgress += item.Progress
	}
	if sum.Total > 0 {
		sum.AverageProgress /= float32(sum.Total)
	}
	return sum
}

func buildOwnerLoads(items []task) []ownerLoad {
	type accumulator struct {
		count    int
		review   int
		progress float32
	}

	acc := map[string]accumulator{}
	for _, item := range items {
		name := item.Owner
		if strings.TrimSpace(name) == "" {
			name = "未分配"
		}
		current := acc[name]
		current.count++
		current.progress += item.Progress
		if item.Status == "review" {
			current.review++
		}
		acc[name] = current
	}

	out := make([]ownerLoad, 0, len(acc))
	for name, item := range acc {
		progress := float32(0)
		if item.count > 0 {
			progress = item.progress / float32(item.count)
		}
		out = append(out, ownerLoad{
			Name:        name,
			Count:       item.count,
			ReviewCount: item.review,
			AvgProgress: progress,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name
	})

	return out
}

func listHeight(compact bool) float32 {
	if compact {
		return 360
	}
	return 520
}

func metricCard(th *ui.Theme, title, value, caption string, accent color.NRGBA, progress float32) ui.Widget {
	return ui.Card(
		ui.Column(
			ui.Text(title, ui.TextSize(13), ui.TextColor(ui.NRGBA(102, 112, 128, 255))),
			ui.Padding(
				ui.Insets{Top: 6},
				ui.Text(value, ui.TextSize(24)),
			),
			ui.Padding(
				ui.Insets{Top: 8},
				ui.ProgressBar(
					clamp01(progress)*100,
					ui.ProgressMin(0),
					ui.ProgressMax(100),
					ui.ProgressTrackColor(withAlpha(th.SurfaceMuted, 55)),
					ui.ProgressFillColor(accent),
				),
			),
			ui.Padding(
				ui.Insets{Top: 8},
				ui.Text(caption, ui.TextSize(12), ui.TextColor(infoColor())),
			),
		),
		ui.CardBorder(th.SurfaceMuted, 1),
	)
}

func taskCard(th *ui.Theme, item task, selected bool, dense bool, onClick func(ctx *ui.Context)) ui.Widget {
	padding := float32(12)
	if dense {
		padding = 10
	}

	background := th.Surface
	borderColor := th.SurfaceMuted
	if selected {
		background = withAlpha(th.Primary, 18)
		borderColor = th.Primary
	}

	return ui.Card(
		ui.Column(
			ui.Row(
				ui.Expanded(
					ui.Column(
						ui.Text(item.Title, ui.TextSize(15)),
						ui.Padding(
							ui.Insets{Top: 6},
							ui.Text(
								fmt.Sprintf("%s  |  %s", item.Owner, item.Due),
								ui.TextSize(12),
								ui.TextColor(infoColor()),
							),
						),
					),
				),
				func() ui.Widget {
					if item.Blocked {
						return blockedBadge()
					}
					return ui.Spacer(0, 0)
				}(),
			),
			ui.Padding(
				ui.Insets{Top: 10},
				ui.Row(
					statusBadge(th, item.Status),
					ui.Padding(ui.Insets{Left: 8}, priorityBadge(item.Priority)),
				),
			),
			ui.Padding(
				ui.Insets{Top: 10},
				ui.ProgressBar(
					item.Progress*100,
					ui.ProgressMin(0),
					ui.ProgressMax(100),
					ui.ProgressTrackColor(withAlpha(th.SurfaceMuted, 50)),
					ui.ProgressFillColor(priorityColor(item.Priority)),
				),
			),
		),
		ui.CardPadding(ui.All(padding)),
		ui.CardBackground(background),
		ui.CardBorder(borderColor, 1),
		ui.CardOnClick(onClick),
	)
}

func buildActivityView(entries []string, th *ui.Theme) ui.Widget {
	if len(entries) == 0 {
		return ui.Text("暂无动态记录。", ui.TextColor(infoColor()))
	}

	items := make([]ui.Widget, 0, len(entries))
	for _, entry := range entries {
		items = append(items,
			ui.Padding(
				ui.Insets{Bottom: 8},
				ui.Container(
					ui.Style{
						Background: withAlpha(th.SurfaceMuted, 20),
						Padding:    ui.Symmetric(8, 10),
						Radius:     8,
					},
					ui.Text(entry, ui.TextSize(12)),
				),
			),
		)
	}

	return ui.FixedHeight(
		320,
		ui.ScrollView(
			ui.Column(items...),
			ui.ScrollVertical(true),
			ui.ScrollAutoToEndKey(len(entries)),
		),
	)
}

func ownerCard(th *ui.Theme, owner ownerLoad) ui.Widget {
	return ui.Card(
		ui.Column(
			ui.Text(owner.Name, ui.TextSize(15)),
			ui.Padding(
				ui.Insets{Top: 6},
				ui.Text(
					fmt.Sprintf("%d 个任务  |  %d 个评审中", owner.Count, owner.ReviewCount),
					ui.TextSize(12),
					ui.TextColor(infoColor()),
				),
			),
			ui.Padding(
				ui.Insets{Top: 10},
				ui.ProgressBar(
					owner.AvgProgress*100,
					ui.ProgressMin(0),
					ui.ProgressMax(100),
					ui.ProgressTrackColor(withAlpha(th.SurfaceMuted, 55)),
					ui.ProgressFillColor(th.Primary),
				),
			),
		),
		ui.CardBorder(th.SurfaceMuted, 1),
	)
}

func buildStatusLane(th *ui.Theme, label string, count int, total int, accent color.NRGBA) ui.Widget {
	progress := float32(count) / maxFloat32(1, float32(total))
	return ui.Column(
		ui.Row(
			ui.Text(label, ui.TextSize(13)),
			ui.Padding(ui.Insets{Left: 8}, ui.Text(fmt.Sprintf("%d", count), ui.TextSize(12), ui.TextColor(infoColor()))),
		),
		ui.Padding(
			ui.Insets{Top: 6},
			ui.ProgressBar(
				progress*100,
				ui.ProgressMin(0),
				ui.ProgressMax(100),
				ui.ProgressTrackColor(withAlpha(th.SurfaceMuted, 55)),
				ui.ProgressFillColor(accent),
			),
		),
	)
}

func buildRiskQueue(th *ui.Theme, items []task) ui.Widget {
	risky := make([]task, 0, len(items))
	for _, item := range items {
		if item.Blocked || item.Status == "review" {
			risky = append(risky, item)
		}
	}
	if len(risky) == 0 {
		return ui.Text("当前示例队列没有评审或阻塞风险。", ui.TextColor(infoColor()))
	}

	sort.Slice(risky, func(i, j int) bool {
		if risky[i].Blocked != risky[j].Blocked {
			return risky[i].Blocked
		}
		return priorityRank(risky[i].Priority) < priorityRank(risky[j].Priority)
	})

	return ui.FixedHeight(
		280,
		ui.ListView(
			len(risky),
			func(ctx *ui.Context, index int) ui.Widget {
				item := risky[index]
				return ui.Card(
					ui.Column(
						ui.Text(item.Title, ui.TextSize(14)),
						ui.Padding(
							ui.Insets{Top: 6},
							ui.Row(
								statusBadge(th, item.Status),
								ui.Padding(ui.Insets{Left: 8}, priorityBadge(item.Priority)),
								func() ui.Widget {
									if item.Blocked {
										return ui.Padding(ui.Insets{Left: 8}, blockedBadge())
									}
									return ui.Spacer(0, 0)
								}(),
							),
						),
						ui.Padding(
							ui.Insets{Top: 8},
							ui.Text(item.Owner+"  |  "+item.Due, ui.TextSize(12), ui.TextColor(infoColor())),
						),
					),
					ui.CardBorder(th.SurfaceMuted, 1),
				)
			},
			ui.ListItemSpacing(8),
		),
	)
}

func stackCards(compact bool, gap float32, cards ...ui.Widget) ui.Widget {
	children := make([]ui.Widget, 0, len(cards)*2)
	for index, card := range cards {
		if card == nil {
			continue
		}
		if index > 0 {
			if compact {
				children = append(children, ui.VSpacer(gap))
			} else {
				children = append(children, ui.HSpacer(gap))
			}
		}
		if compact {
			children = append(children, card)
		} else {
			children = append(children, ui.Expanded(card))
		}
	}

	if compact {
		return ui.Column(children...)
	}
	return ui.Row(children...)
}

func sectionHeader(title, subtitle string) ui.Widget {
	return ui.Column(
		ui.Text(title, ui.TextSize(16)),
		ui.Padding(
			ui.Insets{Top: 4},
			ui.Text(subtitle, ui.TextSize(12), ui.TextColor(ui.NRGBA(102, 112, 128, 255))),
		),
	)
}

func statusBadge(th *ui.Theme, status string) ui.Widget {
	label := statusLabel(status)
	return badge(label, ui.NRGBA(255, 255, 255, 255), statusColor(th, status))
}

func priorityBadge(priority string) ui.Widget {
	label := priorityLabel(priority)
	return badge(label, ui.NRGBA(255, 255, 255, 255), priorityColor(priority))
}

func blockedBadge() ui.Widget {
	return badge("阻塞", ui.NRGBA(35, 39, 46, 255), warnColor())
}

func badge(label string, foreground, background color.NRGBA) ui.Widget {
	return ui.Container(
		ui.Style{
			Background: background,
			Padding:    ui.Symmetric(4, 8),
			Radius:     999,
		},
		ui.Text(label, ui.TextSize(11), ui.TextColor(foreground)),
	)
}

func statusColor(th *ui.Theme, status string) color.NRGBA {
	switch status {
	case "todo":
		return neutralColor()
	case "active":
		return th.Primary
	case "review":
		return warnColor()
	case "done":
		return successColor()
	default:
		return th.SurfaceMuted
	}
}

func priorityColor(priority string) color.NRGBA {
	switch priority {
	case "high":
		return dangerColor()
	case "medium":
		return warnColor()
	default:
		return neutralColor()
	}
}

func priorityRank(priority string) int {
	switch priority {
	case "high":
		return 0
	case "medium":
		return 1
	default:
		return 2
	}
}

func statusRank(status string) int {
	switch status {
	case "active":
		return 0
	case "review":
		return 1
	case "todo":
		return 2
	case "done":
		return 3
	default:
		return 4
	}
}

func normalizePriority(priority string) string {
	switch strings.TrimSpace(strings.ToLower(priority)) {
	case "high":
		return "high"
	case "low":
		return "low"
	default:
		return "medium"
	}
}

func statusLabel(status string) string {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "todo":
		return "待办"
	case "active":
		return "进行中"
	case "review":
		return "评审中"
	case "done":
		return "已完成"
	default:
		return status
	}
}

func priorityLabel(priority string) string {
	switch strings.TrimSpace(strings.ToLower(priority)) {
	case "high":
		return "高"
	case "medium":
		return "中"
	case "low":
		return "低"
	default:
		return priority
	}
}

func pageLabel(page string) string {
	switch strings.TrimSpace(strings.ToLower(page)) {
	case pageWorkspace:
		return "工作台"
	case pageDelivery:
		return "交付"
	case pageSettings:
		return "设置"
	default:
		return page
	}
}

func blockActionLabel(blocked bool) string {
	if blocked {
		return "解除阻塞"
	}
	return "标记阻塞"
}

func appendActivity(entries []string, message string) []string {
	next := append([]string{}, entries...)
	next = append(next, stamped(message))
	if len(next) > 14 {
		next = next[len(next)-14:]
	}
	return next
}

func stamped(message string) string {
	return time.Now().Format("15:04:05") + "  " + message
}

func maxFloat32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func clamp01(value float32) float32 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func withAlpha(col color.NRGBA, alpha uint8) color.NRGBA {
	col.A = alpha
	return col
}

func neutralColor() color.NRGBA {
	return ui.NRGBA(71, 85, 105, 255)
}

func successColor() color.NRGBA {
	return ui.NRGBA(22, 163, 74, 255)
}

func warnColor() color.NRGBA {
	return ui.NRGBA(245, 158, 11, 255)
}

func dangerColor() color.NRGBA {
	return ui.NRGBA(220, 38, 38, 255)
}

func infoColor() color.NRGBA {
	return ui.NRGBA(102, 102, 102, 255)
}
