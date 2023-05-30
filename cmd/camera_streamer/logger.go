package main

import (
	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
)

func setupLogger() *zap.SugaredLogger {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(colorable.NewColorableStdout()), zapcore.DebugLevel)
	baseLogger := zap.New(core)
	return baseLogger.Sugar().Named("main")
}

// LogRequests returns a middleware that logs all requests to this router
func LogRequests(logger *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Infof("%s %s", r.Method, r.RequestURI)

			next.ServeHTTP(w, r)
		})
	}
}
