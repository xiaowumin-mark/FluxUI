//go:build !windows

package main

import "fmt"

func validateMessageBoxSources() error {
	return fmt.Errorf("message box source validation is only available on Windows")
}
