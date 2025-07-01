FROM golang:1.24@sha256:270cd5365c84dd24716c42d7f9f7ddfbc131c8687e163e6748b9c1322c518213 as build

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal/ ./internal/
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12@sha256:b7b9a6953e7bed6baaf37329331051d7bdc1b99c885f6dbeb72d75b1baad54f9

COPY --from=build /go/bin/app /
CMD ["/app"]
