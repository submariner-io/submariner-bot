FROM golang:alpine as builder
RUN mkdir /build
ENV GO111MODULE on
ENV GOFLAGS -mod=vendor
RUN apk add upx
ADD . /build/
WORKDIR /build
RUN go mod vendor
RUN go build -o main pkg/main/main.go
RUN time upx main

FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser

COPY --from=builder /build/main /app/
ENV SSH_KNOWN_HOSTS /dev/null

WORKDIR /app
CMD ["./main"]
EXPOSE 3000/tcp
