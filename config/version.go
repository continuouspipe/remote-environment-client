package config

import "os"

var AppName = os.Args[0]

const KubeCtlName = "kubectl"

//The current version is assigned via ldflags, see Makefile
var CurrentVersion string
