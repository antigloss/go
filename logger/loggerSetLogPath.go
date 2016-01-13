// +build !windows

package logger

import (
	"os"
	"os/user"
)

func (conf *config) setLogPath(logpath string) {
	host, err := os.Hostname()
	if err != nil {
		host = "Unknown"
	}

	username := "Unknown"
	curUser, err := user.Current()
	if err == nil {
		username = curUser.Username
	}

	conf.logPath = logpath + "/"
	conf.pathPrefix = conf.logPath + gProgname + "." + host + "." + username + ".log."

	for i := 0; i != len(gFullSymlinks); i++ {
		gFullSymlinks[i] = conf.logPath + gSymlinks[i]
	}
}