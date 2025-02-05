FROM golang:1.23@sha256:ff4bc68e3a64dbce1a49bdf1220688ed2db8ceaf1bcd46bd1e02f0992cc49538 as build

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
