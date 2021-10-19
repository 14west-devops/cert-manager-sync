FROM golang:1.15 AS builder

WORKDIR /app

# Let's cache modules retrieval - those don't change so often
COPY go.mod go.sum .
RUN go mod download

COPY . .

RUN go build -o certsync *.go

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest AS runtime

LABEL maintainer="devops@14west.us"

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
ENV USERID=1001

# Copy dist folder from builder image for runtime
COPY --chown=$USERID:0 --from=builder /app/certsync $GOPATH/certsync

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 0755 "$GOPATH"

WORKDIR $GOPATH

USER $USERID

ENTRYPOINT [ "/go/certsync" ]