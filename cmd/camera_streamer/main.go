package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SmartFactory-Tec/camera_streamer/pkg/gst"
	"github.com/SmartFactory-Tec/camera_streamer/pkg/webrtcstream"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"net/http"
	"strconv"
)

func main() {
	logger := setupLogger()

	r := chi.NewRouter()

	r.Use(LogRequests(logger))

	logger.Debugw("initializing gstreamer")
	gst.Init()

	config := loadConfig(logger)

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

	streamStore := make(map[int]*webrtcstream.WebRTCStream)
	cameraServiceUrl := fmt.Sprintf("%s:%d/", config.CameraService.Hostname, config.CameraService.Port)

	streamCtx := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			streamIdString := chi.URLParam(r, "streamID")

			cameraId, err := strconv.ParseInt(streamIdString, 10, 32)

			if err != nil {
				logger.Errorw("invalid camera id")
				http.Error(w, "invalid camera id", 400)
				return
			}

			stream, ok := streamStore[int(cameraId)]

			if !ok {
				logger.Debugw("creating camera stream", "camera id", cameraId)
				resp, err := http.Get("http://" + cameraServiceUrl + "cameras/" + strconv.FormatInt(cameraId, 10))
				if err != nil {
					logger.Errorw("could not get camera info from camera service", "err", err)
					w.WriteHeader(500)
					return
				}
				if resp.StatusCode == 404 {
					logger.Errorw("unknown camera stream requested")
					w.WriteHeader(404)
					return
				} else if resp.StatusCode != 200 {
					logger.Errorw("unknown server error")
					w.WriteHeader(500)
					return
				}

				var streamConfig webrtcstream.Config
				bodyDecoder := json.NewDecoder(resp.Body)
				if err := bodyDecoder.Decode(&streamConfig); err != nil {
					logger.Errorw("could not parse camera service response body")
					w.WriteHeader(500)
					return
				}

				newStream, err := webrtcstream.New(streamConfig)
				if err != nil {
					logger.Error("error creating stream: %w", err)
				}
				streamStore[int(cameraId)] = newStream
				stream = newStream
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
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), r)
	logger.Infow("server stopped")

	if !errors.Is(err, http.ErrServerClosed) {
		logger.Panicw("Fatal error", "err", err.Error())
	}

}
