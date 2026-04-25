package main

import (
	"fmt"
	"strings"
	"time"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type user struct {
	ID    string
	Name  string
	Role  string
	Email string
}

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)

		boot := ui.State[bool](ctx)
		allowSettings := ui.State[bool](ctx)
		quickUserID := ui.State[string](ctx)
		routeLog := ui.State[string](ctx)
		users := ui.State[[]user](ctx)

		if !boot.Value() {
			allowSettings.Set(true)
			quickUserID.Set("u1001")
			routeLog.Set("路由示例已启动")
			users.Set(sampleUsers())
			boot.Set(true)
		}

		gotoUserDetail := func(routeCtx *ui.Context, id string, tab string, trans ui.Transition) {
			id = strings.TrimSpace(id)
			if id == "" {
				routeLog.Set("请输入用户 ID")
				return
			}
			path := fmt.Sprintf("/users/%s?tab=%s", id, tab)
			ui.Navigate(routeCtx, path, ui.WithNavTransition(trans))
		}

		routes := []ui.Route{
			{
				Path: "/",
				Builder: func(routeCtx *ui.Context) ui.Widget {
					return ui.ScrollView(
						ui.Column(
							ui.Text("首页", ui.TextSize(20)),
							ui.Padding(
								ui.Insets{Top: 8},
								ui.Text("这个示例覆盖：动态路由参数、查询参数、前进/替换/返回、路由守卫、404 页面和过渡动画。"),
							),
							ui.Padding(
								ui.Insets{Top: 12},
								ui.Row(
									ui.Button(
										ui.Text("进入用户列表"),
										ui.OnClick(func(ctx *ui.Context) {
											ui.Navigate(ctx, "/users")
										}),
									),
									ui.Padding(
										ui.Insets{Left: 8},
										ui.Button(
											ui.Text("尝试设置页"),
											ui.OnClick(func(ctx *ui.Context) {
												ui.Navigate(ctx, "/settings")
											}),
										),
									),
								),
							),
							ui.Padding(
								ui.Insets{Top: 12},
								ui.TextField(
									quickUserID.Value(),
									ui.InputPlaceholder("输入用户ID，例如 u1002"),
									ui.InputOnChange(func(ctx *ui.Context, value string) {
										quickUserID.Set(value)
									}),
								),
							),
							ui.Padding(
								ui.Insets{Top: 8},
								ui.Row(
									ui.Button(
										ui.Text("打开详情(tab=profile)"),
										ui.OnClick(func(ctx *ui.Context) {
											gotoUserDetail(ctx, quickUserID.Value(), "profile", ui.TransitionSlideLeft)
										}),
									),
									ui.Padding(
										ui.Insets{Left: 8},
										ui.Button(
											ui.Text("打开详情(tab=activity,淡入)"),
											ui.OnClick(func(ctx *ui.Context) {
												gotoUserDetail(ctx, quickUserID.Value(), "activity", ui.TransitionFade)
											}),
										),
									),
								),
							),
						),
						ui.ScrollVertical(true),
					)
				},
			},
			{
				Path: "/users",
				Builder: func(routeCtx *ui.Context) ui.Widget {
					rows := make([]ui.Widget, 0, len(users.Value()))
					for _, item := range users.Value() {
						userItem := item
						rows = append(rows,
							ui.Padding(
								ui.Insets{Bottom: 8},
								ui.Card(
									ui.Row(
										ui.Expanded(
											ui.Column(
												ui.Text(userItem.Name, ui.TextSize(15)),
												ui.Padding(ui.Insets{Top: 4}, ui.Text(userItem.ID+" | "+userItem.Role, ui.TextSize(12), ui.TextColor(th.SurfaceMuted))),
											),
										),
										ui.Button(
											ui.Text("详情"),
											ui.ButtonPadding(ui.Symmetric(6, 10)),
											ui.OnClick(func(ctx *ui.Context) {
												gotoUserDetail(ctx, userItem.ID, "profile", ui.TransitionSlideLeft)
											}),
										),
									),
									ui.CardPadding(ui.All(10)),
								),
							),
						)
					}

					return ui.ScrollView(
						ui.Column(
							ui.Text("用户列表", ui.TextSize(18)),
							ui.Padding(ui.Insets{Top: 8}, ui.Text("点击“详情”进入动态路由：/users/:id?tab=...")),
							ui.Padding(ui.Insets{Top: 12}, ui.Column(rows...)),
						),
						ui.ScrollVertical(true),
					)
				},
			},
			{
				Path: "/users/:id",
				Builder: func(routeCtx *ui.Context) ui.Widget {
					params := ui.RouteParams(routeCtx)
					userID := params.Path("id")
					tab := params.Query("tab")
					if tab == "" {
						tab = "overview"
					}

					u, ok := findUser(users.Value(), userID)
					nameText := "未知用户"
					roleText := "-"
					emailText := "-"
					if ok {
						nameText = u.Name
						roleText = u.Role
						emailText = u.Email
					}

					return ui.ScrollView(
						ui.Column(
							ui.Text("用户详情", ui.TextSize(18)),
							ui.Padding(ui.Insets{Top: 8}, ui.Text("路径参数 id: "+userID)),
							ui.Padding(ui.Insets{Top: 4}, ui.Text("查询参数 tab: "+tab)),
							ui.Padding(ui.Insets{Top: 8}, ui.Text("姓名: "+nameText)),
							ui.Padding(ui.Insets{Top: 4}, ui.Text("角色: "+roleText)),
							ui.Padding(ui.Insets{Top: 4}, ui.Text("邮箱: "+emailText)),
							ui.Padding(
								ui.Insets{Top: 12},
								ui.Row(
									ui.Button(
										ui.Text("Replace -> tab=activity"),
										ui.OnClick(func(ctx *ui.Context) {
											path := fmt.Sprintf("/users/%s?tab=activity", userID)
											ui.NavigateReplace(ctx, path, ui.WithNavTransition(ui.TransitionFade))
										}),
									),
									ui.Padding(
										ui.Insets{Left: 8},
										ui.Button(
											ui.Text("Back"),
											ui.OnClick(func(ctx *ui.Context) {
												ui.NavigateBack(ctx)
											}),
										),
									),
								),
							),
							ui.Padding(
								ui.Insets{Top: 8},
								ui.Button(
									ui.Text("回到用户列表（Replace）"),
									ui.OnClick(func(ctx *ui.Context) {
										ui.NavigateReplace(ctx, "/users")
									}),
								),
							),
						),
						ui.ScrollVertical(true),
					)
				},
			},
			{
				Path: "/settings",
				Builder: func(routeCtx *ui.Context) ui.Widget {
					return ui.Center(
						ui.Column(
							ui.Text("设置页", ui.TextSize(18)),
							ui.Padding(ui.Insets{Top: 8}, ui.Text("当前页面可通过路由守卫控制是否允许进入。")),
							ui.Padding(
								ui.Insets{Top: 12},
								ui.Button(
									ui.Text("返回上一页"),
									ui.OnClick(func(ctx *ui.Context) {
										ui.NavigateBack(ctx)
									}),
								),
							),
						),
					)
				},
			},
		}

		routerWidget := ui.Router(
			ctx,
			routes,
			ui.RouterTransition(ui.TransitionSlideLeft),
			ui.RouterTransitionDuration(260*time.Millisecond),
			ui.RouterBeforeEach(func(ctx *ui.Context, from, to string) bool {
				if to == "/settings" && !allowSettings.Value() {
					routeLog.Set(fmt.Sprintf("守卫拦截：%s -> %s（未授权）", from, to))
					return false
				}
				routeLog.Set(fmt.Sprintf("导航通过：%s -> %s", from, to))
				return true
			}),
			ui.RouterNotFound(func(routeCtx *ui.Context) ui.Widget {
				return ui.Center(
					ui.Column(
						ui.Text("404", ui.TextSize(28), ui.TextColor(ui.NRGBA(220, 38, 38, 255))),
						ui.Padding(ui.Insets{Top: 8}, ui.Text("未找到路由: "+ui.CurrentPath(routeCtx))),
						ui.Padding(
							ui.Insets{Top: 12},
							ui.Button(
								ui.Text("回到首页"),
								ui.OnClick(func(ctx *ui.Context) {
									ui.NavigateReplace(ctx, "/")
								}),
							),
						),
					),
				)
			}),
		)

		currentPath := ui.CurrentPath(ctx)
		if currentPath == "" {
			currentPath = "/"
		}

		return ui.Container(
			ui.Style{
				Background: th.Surface,
				Padding:    ui.All(12),
			},
			ui.Column(
				ui.AppBar(
					ui.Text("FluxUI Router 全面示例", ui.TextSize(16)),
					ui.AppBarActions(
						ui.Button(
							ui.Text("Back"),
							ui.ButtonPadding(ui.Symmetric(4, 10)),
							ui.OnClick(func(ctx *ui.Context) {
								ui.NavigateBack(ctx, ui.WithNavTransition(ui.TransitionSlideRight))
							}),
						),
					),
				),
				ui.Padding(
					ui.Insets{Top: 10},
					ui.Card(
						ui.Column(
							ui.Text(fmt.Sprintf("当前路径: %s", currentPath), ui.TextSize(13)),
							ui.Padding(
								ui.Insets{Top: 4},
								ui.Text(fmt.Sprintf("栈深度: %d | 可返回: %v", ui.StackDepth(ctx), ui.CanGoBack(ctx)), ui.TextSize(12), ui.TextColor(th.SurfaceMuted)),
							),
							ui.Padding(
								ui.Insets{Top: 8},
								ui.Row(
									ui.Button(
										ui.Text("首页"),
										ui.ButtonPadding(ui.Symmetric(4, 10)),
										ui.OnClick(func(ctx *ui.Context) {
											ui.Navigate(ctx, "/")
										}),
									),
									ui.Padding(
										ui.Insets{Left: 6},
										ui.Button(
											ui.Text("用户列表"),
											ui.ButtonPadding(ui.Symmetric(4, 10)),
											ui.OnClick(func(ctx *ui.Context) {
												ui.Navigate(ctx, "/users")
											}),
										),
									),
									ui.Padding(
										ui.Insets{Left: 6},
										ui.Button(
											ui.Text("设置"),
											ui.ButtonPadding(ui.Symmetric(4, 10)),
											ui.OnClick(func(ctx *ui.Context) {
												ui.Navigate(ctx, "/settings")
											}),
										),
									),
									ui.Padding(
										ui.Insets{Left: 6},
										ui.Button(
											ui.Text("未知路由"),
											ui.ButtonPadding(ui.Symmetric(4, 10)),
											ui.OnClick(func(ctx *ui.Context) {
												ui.Navigate(ctx, "/missing/path")
											}),
										),
									),
								),
							),
							ui.Padding(
								ui.Insets{Top: 10},
								ui.Row(
									ui.Checkbox(
										"允许进入设置页",
										allowSettings.Value(),
										ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) {
											allowSettings.Set(checked)
										}),
									),
								),
							),
							ui.Padding(
								ui.Insets{Top: 8},
								ui.Text(routeLog.Value(), ui.TextSize(12), ui.TextColor(ui.NRGBA(30, 64, 175, 255))),
							),
						),
						ui.CardPadding(ui.All(12)),
					),
				),
				ui.Padding(
					ui.Insets{Top: 10},
					ui.Expanded(
						ui.Card(
							ui.Fill(
								ui.Padding(
									ui.All(12),
									routerWidget,
								),
							),
							ui.CardBorder(th.SurfaceMuted, 1),
						),
					),
				),
			),
		)
	}, ui.Title("FluxUI Router Example"), ui.Size(980, 760))
}

func sampleUsers() []user {
	return []user{
		{ID: "u1001", Name: "Ava", Role: "产品经理", Email: "ava@fluxui.dev"},
		{ID: "u1002", Name: "Noah", Role: "后端工程师", Email: "noah@fluxui.dev"},
		{ID: "u1003", Name: "Mia", Role: "前端工程师", Email: "mia@fluxui.dev"},
		{ID: "u1004", Name: "Liam", Role: "测试工程师", Email: "liam@fluxui.dev"},
	}
}

func findUser(users []user, id string) (user, bool) {
	for _, item := range users {
		if item.ID == id {
			return item, true
		}
	}
	return user{}, false
}
