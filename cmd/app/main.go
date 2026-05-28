package main

import "os"

func main() {
	c, err := newContainer()
	if err != nil {
		os.Stderr.WriteString("bootstrap " + err.Error() + "\n")
		os.Exit(1)
	}
	os.Exit(c.app.Run())
}
