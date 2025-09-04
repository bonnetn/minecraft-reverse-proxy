FROM golang:1.25@sha256:5e856b892fff06368bd92fe3ac00c457ede6d399e7152575b57f9a14312ab713 as build

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal/ ./internal/
RUN go vet -v
RUN go test -v

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12@sha256:f2ff10a709b0fd153997059b698ada702e4870745b6077eff03a5f4850ca91b6

COPY --from=build /go/bin/app /
CMD ["/app"]
