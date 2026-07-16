FROM golang:1.26@sha256:ae5a2316d12f3e78fd99177dad452e6ad4f240af2d71d57b480c3477f250fec6 as build

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal/ ./internal/
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12@sha256:61b7ccecebc7c474a531717de80a94709d20547cdcdaf740c25876f2a8e38b44

COPY --from=build /go/bin/app /
CMD ["/app"]
