FROM golang:1.17 as builder

RUN apt-get update && apt-get install -y libasound2-dev

ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ENV GO111MODULE=on

WORKDIR /work
ADD . .
RUN GOOS=linux CGO_ENABLED=1 go build -ldflags="-w -s" -o /usr/local/bin/vanguard-store github.com/weiyongsheng/vanguard-store-notice

FROM alpine:3.6
MAINTAINER zc
LABEL maintainer="zc" \
    email="zc2638@qq.com"

ENV TZ="Asia/Shanghai"

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk update && \
    apk --no-cache add tzdata ca-certificates libc6-compat libgcc libstdc++ alsa-lib-dev

COPY --from=builder /usr/local/bin/vanguard-store /usr/local/bin/vanguard-store

WORKDIR /work
CMD ["vanguard-store"]