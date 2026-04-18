package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	ui "github.com/xiaowumin-mark/FluxUI/ui"
)

type githubRepo struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Stars       int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	OpenIssues  int    `json:"open_issues_count"`
	Language    string `json:"language"`
	UpdatedAt   string `json:"updated_at"`
}

type fetchState struct {
	mu      sync.Mutex
	loading bool
	result  *githubRepo
	errMsg  string
	doneAt  time.Time
}

func main() {
	fetcher := &fetchState{}

	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)

		owner := ui.State[string](ctx)
		repo := ui.State[string](ctx)
		inited := ui.State[bool](ctx)

		loading := ui.State[bool](ctx)
		errorMsg := ui.State[string](ctx)
		lastUpdated := ui.State[string](ctx)
		fullName := ui.State[string](ctx)
		description := ui.State[string](ctx)
		language := ui.State[string](ctx)
		stars := ui.State[int](ctx)
		forks := ui.State[int](ctx)
		openIssues := ui.State[int](ctx)

		if !inited.Value() {
			owner.Set("xiaowumin-mark")
			repo.Set("FluxUI")
			inited.Set(true)
		}

		startFetch := func() {
			o := strings.TrimSpace(owner.Value())
			r := strings.TrimSpace(repo.Value())
			if o == "" || r == "" {
				errorMsg.Set("仓库 owner/repo 不能为空")
				return
			}
			if !fetcher.start(o, r) {
				return
			}
			loading.Set(true)
			errorMsg.Set("")
		}

		if loading.Value() {
			done, result, errMsg, doneAt := fetcher.poll()
			if done {
				loading.Set(false)
				if errMsg != "" {
					errorMsg.Set(errMsg)
				} else if result != nil {
					errorMsg.Set("")
					fullName.Set(result.FullName)
					description.Set(strings.TrimSpace(result.Description))
					language.Set(strings.TrimSpace(result.Language))
					stars.Set(result.Stars)
					forks.Set(result.Forks)
					openIssues.Set(result.OpenIssues)

					displayTime := doneAt
					if displayTime.IsZero() {
						displayTime = time.Now()
					}
					lastUpdated.Set(displayTime.Format("2006-01-02 15:04:05"))
				}
			} else {
				ctx.RequestRedraw()
			}
		}

		infoLines := []ui.Widget{
			ui.Text("网络请求示例（异步，不阻塞 UI）", ui.TextSize(22)),
			ui.Padding(
				ui.Insets{Top: 8},
				ui.Text("输入 GitHub 仓库 owner/repo，点击“请求仓库信息”后在后台加载。", ui.TextSize(13), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
			),
		}

		form := ui.Column(
			ui.Padding(
				ui.Insets{Top: 14},
				ui.Text("Owner", ui.TextSize(13)),
			),
			ui.TextField(
				owner.Value(),
				ui.InputPlaceholder("例如：xiaowumin-mark"),
				ui.InputOnChange(func(ctx *ui.Context, value string) {
					owner.Set(value)
				}),
			),
			ui.Padding(
				ui.Insets{Top: 10},
				ui.Text("Repo", ui.TextSize(13)),
			),
			ui.TextField(
				repo.Value(),
				ui.InputPlaceholder("例如：FluxUI"),
				ui.InputOnChange(func(ctx *ui.Context, value string) {
					repo.Set(value)
				}),
			),
		)

		statusText := "空闲"
		statusColor := ui.NRGBA(51, 65, 85, 255)
		if loading.Value() {
			statusText = "加载中..."
			statusColor = ui.NRGBA(2, 132, 199, 255)
		} else if errorMsg.Value() != "" {
			statusText = "请求失败"
			statusColor = ui.NRGBA(220, 38, 38, 255)
		} else if fullName.Value() != "" {
			statusText = "请求成功"
			statusColor = ui.NRGBA(22, 163, 74, 255)
		}

		actions := ui.Row(
			ui.Button(
				ui.Text("请求仓库信息"),
				ui.Disabled(loading.Value()),
				ui.OnClick(func(ctx *ui.Context) {
					startFetch()
				}),
			),
			ui.Padding(
				ui.Insets{Left: 10, Top: 8},
				ui.Text("状态: "+statusText, ui.TextSize(13), ui.TextColor(statusColor)),
			),
		)

		resultPanel := ui.Container(
			ui.Style{
				Background: ui.NRGBA(248, 250, 252, 255),
				Padding:    ui.All(12),
				Radius:     10,
			},
			ui.Column(
				ui.Text("结果", ui.TextSize(16)),
				ui.Padding(ui.Insets{Top: 8}, ui.Text("仓库: "+withFallback(fullName.Value(), "-"), ui.TextSize(13))),
				ui.Padding(ui.Insets{Top: 4}, ui.Text("描述: "+withFallback(description.Value(), "-"), ui.TextSize(13))),
				ui.Padding(ui.Insets{Top: 4}, ui.Text("语言: "+withFallback(language.Value(), "-"), ui.TextSize(13))),
				ui.Padding(ui.Insets{Top: 4}, ui.Text(fmt.Sprintf("Star: %d  Fork: %d  Open Issues: %d", stars.Value(), forks.Value(), openIssues.Value()), ui.TextSize(13))),
				ui.Padding(ui.Insets{Top: 4}, ui.Text("更新时间: "+withFallback(lastUpdated.Value(), "-"), ui.TextSize(12), ui.TextColor(ui.NRGBA(100, 116, 139, 255)))),
				func() ui.Widget {
					if strings.TrimSpace(errorMsg.Value()) == "" {
						return ui.Spacer(0, 0)
					}
					return ui.Padding(
						ui.Insets{Top: 8},
						ui.Text("错误: "+errorMsg.Value(), ui.TextSize(12), ui.TextColor(ui.NRGBA(220, 38, 38, 255))),
					)
				}(),
			),
		)

		return ui.Container(
			ui.Style{
				Background: th.Surface,
				Padding:    ui.All(16),
			},
			ui.ScrollView(
				ui.Column(append(
					infoLines,
					ui.Padding(ui.Insets{Top: 12}, form),
					ui.Padding(ui.Insets{Top: 12}, actions),
					ui.Padding(ui.Insets{Top: 12}, resultPanel),
				)...),
			),
		)
	}, ui.Title("FluxUI Network Request"), ui.Size(760, 520))
}

func withFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func (f *fetchState) start(owner, repo string) bool {
	if f == nil {
		return false
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.loading {
		return false
	}
	f.loading = true
	f.result = nil
	f.errMsg = ""
	f.doneAt = time.Time{}

	go func() {
		result, err := fetchRepo(owner, repo)

		f.mu.Lock()
		defer f.mu.Unlock()
		if err != nil {
			f.errMsg = err.Error()
		} else {
			f.result = result
		}
		f.doneAt = time.Now()
		f.loading = false
	}()
	return true
}

func (f *fetchState) poll() (done bool, result *githubRepo, errMsg string, doneAt time.Time) {
	if f == nil {
		return true, nil, "fetchState is nil", time.Now()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.loading {
		return false, nil, "", time.Time{}
	}
	return true, f.result, f.errMsg, f.doneAt
}

func fetchRepo(owner, repo string) (*githubRepo, error) {
	client := &http.Client{Timeout: 12 * time.Second}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "FluxUI-NetworkRequest-Example")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(body))
		if len(msg) > 160 {
			msg = msg[:160] + "..."
		}
		if msg == "" {
			msg = "unexpected status"
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
	}

	var result githubRepo
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
