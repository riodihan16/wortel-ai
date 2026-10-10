//go:build !windows

package main

import "fmt"

// fatalDialog is a no-op on non-Windows; the console prints the error there.
func fatalDialog(err error) {
	fmt.Println("Error:", err.Error())
}
