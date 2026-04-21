package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

func main() {
	_ = ui.Run(func(ctx *ui.Context) ui.Widget {
		th := ui.UseTheme(ctx)

		owner := ui.State[string](ctx)
		repo := ui.State[string](ctx)
		inited := ui.State[bool](ctx)
		fetch := ui.UseAsync[*githubRepo](ctx)

		if !inited.Value() {
			owner.Set("xiaowumin-mark")
			repo.Set("FluxUI")
			inited.Set(true)
		}

		startFetch := func() {
			o := strings.TrimSpace(owner.Value())
			r := strings.TrimSpace(repo.Value())
			if o == "" || r == "" {
				return
			}
			fetch.Run(func() (*githubRepo, error) {
				return fetchRepo(o, r)
			})
		}

		statusText := "空闲"
		statusColor := ui.NRGBA(51, 65, 85, 255)
		switch fetch.Status() {
		case ui.AsyncLoading:
			statusText = "加载中..."
			statusColor = ui.NRGBA(2, 132, 199, 255)
		case ui.AsyncError:
			statusText = "请求失败"
			statusColor = ui.NRGBA(220, 38, 38, 255)
		case ui.AsyncSuccess:
			statusText = "请求成功"
			statusColor = ui.NRGBA(22, 163, 74, 255)
		}

		result := fetch.Data()

		return ui.Container(
			ui.Style{Background: th.Surface, Padding: ui.All(16)},
			ui.ScrollView(
				ui.Column(
					ui.Text("网络请求示例（异步，不阻塞 UI）", ui.TextSize(22)),
					ui.Padding(
						ui.Insets{Top: 8},
						ui.Text("输入 GitHub 仓库 owner/repo，点击「请求仓库信息」后在后台加载。",
							ui.TextSize(13), ui.TextColor(ui.NRGBA(71, 85, 105, 255))),
					),
					ui.Padding(ui.Insets{Top: 14}, ui.Text("Owner", ui.TextSize(13))),
					ui.TextField(owner.Value(),
						ui.InputPlaceholder("例如：xiaowumin-mark"),
						ui.InputOnChange(func(ctx *ui.Context, v string) { owner.Set(v) }),
					),
					ui.Padding(ui.Insets{Top: 10}, ui.Text("Repo", ui.TextSize(13))),
					ui.TextField(repo.Value(),
						ui.InputPlaceholder("例如：FluxUI"),
						ui.InputOnChange(func(ctx *ui.Context, v string) { repo.Set(v) }),
					),
					ui.Padding(
						ui.Insets{Top: 12},
						ui.Row(
							ui.Button(
								ui.Text("请求仓库信息"),
								ui.Disabled(fetch.Loading()),
								ui.OnClick(func(ctx *ui.Context) { startFetch() }),
							),
							ui.Padding(
								ui.Insets{Left: 10, Top: 8},
								ui.Text("状态: "+statusText, ui.TextSize(13), ui.TextColor(statusColor)),
							),
						),
					),
					ui.Padding(
						ui.Insets{Top: 12},
						ui.Container(
							ui.Style{
								Background: ui.NRGBA(248, 250, 252, 255),
								Padding:    ui.All(12),
								Radius:     10,
							},
							ui.Column(
								ui.Text("结果", ui.TextSize(16)),
								ui.Padding(ui.Insets{Top: 8}, ui.Text("仓库: "+repoField(result, func(r *githubRepo) string { return r.FullName }), ui.TextSize(13))),
								ui.Padding(ui.Insets{Top: 4}, ui.Text("描述: "+repoField(result, func(r *githubRepo) string { return r.Description }), ui.TextSize(13))),
								ui.Padding(ui.Insets{Top: 4}, ui.Text("语言: "+repoField(result, func(r *githubRepo) string { return r.Language }), ui.TextSize(13))),
								ui.Padding(ui.Insets{Top: 4}, ui.Text(repoStats(result), ui.TextSize(13))),
								func() ui.Widget {
									if fetch.Error() == nil {
										return ui.Spacer(0, 0)
									}
									return ui.Padding(
										ui.Insets{Top: 8},
										ui.Text("错误: "+fetch.Error().Error(), ui.TextSize(12), ui.TextColor(ui.NRGBA(220, 38, 38, 255))),
									)
								}(),
							),
						),
					),
				),
			),
		)
	}, ui.Title("FluxUI Network Request"), ui.Size(760, 520))
}

func repoField(r *githubRepo, fn func(*githubRepo) string) string {
	if r == nil {
		return "-"
	}
	v := strings.TrimSpace(fn(r))
	if v == "" {
		return "-"
	}
	return v
}

func repoStats(r *githubRepo) string {
	if r == nil {
		return "Star: 0  Fork: 0  Open Issues: 0"
	}
	return fmt.Sprintf("Star: %d  Fork: %d  Open Issues: %d", r.Stars, r.Forks, r.OpenIssues)
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
