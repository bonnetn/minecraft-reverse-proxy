FROM golang:1.24@sha256:2c89c41fb9efc3807029b59af69645867cfe978d2b877d475be0d72f6c6ce6f6 as build

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
