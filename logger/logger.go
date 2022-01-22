package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Setup(levelSet string, developmentLogger bool) {
	paths := []string{"stderr"}

	level := zapcore.InfoLevel
	if err := level.Set(levelSet); err != nil {
		panic(err)
	}
	config := zap.NewProductionConfig()
	// create logger config (production or development)
	if developmentLogger {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	config.ErrorOutputPaths = paths
	config.OutputPaths = paths
	config.Level.SetLevel(level)
	l, err := config.Build(
		zap.AddStacktrace(zap.ErrorLevel),
	)
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(l)
	zap.RedirectStdLog(l.Named("stdlog"))
}
