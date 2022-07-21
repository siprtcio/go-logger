package logger

import (
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
)

type ErrorHook struct {
}

func (h *ErrorHook) Levels() []logrus.Level {
	// fire only on ErrorLevel (.Error(), .Errorf(), etc.)
	return []logrus.Level{logrus.ErrorLevel}
}

func (h *ErrorHook) Fire(e *logrus.Entry) error {
	if e == nil{
		return nil
	}
	if e.Data == nil{
		return nil
	}
	if e.Level == logrus.DebugLevel || e.Level == logrus.InfoLevel  {
		if e.Data["error"] != ""{
			delete(e.Data, "error")
		}

	}
	// e.Data is a map with all fields attached to entry
	if _, ok := e.Data["error"]; !ok {
		switch e.Data["error"].(type) {
		case error:
			errJSON := eris.ToJSON(e.Data["error"].(error), true)
			e.Data["error"] = errJSON
		default:

		}
	}

	return nil
}