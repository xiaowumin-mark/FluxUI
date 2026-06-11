package main

type docMeta struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Category string   `json:"category"`
	Order    int      `json:"order"`
	Summary  string   `json:"summary"`
	Example  docDemo  `json:"example"`
	APIs     []string `json:"apis"`
}

type docDemo struct {
	ID    string            `json:"id"`
	Props map[string]string `json:"props"`
}

type widgetDoc struct {
	Meta    docMeta
	Content string
	Path    string
}

type menuEntry struct {
	IsCategory bool
	Category   string
	Doc        *widgetDoc
}

type remoteLoadResult struct {
	Docs []widgetDoc
	Err  error
}

type docsRuntimeState struct {
	Docs       []widgetDoc
	Source     string
	LoadErr    error
	OnlineErr  error
	Loading    bool
	ResultCh   <-chan remoteLoadResult
	ResultDone bool
}

type githubContentEntry struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}
