package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/xiaowumin-mark/FluxUI/internal"
	ui "github.com/xiaowumin-mark/FluxUI/ui"

	gioLayout "gioui.org/layout"
	"gioui.org/op"
)

func TestDragDropDocsExposeBrowserExamples(t *testing.T) {
	docs, err := loadWidgetDocs()
	if err != nil {
		t.Fatalf("load docs: %v", err)
	}

	cases := map[string]string{
		"drop_target": "drop_target_basic",
		"drag_source": "drag_source_basic",
	}
	for docID, exampleID := range cases {
		doc := findDocByID(docs, docID)
		if doc == nil {
			t.Fatalf("expected %s doc to be loaded", docID)
		}
		if doc.Meta.Example.ID != exampleID {
			t.Fatalf("%s example id = %q, want %q", docID, doc.Meta.Example.ID, exampleID)
		}
	}
}

func TestSystemDocsExposeBrowserExample(t *testing.T) {
	docs, err := loadWidgetDocs()
	if err != nil {
		t.Fatalf("load docs: %v", err)
	}

	for _, docID := range []string{
		"system_api",
		"window_api",
		"file_dialog_api",
		"message_box_api",
		"notification_api",
		"tray_api",
		"clipboard_shell_api",
		"system_events_api",
	} {
		doc := findDocByID(docs, docID)
		if doc == nil {
			t.Fatalf("expected %s doc to be loaded", docID)
		}
		if doc.Meta.Example.ID != "system_api_basic" {
			t.Fatalf("%s example id = %q, want system_api_basic", docID, doc.Meta.Example.ID)
		}
	}
}

func TestEventSystemDocsExposeBrowserExample(t *testing.T) {
	docs, err := loadWidgetDocs()
	if err != nil {
		t.Fatalf("load docs: %v", err)
	}

	doc := findDocByID(docs, "event_system")
	if doc == nil {
		t.Fatal("expected event_system doc to be loaded")
	}
	if doc.Meta.Example.ID != "event_system_basic" {
		t.Fatalf("event_system example id = %q, want event_system_basic", doc.Meta.Example.ID)
	}
	if !strings.Contains(doc.Content, "什么时候用简单 API") || !strings.Contains(doc.Content, "什么时候用高级事件") {
		t.Fatal("event system guide should explain simple API vs advanced events")
	}

	data, err := os.ReadFile("event_system_demo.go")
	if err != nil {
		t.Fatalf("read event_system_demo.go: %v", err)
	}
	source := string(data)
	for token, label := range map[string]string{
		"PointerAreaElement(":       "pointer area example",
		"PointerOnContextMenu(":     "context menu example",
		"PointerOnWheel(":           "wheel example",
		"KeyboardScopeElement(":     "keyboard scope example",
		"ShortcutOn(":               "local shortcut example",
		"InputOnBeforeInput(":       "beforeinput example",
		"DragSourceElement(":        "drag source example",
		"DropTargetElement(":        "drop target example",
		"DispatchCustomEvent(":      "custom event example",
		"OnEvent(ctx,":              "event listener example",
		"PointerCaptureOnPress(":    "pointer capture example",
		"CoalescedSamples()":        "coalesced pointer example",
		"ev.PreventDefault()":       "preventDefault example",
		"ev.StopPropagation()":      "stopPropagation example",
		"ev.SetPointerCapture(ctx)": "set pointer capture example",
	} {
		if !strings.Contains(source, token) {
			t.Fatalf("missing %s in event_system_demo.go; expected token %q", label, token)
		}
	}
}

func TestRootRoadmapDocsExposeBrowserExamples(t *testing.T) {
	docs, err := loadWidgetDocs()
	if err != nil {
		t.Fatalf("load docs: %v", err)
	}

	cases := map[string]string{
		"material3_design_plan": "material3_showcase",
		"material3_roadmap":     "material3_showcase",
		"system_api_roadmap":    "system_api_basic",
	}
	for docID, exampleID := range cases {
		doc := findDocByID(docs, docID)
		if doc == nil {
			t.Fatalf("expected root roadmap doc %s to be loaded", docID)
		}
		if doc.Meta.Example.ID != exampleID {
			t.Fatalf("%s example id = %q, want %q", docID, doc.Meta.Example.ID, exampleID)
		}
	}
}

func TestDocsBrowserOverviewLoadsFirst(t *testing.T) {
	docs, err := loadWidgetDocs()
	if err != nil {
		t.Fatalf("load docs: %v", err)
	}
	if len(docs) == 0 {
		t.Fatal("expected loaded docs")
	}
	if docs[0].Meta.ID != docsOverviewID {
		t.Fatalf("first doc = %q, want %q", docs[0].Meta.ID, docsOverviewID)
	}
	doc := findDocByID(docs, docsOverviewID)
	if doc == nil {
		t.Fatalf("expected generated overview doc %q", docsOverviewID)
	}
	if doc.Meta.Example.ID != "docs_overview" {
		t.Fatalf("overview example id = %q, want docs_overview", doc.Meta.Example.ID)
	}
	if len(doc.Meta.APIs) == 0 {
		t.Fatal("overview should expose API quick-reference entries")
	}
	if !strings.Contains(doc.Content, "React-style") || !strings.Contains(doc.Content, "```go") {
		t.Fatal("overview content should introduce React-style usage with a Go code block")
	}
}

func TestParseWidgetDocIgnoresMetadataExamples(t *testing.T) {
	docsRoot, err := resolveDocsRootDir()
	if err != nil {
		t.Fatalf("resolve docs root: %v", err)
	}
	path := filepath.Join(docsRoot, "README.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read docs README: %v", err)
	}

	if _, err := parseWidgetDoc(path, string(data)); err == nil {
		t.Fatal("expected README metadata example block to be ignored")
	}
}

func TestAllDocExamplesAreKnownToBrowser(t *testing.T) {
	docs, err := loadWidgetDocs()
	if err != nil {
		t.Fatalf("load docs: %v", err)
	}

	for _, doc := range docs {
		exampleID := doc.Meta.Example.ID
		if exampleID == "" {
			exampleID = doc.Meta.ID
		}
		if !isDocsDemoKnown(exampleID) {
			t.Fatalf("%s references unknown example %q", doc.Meta.ID, exampleID)
		}
	}
}

func TestAllDocExamplesHaveBuildDemoCases(t *testing.T) {
	docs, err := loadWidgetDocs()
	if err != nil {
		t.Fatalf("load docs: %v", err)
	}
	cases := implementedDocsDemoCases(t)

	for _, doc := range docs {
		exampleID := doc.Meta.Example.ID
		if exampleID == "" {
			exampleID = doc.Meta.ID
		}
		if _, ok := cases[exampleID]; !ok {
			t.Fatalf("%s references example %q, but buildDemo has no matching case", doc.Meta.ID, exampleID)
		}
	}
}

func TestDocsBrowserSourceTextHasNoMojibake(t *testing.T) {
	badFragments := []string{
		"\u93c2\u56e8",
		"\u9410\u7470",
		"\u699b\u6a3f",
		"\u8930\u64b3",
		"\u7487\u950b",
		"\u7459\ufe40",
		"\u93c8\ue100",
		"\u9366\u3127",
		"\u6d93\u5b2b",
		"\u701b\u612e",
		"\u95c3\u6751",
		"\u9357\ufe33",
		"\u6d63\u5eb4",
		"\u6942\u6a39",
		"\u7ec0\u8f70",
		"\u934f\u62bd",
		"\u93c4\u5267",
		"\u5a4a\u6c2c",
		"\u68e3\u682d",
		"\u93b4\u6205",
		"\u9358\u719a",
	}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob source files: %v", err)
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		text := string(data)
		for _, fragment := range badFragments {
			if strings.Contains(text, fragment) {
				t.Fatalf("%s contains mojibake fragment %q", path, fragment)
			}
		}
	}
}

func TestDocsBrowserMainStaysFocused(t *testing.T) {
	data, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	text := string(data)
	bannedPrefixes := []string{
		"func buildDocs",
		"func renderMarkdown",
		"func docsRightPanel",
		"func docsLeftPanel",
		"func docsExample",
		"func docsBrowserApp",
		"func docsThemeControls",
		"func docsCategory",
	}
	for _, prefix := range bannedPrefixes {
		if strings.Contains(text, prefix) {
			t.Fatalf("main.go contains %q; keep app UI, renderer, and demo code in focused modules", prefix)
		}
	}
}

func TestDocsBrowserControlDemosStaySplit(t *testing.T) {
	expectedFiles := []string{
		"controls_feedback_demo.go",
		"controls_selection_demo.go",
		"controls_media_demo.go",
		"controls_overlay_demo.go",
		"controls_progress_demo.go",
	}
	for _, path := range expectedFiles {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected focused control demo file %s: %v", path, err)
		}
	}

	data, err := os.ReadFile("controls_feedback_demo.go")
	if err != nil {
		t.Fatalf("read controls_feedback_demo.go: %v", err)
	}
	text := string(data)
	for _, moved := range []string{
		"func docsCheckboxDemo",
		"func docsMenuDemo",
		"func docsImageDemo",
		"func docsProgressIndicatorsDemo",
	} {
		if strings.Contains(text, moved) {
			t.Fatalf("controls_feedback_demo.go contains %q; keep focused demo groups split", moved)
		}
	}
}

func TestDocsRightPanelUsesVirtualizedContentAndExternalPopup(t *testing.T) {
	appData, err := os.ReadFile("docs_browser_app.go")
	if err != nil {
		t.Fatalf("read docs_browser_app.go: %v", err)
	}
	appText := string(appData)
	for _, token := range []string{
		"docsRightPanelList(rightPanelContent)",
		"examplePopup = docsExamplePopup(",
		"ui.UseMemo(ctx, []any{currentDocID, currentDocContent",
	} {
		if !strings.Contains(appText, token) {
			t.Fatalf("docs_browser_app.go should contain %q", token)
		}
	}

	rightData, err := os.ReadFile("right_panel.go")
	if err != nil {
		t.Fatalf("read right_panel.go: %v", err)
	}
	rightText := string(rightData)
	if strings.Contains(rightText, "docsExamplePopup(") {
		t.Fatal("right panel content should not place popup inside the scroll/list content")
	}
	if !strings.Contains(rightText, "ui.ListViewElement(") || !strings.Contains(rightText, "ui.ListVirtualized(true)") {
		t.Fatal("right panel content should be rendered through a virtualized list")
	}
}

func TestDocsDemoPresentationRules(t *testing.T) {
	cases := map[string]docsDemoPresentation{
		"button_basic":       {Height: 230, Center: true},
		"animation_basic":    {Height: 300, Scroll: true},
		"material3_showcase": {Height: 440, Scroll: true},
		"system_api_basic":   {Height: 360},
		"event_system_basic": {Height: 440},
	}
	for exampleID, want := range cases {
		got := docsDemoPresentationFor(exampleID)
		if got != want {
			t.Fatalf("%s presentation = %#v, want %#v", exampleID, got, want)
		}
	}
}

func TestSystemAPIDemoDoesNotProbeDuringCapabilityCardRender(t *testing.T) {
	data, err := os.ReadFile("system_api_demo.go")
	if err != nil {
		t.Fatalf("read system_api_demo.go: %v", err)
	}
	text := string(data)
	cardStart := strings.Index(text, "func systemCapabilityCard(")
	if cardStart < 0 {
		t.Fatal("systemCapabilityCard not found")
	}
	cardEnd := strings.Index(text[cardStart:], "\nfunc ")
	if cardEnd < 0 {
		t.Fatal("systemCapabilityCard end not found")
	}
	cardBody := text[cardStart : cardStart+cardEnd]
	if strings.Contains(cardBody, "system.Capabilities(") || strings.Contains(cardBody, "system.Availability(") {
		t.Fatal("systemCapabilityCard must use cached probe state, not live system probing")
	}
	if !strings.Contains(text, "func docsSystemProbeSnapshot(") ||
		!strings.Contains(text, "system.Capabilities()") ||
		!strings.Contains(text, "system.Availability(") {
		t.Fatal("expected docsSystemProbeSnapshot to own system capability probing")
	}
}

func TestRenderMarkdownDocumentBuildsCodeBlocks(t *testing.T) {
	th := docsBrowserTheme(defaultDocsThemeSeed, false)
	state := &testStringState{}
	got := renderMarkdownDocument("# Title\n\n```go\nfmt.Println(\"ok\")\n```\n\n- item", th, state)
	if len(got) < 4 {
		t.Fatalf("expected rendered markdown elements, got %d", len(got))
	}
}

func TestRenderMarkdownDocumentBuildsNestedListsAndDeepHeadings(t *testing.T) {
	th := docsBrowserTheme(defaultDocsThemeSeed, false)
	got := renderMarkdownDocument("#### Phase\n\n- parent\n  - child\n- [x] done\n- [ ] todo\n", th, &testStringState{})
	if len(got) < 8 {
		t.Fatalf("expected rendered advanced markdown elements, got %d", len(got))
	}

	if level, text, ok := markdownHeadingInfo("#### Phase"); !ok || level != 4 || text != "Phase" {
		t.Fatalf("heading info = %d, %q, %v; want level 4 Phase", level, text, ok)
	}
	marker, text, indent, ok := markdownListInfo("  - child")
	if !ok || marker != "-" || text != "child" || indent != 1 {
		t.Fatalf("nested list info = %q, %q, %d, %v", marker, text, indent, ok)
	}
	marker, text, indent, ok = markdownListInfo("- [x] done")
	if !ok || marker != "[x]" || text != "done" || indent != 0 {
		t.Fatalf("task list info = %q, %q, %d, %v", marker, text, indent, ok)
	}
}

func TestRenderMarkdownDocumentBuildsTables(t *testing.T) {
	th := docsBrowserTheme(defaultDocsThemeSeed, false)
	got := renderMarkdownDocument("| API | Use |\n| --- | --- |\n| `RunElement` | docs browser |\n", th, &testStringState{})
	if len(got) < 2 {
		t.Fatalf("expected rendered table elements, got %d", len(got))
	}
	if !isMarkdownTableStart([]string{"| A | B |", "| --- | --- |"}, 0) {
		t.Fatal("expected table start detection")
	}
	if !isMarkdownTableSeparator("| --- | :---: |") {
		t.Fatal("expected table separator detection")
	}
}

func TestMarkdownTableColumnWidthsStayFinite(t *testing.T) {
	widths := markdownTableColumnWidths([][]string{
		{"API", "Use"},
		{"RunElement", "docs browser"},
		{"VeryLongColumnValueThatShouldBeClampedInsteadOfExpandingForever", "short"},
	})
	if len(widths) != 2 {
		t.Fatalf("column width count = %d, want 2", len(widths))
	}
	for index, width := range widths {
		if width < 88 || width > 280 {
			t.Fatalf("column %d width = %.1f, want clamped finite width", index, width)
		}
	}
	if total := widths[0] + widths[1]; total >= 1_000_000 {
		t.Fatalf("table content width = %.1f, should not inherit scroll measurement limit", total)
	}
}

func TestMaterial3ShowcaseElementBuilds(t *testing.T) {
	th := docsBrowserTheme(defaultDocsThemeSeed, false)
	if buildDocsMaterial3Showcase(th) == nil {
		t.Fatal("expected material3 showcase element")
	}
}

func TestDocsOverviewDemoElementBuilds(t *testing.T) {
	th := docsBrowserTheme(defaultDocsThemeSeed, false)
	if buildDocsOverviewDemo(th) == nil {
		t.Fatal("expected docs overview element")
	}
}

func TestSystemAPIDemoElementBuilds(t *testing.T) {
	th := docsBrowserTheme(defaultDocsThemeSeed, false)
	status := &testStringState{value: "ready"}
	if buildDocsSystemAPIDemo(nil, status, th) == nil {
		t.Fatal("expected system API demo element")
	}
	if docsSystemSingleInstanceSection(th) == nil {
		t.Fatal("expected single instance section")
	}
	if docsSystemGlobalShortcutSection(th) == nil {
		t.Fatal("expected global shortcut section")
	}
	if docsSystemRegistrationSection(th) == nil {
		t.Fatal("expected system registration section")
	}
	if docsSystemDragDropProbeSection(th) == nil {
		t.Fatal("expected drag/drop probe section")
	}
}

func TestDocsBrowserDemosUseRepresentativeReactAPIs(t *testing.T) {
	required := map[string]string{
		"GridViewElement(":                    "GridViewElement dynamic grid demo",
		"GridMinItemWidth(":                   "GridMinItemWidth responsive grid option",
		"GridOnReachEnd(":                     "GridOnReachEnd pagination callback",
		"ScrollHorizontal(":                   "ScrollHorizontal demo",
		"ScrollOnChange(":                     "ScrollOnChange offset callback demo",
		"ScrollAutoToEndKey(":                 "ScrollAutoToEndKey auto-scroll demo",
		"ScrollAttachRef(":                    "ScrollAttachRef command demo",
		"ListVirtualized(":                    "ListVirtualized list demo",
		"ListPadding(":                        "ListPadding list demo",
		"ListDecoration(":                     "ListDecoration list demo",
		"ListAxis(":                           "ListAxis horizontal list demo",
		"UseMemo(":                            "UseMemo runtime demo",
		"UseRef(":                             "UseRef runtime demo",
		"UseCallback(":                        "UseCallback runtime demo",
		"UseAsync[string](":                   "UseAsync runtime demo",
		"UseInterval(":                        "UseInterval runtime demo",
		"AsyncLoading":                        "AsyncStatus loading demo",
		"Provider[string](":                   "typed Provider runtime demo",
		"UseContext(":                         "UseContext runtime demo",
		"Fragment(":                           "Fragment composition demo",
		"WithFontElement(":                    "WithFontElement scoped font demo",
		"FromWidget(":                         "FromWidget legacy bridge demo",
		"RouterElement(":                      "RouterElement React-style router demo",
		"RouteElement(":                       "RouteElement route declaration demo",
		"RouteName(":                          "RouteName route metadata demo",
		"RouteTitle(":                         "RouteTitle route metadata demo",
		"RouteKey(":                           "RouteKey route identity demo",
		"RouteMeta(":                          "RouteMeta route metadata demo",
		"RouteMetaMap(":                       "RouteMetaMap route metadata demo",
		"RouteBeforeEnter(":                   "RouteBeforeEnter route guard demo",
		"RouterNotFoundElement(":              "RouterNotFoundElement 404 demo",
		"UseNavigate(":                        "UseNavigate router hook demo",
		"UseLocation(":                        "UseLocation router hook demo",
		"UseParams(":                          "UseParams router hook demo",
		"UseRoute(":                           "UseRoute metadata hook demo",
		"ui.Navigate(":                        "Navigate direct navigation demo",
		"ui.NavigateReplace(":                 "NavigateReplace navigation demo",
		"ui.NavigateBack(":                    "NavigateBack navigation demo",
		"ui.CanGoBack(":                       "CanGoBack stack state demo",
		"ui.CurrentPath(":                     "CurrentPath route state demo",
		"ui.StackDepth(":                      "StackDepth route state demo",
		".AllPathParams(":                     "AllPathParams route params demo",
		".AllQueryParams(":                    "AllQueryParams query params demo",
		".HasParam(":                          "HasParam route params demo",
		".HasQuery(":                          "HasQuery route params demo",
		"FilledButtonElement(":                "FilledButtonElement button variant demo",
		"FilledTonalButtonElement(":           "FilledTonalButtonElement button variant demo",
		"ElevatedButtonElement(":              "ElevatedButtonElement button variant demo",
		"OutlinedTextFieldElement(":           "OutlinedTextFieldElement input variant demo",
		"FilledTextFieldElement(":             "FilledTextFieldElement input variant demo",
		"FilledCardElement":                   "FilledCardElement card variant demo",
		"ElevatedCardElement":                 "ElevatedCardElement card variant demo",
		"OutlinedCardElement":                 "OutlinedCardElement card variant demo",
		"NewButtonRef(":                       "ButtonRef command demo",
		"ButtonAttachRef(":                    "ButtonAttachRef demo",
		"OnHover(":                            "Button hover callback demo",
		"ButtonRadius(":                       "Button radius option demo",
		"NewInputRef(":                        "InputRef command demo",
		"InputAttachRef(":                     "InputAttachRef demo",
		"InputFontFamily(":                    "InputFontFamily typography demo",
		".SetText(":                           "InputRef SetText demo",
		".Append(":                            "InputRef Append demo",
		".Clear(":                             "InputRef Clear demo",
		".Focus(":                             "InputRef Focus demo",
		".Blur(":                              "InputRef Blur demo",
		"NewCheckboxRef(":                     "CheckboxRef command demo",
		"CheckboxAttachRef(":                  "CheckboxAttachRef demo",
		"CheckboxDecoration(":                 "CheckboxDecoration demo",
		"NewSwitchRef(":                       "SwitchRef command demo",
		"SwitchAttachRef(":                    "SwitchAttachRef demo",
		"SwitchDecoration(":                   "SwitchDecoration demo",
		"NewSliderRef(":                       "SliderRef command demo",
		"SliderAttachRef(":                    "SliderAttachRef demo",
		".StepBy(":                            "SliderRef StepBy demo",
		"SliderDecoration(":                   "SliderDecoration demo",
		"NewRadioGroupRef(":                   "RadioGroupRef command demo",
		"RadioGroupAttachRef(":                "RadioGroupAttachRef demo",
		"NewSelectRef[string](":               "SelectRef command demo",
		"SelectSearchable[string](":           "SelectSearchable demo",
		"SelectOnOpenChange[string](":         "SelectOnOpenChange demo",
		"MenuElement(":                        "MenuElement static menu demo",
		"MenuWidth(":                          "MenuWidth static menu demo",
		"MenuQuick(":                          "MenuQuick static menu demo",
		"DropdownMenuWidth(":                  "DropdownMenuWidth demo",
		"DropdownMenuXOffset(":                "DropdownMenuXOffset demo",
		"DropdownMenuYOffset(":                "DropdownMenuYOffset demo",
		"DropdownMenuAnchorCorner(":           "DropdownMenuAnchorCorner demo",
		"DropdownMenuMenuCorner(":             "DropdownMenuMenuCorner demo",
		"DropdownMenuDefaultFocusOf(":         "DropdownMenuDefaultFocus demo",
		"DropdownMenuTypeaheadDelay(":         "DropdownMenuTypeaheadDelay demo",
		"DropdownMenuNoNavigationWrap(":       "DropdownMenuNoNavigationWrap demo",
		"DropdownMenuDecoration(":             "DropdownMenuDecoration demo",
		"NewTabsRef(":                         "TabsRef command demo",
		"TabsAttachRef(":                      "TabsAttachRef demo",
		"TabsScrollable(":                     "TabsScrollable demo",
		"NewDialogRef(":                       "DialogRef command demo",
		"DialogAttachRef(":                    "DialogAttachRef demo",
		"DialogMaskColor(":                    "Dialog mask color demo",
		"NewPopupRef(":                        "PopupRef command demo",
		"PopupAttachRef(":                     "PopupAttachRef demo",
		"PopupMaskColor(":                     "Popup mask color demo",
		"ProgressIndeterminate(":              "Progress indeterminate demo",
		"ProgressSize(":                       "ProgressSize circular progress demo",
		"SearchBarTrailing(":                  "SearchBar trailing slot demo",
		"SearchBarInputOptions(":              "SearchBar input option forwarding demo",
		"BadgeOffset(":                        "Badge offset demo",
		"TooltipOffset(":                      "Tooltip offset demo",
		"ImageAttachRef(":                     "Image ref demo",
		"ImageOnClick(":                       "Image click demo",
		"IconAttachRef(":                      "Icon ref demo",
		"IconOnClick(":                        "Icon click demo",
		"ListItemElement(":                    "ListItemElement compact demo",
		"ListItemMinHeight(":                  "ListItemMinHeight demo",
		"IconButtonSize(":                     "IconButtonSize demo",
		"IconButtonBackground(":               "IconButtonBackground demo",
		"IconButtonDecoration(":               "IconButtonDecoration demo",
		"FloatingActionButtonBackground(":     "FloatingActionButtonBackground demo",
		"FloatingActionButtonDecoration(":     "FloatingActionButtonDecoration demo",
		"FilterChipElementWithSlots(":         "FilterChipElementWithSlots demo",
		"ChipElevated(":                       "ChipElevated demo",
		"ChipSoftDisabled(":                   "ChipSoftDisabled demo",
		"ChipRemovable(":                      "ChipRemovable demo",
		"ChipOnRemove(":                       "ChipOnRemove demo",
		"ChipBackground(":                     "ChipBackground demo",
		"ChipDecoration(":                     "ChipDecoration demo",
		"AppBarElement(":                      "AppBarElement configured demo",
		"AppBarLeading(":                      "AppBar leading option demo",
		"AppBarDecoration(":                   "AppBar decoration option demo",
		"NewBottomNavRef(":                    "BottomNavRef command demo",
		"BottomNavAttachRef(":                 "BottomNavAttachRef demo",
		"BottomNavDecoration(":                "BottomNavDecoration demo",
		"BottomNavAlignmentOf(":               "BottomNavAlignmentOf demo",
		"BottomNavAlignSpaceEvenly":           "BottomNavAlignSpaceEvenly demo",
		"NavigationRailHeader(":               "NavigationRail header demo",
		"NavigationRailFooter(":               "NavigationRail footer demo",
		"NavigationRailActiveColor(":          "NavigationRailActiveColor demo",
		"NavigationDrawerHeader(":             "NavigationDrawer header demo",
		"NavigationDrawerFooter(":             "NavigationDrawer footer demo",
		"NavigationDrawerActiveColor(":        "NavigationDrawerActiveColor demo",
		"Only(":                               "Insets Only helper demo",
		"LeftRight(":                          "Insets LeftRight helper demo",
		"BorderDeco(":                         "BorderDeco helper demo",
		"Circle(":                             "Circle decoration helper demo",
		"HoverBg(":                            "HoverBg decoration helper demo",
		"PressedBg(":                          "PressedBg decoration helper demo",
		"Focused(":                            "Focused decoration helper demo",
		"DisabledDeco(":                       "DisabledDeco decoration helper demo",
		"ContainerDecorationDisabled(":        "ContainerDecorationDisabled demo",
		"OnDecoClick(":                        "OnDecoClick demo",
		"OnDecoHoverEnter(":                   "OnDecoHoverEnter demo",
		"OnDecoHoverLeave(":                   "OnDecoHoverLeave demo",
		"ScaleDeco(":                          "ScaleDeco transform helper demo",
		"TranslateDeco(":                      "TranslateDeco transform helper demo",
		"TextLineHeight(":                     "TextLineHeight demo",
		"TextFont(":                           "TextFont demo",
		"DragSourceEventStarted":              "DragSourceEventStarted demo",
		"DragSourceEventRequested":            "DragSourceEventRequested demo",
		"DragSourceEventCompleted":            "DragSourceEventCompleted demo",
		"DragSourceEventCancelled":            "DragSourceEventCancelled demo",
		"Capabilities(":                       "system.Capabilities capability grid demo",
		"Availability(":                       "system.Availability capability grid demo",
		"CurrentWindowNativeHandle(":          "current native owner handle demo",
		"WindowSetResizable(":                 "WindowSetResizable demo",
		"WindowSetDecorated(":                 "WindowSetDecorated demo",
		"WindowSetMinSize(":                   "WindowSetMinSize demo",
		"WindowSetMaxSize(":                   "WindowSetMaxSize demo",
		"WindowSetWindowsFrameStyle(":         "WindowSetWindowsFrameStyle demo",
		"WindowStartDragMove(":                "WindowStartDragMove demo",
		"WindowDragAreaElement(":              "WindowDragAreaElement demo",
		"WindowMaximizeButtonElement(":        "WindowMaximizeButtonElement Snap Flyout demo",
		"WindowMaximizeButtonDisabled(":       "WindowMaximizeButtonDisabled demo",
		"ProbeWindowsChrome(":                 "ProbeWindowsChrome demo",
		".SetCloseRequestedHandler(":          "WindowHandle close-request guard demo",
		".PollEvents(":                        "WindowHandle PollEvents demo",
		".SubscribeEvents(":                   "WindowHandle SubscribeEvents demo",
		"FileDialogDefaultDir(":               "FileDialogDefaultDir demo",
		"FileDialogOwner(":                    "FileDialogOwner demo",
		"FileDialogAllowCreateDirs(":          "FileDialogAllowCreateDirs demo",
		"FileDialogAllowMissingPath(":         "FileDialogAllowMissingPath demo",
		"FileDialogOverwritePrompt(":          "FileDialogOverwritePrompt demo",
		"ShowMessageBoxAsyncContext(":         "ui.ShowMessageBoxAsyncContext demo",
		"ShowMessageBoxDetailedAsyncContext(": "ui.ShowMessageBoxDetailedAsyncContext demo",
		"MessageBoxOwner(":                    "MessageBoxOwner demo",
		"MessageBoxExpandedDetailsByDefault(": "MessageBoxExpandedDetailsByDefault demo",
		"MessageBoxCommandLinksNoIcon(":       "MessageBoxCommandLinksNoIcon demo",
		"NotificationBackendPath(":            "NotificationBackendPath demo",
		"NotificationIcon(":                   "NotificationIcon demo",
		"NotificationTimeout(":                "NotificationTimeout demo",
		"NotificationBackendBalloon":          "NotificationBackendBalloon demo",
		"TrayIcon(":                           "TrayIcon path demo",
		"TrayIconBytes(":                      "TrayIconBytes demo",
		"TrayIconResource(":                   "TrayIconResource demo",
		"TrayMenuItems(":                      "TrayMenuItems demo",
		".SetTooltip(":                        "Tray SetTooltip demo",
		".SetMenu(":                           "Tray SetMenu demo",
		".SetMenuProvider(":                   "Tray SetMenuProvider demo",
		".SetIcon(":                           "Tray SetIcon demo",
		".SetIconBytes(":                      "Tray SetIconBytes demo",
		".SetIconResource(":                   "Tray SetIconResource demo",
	}

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob source files: %v", err)
	}
	var source strings.Builder
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		source.Write(data)
		source.WriteByte('\n')
	}

	text := source.String()
	for token, label := range required {
		if !strings.Contains(text, token) {
			t.Fatalf("missing %s in docs browser demo source; expected token %q", label, token)
		}
	}
}

func TestAPIIndexElementsBuild(t *testing.T) {
	th := docsBrowserTheme(defaultDocsThemeSeed, false)
	state := &testStringState{}
	apis := []string{
		"ButtonElement(child Element, opts ...ButtonOption) Element",
		"TextElement(content string, opts ...TextOption) Element",
	}
	if docsAPIIndexSection(apis, state, th) == nil {
		t.Fatal("expected API index section")
	}
	if docsAPIIndexRow(apis[0], state, th) == nil {
		t.Fatal("expected API index row")
	}
	if got := apiCountLabel(1); got != "1 个 API" {
		t.Fatalf("apiCountLabel(1) = %q", got)
	}
	if got := apiCountLabel(2); got != "2 个 API" {
		t.Fatalf("apiCountLabel(2) = %q", got)
	}
}

func TestRightPanelContentBuilds(t *testing.T) {
	th := docsBrowserTheme(defaultDocsThemeSeed, false)
	open := &testBoolState{}
	status := &testStringState{}
	doc := &widgetDoc{
		Meta: docMeta{
			ID:       "button",
			Title:    "Button",
			Category: "Input",
			Summary:  "Button docs",
			Example:  docDemo{ID: "button_basic"},
			APIs:     []string{"ButtonElement(child Element, opts ...ButtonOption) Element"},
		},
		Content: "# Button\n\n```go\nui.ButtonElement(ui.TextElement(\"OK\"))\n```",
	}

	markdown := renderMarkdownDocument(doc.Content, th, status)
	content := docsRightPanelContent(doc, th, open, status, markdown, func(*widgetDoc) ui.Element {
		return ui.TextElement("demo")
	})
	if len(content) < 6 {
		t.Fatalf("expected right panel content, got %d elements", len(content))
	}
	if docsRightPanelList(content) == nil {
		t.Fatal("expected virtualized right panel list")
	}

	empty := docsRightPanelContent(nil, th, open, status, nil, nil)
	if len(empty) == 0 {
		t.Fatal("expected empty state when no document is selected")
	}
}

func TestEveryLoadedDocBuildsDemoAndRightPanel(t *testing.T) {
	docs, err := loadWidgetDocs()
	if err != nil {
		t.Fatalf("load docs: %v", err)
	}

	ctx, cleanup := newDocsTestContext()
	defer cleanup()
	th := docsBrowserTheme(defaultDocsThemeSeed, false)
	state := newTestDocsDemoState(th)

	for i := range docs {
		doc := &docs[i]
		t.Run(doc.Meta.ID, func(t *testing.T) {
			demo := buildDocsDemo(ctx, doc, state)
			if demo == nil {
				t.Fatalf("%s built nil demo", doc.Meta.ID)
			}

			content := docsRightPanelContent(
				doc,
				th,
				&testBoolState{},
				&testStringState{},
				renderMarkdownDocument(doc.Content, th, &testStringState{}),
				func(doc *widgetDoc) ui.Element {
					return buildDocsDemo(ctx, doc, state)
				},
			)
			if len(content) < 6 {
				t.Fatalf("%s right panel content too small: %d elements", doc.Meta.ID, len(content))
			}
		})
	}
}

func TestDocsBrowserAppRootBuilds(t *testing.T) {
	docs, err := loadWidgetDocs()
	if err != nil {
		t.Fatalf("load docs: %v", err)
	}
	ctx, cleanup := newDocsTestContext()
	defer cleanup()

	root := docsBrowserApp(ctx, &docsRuntimeState{
		Docs:   docs,
		Source: "local",
	})
	if root == nil {
		t.Fatal("expected docs browser root element")
	}
	if widget := ui.RenderElement(root); widget == nil {
		t.Fatal("expected docs browser root widget")
	}
}

func TestExampleViewerElementsBuild(t *testing.T) {
	th := docsBrowserTheme(defaultDocsThemeSeed, false)
	open := &testBoolState{}
	demo := ui.TextElement("demo")

	if docsExampleSectionHeader("Button", "button_basic", open, th) == nil {
		t.Fatal("expected example section header")
	}
	if docsInlineExampleFrame(120, demo, th) == nil {
		t.Fatal("expected inline example frame")
	}
	if docsExamplePopup(true, "Button", "button_basic", demo, open, th) == nil {
		t.Fatal("expected example popup")
	}
}

func TestDocsCategoryFiltering(t *testing.T) {
	docs := []widgetDoc{
		{Meta: docMeta{ID: "button", Category: "Components"}},
		{Meta: docMeta{ID: "system", Category: "Guides"}},
		{Meta: docMeta{ID: "misc"}},
	}

	categories := docsCategories(docs)
	if len(categories) != 3 {
		t.Fatalf("expected three categories, got %v", categories)
	}
	orderedCategories := docsCategories([]widgetDoc{
		{Meta: docMeta{ID: "system", Category: "系统"}},
		{Meta: docMeta{ID: "guide", Category: "使用指南"}},
		{Meta: docMeta{ID: "input", Category: "输入交互"}},
	})
	if strings.Join(orderedCategories, ",") != "使用指南,输入交互,系统" {
		t.Fatalf("category order = %v", orderedCategories)
	}
	filtered := filterDocsByCategory(docs, "Guides")
	if len(filtered) != 1 || filtered[0].Meta.ID != "system" {
		t.Fatalf("unexpected category filter result: %#v", filtered)
	}
	uncategorized := filterDocsByCategory(docs, "未分类")
	if len(uncategorized) != 1 || uncategorized[0].Meta.ID != "misc" {
		t.Fatalf("unexpected uncategorized filter result: %#v", uncategorized)
	}
	all := filterDocsByCategory(docs, docsAllCategories)
	if len(all) != len(docs) {
		t.Fatalf("all category should keep docs, got %d", len(all))
	}
	counts := docsCategoryCounts(docs)
	if counts["Components"] != 1 || counts["Guides"] != 1 || counts["未分类"] != 1 {
		t.Fatalf("unexpected category counts: %#v", counts)
	}
	if got := docsCategoryLabel("Guides", 12); got != "Guides 12" {
		t.Fatalf("category label = %q", got)
	}
}

func TestDocsSearchCoversMetadataAPIAndContent(t *testing.T) {
	docs := []widgetDoc{
		{
			Meta: docMeta{
				ID:       "button",
				Title:    "Button",
				Category: "Input",
				Summary:  "Click actions",
				APIs:     []string{"ButtonElement(child Element, opts ...ButtonOption) Element"},
			},
			Content: "Primary action docs",
		},
		{
			Meta:    docMeta{ID: "router", Title: "Router", Category: "Navigation"},
			Content: "Route guard and transition examples",
		},
	}

	apiMatches := filterDocs(docs, "ButtonOption")
	if len(apiMatches) != 1 || apiMatches[0].Meta.ID != "button" {
		t.Fatalf("API search did not find button doc: %#v", apiMatches)
	}

	multiTermMatches := filterDocs(docs, "button primary")
	if len(multiTermMatches) != 1 || multiTermMatches[0].Meta.ID != "button" {
		t.Fatalf("multi-term search did not find button doc: %#v", multiTermMatches)
	}

	contentMatches := filterDocs(docs, "transition")
	if len(contentMatches) != 1 || contentMatches[0].Meta.ID != "router" {
		t.Fatalf("content search did not find router doc: %#v", contentMatches)
	}
	if matches := filterDocs(docs, "button transition"); len(matches) != 0 {
		t.Fatalf("multi-term AND search should reject split docs, got %#v", matches)
	}
	if terms := docsSearchTerms("  Button   Primary "); len(terms) != 2 || terms[0] != "button" || terms[1] != "primary" {
		t.Fatalf("unexpected search terms: %#v", terms)
	}
}

func TestDocsAPISearchSummaryBuildsQuickReference(t *testing.T) {
	docs := []widgetDoc{
		{
			Meta: docMeta{
				ID:    "button",
				Title: "Button",
				APIs: []string{
					"ButtonElement(child Element, opts ...ButtonOption) Element",
					"FilledButtonElement(child Element, opts ...ButtonOption) Element",
				},
			},
		},
		{
			Meta: docMeta{
				ID:    "router",
				Title: "Router",
				APIs:  []string{"UseNavigate(ctx *Context) NavigateFunc"},
			},
		},
	}

	matches, total := docsAPISearchMatches(docs, "button", 1)
	if total != 2 {
		t.Fatalf("api search total = %d, want 2", total)
	}
	if len(matches) != 1 || matches[0].DocID != "button" {
		t.Fatalf("api search preview = %#v, want one button match", matches)
	}
	matches, total = docsAPISearchMatches(docs, "filled buttonoption", 5)
	if total != 1 || len(matches) != 1 || matches[0].API != "FilledButtonElement(child Element, opts ...ButtonOption) Element" {
		t.Fatalf("multi-term api search = total %d matches %#v", total, matches)
	}
	if got := docsAPIShortSignature("abcdefghijklmnopqrstuvwxyz", 8); got != "abcde..." {
		t.Fatalf("short signature = %q", got)
	}

	th := docsBrowserTheme(defaultDocsThemeSeed, false)
	if docsAPISearchSummary(docs, "button", &testStringState{}, th) == nil {
		t.Fatal("expected API search summary element")
	}
	if docsAPISearchResultRow(matches[0], &testStringState{}, th) == nil {
		t.Fatal("expected API search result row element")
	}
}

func TestDocsThemeSeedSwitching(t *testing.T) {
	flux := docsBrowserTheme("flux", false)
	mint := docsBrowserTheme("mint", false)
	dark := docsBrowserTheme("flux", true)
	if flux == nil || mint == nil || dark == nil {
		t.Fatal("expected themes for all docs browser variants")
	}
	if flux.Primary == mint.Primary {
		t.Fatalf("theme seed should change primary color: %v", flux.Primary)
	}
	if flux.Surface == dark.Surface {
		t.Fatalf("dark mode should change surface color: %v", flux.Surface)
	}
}

func TestDocsLeftPanelUsesThemeSemanticColors(t *testing.T) {
	data, err := os.ReadFile("left_panel.go")
	if err != nil {
		t.Fatalf("read left panel source: %v", err)
	}
	text := string(data)
	for _, token := range []string{
		"th.Colors.SurfaceContainerLow",
		"th.Colors.OnSurfaceVariant",
		"th.Colors.SecondaryContainer",
		"th.Colors.OnSecondaryContainer",
		"th.Colors.OutlineVariant",
	} {
		if !strings.Contains(text, token) {
			t.Fatalf("left panel should use theme semantic color %s", token)
		}
	}
	for _, oldColor := range []string{
		"ui.NRGBA(248, 250, 252, 255)",
		"ui.NRGBA(226, 232, 240, 255)",
		"ui.NRGBA(203, 213, 225, 255)",
		"ui.NRGBA(100, 116, 139, 255)",
	} {
		if strings.Contains(text, oldColor) {
			t.Fatalf("left panel still uses hard-coded light color %s", oldColor)
		}
	}
}

type testStringState struct {
	value string
}

func (s *testStringState) Value() string {
	return s.value
}

func (s *testStringState) Set(value string) {
	s.value = value
}

type testBoolState struct {
	value bool
}

func (s *testBoolState) Value() bool {
	return s.value
}

func (s *testBoolState) Set(value bool) {
	s.value = value
}

type testIntState struct {
	value int
}

func (s *testIntState) Value() int {
	return s.value
}

func (s *testIntState) Set(value int) {
	s.value = value
}

type testFloat32State struct {
	value float32
}

func (s *testFloat32State) Value() float32 {
	return s.value
}

func (s *testFloat32State) Set(value float32) {
	s.value = value
}

type testStringSliceState struct {
	value []string
}

func (s *testStringSliceState) Value() []string {
	return s.value
}

func (s *testStringSliceState) Set(value []string) {
	s.value = value
}

type testDragDropStateHandle struct {
	value docsDragDropState
}

func (s *testDragDropStateHandle) Value() docsDragDropState {
	return s.value
}

func (s *testDragDropStateHandle) Set(value docsDragDropState) {
	s.value = value
}

func newDocsTestContext() (*ui.Context, func()) {
	runtime := internal.NewRuntime(ui.NewTheme(ui.LightColors()))
	var ops op.Ops
	ctx := runtime.Frame(gioLayout.Context{Ops: &ops})
	return ctx, runtime.Dispose
}

func newTestDocsDemoState(th *ui.Theme) docsDemoState {
	stringState := func(value string) *testStringState { return &testStringState{value: value} }
	boolState := func(value bool) *testBoolState { return &testBoolState{value: value} }
	intState := func(value int) *testIntState { return &testIntState{value: value} }

	return docsDemoState{
		Theme:                   th,
		SystemDemoStatus:        stringState("ready"),
		ButtonCount:             intState(0),
		InputValue:              stringState("FluxUI"),
		CheckboxValue:           boolState(true),
		SwitchValue:             boolState(true),
		SliderValue:             &testFloat32State{value: 40},
		StyleDemoActive:         boolState(false),
		AnimationDemoActive:     boolState(false),
		ThemeDemoDark:           boolState(false),
		RadioValue:              stringState("layout"),
		SelectValue:             stringState("medium"),
		TabValue:                stringState("overview"),
		DialogOpen:              boolState(false),
		PopupOpen:               boolState(false),
		ToastMessage:            stringState(""),
		SnackbarMessage:         stringState(""),
		SnackbarSerial:          intState(0),
		SnackbarActionCount:     intState(0),
		ChipSelected:            boolState(true),
		SearchBarValue:          stringState(""),
		BottomNavValue:          stringState("home"),
		MenuOpen:                boolState(false),
		MenuValue:               stringState("copy"),
		ListItemSelected:        stringState("inbox"),
		IconButtonSelected:      boolState(true),
		FabCount:                intState(0),
		RailValue:               stringState("home"),
		DrawerValue:             stringState("inbox"),
		ClickCount:              intState(0),
		AppbarActionCount:       intState(0),
		ListReachEndCount:       intState(0),
		HookDemoCount:           intState(0),
		HookDemoShowChild:       boolState(false),
		HookDemoLogs:            &testStringSliceState{value: []string{}},
		RouterDemoAllowSettings: boolState(true),
		RouterDemoUserID:        stringState("u1001"),
		RouterDemoLog:           stringState("Router demo ready"),
		DragDropDemoState:       &testDragDropStateHandle{value: defaultDocsDragDropState()},
	}
}

func implementedDocsDemoCases(t *testing.T) map[string]struct{} {
	t.Helper()
	cases := map[string]struct{}{}
	collectDemoCasesFromFile(t, "demo_registry.go", cases)
	if len(cases) == 0 {
		t.Fatal("expected buildDemo cases to be discovered")
	}
	return cases
}

func collectDemoCasesFromFile(t *testing.T, path string, cases map[string]struct{}) {
	t.Helper()
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	ast.Inspect(file, func(node ast.Node) bool {
		decl, ok := node.(*ast.FuncDecl)
		if !ok || decl.Name == nil || decl.Name.Name != "buildDocsDemo" {
			return true
		}
		ast.Inspect(decl.Body, func(inner ast.Node) bool {
			clause, ok := inner.(*ast.CaseClause)
			if !ok {
				return true
			}
			for _, expr := range clause.List {
				basic, ok := expr.(*ast.BasicLit)
				if !ok || basic.Kind != token.STRING {
					continue
				}
				value, err := strconv.Unquote(basic.Value)
				if err != nil {
					t.Fatalf("unquote buildDocsDemo case %s: %v", basic.Value, err)
				}
				cases[value] = struct{}{}
			}
			return true
		})
		return true
	})
}
