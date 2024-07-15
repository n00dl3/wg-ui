package main

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	logLevel          = kingpin.Flag("log-level", "The level of logging").Default("info").Enum("debug", "info", "warn", "error", "panic", "fatal")
	startServerCMD    = kingpin.Command("server", "Start server.").Default()
	passwdCMD         = kingpin.Command("passwd", "Generate password hash.")
	passwdCMDPassword = passwdCMD.Arg("password", "The password to hash").Required().String()
)

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.CommandLine.DefaultEnvars()
	cmd := kingpin.Parse()

	switch strings.ToLower(*logLevel) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	switch cmd {
	case "passwd":
		bytes, err := bcrypt.GenerateFromPassword([]byte(*passwdCMDPassword), 14)
		if err != nil {
			log.Fatalf("generate password error: %v", err)
		}
		log.Infof("Password Hash: %s", string(bytes))
		return
	case "server":
		runServer()
	default:
		log.Fatal("Unknown command")
	}

}
