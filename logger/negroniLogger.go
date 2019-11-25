package logger

// NegroniLogger is implementation ALogger interface from negroni package
type NegroniLogger struct {
	Logger Logger
}

// NewNegroniLogger creates new instance of NegroniLogger
func NewNegroniLogger(logger Logger) *NegroniLogger {
	return &NegroniLogger{
		Logger: logger,
	}
}

// Println log message
func (l *NegroniLogger) Println(v ...interface{}) {
	l.Logger.Debugf("", v...)
}

// Printf log message with format
func (l *NegroniLogger) Printf(format string, v ...interface{}) {
	l.Logger.Debugf(format, v)
}
