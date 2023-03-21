package main

import (
	"camera_server/pkg/gst"
	"camera_server/pkg/webrtcstream"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"net/http"
)

func main() {
	logger := setupLogger()

	r := chi.NewRouter()

	r.Use(LogRequests(logger))

	logger.Debugw("initializing gstreamer")
	gst.Init()

	config, err := NewConfig()

	var allowedOrigins []string

	if !config.Cors.AllowAllOrigins {
		allowedOrigins = config.Cors.AllowedOrigins
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		ExposedHeaders: []string{"*"},
	}))
	streamStore := make(map[string]*webrtcstream.WebRTCStream)

	for _, streamConfig := range config.Streams {
		logger.Debugw("creating stream", "name", streamConfig.Name)
		stream, err := webrtcstream.New(streamConfig)
		if err != nil {
			logger.Error("error creating stream: %w", err)
		}
		streamStore[streamConfig.Id] = stream
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

	r.Route("/{streamID}", func(r chi.Router) {
		r.Use(streamCtx)
		r.Get("/", makeGetStreamHandler(logger))
	})

	logger.Infow("starting web server", "port", config.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.Port), r)
	logger.Infow("server stopped")

	if !errors.Is(err, http.ErrServerClosed) {
		logger.Panicw("Fatal error", "err", err.Error())
	}

}
