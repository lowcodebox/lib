package logger

//goland:noinspection GoUnusedExportedFunction
func SetupDefaultLogger(namespace string, options ...ConfigOption) {
	logger := initLogger(options...)
	defaultLogger = New(logger.Named(namespace))
}
