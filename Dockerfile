FROM golang:1.25@sha256:9e56f0d0f043a68bb8c47c819e47dc29f6e8f5129b8885bed9d43f058f7f3ed6 as build

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal/ ./internal/
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12@sha256:2e114d20aa6371fd271f854aa3d6b2b7d2e70e797bb3ea44fb677afec60db22c

COPY --from=build /go/bin/app /
CMD ["/app"]
