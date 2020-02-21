FROM golang:1.13-alpine as builder

COPY . /home

WORKDIR /home

RUN go build -v -mod=vendor

###########
FROM alpine

RUN apk --no-cache add ca-certificates

LABEL NAME="LinkSmart Service Registrator"

WORKDIR /home
COPY --from=builder /home/service-registrator .

ENTRYPOINT ["./service-registrator"]