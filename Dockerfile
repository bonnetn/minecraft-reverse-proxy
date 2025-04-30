FROM golang:1.24@sha256:30baaea08c5d1e858329c50f29fe381e9b7d7bced11a0f5f1f69a1504cdfbf5e as build

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
