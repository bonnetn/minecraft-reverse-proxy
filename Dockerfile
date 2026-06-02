FROM golang:1.26@sha256:98fc714bfe32e7d3c539d63bda9b9cd089fd699dc3cbd1c534fec3c4deb9ca98 as build

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
