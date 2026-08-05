package main

import "os"

func main() {
	if len(os.Args) < 2 {
		runServe(nil)
		return
	}
	switch os.Args[1] {
	case "init":
		runInit(os.Args[2:])
	case "scan":
		runScan(os.Args[2:])
	case "serve":
		runServe(os.Args[2:])
	default:
		// `lvs --port 8080` 也视为 serve
		runServe(os.Args[1:])
	}
}
