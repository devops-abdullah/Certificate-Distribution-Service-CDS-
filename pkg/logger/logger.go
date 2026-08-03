package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func Init(level string) {

	Log.SetOutput(os.Stdout)

	Log.SetFormatter(&logrus.JSONFormatter{})

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}

	Log.SetLevel(lvl)

	Log.Info("Logger initialized")
}