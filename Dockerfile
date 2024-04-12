FROM golang:alpine AS builder

WORKDIR /usr/local/src

RUN apk add bash gcc musl-dev gettext

COPY ["./go.mod", "./go.sum", "./"]
RUN go mod download

COPY ./internal ./internal
COPY ./pkg ./pkg
COPY ./cmd ./cmd
RUN go build -o ./bin/app cmd/app/main.go

FROM alpine

COPY --from=builder /usr/local/src/bin/app ./

COPY ./templates ./templates
COPY ./configs ./configs
COPY .env .

CMD ["./app"]