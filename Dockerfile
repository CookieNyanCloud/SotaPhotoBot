FROM golang:1.17.0-bullseye

RUN go version

ENV GOPATH=/

COPY ./ ./

RUN go mod download

RUN go build -o photosotabot ./cmd/main.go

CMD ["./photosotabot"]