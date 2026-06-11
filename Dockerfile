FROM golang:1.26@sha256:d47ca13cd596f3a338c1be5f79af628f42bedcf89455266211a9ab4f95da2828 as build

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal/ ./internal/
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12@sha256:9c346e4be81b5ca7ff31a0d89eaeade58b0f95cfd3baed1f36083ddb47ca3160

COPY --from=build /go/bin/app /
CMD ["/app"]
