FROM golang:1.26@sha256:0f70d7d828acd8456022127f31975364e58d792999a7e92af6fc972e124bb6b0 as build

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal/ ./internal/
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12@sha256:22fd79fd75eab2372585b44517f8a094349938919dc613aafc37e4bdc9967c82

COPY --from=build /go/bin/app /
CMD ["/app"]
