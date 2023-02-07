package main

import (
	"camera_server/pkg/gst"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"time"
)

func setupLogger() *zap.SugaredLogger {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(colorable.NewColorableStdout()), zapcore.DebugLevel)
	baseLogger := zap.New(core)
	return baseLogger.Sugar().Named("main")
}

func main() {
	logger := setupLogger()

	r := chi.NewRouter()

	r.Use(middleware.Timeout(10 * time.Second))

	logger.Debugw("Initializing GStreamer")
	gst.Init()

	logger.Infow("Loading configuration from file", "filename", "config.toml")
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

	var streams []*Stream
	streamMap := make(map[string]*Stream)

	logger.Debugw("Creating camera streams from configuration file")
	// Convert camera configuration entries to camera streams
	for _, cameraConfig := range config.Cameras {
		streamSrc, err := NewUriDecodeBinWithAuth(cameraConfig.Id, cameraConfig.Hostname, cameraConfig.Port, cameraConfig.Path, cameraConfig.User, cameraConfig.Password)
		if err != nil {
			logger.Panicw(err.Error())
		}

		stream, err := NewStream(cameraConfig.Name, cameraConfig.Id, streamSrc, logger)
		if err != nil {
			logger.Panicw(err.Error())
		}

		streams = append(streams, &stream)
		streamMap[cameraConfig.Id] = &stream
	}

	getStreams := func(w http.ResponseWriter, r *http.Request) {
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

		streamID := chi.URLParam(r, "streamID")

		stream, ok := streamMap[streamID]

		if !ok {
			logger.Errorw("stream not found", "id", streamID)
			w.WriteHeader(404)
			if _, err := w.Write([]byte("stream not found")); err != nil {
				logger.Errorw("could not send error response")
			}
		}

		json, err := json.Marshal(stream)

		if err != nil {
			logger.Errorw("error marshaling stream")
		}

		w.Write(json)
	}

	connectToStream := func(w http.ResponseWriter, r *http.Request) {
		logger := logger.Named("connectToStream")

		streamID := chi.URLParam(r, "streamID")

		stream, ok := streamMap[streamID]
		if !ok {
			logger.Errorw("unknown camera requested", "streamID", streamID)
			w.WriteHeader(404)
			return
		}
		streamer := NewCameraStreamer(stream, logger)

		streamer.Begin(w, r)
	}

	r.Route("/streams", func(r chi.Router) {
		r.Get("/", getStreams)
		r.Get("/{streamID}", getStream)
		r.Get("/{streamID}/video", connectToStream)
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
