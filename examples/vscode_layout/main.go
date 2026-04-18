package main

import (
	"fmt"
	"strings"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
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

type menuItem struct {
	Key      string
	Label    string
	Command  string
	Children []menuItem
}

type topMenu struct {
	Key   string
	Label string
	Items []menuItem
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

	menus := []topMenu{
		{
			Key:   "file",
			Label: "File",
			Items: []menuItem{
				{Key: "file.new", Label: "New File", Command: "file.new"},
				{
					Key:   "file.open_recent",
					Label: "Open Recent",
					Children: []menuItem{
						{Key: "recent.alpha", Label: "alpha-service", Command: "recent.alpha"},
						{Key: "recent.beta", Label: "beta-admin", Command: "recent.beta"},
						{Key: "recent.docs", Label: "docs-site", Command: "recent.docs"},
					},
				},
				{Key: "file.save", Label: "Save", Command: "file.save"},
				{Key: "file.exit", Label: "Exit", Command: "file.exit"},
			},
		},
		{
			Key:   "edit",
			Label: "Edit",
			Items: []menuItem{
				{Key: "edit.undo", Label: "Undo", Command: "edit.undo"},
				{Key: "edit.redo", Label: "Redo", Command: "edit.redo"},
				{Key: "edit.find", Label: "Find", Command: "edit.find"},
			},
		},
		{
			Key:   "view",
			Label: "View",
			Items: []menuItem{
				{
					Key:   "view.appearance",
					Label: "Appearance",
					Children: []menuItem{
						{Key: "view.zen", Label: "Zen Mode", Command: "view.zen"},
						{Key: "view.sidebar", Label: "Toggle Side Bar", Command: "view.sidebar"},
					},
				},
				{Key: "view.command_palette", Label: "Command Palette", Command: "view.command_palette"},
			},
		},
		{
			Key:   "terminal",
			Label: "Terminal",
			Items: []menuItem{
				{Key: "terminal.new", Label: "New Terminal", Command: "terminal.new"},
				{Key: "terminal.split", Label: "Split Terminal", Command: "terminal.split"},
			},
		},
		{
			Key:   "help",
			Label: "Help",
			Items: []menuItem{
				{Key: "help.docs", Label: "Documentation", Command: "help.docs"},
				{Key: "help.about", Label: "About", Command: "help.about"},
			},
		},
	}

	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)

		activeTool := ui.State[string](ctx)
		selectedFile := ui.State[string](ctx)
		codeText := ui.State[string](ctx)
		lastCommand := ui.State[string](ctx)
		openTopMenu := ui.State[string](ctx)
		openSubMenu := ui.State[string](ctx)

		if activeTool.Value() == "" {
			activeTool.Set("explorer")
		}
		if selectedFile.Value() == "" && len(files) > 0 {
			selectedFile.Set(files[0].Name)
			codeText.Set(files[0].Content)
		}
		if lastCommand.Value() == "" {
			lastCommand.Set("Ready")
		}

		const brandWidth = float32(120)
		const topMenuButtonWidth = float32(82)
		const topMenuButtonGap = float32(4)

		menuButtons := make([]ui.Widget, 0, len(menus))
		for i := range menus {
			menu := menus[i]
			isOpen := openTopMenu.Value() == menu.Key

			bg := ui.NRGBA(0, 0, 0, 0)
			fg := ui.NRGBA(204, 204, 204, 255)
			if isOpen {
				bg = ui.NRGBA(63, 63, 70, 255)
				fg = ui.NRGBA(255, 255, 255, 255)
			}

			right := float32(0)
			if i < len(menus)-1 {
				right = topMenuButtonGap
			}

			menuButtons = append(menuButtons,
				ui.Padding(
					ui.Insets{Right: right},
					ui.FixedWidth(
						topMenuButtonWidth,
						ui.Button(
							ui.FillWidth(
								ui.Container(
									ui.Style{
										Background: bg,
										Padding:    ui.Symmetric(7, 0),
										Radius:     4,
									},
									ui.Center(ui.Text(menu.Label, ui.TextColor(fg), ui.TextSize(13))),
								),
							),
							ui.ButtonBackground(ui.NRGBA(0, 0, 0, 0)),
							ui.ButtonPadding(ui.All(0)),
							ui.OnClick(func(ctx *ui.Context) {
								if openTopMenu.Value() == menu.Key {
									openTopMenu.Set("")
									openSubMenu.Set("")
									return
								}
								openTopMenu.Set(menu.Key)
								openSubMenu.Set("")
							}),
						),
					),
				),
			)
		}

		topMenuBar := ui.FixedHeight(
			44,
			ui.Container(
				ui.Style{
					Background: ui.NRGBA(37, 37, 38, 255),
					Padding:    ui.Symmetric(8, 10),
				},
				ui.Row(
					ui.FixedWidth(
						brandWidth,
						ui.Text("FluxUI IDE", ui.TextColor(ui.NRGBA(229, 229, 229, 255)), ui.TextSize(14)),
					),
					ui.Row(menuButtons...),
					ui.Expanded(ui.Spacer(0, 0)),
					ui.Text("Workspace", ui.TextColor(ui.NRGBA(140, 140, 140, 255)), ui.TextSize(12)),
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
			ui.FillWidth(
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
						ui.Padding(
							ui.Insets{Left: 14},
							ui.Text(
								fmt.Sprintf("Command: %s", lastCommand.Value()),
								ui.TextColor(ui.NRGBA(255, 255, 255, 255)),
								ui.TextSize(12),
							),
						),
					),
				),
			),
		)

		page := ui.Container(
			ui.Style{
				Background: th.Surface,
			},
			ui.Column(
				topMenuBar,
				mainArea,
				statusBar,
			),
		)

		layers := []ui.Widget{page}
		activeMenu, activeMenuIndex, hasActive := findTopMenu(menus, openTopMenu.Value())
		if hasActive {
			left := brandWidth + float32(activeMenuIndex)*(topMenuButtonWidth+topMenuButtonGap)

			layers = append(layers,
				ui.Padding(
					ui.Insets{Top: 44},
					ui.ClickArea(
						ui.Fill(ui.Spacer(0, 0)),
						func(ctx *ui.Context) {
							openTopMenu.Set("")
							openSubMenu.Set("")
						},
					),
				),
			)

			layers = append(layers,
				ui.Padding(
					ui.Insets{Top: 44, Left: left},
					buildMenuPanel(
						activeMenu.Items,
						openSubMenu.Value(),
						func(item menuItem) {
							if len(item.Children) > 0 {
								if openSubMenu.Value() == item.Key {
									openSubMenu.Set("")
								} else {
									openSubMenu.Set(item.Key)
								}
								return
							}
							openTopMenu.Set("")
							openSubMenu.Set("")
							execCommand(item.Command, selectedFile.Set, codeText.Set, lastCommand.Set)
						},
					),
				),
			)

			if children, ok := findSubMenu(activeMenu.Items, openSubMenu.Value()); ok {
				layers = append(layers,
					ui.Padding(
						ui.Insets{Top: 44, Left: left + 228},
						buildMenuPanel(
							children,
							"",
							func(item menuItem) {
								openTopMenu.Set("")
								openSubMenu.Set("")
								execCommand(item.Command, selectedFile.Set, codeText.Set, lastCommand.Set)
							},
						),
					),
				)
			}
		}

		return ui.Stack(layers...)
	}, ui.Title("FluxUI VSCode Layout"), ui.Size(1200, 780))
}

func buildMenuPanel(items []menuItem, openedKey string, onClick func(item menuItem)) ui.Widget {
	rows := make([]ui.Widget, 0, len(items))
	for i := range items {
		item := items[i]
		hasChildren := len(item.Children) > 0
		expanded := hasChildren && openedKey == item.Key

		bg := ui.NRGBA(45, 45, 48, 255)
		if expanded {
			bg = ui.NRGBA(14, 99, 156, 255)
		}

		arrow := ""
		if hasChildren {
			arrow = ">"
		}

		row := ui.FillWidth(
			ui.Button(
				ui.FillWidth(
					ui.Container(
						ui.Style{
							Background: bg,
							Padding:    ui.Symmetric(8, 10),
							Radius:     6,
						},
						ui.Row(
							ui.Text(item.Label, ui.TextColor(ui.NRGBA(220, 220, 220, 255)), ui.TextSize(12)),
							ui.Expanded(ui.Spacer(0, 0)),
							ui.Text(arrow, ui.TextColor(ui.NRGBA(180, 180, 180, 255)), ui.TextSize(12)),
						),
					),
				),
				ui.ButtonBackground(ui.NRGBA(0, 0, 0, 0)),
				ui.ButtonPadding(ui.All(0)),
				ui.OnClick(func(ctx *ui.Context) {
					onClick(item)
				}),
			),
		)

		if i < len(items)-1 {
			row = ui.Padding(ui.Insets{Bottom: 6}, row)
		}
		rows = append(rows, row)
	}

	return ui.FixedWidth(
		220,
		ui.Container(
			ui.Style{
				Background: ui.NRGBA(37, 37, 38, 255),
				Padding:    ui.All(8),
				Radius:     8,
			},
			ui.Column(rows...),
		),
	)
}

func execCommand(command string, setSelectedFile, setCodeText, setLastCommand func(string)) {
	switch command {
	case "file.new":
		setSelectedFile("untitled.txt")
		setCodeText("")
		setLastCommand("file.new")
	case "file.save":
		setLastCommand("file.save")
	case "file.exit":
		setLastCommand("file.exit")
	case "recent.alpha":
		setSelectedFile("alpha-service/main.go")
		setCodeText("// opened recent workspace: alpha-service\n")
		setLastCommand("recent.alpha")
	case "recent.beta":
		setSelectedFile("beta-admin/app.go")
		setCodeText("// opened recent workspace: beta-admin\n")
		setLastCommand("recent.beta")
	case "recent.docs":
		setSelectedFile("docs-site/README.md")
		setCodeText("# docs-site\n")
		setLastCommand("recent.docs")
	default:
		setLastCommand(command)
	}
}

func findTopMenu(menus []topMenu, key string) (topMenu, int, bool) {
	for i := range menus {
		if menus[i].Key == key {
			return menus[i], i, true
		}
	}
	return topMenu{}, -1, false
}

func findSubMenu(items []menuItem, key string) ([]menuItem, bool) {
	for i := range items {
		if items[i].Key == key && len(items[i].Children) > 0 {
			return items[i].Children, true
		}
	}
	return nil, false
}

func activeToolLabel(key string, tools []sideTool) string {
	for i := range tools {
		if tools[i].Key == key {
			return tools[i].Label
		}
	}
	return key
}
