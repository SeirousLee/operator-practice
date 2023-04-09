FROM golang:1.19 as builder

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -o kernel-params main.go

FROM alpine:3.15.3

WORKDIR /app

COPY --from=builder /app/kernel-params .

CMD ["./kernel-params"]