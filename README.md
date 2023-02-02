# Smart Factory camera server
***
This repository contains a camera server written in Go, that aims to convert RTSP video streams from IP cameras to a WebRTC compatible format, allowing the display of the 
camera streams in a web browser.

## System Requirements
The project is mostly written in the Go programming language. It is intended to be compiled using Go version 1.19. If
not already installed, [the official Go site](https://go.dev/doc/install) contains instructions as to how to install it.

The camera server also requires the GStreamer libraries to be available in the system. This can be installes following
the instructions available [in their official wiki](https://gstreamer.freedesktop.org/documentation/installing/index.html?gi-language=c).

## Fetch the sources
You can fetch the latest development release by directly cloning the repository:
```shell
git clone https://github.com/SmartFactory-Tec/camera_server.git
```

## Building and running the server
First, enter the repository's folder and download all of the project's dependencies:

```shell
cd camera_server
go mod download
```

Next, to run the program run the following command:

```shell
go run camera_server/cmd/camera_server
```

If you otherwise want to only build the server and get it's executable, use the following command (don't forget to give 
the executable a name):

```shell
go build -o {EXECUTABLE NAME} camera_server/cmd/camera_server
```

## Configuring the server
The server automatically generates a default configuration file in your distribution's default location. Normally,
this is `~/.config/camera_server/config.toml`. This can be overriden by setting the `SERVER_CONFIG_PATH` environment 
variable to any other folder.

## Running in docker
An image for running this server is available in the GitHub Container Registry. Pull it with the following command:

```shell
docker pull ghcr.io/smartfactory-tec/camera_server:latest
```

To configure the server inside the docker image, override the `/config` folder inside the image (with a volume or bind
mount). 

By default, the image exposes the 3000 port.
