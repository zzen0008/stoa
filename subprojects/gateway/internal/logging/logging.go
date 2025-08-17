package logging

import (
	"github.com/sirupsen/logrus"
	"go.elastic.co/ecslogrus"
)

func NewLogger() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&ecslogrus.Formatter{})
	return log
}
