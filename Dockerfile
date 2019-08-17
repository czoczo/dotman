############################
# STEP 1 build executable binary
############################

FROM golang:alpine AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

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

# Create and set ssh known_hosts file location
ENV SSH_KNOWN_HOSTS=/known_hosts
RUN touch $SSH_KNOWN_HOSTS

# Run the hello binary.
ENTRYPOINT ["/go/bin/dotman"]
