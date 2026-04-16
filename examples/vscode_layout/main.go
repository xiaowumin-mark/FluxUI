package main

import (
	"fmt"
	"strings"

	"fluxui/ui"
)

type fileItem struct {
	Name    string
	Content string
}

type sideTool struct {
	Key   string
	Icon  string
	Label string
}

func main() {
	files := []fileItem{
		{
			Name: "main.go",
			Content: strings.Join([]string{
				"package main",
				"",
				"import \"fmt\"",
				"",
				"func main() {",
				"\tfmt.Println(\"Hello FluxUI\")",
				"}",
				"",
			}, "\n"),
		},
		{
			Name: "app.go",
			Content: strings.Join([]string{
				"package app",
				"",
				"type App struct {",
				"\tName string",
				"}",
				"",
				"func New(name string) *App {",
				"\treturn &App{Name: name}",
				"}",
				"",
			}, "\n"),
		},
		{
			Name: "theme.go",
			Content: strings.Join([]string{
				"package theme",
				"",
				"type Theme struct {",
				"\tPrimary string",
				"\tSurface string",
				"}",
				"",
			}, "\n"),
		},
		{
			Name: "README.md",
			Content: strings.Join([]string{
				"# FluxUI Workspace",
				"",
				"- Declarative widgets",
				"- Frame-based rendering",
				"- State + animation + event system",
				"",
			}, "\n"),
		},
	}

	tools := []sideTool{
		{Key: "explorer", Icon: "E", Label: "Explorer"},
		{Key: "search", Icon: "S", Label: "Search"},
		{Key: "git", Icon: "G", Label: "Source"},
		{Key: "run", Icon: "R", Label: "Run"},
	}

	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)

		activeTool := ui.State[string](ctx)
		selectedFile := ui.State[string](ctx)
		codeText := ui.State[string](ctx)

		if activeTool.Value() == "" {
			activeTool.Set("explorer")
		}

		if selectedFile.Value() == "" && len(files) > 0 {
			selectedFile.Set(files[0].Name)
			codeText.Set(files[0].Content)
		}

		topMenu := ui.FixedHeight(
			44,
			ui.Container(
				ui.Style{
					Background: ui.NRGBA(37, 37, 38, 255),
					Padding:    ui.Symmetric(8, 10),
				},
				ui.Row(
					ui.Text("FluxUI IDE", ui.TextColor(ui.NRGBA(229, 229, 229, 255)), ui.TextSize(14)),
					ui.Padding(ui.Insets{Left: 18}, ui.Text("File", ui.TextColor(ui.NRGBA(204, 204, 204, 255)), ui.TextSize(13))),
					ui.Padding(ui.Insets{Left: 12}, ui.Text("Edit", ui.TextColor(ui.NRGBA(204, 204, 204, 255)), ui.TextSize(13))),
					ui.Padding(ui.Insets{Left: 12}, ui.Text("Selection", ui.TextColor(ui.NRGBA(204, 204, 204, 255)), ui.TextSize(13))),
					ui.Padding(ui.Insets{Left: 12}, ui.Text("View", ui.TextColor(ui.NRGBA(204, 204, 204, 255)), ui.TextSize(13))),
					ui.Padding(ui.Insets{Left: 12}, ui.Text("Terminal", ui.TextColor(ui.NRGBA(204, 204, 204, 255)), ui.TextSize(13))),
					ui.Padding(ui.Insets{Left: 12}, ui.Text("Help", ui.TextColor(ui.NRGBA(204, 204, 204, 255)), ui.TextSize(13))),
				),
			),
		)

		leftSidebar := ui.FixedWidth(
			68,
			ui.Container(
				ui.Style{
					Background: ui.NRGBA(45, 45, 48, 255),
					Padding:    ui.Insets{Top: 8, Left: 6, Right: 6, Bottom: 8},
				},
				ui.Column(func() []ui.Widget {
					items := make([]ui.Widget, 0, len(tools)+1)
					for idx := range tools {
						tool := tools[idx]
						isActive := tool.Key == activeTool.Value()

						bg := ui.NRGBA(0, 0, 0, 0)
						fg := ui.NRGBA(166, 166, 166, 255)
						if isActive {
							bg = ui.NRGBA(14, 99, 156, 255)
							fg = ui.NRGBA(255, 255, 255, 255)
						}

						items = append(items,
							ui.Padding(
								ui.Insets{Bottom: 8},
								ui.FillWidth(
									ui.Button(
										ui.Container(
											ui.Style{
												Background: bg,
												Padding:    ui.Symmetric(8, 0),
												Radius:     8,
											},
											ui.Center(
												ui.Text(tool.Icon, ui.TextColor(fg), ui.TextSize(14)),
											),
										),
										ui.ButtonBackground(ui.NRGBA(0, 0, 0, 0)),
										ui.ButtonPadding(ui.All(0)),
										ui.OnClick(func(ctx *ui.Context) {
											activeTool.Set(tool.Key)
										}),
									),
								),
							),
						)
					}
					items = append(items, ui.Expanded(ui.Spacer(0, 0)))
					return items
				}()...),
			),
		)

		filePanel := ui.FixedWidth(
			240,
			ui.Container(
				ui.Style{
					Background: ui.NRGBA(37, 37, 38, 255),
					Padding:    ui.Insets{Top: 12, Left: 10, Right: 10, Bottom: 10},
				},
				ui.Column(
					ui.Text("FILES", ui.TextColor(ui.NRGBA(204, 204, 204, 255)), ui.TextSize(12)),
					ui.Padding(
						ui.Insets{Top: 10},
						ui.Expanded(
							ui.ListView(
								len(files),
								func(ctx *ui.Context, index int) ui.Widget {
									entry := files[index]
									isActive := entry.Name == selectedFile.Value()

									bg := ui.NRGBA(0, 0, 0, 0)
									fg := ui.NRGBA(212, 212, 212, 255)
									if isActive {
										bg = ui.NRGBA(46, 46, 46, 255)
										fg = ui.NRGBA(255, 255, 255, 255)
									}

									return ui.FillWidth(
										ui.Button(
											ui.FillWidth(
												ui.Container(
													ui.Style{
														Background: bg,
														Padding:    ui.Symmetric(8, 10),
														Radius:     6,
													},
													ui.Text(entry.Name, ui.TextColor(fg), ui.TextSize(13)),
												),
											),
											ui.ButtonBackground(ui.NRGBA(0, 0, 0, 0)),
											ui.ButtonPadding(ui.All(0)),
											ui.ButtonRadius(6),
											ui.OnClick(func(ctx *ui.Context) {
												selectedFile.Set(entry.Name)
												codeText.Set(entry.Content)
											}),
										),
									)
								},
								ui.ListItemSpacing(6),
							),
						),
					),
				),
			),
		)

		codePanel := ui.Container(
			ui.Style{
				Background: ui.NRGBA(30, 30, 30, 255),
				Padding:    ui.Insets{Top: 12, Left: 12, Right: 12, Bottom: 12},
			},
			ui.Column(
				ui.Row(
					ui.Text(selectedFile.Value(), ui.TextColor(ui.NRGBA(220, 220, 220, 255)), ui.TextSize(13)),
					ui.Padding(ui.Insets{Left: 10}, ui.Text("UTF-8", ui.TextColor(ui.NRGBA(120, 120, 120, 255)), ui.TextSize(12))),
				),
				ui.Expanded(
					ui.Padding(
						ui.Insets{Top: 10},
						ui.Fill(
							ui.TextField(
								codeText.Value(),
								ui.InputSingleLine(false),
								ui.InputPlaceholder("// input code here"),
								ui.InputBackground(ui.NRGBA(30, 30, 30, 255)),
								ui.InputForeground(ui.NRGBA(212, 212, 212, 255)),
								ui.InputBorder(ui.NRGBA(62, 62, 66, 255)),
								ui.InputBorderFocus(ui.NRGBA(14, 99, 156, 255)),
								ui.InputPadding(ui.All(12)),
								ui.InputOnChange(func(ctx *ui.Context, value string) {
									codeText.Set(value)
								}),
							),
						),
					),
				),
			),
		)

		mainArea := ui.Expanded(
			ui.Row(
				leftSidebar,
				filePanel,
				ui.Expanded(codePanel),
			),
		)

		statusBar := ui.FixedHeight(
			28,
			ui.Container(
				ui.Style{
					Background: ui.NRGBA(0, 122, 204, 255),
					Padding:    ui.Symmetric(6, 10),
				},
				ui.Row(
					ui.Text(
						fmt.Sprintf("Tool: %s", activeToolLabel(activeTool.Value(), tools)),
						ui.TextColor(ui.NRGBA(255, 255, 255, 255)),
						ui.TextSize(12),
					),
					ui.Padding(
						ui.Insets{Left: 14},
						ui.Text(
							fmt.Sprintf("File: %s", selectedFile.Value()),
							ui.TextColor(ui.NRGBA(255, 255, 255, 255)),
							ui.TextSize(12),
						),
					),
				),
			),
		)

		return ui.Container(
			ui.Style{
				Background: th.Surface,
			},
			ui.Column(
				topMenu,
				mainArea,
				statusBar,
			),
		)
	}, ui.Title("FluxUI VSCode Layout"), ui.Size(1200, 780))
}

func activeToolLabel(key string, tools []sideTool) string {
	for i := range tools {
		if tools[i].Key == key {
			return tools[i].Label
		}
	}
	return key
}
