FROM golang:1.24@sha256:065cb8cf9978e8a50860c20dc4e4124636e80b5d2bd175347f57145c774910a7 as build

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal/ ./internal/
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12@sha256:3d0f463de06b7ddff27684ec3bfd0b54a425149d0f8685308b1fdf297b0265e9

COPY --from=build /go/bin/app /
CMD ["/app"]
