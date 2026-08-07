package logger

import (
	"fmt"
	"github.com/rs/zerolog/log"
)

func LogErrorWithStack(err error, message string) {
	if err == nil {
		return
	}

	l := log.Error().Err(err).Str("message", message)
	stackStr := fmt.Sprintf("%+v", err)
	l = l.Str("stack", stackStr)

	l.Send()
}
