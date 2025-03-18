FROM golang:1.24@sha256:762bb9cb6d35eb03537551112a3519cf0e6bfc66891530ce7dc7d6169ea1eeb3 as build

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal/ ./internal/
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12@sha256:95ea148e8e9edd11cc7f639dc11825f38af86a14e5c7361753c741ceadef2167

COPY --from=build /go/bin/app /
CMD ["/app"]
