package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func buildMenuEntries(docs []widgetDoc) []menuEntry {
	entries := make([]menuEntry, 0, len(docs)+8)
	lastCategory := ""
	for i := range docs {
		doc := &docs[i]
		category := doc.Meta.Category
		if category == "" {
			category = "未分类"
		}
		if category != lastCategory {
			entries = append(entries, menuEntry{
				IsCategory: true,
				Category:   category,
			})
			lastCategory = category
		}
		entries = append(entries, menuEntry{
			Doc: doc,
		})
	}
	return entries
}

func findDocByID(docs []widgetDoc, id string) *widgetDoc {
	for i := range docs {
		if docs[i].Meta.ID == id {
			return &docs[i]
		}
	}
	return nil
}

func sortDocs(docs []widgetDoc) {
	sort.Slice(docs, func(i, j int) bool {
		a := docs[i].Meta
		b := docs[j].Meta
		aRank := categoryRank(a.Category)
		bRank := categoryRank(b.Category)
		if aRank != bRank {
			return aRank < bRank
		}
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		if a.Order != b.Order {
			return a.Order < b.Order
		}
		return a.Title < b.Title
	})
}

func categoryRank(category string) int {
	switch category {
	case "使用指南":
		return 10
	case "样式系统":
		return 20
	case "主题系统":
		return 30
	case "状态与副作用":
		return 40
	case "布局系统":
		return 50
	case "基础显示":
		return 60
	case "输入交互":
		return 70
	case "反馈组件":
		return 80
	case "导航组件":
		return 90
	case "系统":
		return 100
	default:
		return 999
	}
}

func parseWidgetDoc(path string, raw string) (widgetDoc, error) {
	content := strings.TrimPrefix(raw, "\uFEFF")
	startMarker := "<!-- fluxui-doc-meta"
	leading := strings.TrimLeft(content, " \t\r\n")
	if !strings.HasPrefix(leading, startMarker) {
		return widgetDoc{}, fmt.Errorf("文档 %s 缺少 fluxui-doc-meta", filepath.Base(path))
	}
	start := len(content) - len(leading)

	rest := content[start+len(startMarker):]
	endRel := strings.Index(rest, "-->")
	if endRel < 0 {
		return widgetDoc{}, fmt.Errorf("文档 %s 的元数据注释未闭合", filepath.Base(path))
	}

	metaText := strings.TrimSpace(rest[:endRel])
	var meta docMeta
	if err := json.Unmarshal([]byte(metaText), &meta); err != nil {
		return widgetDoc{}, fmt.Errorf("文档 %s 的元数据 JSON 解析失败: %w", filepath.Base(path), err)
	}

	if meta.ID == "" {
		meta.ID = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	if meta.Title == "" {
		meta.Title = meta.ID
	}
	if meta.Category == "" {
		meta.Category = "未分类"
	}
	if meta.Order == 0 {
		meta.Order = 9999
	}
	if meta.Example.ID == "" {
		meta.Example.ID = meta.ID
	}

	body := strings.TrimSpace(content[:start] + rest[endRel+3:])
	return widgetDoc{
		Meta:    meta,
		Content: body,
		Path:    path,
	}, nil
}
