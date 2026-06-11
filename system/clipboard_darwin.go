//go:build darwin

package system

func unixClipboardCandidates() []unixClipboardProvider {
	return []unixClipboardProvider{
		{
			name:  "pbcopy",
			read:  unixClipboardCommand{name: "pbpaste"},
			write: unixClipboardCommand{name: "pbcopy"},
		},
	}
}
