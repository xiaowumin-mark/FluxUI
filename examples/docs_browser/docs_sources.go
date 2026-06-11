package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	githubDocsAPIBase = "https://api.github.com/repos/xiaowumin-mark/FluxUI/contents/docs"
	githubDocsRawBase = "https://raw.githubusercontent.com/xiaowumin-mark/FluxUI/main/docs"
)

var docsSubdirs = []string{"", "widgets", "style", "theme", "guides"}

func applyRemoteDocsResult(state *docsRuntimeState) {
	if state == nil || !state.Loading || state.ResultDone || state.ResultCh == nil {
		return
	}
	select {
	case result := <-state.ResultCh:
		state.ResultDone = true
		state.Loading = false
		state.OnlineErr = result.Err
		if len(result.Docs) > 0 {
			state.Docs = result.Docs
			state.LoadErr = nil
			return
		}
		state.Docs = []widgetDoc{buildOnlineLoadFailedDoc(state.LoadErr, state.OnlineErr)}
		if state.LoadErr != nil && state.OnlineErr != nil {
			state.LoadErr = fmt.Errorf("本地加载失败: %v；在线加载失败: %v", state.LoadErr, state.OnlineErr)
		} else if state.OnlineErr != nil {
			state.LoadErr = fmt.Errorf("在线加载失败: %w", state.OnlineErr)
		}
	default:
	}
}

func loadWidgetDocs() ([]widgetDoc, error) {
	docsRoot, err := resolveDocsRootDir()
	if err != nil {
		return nil, err
	}

	docs := make([]widgetDoc, 0, 48)
	for _, subdir := range docsSubdirs {
		dir := filepath.Join(docsRoot, subdir)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
				continue
			}

			path := filepath.Join(dir, entry.Name())
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				continue
			}

			doc, parseErr := parseWidgetDoc(path, string(data))
			if parseErr != nil {
				continue
			}
			docs = append(docs, doc)
		}
	}

	if len(docs) == 0 {
		return nil, errors.New("docs 下没有可解析的 Markdown 文档")
	}

	docs = append(docs, buildDocsBrowserOverviewDoc())
	sortDocs(docs)

	return docs, nil
}

func loadWidgetDocsFromGitHub() ([]widgetDoc, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	docs := make([]widgetDoc, 0, 48)
	var firstErr error
	for _, subdir := range docsSubdirs {
		entries, err := fetchGitHubDocsEntries(client, subdir)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}

		for _, entry := range entries {
			if entry.Type != "file" {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(entry.Name), ".md") {
				continue
			}

			url := strings.TrimSpace(entry.DownloadURL)
			if url == "" {
				url = docsGitHubRawURL(subdir, entry.Name)
			}

			text, err := fetchHTTPText(client, url)
			if err != nil {
				continue
			}
			doc, err := parseWidgetDoc(url, text)
			if err != nil {
				continue
			}
			docs = append(docs, doc)
		}
	}

	if len(docs) == 0 {
		if firstErr != nil {
			return nil, firstErr
		}
		return nil, errors.New("GitHub docs 下没有可解析的 Markdown 文档")
	}

	docs = append(docs, buildDocsBrowserOverviewDoc())
	sortDocs(docs)

	return docs, nil
}

func fetchGitHubDocsEntries(client *http.Client, subdir string) ([]githubContentEntry, error) {
	body, err := fetchHTTPBytes(client, docsGitHubContentsURL(subdir), true)
	if err != nil {
		return nil, err
	}
	var entries []githubContentEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("解析 GitHub 文档目录失败: %w", err)
	}
	return entries, nil
}

func docsGitHubContentsURL(subdir string) string {
	subdir = cleanDocsSubdir(subdir)
	if subdir == "" {
		return githubDocsAPIBase + "?ref=main"
	}
	return githubDocsAPIBase + "/" + subdir + "?ref=main"
}

func docsGitHubRawURL(subdir string, name string) string {
	subdir = cleanDocsSubdir(subdir)
	name = strings.TrimLeft(strings.ReplaceAll(name, "\\", "/"), "/")
	if subdir == "" {
		return githubDocsRawBase + "/" + name
	}
	return githubDocsRawBase + "/" + subdir + "/" + name
}

func cleanDocsSubdir(subdir string) string {
	subdir = strings.ReplaceAll(subdir, "\\", "/")
	return strings.Trim(subdir, "/.")
}

func fetchHTTPText(client *http.Client, url string) (string, error) {
	data, err := fetchHTTPBytes(client, url, false)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func fetchHTTPBytes(client *http.Client, url string, api bool) ([]byte, error) {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "FluxUI-DocsBrowser")
	if api {
		req.Header.Set("Accept", "application/vnd.github+json")
	} else {
		req.Header.Set("Accept", "text/plain, */*")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := strings.TrimSpace(string(data))
		if len(snippet) > 160 {
			snippet = snippet[:160] + "..."
		}
		if snippet != "" {
			return nil, fmt.Errorf("请求 %s 失败: HTTP %d, %s", url, resp.StatusCode, snippet)
		}
		return nil, fmt.Errorf("请求 %s 失败: HTTP %d", url, resp.StatusCode)
	}
	return data, nil
}

func buildOnlineLoadingDoc(localErr error) widgetDoc {
	lines := []string{
		"# 正在加载在线文档",
		"",
		"本地 `docs` 文档目录不可用，正在从 GitHub 拉取在线 Markdown 文档。",
		"",
		"在线来源：",
		"- https://github.com/xiaowumin-mark/FluxUI/tree/main/docs",
	}
	if localErr != nil {
		lines = append(lines, "", "本地错误："+localErr.Error())
	}
	return widgetDoc{
		Meta: docMeta{
			ID:       "loading_online_docs",
			Title:    "在线文档加载中",
			Category: "系统",
			Order:    1,
			Summary:  "本地文档不可用，正在异步加载在线文档。",
			Example:  docDemo{ID: "fallback"},
		},
		Content: strings.Join(lines, "\n"),
		Path:    githubDocsAPIBase,
	}
}

func buildOnlineLoadFailedDoc(localErr, onlineErr error) widgetDoc {
	lines := []string{
		"# 文档加载失败",
		"",
		"本地与在线文档都未能加载，请检查：",
		"- 本地 `docs` 目录是否存在且可读",
		"- 网络连接是否可访问 GitHub",
		"- 文档文件是否包含 `fluxui-doc-meta` 元数据块",
	}
	if localErr != nil {
		lines = append(lines, "", "本地错误："+localErr.Error())
	}
	if onlineErr != nil {
		lines = append(lines, "在线错误："+onlineErr.Error())
	}
	return widgetDoc{
		Meta: docMeta{
			ID:       "load_error",
			Title:    "文档加载失败",
			Category: "系统",
			Order:    1,
			Summary:  "本地与在线文档均不可用。",
			Example:  docDemo{ID: "fallback"},
		},
		Content: strings.Join(lines, "\n"),
		Path:    githubDocsAPIBase,
	}
}

func resolveDocsRootDir() (string, error) {
	candidates := make([]string, 0, 12)
	candidates = append(candidates,
		"docs",
		filepath.Join("..", "docs"),
		filepath.Join("..", "..", "docs"),
	)

	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, "docs"),
			filepath.Join(cwd, "..", "docs"),
			filepath.Join(cwd, "..", "..", "docs"),
		)
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "docs"),
			filepath.Join(exeDir, "..", "docs"),
			filepath.Join(exeDir, "..", "..", "docs"),
			filepath.Join(exeDir, "..", "..", "..", "docs"),
		)
	}

	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		cleaned := filepath.Clean(candidate)
		if _, ok := seen[cleaned]; ok {
			continue
		}
		seen[cleaned] = struct{}{}
		info, err := os.Stat(cleaned)
		if err != nil {
			continue
		}
		if info.IsDir() {
			return cleaned, nil
		}
	}
	return "", errors.New("未找到 docs 文档目录")
}
