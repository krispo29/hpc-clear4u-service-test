# Accept the Go version for the image to be set as a build argument.
ARG GO_VERSION=1.22.5

# First stage: build the executable.
FROM golang:${GO_VERSION}-alpine AS builder

# Create the user and group files that will be used in the running container to
# run the process as an unprivileged user.
RUN mkdir /user && \
    echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
    echo 'nobody:x:65534:' > /user/group
RUN mkdir /assets

COPY assets/* /assets/

# Install the Certificate-Authority certificates for the app to be able to make
# calls to HTTPS endpoints.
RUN apk add --no-cache ca-certificates

# Set the environment variables for the go command:
# * CGO_ENABLED=0 to build a statically-linked executable
# * GOFLAGS=-mod=vendor to force `go build` to look into the `/vendor` folder.
ENV CGO_ENABLED=0

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /src

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Import the code from the context.
COPY ./ ./

# Build the executable to `/app`. Mark the build as statically linked.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -installsuffix 'static' \
    -o /app .
RUN chmod a+x /app 

FROM alpine:latest as alpine
RUN apk --no-cache add tzdata zip ca-certificates
WORKDIR /usr/share/zoneinfo
# -0 means no compression.  Needed because go's
# tz loader doesn't handle compressed data.
RUN zip -r -0 /zoneinfo.zip .

# Final stage: the running container.
FROM scratch AS final

# Import the user and group files from the first stage.
COPY --from=builder /user/group /user/passwd /etc/

# Import the Certificate-Authority certificates for enabling HTTPS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Import the compiled executable from the second stage.
COPY --from=builder /app /app

COPY --from=builder /assets /assets

# the timezone data:
ENV ZONEINFO /zoneinfo.zip
COPY --from=alpine /zoneinfo.zip /

ENV TZ=Asia/Bangkok
ENV ZONEINFO=/zoneinfo.zip

# # Import the Google GCS Credential
# COPY --from=builder /src/SMOOTH-EXPRESS-d17b8b9044bb.json /gcs-credentials/

# # SET Google Credentials
# ENV GOOGLE_APPLICATION_CREDENTIALS=/gcs-credentials/SMOOTH-EXPRESS-d17b8b9044bb.json

# Declare the port on which the webserver will be exposed.
# As we're going to run the executable as an unprivileged user, we can't bind
# to ports below 1024.
EXPOSE 6200

# Perform any further action as an unprivileged user.
USER nobody:nobody

# Run the compiled binary.
ENTRYPOINT ["/app"]