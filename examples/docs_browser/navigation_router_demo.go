package main

import (
	"fmt"
	"strings"
	"time"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func docsAppBarDemo(actionCount docsIntState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		return ui.ColumnElement(
			ui.AppBarElementWithSlots(
				ui.TextElement("文档 AppBar", ui.TextSize(14), ui.TextColor(th.Colors.OnSurface)),
				nil,
				[]ui.Element{
					ui.ButtonElement(
						ui.TextElement("操作"),
						ui.ButtonPadding(ui.Symmetric(4, 8)),
						ui.OnClick(func(ctx *ui.Context) {
							actionCount.Set(actionCount.Value() + 1)
						}),
					),
				},
			),
			ui.VSpacerElement(8),
			ui.AppBarElement(
				ui.TextElement("已配置的 AppBar", ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
				ui.AppBarLeading(ui.Icon("arrow_back", ui.IconSize(16))),
				ui.AppBarActions(
					ui.IconButton(ui.Icon("search"), ui.IconButtonOnClick(func(ctx *ui.Context) {
						actionCount.Set(actionCount.Value() + 1)
					})),
				),
				ui.AppBarHeight(48),
				ui.AppBarBackground(th.Colors.SurfaceContainerLow),
				ui.AppBarDecoration(ui.Bg(th.Colors.SurfaceContainerLow).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant})),
			),
			ui.PaddingElement(
				ui.Insets{Top: 8},
				ui.TextElement(fmt.Sprintf("操作点击：%d", actionCount.Value()), ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
			),
		)
	})
}

func docsBottomNavigationDemo(value docsStringState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		ref := ui.UseRef(ctx, ui.NewBottomNavRef())
		if ref.Current == nil {
			ref.Current = ui.NewBottomNavRef()
		}
		return ui.FixedHeightElement(
			210,
			ui.ColumnElement(
				ui.ExpandedElement(
					ui.CenterElement(
						ui.TextElement("当前页面："+value.Value(), ui.TextSize(14), ui.TextColor(th.Colors.OnSurface)),
					),
				),
				ui.RowElement(
					docsDemoControlButton("通过引用设置文档", func(ctx *ui.Context) {
						value.Set("docs")
						ref.Current.SetActive("docs")
					}),
					ui.HSpacerElement(8),
					ui.TextElement("BottomNavRef.SetActive", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
				),
				ui.VSpacerElement(8),
				ui.BottomNavigationElement(
					value.Value(),
					[]ui.ElementNavItem{
						{Key: "home", Label: "首页", Icon: ui.IconElement("home", ui.IconSize(18))},
						{Key: "docs", Label: "文档", Icon: ui.IconElement("description", ui.IconSize(18))},
						{Key: "profile", Label: "个人资料", Icon: ui.IconElement("person", ui.IconSize(18))},
					},
					ui.BottomNavAlignmentOf(ui.BottomNavAlignSpaceEvenly),
					ui.BottomNavBackground(th.Colors.SurfaceContainerLow),
					ui.BottomNavActiveColor(th.Colors.Primary),
					ui.BottomNavInactiveColor(th.Colors.OnSurfaceVariant),
					ui.BottomNavDecoration(ui.Bg(th.Colors.SurfaceContainerLow).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant})),
					ui.BottomNavAttachRef(ref.Current),
					ui.BottomNavOnChange(func(ctx *ui.Context, key string) {
						value.Set(key)
					}),
				),
			),
		)
	})
}

func docsNavigationRailDemo(value docsStringState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		return ui.FixedHeightElement(
			240,
			ui.RowElement(
				ui.NavigationRailElement(
					value.Value(),
					[]ui.ElementNavItem{
						{Key: "home", Label: "首页", Icon: ui.IconElement("home")},
						{Key: "search", Label: "搜索", Icon: ui.IconElement("search")},
						{Key: "settings", Label: "设置", Icon: ui.IconElement("settings")},
					},
					ui.NavigationRailWidth(96),
					ui.NavigationRailHeader(ui.Text("菜单", ui.TextSize(12), ui.TextColor(th.Colors.OnSurface))),
					ui.NavigationRailFooter(ui.Text("v1", ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant))),
					ui.NavigationRailActiveColor(th.Colors.Primary),
					ui.NavigationRailInactiveColor(th.Colors.OnSurfaceVariant),
					ui.NavigationRailDecoration(ui.Bg(th.Colors.SurfaceContainerLow).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant})),
					ui.NavigationRailOnChange(func(ctx *ui.Context, key string) {
						value.Set(key)
					}),
				),
				ui.ExpandedElement(
					ui.CenterElement(ui.TextElement("Rail 页面："+value.Value(), ui.TextSize(14), ui.TextColor(th.Colors.OnSurface))),
				),
			),
		)
	})
}

func docsNavigationDrawerDemo(value docsStringState) ui.Element {
	return ui.ComponentElement(func(ctx *ui.Context) ui.Element {
		th := ui.UseTheme(ctx)
		return ui.FixedHeightElement(
			240,
			ui.RowElement(
				ui.NavigationDrawerElement(
					value.Value(),
					[]ui.ElementNavItem{
						{Key: "inbox", Label: "收件箱", Icon: ui.IconElement("inbox")},
						{Key: "sent", Label: "已发送", Icon: ui.IconElement("send")},
						{Key: "drafts", Label: "草稿箱", Icon: ui.IconElement("drafts")},
					},
					ui.NavigationDrawerWidth(280),
					ui.NavigationDrawerHeader(ui.Text("邮箱", ui.TextSize(14), ui.TextColor(th.Colors.OnSurface))),
					ui.NavigationDrawerFooter(ui.Text("3 个文件夹", ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant))),
					ui.NavigationDrawerActiveColor(th.Colors.Primary),
					ui.NavigationDrawerInactiveColor(th.Colors.OnSurfaceVariant),
					ui.NavigationDrawerDecoration(ui.Bg(th.Colors.SurfaceContainerLow).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant})),
					ui.NavigationDrawerOnChange(func(ctx *ui.Context, key string) {
						value.Set(key)
					}),
				),
				ui.ExpandedElement(
					ui.CenterElement(ui.TextElement("Drawer 页面："+value.Value(), ui.TextSize(14), ui.TextColor(th.Colors.OnSurface))),
				),
			),
		)
	})
}

func docsRouterDemo(
	ctx *ui.Context,
	allowSettings docsBoolState,
	userID docsStringState,
	log docsStringState,
) ui.Element {
	th := ui.UseTheme(ctx)
	routerUsers := []struct {
		ID   string
		Name string
	}{
		{ID: "u1001", Name: "Ava"},
		{ID: "u1002", Name: "Noah"},
		{ID: "u1003", Name: "Mia"},
	}
	findUserName := func(id string) string {
		for _, item := range routerUsers {
			if item.ID == id {
				return item.Name
			}
		}
		return "未知"
	}

	homePage := func(routeCtx *ui.Context) ui.Element {
		th := ui.UseTheme(routeCtx)
		navigate := ui.UseNavigate(routeCtx)
		location := ui.UseLocation(routeCtx)
		route := ui.UseRoute(routeCtx)
		path := "/"
		if location != nil && location.Pathname != "" {
			path = location.Pathname
		}
		title := "未命名"
		if route != nil && route.Title != "" {
			title = route.Title
		}
		return ui.ColumnElement(
			ui.TextElement("RouterElement 首页", ui.TextSize(14), ui.TextColor(th.Colors.OnSurface)),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement("位置："+path, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement("路由标题："+title, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			ui.PaddingElement(
				ui.Insets{Top: 8},
				ui.TextFieldElement(
					userID.Value(),
					ui.InputPlaceholder("用户 ID，如 u1002"),
					ui.InputSingleLine(true),
					ui.InputOnChange(func(ctx *ui.Context, value string) {
						userID.Set(value)
					}),
				),
			),
			ui.PaddingElement(
				ui.Insets{Top: 8},
				ui.RowElement(
					ui.ButtonElement(
						ui.TextElement("详情"),
						ui.ButtonPadding(ui.Symmetric(4, 8)),
						ui.OnClick(func(ctx *ui.Context) {
							id := strings.TrimSpace(userID.Value())
							if id == "" {
								log.Set("空用户 ID")
								return
							}
							navigate("/user/"+id+"?tab=profile", ui.WithNavTransition(ui.TransitionSlideLeft))
						}),
					),
					ui.PaddingElement(
						ui.Insets{Left: 6},
						ui.ButtonElement(
							ui.TextElement("设置"),
							ui.ButtonPadding(ui.Symmetric(4, 8)),
							ui.OnClick(func(ctx *ui.Context) {
								ui.Navigate(ctx, "/settings", ui.WithNavTransition(ui.TransitionFade))
							}),
						),
					),
					ui.PaddingElement(
						ui.Insets{Left: 6},
						ui.ButtonElement(
							ui.TextElement("404"),
							ui.ButtonPadding(ui.Symmetric(4, 8)),
							ui.OnClick(func(ctx *ui.Context) {
								navigate("/not-found")
							}),
						),
					),
				),
			),
		)
	}

	userPage := func(routeCtx *ui.Context) ui.Element {
		th := ui.UseTheme(routeCtx)
		params := ui.UseParams(routeCtx)
		location := ui.UseLocation(routeCtx)
		route := ui.UseRoute(routeCtx)
		id := params.Path("id")
		tab := "overview"
		if params.Query("tab") != "" {
			tab = params.Query("tab")
		} else if location != nil && location.Query("tab") != "" {
			tab = location.Query("tab")
		}
		routeName := "未命名"
		section := "无"
		if route != nil {
			if route.Name != "" {
				routeName = route.Name
			}
			if route.Meta != nil {
				if value, ok := route.Meta["section"].(string); ok && value != "" {
					section = value
				}
			}
		}
		locationQueries := 0
		if location != nil {
			locationQueries = len(location.AllQueryParams())
		}
		canBack := ui.CanGoBack(routeCtx)
		return ui.ColumnElement(
			ui.TextElement("用户详情", ui.TextSize(14), ui.TextColor(th.Colors.OnSurface)),
			ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("ID："+id, ui.TextColor(th.Colors.OnSurface))),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement("名称："+findUserName(id), ui.TextColor(th.Colors.OnSurface))),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement("标签："+tab, ui.TextColor(th.Colors.OnSurface))),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement(fmt.Sprintf("route=%s section=%s", routeName, section), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement(fmt.Sprintf("params path=%d query=%d location query=%d", len(params.AllPathParams()), len(params.AllQueryParams()), locationQueries), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement(fmt.Sprintf("has id=%t has tab=%t can back=%t", params.HasParam("id"), params.HasQuery("tab"), canBack), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			ui.PaddingElement(
				ui.Insets{Top: 8},
				ui.RowElement(
					ui.ButtonElement(
						ui.TextElement("替换标签=activity"),
						ui.ButtonPadding(ui.Symmetric(4, 8)),
						ui.OnClick(func(ctx *ui.Context) {
							ui.NavigateReplace(ctx, "/user/"+id+"?tab=activity", ui.WithNavTransition(ui.TransitionFade))
						}),
					),
					ui.PaddingElement(
						ui.Insets{Left: 6},
						ui.ButtonElement(
							ui.TextElement("返回"),
							ui.ButtonPadding(ui.Symmetric(4, 8)),
							ui.OnClick(func(ctx *ui.Context) {
								if ui.CanGoBack(ctx) {
									ui.NavigateBack(ctx, ui.WithNavTransition(ui.TransitionSlideRight))
								}
							}),
						),
					),
				),
			),
		)
	}

	settingsPage := func(routeCtx *ui.Context) ui.Element {
		th := ui.UseTheme(routeCtx)
		route := ui.UseRoute(routeCtx)
		routeTitle := "设置"
		if route != nil && route.Title != "" {
			routeTitle = route.Title
		}
		return ui.ColumnElement(
			ui.TextElement("设置", ui.TextSize(14), ui.TextColor(th.Colors.OnSurface)),
			ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("RouteBeforeEnter 守卫此路由。", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement("路由标题："+routeTitle, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement(fmt.Sprintf("can back=%t", ui.CanGoBack(routeCtx)), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			ui.PaddingElement(
				ui.Insets{Top: 8},
				ui.ButtonElement(
					ui.TextElement("Back"),
					ui.ButtonPadding(ui.Symmetric(4, 8)),
					ui.OnClick(func(ctx *ui.Context) {
						if ui.CanGoBack(ctx) {
							ui.NavigateBack(ctx)
						}
					}),
				),
			),
		)
	}

	notFoundPage := func(routeCtx *ui.Context) ui.Element {
		th := ui.UseTheme(routeCtx)
		navigate := ui.UseNavigate(routeCtx)
		return ui.ColumnElement(
			ui.TextElement("404", ui.TextSize(16), ui.TextColor(th.Colors.Error)),
			ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("路径："+ui.CurrentPath(routeCtx), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			ui.PaddingElement(
				ui.Insets{Top: 8},
				ui.ButtonElement(
					ui.TextElement("回首页"),
					ui.ButtonPadding(ui.Symmetric(4, 8)),
					ui.OnClick(func(ctx *ui.Context) {
						navigate("/", ui.WithNavTransition(ui.TransitionFade))
					}),
				),
			),
		)
	}

	routerElement := ui.RouterElement(
		ui.RouteElement("/", homePage, ui.RouteName("home"), ui.RouteTitle("Home")),
		ui.RouteElement(
			"/user/:id",
			userPage,
			ui.RouteName("user-detail"),
			ui.RouteTitle("User Detail"),
			ui.RouteKey("user-route"),
			ui.RouteMeta("section", "users"),
			ui.RouteMetaMap(map[string]any{
				"detail": true,
				"scope":  "profile",
			}),
		),
		ui.RouteElement("/settings", settingsPage, ui.RouteName("settings"), ui.RouteTitle("设置"), ui.RouteBeforeEnter(func(ctx *ui.Context, from, to string) bool {
			if !allowSettings.Value() {
				log.Set(fmt.Sprintf("RouteBeforeEnter 已阻止：%s -> %s", from, to))
				return false
			}
			return true
		})),
	).With(
		ui.RouterTransition(ui.TransitionSlideLeft),
		ui.RouterTransitionDuration(220*time.Millisecond),
		ui.RouterBeforeEach(func(ctx *ui.Context, from, to string) bool {
			log.Set(fmt.Sprintf("已导航：%s -> %s", from, to))
			return true
		}),
		ui.RouterNotFoundElement(notFoundPage),
	)

	current := ui.CurrentPath(ctx)
	if current == "" {
		current = "/"
	}

	return ui.ColumnElement(
		ui.TextElement(fmt.Sprintf("path=%s | depth=%d", current, ui.StackDepth(ctx)), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		ui.PaddingElement(
			ui.Insets{Top: 6},
			ui.CheckboxElement(
				"允许设置",
				allowSettings.Value(),
				ui.CheckboxOnChange(func(ctx *ui.Context, checked bool) {
					allowSettings.Set(checked)
				}),
			),
		),
		ui.PaddingElement(
			ui.Insets{Top: 6},
			ui.TextElement(log.Value(), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant)),
		),
		ui.PaddingElement(
			ui.Insets{Top: 8},
			ui.ExpandedElement(
				ui.ContainerDecorationElement(
					ui.Bg(th.Colors.SurfaceContainerLow).WithPad(ui.All(8)).WithRad(6),
					routerElement,
				),
			),
		),
	)
}
