//go:build linux

package system

func unixClipboardCandidates() []unixClipboardProvider {
	return []unixClipboardProvider{
		{
			name:  "wl-clipboard",
			read:  unixClipboardCommand{name: "wl-paste", args: []string{"--no-newline"}},
			write: unixClipboardCommand{name: "wl-copy"},
		},
		{
			name:  "xclip",
			read:  unixClipboardCommand{name: "xclip", args: []string{"-selection", "clipboard", "-out"}},
			write: unixClipboardCommand{name: "xclip", args: []string{"-selection", "clipboard", "-in"}},
		},
		{
			name:  "xsel",
			read:  unixClipboardCommand{name: "xsel", args: []string{"--clipboard", "--output"}},
			write: unixClipboardCommand{name: "xsel", args: []string{"--clipboard", "--input"}},
		},
	}
}
