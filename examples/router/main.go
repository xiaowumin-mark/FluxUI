package main

import (
	"fmt"
	"strings"
	"time"

	statepkg "github.com/xiaowumin-mark/FluxUI/state"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type user struct {
	ID    string
	Name  string
	Role  string
	Email string
}

func samples() []user {
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

func App(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)
	allowSettings := ui.UseState(ctx, true)
	routeLog := ui.UseState(ctx, "路由示例已启动")
	quickUserID := ui.UseState(ctx, "u1001")
	users := ui.UseState(ctx, samples())

	currentPath := ui.CurrentPath(ctx)
	if currentPath == "" {
		currentPath = "/"
	}

	return ui.ContainerDecorationElement(
		ui.Bg(th.Surface).WithPad(ui.All(12)),
		ui.ColumnElement(
			ui.ContainerDecorationElement(
				ui.Bg(th.Primary).WithPad(ui.Symmetric(12, 14)).WithRad(10),
				ui.RowElement(
					ui.TextElement("FluxUI Router 全面示例", ui.TextSize(16), ui.TextColor(th.TextOnPrimary)),
					ui.ExpandedElement(ui.SpacerElement(0, 0)),
					ui.ButtonElement(
						ui.TextElement("Back"),
						ui.ButtonPadding(ui.Symmetric(4, 10)),
						ui.ButtonBackground(ui.NRGBA(255, 255, 255, 30)),
						ui.ButtonForeground(th.TextOnPrimary),
						ui.OnClick(func(ctx *ui.Context) { ui.NavigateBack(ctx, ui.WithNavTransition(ui.TransitionSlideRight)) }),
					),
				),
			),
			ui.SpacerElement(0, 10),
			ui.CardElement(
				ui.ColumnElement(
					ui.TextElement(fmt.Sprintf("当前路径: %s", currentPath), ui.TextSize(13)),
					ui.SpacerElement(0, 4),
					ui.TextElement(fmt.Sprintf("栈深度: %d | 可返回: %v", ui.StackDepth(ctx), ui.CanGoBack(ctx)), ui.TextSize(12), ui.TextColor(th.SurfaceMuted)),
					ui.SpacerElement(0, 8),
					ui.RowElement(
						ui.ButtonElement(ui.TextElement("首页"), ui.ButtonPadding(ui.Symmetric(4, 10)), ui.OnClick(func(ctx *ui.Context) { ui.Navigate(ctx, "/") })),
						ui.SpacerElement(6, 0),
						ui.ButtonElement(ui.TextElement("用户列表"), ui.ButtonPadding(ui.Symmetric(4, 10)), ui.OnClick(func(ctx *ui.Context) { ui.Navigate(ctx, "/users") })),
						ui.SpacerElement(6, 0),
						ui.ButtonElement(ui.TextElement("设置"), ui.ButtonPadding(ui.Symmetric(4, 10)), ui.OnClick(func(ctx *ui.Context) {
							if allowSettings.Value() {
								ui.Navigate(ctx, "/settings")
							} else {
								routeLog.Set("守卫拦截: 当前不允许进入设置页")
							}
						})),
						ui.SpacerElement(6, 0),
						ui.ButtonElement(ui.TextElement("未知页面"), ui.ButtonPadding(ui.Symmetric(4, 10)), ui.OnClick(func(ctx *ui.Context) { ui.Navigate(ctx, "/missing/path") })),
					),
					ui.SpacerElement(0, 10),
					ui.CheckboxElement(
						"允许进入设置页",
						allowSettings.Value(),
						ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) { allowSettings.Set(checked) }),
					),
					ui.SpacerElement(0, 8),
					ui.TextElement(routeLog.Value(), ui.TextSize(12), ui.TextColor(ui.NRGBA(30, 64, 175, 255))),
				),
				ui.CardPadding(ui.All(12)),
			),
			ui.SpacerElement(0, 10),
			ui.ExpandedElement(
				ui.CardElement(
					ui.PaddingElement(
						ui.All(12),
						ui.RouterElement(
							ui.RouteElement("/", HomePage(users, quickUserID), ui.RouteName("home"), ui.RouteTitle("首页")),
							ui.RouteElement("/users", UserListPage(users, th), ui.RouteName("users"), ui.RouteTitle("用户列表")),
							ui.RouteElement("/users/:id", UserDetailPage(users, routeLog), ui.RouteName("user-detail"), ui.RouteTitle("用户详情"), ui.RouteMeta("section", "users")),
							ui.RouteElement("/settings", SettingsPage(), ui.RouteName("settings"), ui.RouteTitle("设置"), ui.RouteBeforeEnter(func(ctx *ui.Context, from, to string) bool {
								if !allowSettings.Value() {
									routeLog.Set("路由守卫拦截: " + from + " -> " + to)
									return false
								}
								return true
							})),
						).With(
							ui.RouterTransition(ui.TransitionSlideLeft),
							ui.RouterTransitionDuration(220*time.Millisecond),
							ui.RouterNotFoundElement(NotFoundPage),
						),
					),
					ui.CardBorder(th.SurfaceMuted, 1),
				),
			),
		),
	)
}

func HomePage(users *statepkg.State[[]user], quickUserID *statepkg.State[string]) ui.Component {
	return func(ctx *ui.Context) ui.Element {
		navigate := ui.UseNavigate(ctx)
		return ui.ScrollViewElement(
			ui.ColumnElement(
				ui.TextElement("首页", ui.TextSize(20)),
				ui.SpacerElement(0, 8),
				ui.TextElement("动态路由参数、查询参数、前进/替换/返回、路由守卫和过渡动画。"),
				ui.SpacerElement(0, 12),
				ui.RowElement(
					ui.ButtonElement(ui.TextElement("进入用户列表"), ui.OnClick(func(ctx *ui.Context) { navigate("/users") })),
					ui.SpacerElement(8, 0),
					ui.ButtonElement(ui.TextElement("尝试设置页"), ui.OnClick(func(ctx *ui.Context) { navigate("/settings") })),
				),
				ui.SpacerElement(0, 12),
				ui.TextFieldElement(
					quickUserID.Value(),
					ui.InputPlaceholder("输入用户ID, 例如 u1002"),
					ui.InputOnChange(func(ctx *ui.Context, v string) { quickUserID.Set(v) }),
				),
				ui.SpacerElement(0, 8),
				ui.RowElement(
					ui.ButtonElement(ui.TextElement("打开详情(tab=profile)"), ui.OnClick(func(ctx *ui.Context) {
						gotoUserDetail(navigate, quickUserID.Value(), "profile", ui.TransitionSlideLeft)
					})),
					ui.SpacerElement(8, 0),
					ui.ButtonElement(ui.TextElement("打开详情(tab=activity,淡入)"), ui.OnClick(func(ctx *ui.Context) {
						gotoUserDetail(navigate, quickUserID.Value(), "activity", ui.TransitionFade)
					})),
				),
			),
			ui.ScrollVertical(true),
		)
	}
}

func UserListPage(users *statepkg.State[[]user], th *ui.Theme) ui.Component {
	return func(ctx *ui.Context) ui.Element {
		navigate := ui.UseNavigate(ctx)
		var rows []ui.Element
		for _, item := range users.Value() {
			userItem := item
			rows = append(rows,
				ui.PaddingElement(
					ui.Insets{Bottom: 8},
					ui.CardElement(
						ui.RowElement(
							ui.ExpandedElement(
								ui.ColumnElement(
									ui.TextElement(userItem.Name, ui.TextSize(15)),
									ui.SpacerElement(0, 4),
									ui.TextElement(userItem.ID+" | "+userItem.Role, ui.TextSize(12), ui.TextColor(th.SurfaceMuted)),
								),
							),
							ui.ButtonElement(
								ui.TextElement("详情"),
								ui.ButtonPadding(ui.Symmetric(6, 10)),
								ui.OnClick(func(ctx *ui.Context) {
									gotoUserDetail(navigate, userItem.ID, "profile", ui.TransitionSlideLeft)
								}),
							),
						),
						ui.CardPadding(ui.All(10)),
					),
				),
			)
		}

		return ui.ScrollViewElement(
			ui.ColumnElement(
				ui.TextElement("用户列表", ui.TextSize(18)),
				ui.SpacerElement(0, 8),
				ui.TextElement("点击详情进入动态路由: /users/:id?tab=..."),
				ui.SpacerElement(0, 12),
				ui.ColumnElement(rows...),
			),
			ui.ScrollVertical(true),
		)
	}
}

func UserDetailPage(users *statepkg.State[[]user], routeLog *statepkg.State[string]) ui.Component {
	return func(ctx *ui.Context) ui.Element {
		params := ui.UseParams(ctx)
		route := ui.UseRoute(ctx)
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

		routeLog.Set(fmt.Sprintf("导航通过: /users/%s?tab=%s", userID, tab))

		return ui.ScrollViewElement(
			ui.ColumnElement(
				ui.TextElement("用户详情", ui.TextSize(18)),
				ui.SpacerElement(0, 4),
				ui.TextElement(fmt.Sprintf("路由: %s / %s", route.Name, route.Title), ui.TextSize(12)),
				ui.SpacerElement(0, 8),
				ui.TextElement("路径参数 id: "+userID),
				ui.SpacerElement(0, 4),
				ui.TextElement("查询参数 tab: "+tab),
				ui.SpacerElement(0, 8),
				ui.TextElement("姓名: "+nameText),
				ui.SpacerElement(0, 4),
				ui.TextElement("角色: "+roleText),
				ui.SpacerElement(0, 4),
				ui.TextElement("邮箱: "+emailText),
				ui.SpacerElement(0, 12),
				ui.RowElement(
					ui.ButtonElement(ui.TextElement("Replace -> tab=activity"), ui.OnClick(func(ctx *ui.Context) {
						path := fmt.Sprintf("/users/%s?tab=activity", userID)
						ui.NavigateReplace(ctx, path, ui.WithNavTransition(ui.TransitionFade))
					})),
					ui.SpacerElement(8, 0),
					ui.ButtonElement(ui.TextElement("Back"), ui.OnClick(func(ctx *ui.Context) { ui.NavigateBack(ctx) })),
				),
				ui.SpacerElement(0, 8),
				ui.ButtonElement(ui.TextElement("回到用户列表 (Replace)"), ui.OnClick(func(ctx *ui.Context) {
					ui.NavigateReplace(ctx, "/users")
				})),
			),
			ui.ScrollVertical(true),
		)
	}
}

func SettingsPage() ui.Component {
	return func(ctx *ui.Context) ui.Element {
		return ui.CenterElement(
			ui.ColumnElement(
				ui.TextElement("设置页", ui.TextSize(18)),
				ui.SpacerElement(0, 8),
				ui.TextElement("通过组件内路由守卫控制是否允许进入。"),
				ui.SpacerElement(0, 12),
				ui.ButtonElement(
					ui.TextElement("返回上一页"),
					ui.OnClick(func(ctx *ui.Context) { ui.NavigateBack(ctx) }),
				),
			),
		)
	}
}

func NotFoundPage(ctx *ui.Context) ui.Element {
	return ui.CenterElement(
		ui.ColumnElement(
			ui.TextElement("404", ui.TextSize(22), ui.TextColor(ui.NRGBA(220, 38, 38, 255))),
			ui.SpacerElement(0, 8),
			ui.TextElement("未找到路径: "+ui.CurrentPath(ctx)),
			ui.SpacerElement(0, 12),
			ui.ButtonElement(ui.TextElement("返回首页"), ui.OnClick(func(ctx *ui.Context) {
				ui.NavigateReplace(ctx, "/", ui.WithNavTransition(ui.TransitionSlideRight))
			})),
		),
	)
}

func gotoUserDetail(navigate func(string, ...ui.NavigateOption), id string, tab string, trans ui.Transition) {
	id = strings.TrimSpace(id)
	if id == "" {
		return
	}
	path := fmt.Sprintf("/users/%s?tab=%s", id, tab)
	navigate(path, ui.WithNavTransition(trans))
}

func main() {
	_ = ui.RunElement(App, ui.Title("FluxUI Router Example"), ui.Size(980, 760))
}
