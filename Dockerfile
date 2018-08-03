FROM golang:1.11beta2-alpine
RUN apk add --update git gcc musl-dev
WORKDIR /usr/s3-exporter

ADD go.mod go.sum ./
RUN go mod -sync
ADD . .
RUN CGO_ENABLED=0 go install ./...

FROM busybox
COPY --from=0 /go/bin/* /usr/local/bin/
ENTRYPOINT [ "s3-exporter" ]
EXPOSE 8080
USER nobody
