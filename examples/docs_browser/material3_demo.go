package main

import (
	"image/color"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func buildDocsMaterial3Showcase(th *ui.Theme) ui.Element {
	if th == nil {
		th = docsBrowserTheme(defaultDocsThemeSeed, false)
	}
	return ui.ColumnElement(
		ui.TextElement("Material Design 3 in FluxUI", ui.TextType(th.Types.TitleMedium), ui.TextColor(th.Colors.OnSurface)),
		ui.VSpacerElement(8),
		ui.TextElement("主题令牌、种子色彩角色、组件变体、状态层、形状和排版可通过 React 风格 Element API 使用。", ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
		ui.VSpacerElement(12),
		ui.RowElement(
			ui.ExpandedElement(material3TokenCard("主色", th.Colors.Primary, th.Colors.OnPrimary, th)),
			ui.HSpacerElement(8),
			ui.ExpandedElement(material3TokenCard("次要色", th.Colors.Secondary, th.Colors.OnSecondary, th)),
			ui.HSpacerElement(8),
			ui.ExpandedElement(material3TokenCard("Surface", th.Colors.SurfaceContainerHigh, th.Colors.OnSurface, th)),
		),
		ui.VSpacerElement(12),
		ui.ContainerDecorationElement(
			ui.Bg(th.Colors.SurfaceContainer).WithPad(ui.All(12)).WithRad(10),
			ui.ColumnElement(
				ui.TextElement("按钮变体", ui.TextType(th.Types.LabelLarge), ui.TextColor(th.Colors.OnSurface)),
				ui.VSpacerElement(8),
				ui.RowElement(
					ui.PaddingElement(ui.Insets{Right: 8, Bottom: 8}, ui.FilledButtonElement(ui.TextElement("Filled"))),
					ui.PaddingElement(ui.Insets{Right: 8, Bottom: 8}, ui.FilledTonalButtonElement(ui.TextElement("Tonal"))),
					ui.PaddingElement(ui.Insets{Right: 8, Bottom: 8}, ui.OutlinedButtonElement(ui.TextElement("Outlined"))),
					ui.PaddingElement(ui.Insets{Right: 8, Bottom: 8}, ui.TextButtonElement(ui.TextElement("Text"))),
				),
				ui.VSpacerElement(8),
				ui.RowElement(
					ui.ExpandedElement(ui.OutlinedTextFieldElement("React-style Element API", ui.InputPlaceholder("Outlined"))),
					ui.HSpacerElement(10),
					ui.ExpandedElement(ui.FilledTextFieldElement("MD3 default styling", ui.InputPlaceholder("Filled"))),
				),
			),
		),
		ui.VSpacerElement(12),
		ui.RowElement(
			ui.ExpandedElement(material3ComponentCard("Filled 卡片", "SurfaceContainer 角色与 MD3 形状。", ui.FilledCardElement, th)),
			ui.HSpacerElement(10),
			ui.ExpandedElement(material3ComponentCard("Elevated 卡片", "海拔和状态层保持令牌驱动。", ui.ElevatedCardElement, th)),
			ui.HSpacerElement(10),
			ui.ExpandedElement(material3ComponentCard("Outlined 卡片", "OutlineVariant 与圆角感知主题。", ui.OutlinedCardElement, th)),
		),
		ui.VSpacerElement(12),
		ui.RowElement(
			ui.FilterChipElement("已选 Chip", ui.ChipSelected(true)),
			ui.HSpacerElement(8),
			ui.AssistChipElement("Assist"),
			ui.HSpacerElement(8),
			ui.BadgeElement(ui.IconButtonElement(ui.IconElement("mail")), "3"),
			ui.HSpacerElement(8),
			ui.CircularProgressElement(0.72),
		),
	)
}

func material3TokenCard(label string, bg color.NRGBA, fg color.NRGBA, th *ui.Theme) ui.Element {
	return ui.ContainerDecorationElement(
		ui.Bg(bg).WithPad(ui.All(12)).WithRad(10),
		ui.ColumnElement(
			ui.TextElement(label, ui.TextType(th.Types.LabelLarge), ui.TextColor(fg)),
			ui.VSpacerElement(8),
			ui.TextElement(material3ColorText(bg), ui.TextType(th.Types.BodySmall), ui.TextColor(fg)),
		),
	)
}

func material3ComponentCard(label string, body string, card func(ui.Element, ...ui.CardOption) ui.Element, th *ui.Theme) ui.Element {
	return card(
		ui.ContainerDecorationElement(
			ui.Bg(color.NRGBA{}).WithPad(ui.All(12)),
			ui.ColumnElement(
				ui.TextElement(label, ui.TextType(th.Types.LabelLarge), ui.TextColor(th.Colors.OnSurface)),
				ui.VSpacerElement(6),
				ui.TextElement(body, ui.TextType(th.Types.BodySmall), ui.TextColor(th.Colors.OnSurfaceVariant)),
			),
		),
	)
}

func material3ColorText(c color.NRGBA) string {
	return "#" + hexByte(c.R) + hexByte(c.G) + hexByte(c.B)
}

func hexByte(v uint8) string {
	const digits = "0123456789ABCDEF"
	return string([]byte{digits[v>>4], digits[v&0x0f]})
}
