package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/epos-eu/epos-opensource/cmd"
	"github.com/epos-eu/epos-opensource/common/configdir"
)

func init() {
	logFile, err := os.OpenFile(filepath.Join(configdir.GetPath(), "log.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalln("Failed to open log file:", err)
	}
	log.SetOutput(logFile)
}

func main() {
	cmd.Execute()
}
