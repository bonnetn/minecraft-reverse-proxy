FROM golang:1.24@sha256:18a1f2d1e1d3c49f27c904e9182375169615c65852ace724987929b910195b2c as build

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
