//This package modifies the functionality provided by glog
//The code present on this file has been commented out on either glog.go or glog_file.go
//It has been necessary as it wasn't possible to set the logDir using flags since it was getting overwritten by other
//dependencies (possibly cobra)
package cplogs

import (
	"flag"
	"fmt"
	"os"
	"time"
)

//low timeout but we want to see the logs if something went wrong
const flushInterval = 2 * time.Second

const defaultLogDir = "./cp-remote-logs/"

var logDir string

func createLogDirs() {
	if logDir != "" {
		logDirs = append(logDirs, logDir)
	}
	logDirs = append(logDirs, os.TempDir())
}

func GetLogInfoFile() string {
	return logDir + program + ".INFO"
}

func init() {
	//required by glog.go
	flag.Parse()

	//create the log folder if is not there already
	if err := os.MkdirAll(defaultLogDir, 0775); err != nil {
		fmt.Errorf("Error creating cp remote logs directory: %s", defaultLogDir)
		return
	}

	logDir = defaultLogDir

	//overrides settings that glog usually sets by flag
	logging.toStderr = false
	logging.alsoToStderr = false

	//default stderrThreshold is ERROR.
	logging.stderrThreshold = errorLog

	//set maximum level of verbose
	logging.setVState(5, nil, false)
	go logging.flushDaemon()
}
