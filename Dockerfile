FROM golang:1.25@sha256:10a15b9d650c559eff6cb070f3177f1e2fc067cd7412e5ca97c9cb8167a924b7 as build

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
