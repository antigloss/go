package logger

import (
	"os"
	"os/user"
	"strings"
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
	usernames := strings.Split(username,"\\")

	conf.logPath = logpath + "\\"
	conf.pathPrefix = conf.logPath + gProgname + "." + host + "." + usernames[len(usernames)-1] + ".log."

	for i := 0; i != len(gFullSymlinks); i++ {
		gFullSymlinks[i] = conf.logPath + gSymlinks[i]
	}
}
