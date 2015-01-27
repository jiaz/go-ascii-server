# Base image contains ffmpeg and go
# GOPATH is /gopath
# GOROOT is /goroot
FROM jiaz/golang:v3

# Copy the local package files to the container's workspace.
ADD . /gopath/src/github.com/jiaz/go-ascii-server

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get github.com/jiaz/go-ascii-server
RUN go install github.com/jiaz/go-ascii-server

# Get mov
COPY resources/demo.m4v /gopath/bin/resources/demo.m4v

# Set cwd to gopath/bin
WORKDIR /gopath/bin

# Run the command by default when the container starts.
ENTRYPOINT ["/gopath/bin/go-ascii-server"]

# Document that the service listens on port 5555.
EXPOSE 5555
