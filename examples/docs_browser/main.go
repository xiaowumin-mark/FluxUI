package main

import (
	_ "github.com/xiaowumin-mark/FluxUI/icons/md3"
	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

func main() {
	if maybeRunDocsBrowserToastActivator() {
		return
	}

	initialDocs, loadErr := loadWidgetDocs()
	runtimeState := &docsRuntimeState{
		Docs:    initialDocs,
		Source:  "local",
		LoadErr: loadErr,
	}
	if len(runtimeState.Docs) == 0 {
		resultCh := make(chan remoteLoadResult, 1)
		runtimeState.Source = "online"
		runtimeState.Loading = true
		runtimeState.ResultCh = resultCh
		runtimeState.Docs = []widgetDoc{buildOnlineLoadingDoc(loadErr)}
		go func() {
			remoteDocs, err := loadWidgetDocsFromGitHub()
			resultCh <- remoteLoadResult{Docs: remoteDocs, Err: err}
		}()
	}

	app := func(ctx *ui.Context) ui.Element {
		return docsBrowserApp(ctx, runtimeState)
	}

	_ = ui.RunElement(app, ui.Title("FluxUI Docs Browser"), ui.Size(1360, 880))
}
