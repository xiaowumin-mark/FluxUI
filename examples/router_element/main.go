package main

import (
	"fmt"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	_ = ui.RunElement(App, ui.Title("FluxUI RouterElement Example"), ui.Size(720, 520))
}

func App(ctx *ui.Context) ui.Element {
	th := ui.UseTheme(ctx)

	return ui.ContainerElement(
		ui.Style{
			Background: th.Surface,
			Padding:    ui.All(16),
		},
		ui.ColumnElement(
			ui.TextElement("RouterElement comparison", ui.TextSize(24)),
			ui.PaddingElement(
				ui.Insets{Top: 8},
				ui.TextElement("This example mirrors the legacy router example with React-style route components and hooks."),
			),
			ui.PaddingElement(
				ui.Insets{Top: 16},
				ui.RouterElement(
					ui.RouteElement("/", HomePage),
					ui.RouteElement("/users/:id", UserPage),
					ui.RouteElement("/settings", SettingsPage),
				),
			),
		),
	)
}

func HomePage(ctx *ui.Context) ui.Element {
	location := ui.UseLocation(ctx)
	navigate := ui.UseNavigate(ctx)

	return pageShell(
		"Home",
		ui.TextElement("UseNavigate drives RouterElement routes from function components."),
		routeInfo(location),
		buttonRow(
			linkButton("Open user u1001", func() {
				navigate("/users/u1001?tab=profile", ui.WithNavTransition(ui.TransitionSlideLeft))
			}),
			linkButton("Open user u1002", func() {
				navigate("/users/u1002?tab=activity", ui.WithNavTransition(ui.TransitionFade))
			}),
			linkButton("Settings", func() {
				navigate("/settings")
			}),
		),
	)
}

func UserPage(ctx *ui.Context) ui.Element {
	location := ui.UseLocation(ctx)
	params := ui.UseParams(ctx)
	navigate := ui.UseNavigate(ctx)

	userID := params.Get("id")
	tab := location.Query("tab")
	if tab == "" {
		tab = "overview"
	}

	return pageShell(
		"User detail",
		ui.TextElement("UseParams reads :id and UseLocation reads the query string."),
		routeInfo(location),
		ui.TextElement(fmt.Sprintf("Path param id: %s", userID)),
		ui.TextElement(fmt.Sprintf("Query tab: %s", tab)),
		ui.TextElement("Route component identity defaults to the route pattern unless RouteKey changes.", ui.TextSize(13)),
		buttonRow(
			linkButton("Replace tab=activity", func() {
				navigate(fmt.Sprintf("/users/%s?tab=activity", userID), ui.WithNavTransition(ui.TransitionFade))
			}),
			linkButton("Open u1002", func() {
				navigate("/users/u1002?tab=profile", ui.WithNavTransition(ui.TransitionSlideLeft))
			}),
			linkButton("Home", func() {
				navigate("/")
			}),
		),
	)
}

func SettingsPage(ctx *ui.Context) ui.Element {
	location := ui.UseLocation(ctx)
	navigate := ui.UseNavigate(ctx)

	return pageShell(
		"Settings",
		ui.TextElement("This route is declared with RouteElement and rendered by RouterElement."),
		routeInfo(location),
		buttonRow(
			linkButton("Home", func() {
				navigate("/")
			}),
			linkButton("User u1001", func() {
				navigate("/users/u1001?tab=settings")
			}),
		),
	)
}

func pageShell(title string, children ...ui.Element) ui.Element {
	items := make([]ui.Element, 0, len(children)+1)
	items = append(items, ui.TextElement(title, ui.TextSize(20)))
	items = append(items, children...)

	return ui.ContainerElement(
		ui.Style{
			Background: ui.NRGBA(255, 255, 255, 255),
			Padding:    ui.All(16),
			Radius:     10,
		},
		ui.ColumnElement(items...),
	)
}

func routeInfo(location *ui.Location) ui.Element {
	path := ""
	pathname := ""
	if location != nil {
		path = location.Path
		pathname = location.Pathname
	}

	return ui.ColumnElement(
		ui.PaddingElement(ui.Insets{Top: 8}, ui.TextElement("Location path: "+path, ui.TextSize(13))),
		ui.PaddingElement(ui.Insets{Top: 4}, ui.TextElement("Location pathname: "+pathname, ui.TextSize(13))),
	)
}

func buttonRow(children ...ui.Element) ui.Element {
	spaced := make([]ui.Element, 0, len(children)*2)
	for index, child := range children {
		if index > 0 {
			spaced = append(spaced, ui.SpacerElement(8, 0))
		}
		spaced = append(spaced, child)
	}

	return ui.PaddingElement(ui.Insets{Top: 12}, ui.RowElement(spaced...))
}

func linkButton(label string, onClick func()) ui.Element {
	return ui.ButtonElement(
		ui.TextElement(label),
		ui.ButtonPadding(ui.Symmetric(6, 10)),
		ui.OnClick(func(ctx *ui.Context) {
			onClick()
		}),
	)
}
