//go:build darwin || linux

package system

// resolveTrayMenu invokes user code, so callers must not hold tray locks.
func resolveTrayMenu(menu TrayMenu, provider func() TrayMenu) TrayMenu {
	if provider != nil {
		menu = provider()
	}
	return cloneTrayMenu(menu)
}

func findTrayMenuItem(menu TrayMenu, id string) (TrayMenuItem, bool) {
	for _, item := range menu {
		if item.ID == id {
			return item, true
		}
		if child, ok := findTrayMenuItem(item.Children, id); ok {
			return child, true
		}
	}
	return TrayMenuItem{}, false
}
