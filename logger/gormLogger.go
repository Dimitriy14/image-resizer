package logger

// NewGormLogger creates new instance of GormLoggerImpl
func NewGormLogger(logger Logger) GormLoggerImpl {
	return GormLoggerImpl{log: logger}
}

// GormLoggerImpl is an implementation of logger from gorm package
type GormLoggerImpl struct {
	log Logger
}

// Print log message
func (l GormLoggerImpl) Print(v ...interface{}) {
	l.log.Debugf("", v)
}
