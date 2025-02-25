FROM golang:1.24@sha256:5255fad61a7e8880e742ee3e30ac54d3fdc48ea5236d0bcf14bfedb6643cbeae as build

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal/ ./internal/
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12@sha256:3f2b64ef97bd285e36132c684e6b2ae8f2723293d09aae046196cca64251acac

COPY --from=build /go/bin/app /
CMD ["/app"]
