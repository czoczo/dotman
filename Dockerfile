############################
# STEP 1 build executable binary
############################

FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

ENV GO111MODULE=on

WORKDIR $GOPATH/src/dotman
COPY . .

# Fetch dependencies.

# Using go get.
RUN go get -d -v

# Build the binary.
RUN go build -o /go/bin/dotman

############################
# STEP 2 build a small image
############################
FROM alpine

# Copy our static executable.
COPY --from=builder /go/bin/dotman /go/bin/dotman

# Run the hello binary.
ENTRYPOINT ["/go/bin/dotman"]
