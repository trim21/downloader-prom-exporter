package logger

import "github.com/sirupsen/logrus"

func Setup() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceQuote:       true,
		TimestampFormat:  "01-02 15:04:05",
		DisableSorting:   false,
		ForceColors:      true,
		FullTimestamp:    true,
		PadLevelText:     true,
		QuoteEmptyFields: true,
	})
}
