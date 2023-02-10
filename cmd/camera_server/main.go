package main

import (
	"camera_server/pkg/gst"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
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

// SugaredLogger inserts a sugared logger into the request context, as well as logging all requests to the chi server
func SugaredLogger(logger *zap.SugaredLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Infof("%s %s", r.Method, r.RequestURI)

			ctx := context.WithValue(r.Context(), "logger", logger.Named("request").With("method", r.Method, "uri", r.RequestURI))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func main() {
	logger := setupLogger()

	r := chi.NewRouter()

	r.Use(SugaredLogger(logger))

	logger.Debugw("initializing gstreamer")
	gst.Init()

	config, err := NewConfig()

	var allowedOrigins []string

	if config.AllowAllOrigins && !config.HTTPSOriginOnly {
		allowedOrigins = []string{"http://*", "https://*"}
	} else if config.AllowAllOrigins {
		allowedOrigins = []string{"https://"}
	} else if config.HTTPSOriginOnly {
		allowedOrigins = []string{fmt.Sprintf("https://%s", config.ClientOrigin)}
	} else {
		allowedOrigins = []string{fmt.Sprintf("https://%s", config.ClientOrigin), fmt.Sprintf("http://%s", config.ClientOrigin)}
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		ExposedHeaders: []string{"*"},
	}))

	streamStore := NewStreamStoreFromConfigs(config.StreamConfigs)

	getStreams := func(w http.ResponseWriter, r *http.Request) {
		streams := make([]*Stream, len(streamStore))

		idx := 0
		for _, v := range streamStore {
			streams[idx] = v
			idx++
		}

		list, err := json.Marshal(streams)

		if err != nil {
			logger.Errorw("Error marshaling camera names")
			return
		}

		_, err = w.Write(list)
		if err != nil {
			logger.Errorw("Error sending camera list")
			return
		}
	}

	getStream := func(w http.ResponseWriter, r *http.Request) {
		logger := logger.Named("getStream")
		ctx := r.Context()

		streamJson, err := json.Marshal(ctx.Value("stream").(*Stream))

		if err != nil {
			logger.Errorw("error marshaling stream")
		}

		_, err = w.Write(streamJson)
		if err != nil {
			panic(err)
		}
	}

	streamCtx := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			streamID := chi.URLParam(r, "streamID")

			stream, ok := streamStore[streamID]

			if !ok {
				logger.Errorw("unknown camera requested", "streamID", streamID)
				w.WriteHeader(404)
				return
			}

			ctx := context.WithValue(r.Context(), "stream", stream)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	r.Route("/streams", func(r chi.Router) {
		r.Get("/", getStreams)
		r.Route("/{streamID}", func(r chi.Router) {
			r.Use(streamCtx)
			r.Get("/", getStream)
			r.Get("/socket", WebRTCHandler(config.ClientOrigin, config.HTTPSOriginOnly, config.AllowAllOrigins))
		})
	})

	logger.Infow("starting web server", "port", config.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.Port), r)
	logger.Infow("server stopped")

	if !errors.Is(err, http.ErrServerClosed) {
		logger.Panicw("Fatal error", "err", err.Error())
	}

}

func panicIfError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
