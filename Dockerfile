FROM golang:1.26@sha256:f7159064a17ccc65d0e10e342ae8783026182704bf4af8f6df8d5ba9af2be2fd as build

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal/ ./internal/
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12@sha256:20bc6c0bc4d625a22a8fde3e55f6515709b32055ef8fb9cfbddaa06d1760f838

COPY --from=build /go/bin/app /
CMD ["/app"]
