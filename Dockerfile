# Specifies a parent image
FROM golang:1.19.2-bullseye
ENV SERVER_CONFIG_PATH=/config

# Creates an app directory to hold your app’s source code
WORKDIR /camera_server

# Copies everything from your root directory into /app
COPY . .

RUN apt-get update && apt-get install -y  \
    libgstreamer1.0-dev  \
    libgstreamer-plugins-base1.0-dev  \
    libgstreamer-plugins-bad1.0-dev  \
    gstreamer1.0-plugins-base  \
    gstreamer1.0-plugins-good  \
    gstreamer1.0-plugins-bad  \
    gstreamer1.0-plugins-ugly  \
    gstreamer1.0-libav  \
    gstreamer1.0-tools  \
    gstreamer1.0-x  \
    gstreamer1.0-alsa  \
    gstreamer1.0-gl  \
    gstreamer1.0-gtk3  \
    gstreamer1.0-qt5  \
    gstreamer1.0-pulseaudio

# Installs Go dependencies
RUN go mod download

# Builds your app with optional configuration
RUN go build -buildvcs=false -o ./camera_server camera_server/cmd/camera_server

# Tells Docker which network port your container listens on
EXPOSE 3000

# Specifies the executable command that runs when the container starts
ENTRYPOINT [ "/camera_server/camera_server" ]