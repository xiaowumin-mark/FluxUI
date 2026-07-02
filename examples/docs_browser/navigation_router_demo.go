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
				ui.TextElement("Docs AppBar", ui.TextSize(14), ui.TextColor(th.Colors.OnSurface)),
				nil,
				[]ui.Element{
					ui.ButtonElement(
						ui.TextElement("Action"),
						ui.ButtonPadding(ui.Symmetric(4, 8)),
						ui.OnClick(func(ctx *ui.Context) {
							actionCount.Set(actionCount.Value() + 1)
						}),
					),
				},
			),
			ui.VSpacerElement(8),
			ui.AppBarElement(
				ui.TextElement("Configured AppBar", ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
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
				ui.TextElement(fmt.Sprintf("Action clicks: %d", actionCount.Value()), ui.TextSize(13), ui.TextColor(th.Colors.OnSurface)),
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
						ui.TextElement("Current page: "+value.Value(), ui.TextSize(14), ui.TextColor(th.Colors.OnSurface)),
					),
				),
				ui.RowElement(
					docsDemoControlButton("Set docs via ref", func(ctx *ui.Context) {
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
						{Key: "home", Label: "Home", Icon: ui.IconElement("home", ui.IconSize(18))},
						{Key: "docs", Label: "Docs", Icon: ui.IconElement("description", ui.IconSize(18))},
						{Key: "profile", Label: "Profile", Icon: ui.IconElement("person", ui.IconSize(18))},
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
						{Key: "home", Label: "Home", Icon: ui.IconElement("home")},
						{Key: "search", Label: "Search", Icon: ui.IconElement("search")},
						{Key: "settings", Label: "Settings", Icon: ui.IconElement("settings")},
					},
					ui.NavigationRailWidth(96),
					ui.NavigationRailHeader(ui.Text("Menu", ui.TextSize(12), ui.TextColor(th.Colors.OnSurface))),
					ui.NavigationRailFooter(ui.Text("v1", ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant))),
					ui.NavigationRailActiveColor(th.Colors.Primary),
					ui.NavigationRailInactiveColor(th.Colors.OnSurfaceVariant),
					ui.NavigationRailDecoration(ui.Bg(th.Colors.SurfaceContainerLow).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant})),
					ui.NavigationRailOnChange(func(ctx *ui.Context, key string) {
						value.Set(key)
					}),
				),
				ui.ExpandedElement(
					ui.CenterElement(ui.TextElement("Rail page: "+value.Value(), ui.TextSize(14), ui.TextColor(th.Colors.OnSurface))),
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
						{Key: "inbox", Label: "Inbox", Icon: ui.IconElement("inbox")},
						{Key: "sent", Label: "Sent", Icon: ui.IconElement("send")},
						{Key: "drafts", Label: "Drafts", Icon: ui.IconElement("drafts")},
					},
					ui.NavigationDrawerWidth(280),
					ui.NavigationDrawerHeader(ui.Text("Mailbox", ui.TextSize(14), ui.TextColor(th.Colors.OnSurface))),
					ui.NavigationDrawerFooter(ui.Text("3 folders", ui.TextSize(11), ui.TextColor(th.Colors.OnSurfaceVariant))),
					ui.NavigationDrawerActiveColor(th.Colors.Primary),
					ui.NavigationDrawerInactiveColor(th.Colors.OnSurfaceVariant),
					ui.NavigationDrawerDecoration(ui.Bg(th.Colors.SurfaceContainerLow).WithBorder(ui.Border{Width: 1, Color: th.Colors.OutlineVariant})),
					ui.NavigationDrawerOnChange(func(ctx *ui.Context, key string) {
						value.Set(key)
					}),
				),
				ui.ExpandedElement(
					ui.CenterElement(ui.TextElement("Drawer page: "+value.Value(), ui.TextSize(14), ui.TextColor(th.Colors.OnSurface))),
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
		return "unknown"
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
		title := "untitled"
		if route != nil && route.Title != "" {
			title = route.Title
		}
		return ui.ColumnElement(
			ui.TextElement("RouterElement home", ui.TextSize(14), ui.TextColor(th.Colors.OnSurface)),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement("location: "+path, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement("route title: "+title, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			ui.PaddingElement(
				ui.Insets{Top: 8},
				ui.TextFieldElement(
					userID.Value(),
					ui.InputPlaceholder("user id, e.g. u1002"),
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
						ui.TextElement("Detail"),
						ui.ButtonPadding(ui.Symmetric(4, 8)),
						ui.OnClick(func(ctx *ui.Context) {
							id := strings.TrimSpace(userID.Value())
							if id == "" {
								log.Set("empty user id")
								return
							}
							navigate("/user/"+id+"?tab=profile", ui.WithNavTransition(ui.TransitionSlideLeft))
						}),
					),
					ui.PaddingElement(
						ui.Insets{Left: 6},
						ui.ButtonElement(
							ui.TextElement("Settings"),
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
		routeName := "unnamed"
		section := "none"
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
			ui.TextElement("User Detail", ui.TextSize(14), ui.TextColor(th.Colors.OnSurface)),
			ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("id: "+id, ui.TextColor(th.Colors.OnSurface))),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement("name: "+findUserName(id), ui.TextColor(th.Colors.OnSurface))),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement("tab: "+tab, ui.TextColor(th.Colors.OnSurface))),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement(fmt.Sprintf("route=%s section=%s", routeName, section), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement(fmt.Sprintf("params path=%d query=%d location query=%d", len(params.AllPathParams()), len(params.AllQueryParams()), locationQueries), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement(fmt.Sprintf("has id=%t has tab=%t can back=%t", params.HasParam("id"), params.HasQuery("tab"), canBack), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			ui.PaddingElement(
				ui.Insets{Top: 8},
				ui.RowElement(
					ui.ButtonElement(
						ui.TextElement("Replace tab=activity"),
						ui.ButtonPadding(ui.Symmetric(4, 8)),
						ui.OnClick(func(ctx *ui.Context) {
							ui.NavigateReplace(ctx, "/user/"+id+"?tab=activity", ui.WithNavTransition(ui.TransitionFade))
						}),
					),
					ui.PaddingElement(
						ui.Insets{Left: 6},
						ui.ButtonElement(
							ui.TextElement("Back"),
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
		routeTitle := "Settings"
		if route != nil && route.Title != "" {
			routeTitle = route.Title
		}
		return ui.ColumnElement(
			ui.TextElement("Settings", ui.TextSize(14), ui.TextColor(th.Colors.OnSurface)),
			ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("RouteBeforeEnter guards this route.", ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement("route title: "+routeTitle, ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
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
			ui.PaddingElement(ui.Insets{Top: 6}, ui.TextElement("path: "+ui.CurrentPath(routeCtx), ui.TextSize(12), ui.TextColor(th.Colors.OnSurfaceVariant))),
			ui.PaddingElement(
				ui.Insets{Top: 8},
				ui.ButtonElement(
					ui.TextElement("Go Home"),
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
		ui.RouteElement("/settings", settingsPage, ui.RouteName("settings"), ui.RouteTitle("Settings"), ui.RouteBeforeEnter(func(ctx *ui.Context, from, to string) bool {
			if !allowSettings.Value() {
				log.Set("RouteBeforeEnter blocked: " + from + " -> " + to)
				return false
			}
			return true
		})),
	).With(
		ui.RouterTransition(ui.TransitionSlideLeft),
		ui.RouterTransitionDuration(220*time.Millisecond),
		ui.RouterBeforeEach(func(ctx *ui.Context, from, to string) bool {
			log.Set("navigated: " + from + " -> " + to)
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
				"allow settings",
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
