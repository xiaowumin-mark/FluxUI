//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	reStyle = regexp.MustCompile(`ui\.ContainerElement\(\s*ui\.Style\{`)
	reBg    = regexp.MustCompile(`Background:\s*([^,\n]+)`)
	rePad   = regexp.MustCompile(`Padding:\s*([^,\n]+)`)
	reRad   = regexp.MustCompile(`Radius:\s*([^,\n]+)`)
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run tools/convert_style.go <file>")
		return
	}
	path := os.Args[1]
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	content := string(data)

	lines := strings.Split(content, "\n")
	var out []string
	i := 0
	for i < len(lines) {
		line := lines[i]
		if !strings.Contains(line, "ui.ContainerElement(") || !strings.Contains(line, "ui.Style{") {
			out = append(out, line)
			i++
			continue
		}

		indent := ""
		for _, ch := range line {
			if ch == ' ' || ch == '\t' {
				indent += string(ch)
			} else {
				break
			}
		}
		trimmed := strings.TrimSpace(line)

		if trimmed == "ui.ContainerElement(" {
			out = append(out, indent+"ui.ContainerDecorationElement(")
			i++
			for i < len(lines) && strings.TrimSpace(lines[i]) != "ui.Style{" {
				out = append(out, lines[i])
				i++
			}
		} else {
			out = append(out, strings.Replace(line, "ui.ContainerElement(", "ui.ContainerDecorationElement(", 1))
			i++
			for i < len(lines) && strings.TrimSpace(lines[i]) == "ui.Style{" {
				i++
			}
		}

		if i < len(lines) && strings.TrimSpace(lines[i]) == "ui.Style{" {
			i++

			var bg, pad, rad, border string
		styleLoop:
			for i < len(lines) {
				sline := lines[i]
				trim := strings.TrimSuffix(strings.TrimSpace(sline), ",")
				switch {
				case strings.HasPrefix(trim, "Background:"):
					bg = strings.TrimSpace(strings.TrimPrefix(trim, "Background:"))
				case strings.HasPrefix(trim, "Padding:"):
					pad = strings.TrimSpace(strings.TrimPrefix(trim, "Padding:"))
				case strings.HasPrefix(trim, "Radius:"):
					rad = strings.TrimSpace(strings.TrimPrefix(trim, "Radius:"))
				case strings.HasPrefix(trim, "Border:"):
					border = strings.TrimSpace(strings.TrimPrefix(trim, "Border:"))
				case trim == "}" || trim == "},":
					break styleLoop
				}
				i++
			}

			var parts []string
			if bg != "" && bg != "ui.NRGBA(0, 0, 0, 0)" {
				parts = append(parts, "ui.Bg("+bg+")")
			}
			if pad != "" {
				parts = append(parts, "WithPad("+pad+")")
			}
			if rad != "" && rad != "0" {
				parts = append(parts, "WithRad("+rad+")")
			}
			if border != "" {
				parts = append(parts, "WithBorder("+border+")")
			}

			if len(parts) > 0 {
				out = append(out, indent+"\t"+strings.Join(parts, ".")+",")
			}
		}
	}

	newContent := strings.Join(out, "\n")

	var buf bytes.Buffer
	for _, l := range strings.Split(newContent, "\n") {
		if strings.TrimSpace(l) == "ui.ContainerDecorationElement(" {
			continue
		}
		buf.WriteString(l + "\n")
	}

	if err := os.WriteFile(path, []byte(strings.TrimRight(buf.String(), "\n")+"\n"), 0644); err != nil {
		panic(err)
	}
	fmt.Printf("converted %s\n", path)
}
