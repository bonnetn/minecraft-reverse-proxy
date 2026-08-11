FROM golang:1.26@sha256:7caba5286b4c3613a337b709c573047d8ae62ee76106647313b61e72b99f20af as build

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal/ ./internal/
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12@sha256:6447365a6337c3732f412d1b74357b30a633831955b2bc45552b0086be907687

COPY --from=build /go/bin/app /
CMD ["/app"]
