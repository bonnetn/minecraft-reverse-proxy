FROM golang:1.24@sha256:d1db785fb37feb87d9140d78ed4fb7c75ee787360366a9c5efe39c7a841a0277 as build

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal/ ./internal/
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12@sha256:d9f9472a8f4541368192d714a995eb1a99bab1f7071fc8bde261d7eda3b667d8

COPY --from=build /go/bin/app /
CMD ["/app"]
