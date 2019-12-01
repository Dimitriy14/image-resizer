package logger

func NewMokLogger() Logger {
	return loggerMock{}
}

type loggerMock struct {
}

func (loggerMock) Debugf(format string, v ...interface{}) {}
func (loggerMock) Infof(format string, v ...interface{})  {}
func (loggerMock) Warnf(format string, v ...interface{})  {}
func (loggerMock) Errorf(format string, v ...interface{}) {}
