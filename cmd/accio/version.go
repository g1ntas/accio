package main

import (
	"fmt"
	"runtime"
)

var (
	buildTag    = "unknown"
	buildDate   = "unknown"
	buildCommit = "unknown"
)

func init() {
	rootCmd.Version = buildTag
	rootCmd.SetVersionTemplate(fmt.Sprintf(`Accio version %s
Build date: %s
Commit: %s
built with: %s
`, buildTag, buildDate, buildCommit, runtime.Version()))
}
