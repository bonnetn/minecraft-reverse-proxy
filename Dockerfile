FROM golang:1.26@sha256:983a0823d3dab83604654972fe6bbda13142a7c57f987804fbdddb9d47dad9ec as build

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
