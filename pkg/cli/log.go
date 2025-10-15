package cli

import "github.com/sirupsen/logrus"

func setLogLevel(logLevel string) {
	if logLevel == "" {
		return
	}
	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"log_level": logLevel,
		}).WithError(err).Error("the log level is invalid")
	}
	logrus.SetLevel(lvl)
}
