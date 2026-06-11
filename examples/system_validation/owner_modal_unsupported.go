//go:build !windows

package main

import "fmt"

func validateOwnerModal() error {
	return fmt.Errorf("owner modal validation is only available on Windows")
}
