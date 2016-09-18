FROM golang:latest

# Copy the app
ADD . /app
WORKDIR /app

# Test it
RUN go test -v

# Build it
RUN go build -v nudger.go

# Run it
CMD ./nudger
