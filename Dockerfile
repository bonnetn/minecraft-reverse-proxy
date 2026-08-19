FROM golang:1.26@sha256:45a5f7a810238aabcbad211d70b9ae082022d96f7c7259e94041ad1b933575ac as build

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
