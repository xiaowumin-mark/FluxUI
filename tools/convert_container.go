//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run convert_container.go <file>")
		return
	}
	for _, path := range os.Args[1:] {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("skip %s: %v\n", path, err)
			continue
		}
		content := string(data)
		result := convertFile(content)
		if result != content {
			if err := os.WriteFile(path, []byte(result), 0644); err != nil {
				fmt.Printf("write %s: %v\n", path, err)
			} else {
				fmt.Printf("fixed %s\n", path)
			}
		}
	}
}

func convertFile(content string) string {
	re := regexp.MustCompile(`ui\.ContainerElement\(\s*\n(\s*)ui\.Style\{([^}]*)\}`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		sub := re.FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		styleBody := sub[2]

		bg := fieldValue(styleBody, "Background")
		pad := fieldValue(styleBody, "Padding")
		rad := fieldValue(styleBody, "Radius")
		border := fieldValue(styleBody, "Border")
		opacity := fieldValue(styleBody, "Opacity")

		if bg == "" && pad == "" && rad == "" && border == "" && opacity == "" {
			return strings.Replace(match, "ui.ContainerElement(", "ui.ContainerDecorationElement(", 1)
		}

		var parts []string
		if bg != "" {
			parts = append(parts, "ui.Bg("+bg+")")
		}
		if pad != "" {
			parts = append(parts, "WithPad("+pad+")")
		}
		if rad != "" {
			parts = append(parts, "WithRad("+rad+")")
		}
		if border != "" {
			parts = append(parts, "WithBorder("+border+")")
		}
		if opacity != "" {
			parts = append(parts, "WithOpacity("+opacity+")")
		}

		chain := strings.Join(parts, ".")
		result := "ui.ContainerDecorationElement(\n" + sub[1] + chain
		return result
	})
}

func fieldValue(body, name string) string {
	idx := strings.Index(body, name+":")
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(body[idx+len(name)+1:])
	rest = strings.TrimSpace(rest)
	end := strings.IndexAny(rest, ",\n")
	if end < 0 {
		return ""
	}
	val := rest[:end]
	val = strings.TrimSuffix(val, ",")
	return strings.TrimSpace(val)
}

// Also fix ContainerElement(Style{...}) on a single line
func fixSingleLine(content string) string {
	re := regexp.MustCompile(`ui\.ContainerElement\(ui\.Style\{([^}]*)\},`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		sub := re.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		styleBody := sub[1]
		bg := fieldValue(styleBody, "Background")
		pad := fieldValue(styleBody, "Padding")
		rad := fieldValue(styleBody, "Radius")

		if bg == "" && pad == "" && rad == "" {
			return match
		}

		var parts []string
		if bg != "" {
			parts = append(parts, "ui.Bg("+bg+")")
		}
		if pad != "" {
			parts = append(parts, "WithPad("+pad+")")
		}
		if rad != "" {
			parts = append(parts, "WithRad("+rad+")")
		}

		chain := strings.Join(parts, ".")
		result := bytes.NewBufferString("ui.ContainerDecorationElement(" + chain + ",")
		return result.String()
	})
}
