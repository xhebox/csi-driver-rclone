FROM golang:alpine as builder

ADD https://raw.fastgit.org/njhallett/apk-fastest-mirror/main/apk-fastest-mirror.sh /
RUN sed -i 's/https/http/g' /apk-fastest-mirror.sh && sh /apk-fastest-mirror.sh -t 50 && apk add --no-cache --progress git make
ARG GOPROXY
ADD . /build
RUN export GOPROXY=${GOPROXY} && cd /build && go mod download -x
RUN export GOPROXY=${GOPROXY} && cd /build && make

FROM alpine:latest

ADD https://raw.fastgit.org/njhallett/apk-fastest-mirror/main/apk-fastest-mirror.sh /
RUN sed -i 's/https/http/g' /apk-fastest-mirror.sh && sh /apk-fastest-mirror.sh -t 50 && apk add --no-cache --progress fuse rclone
COPY --from=builder /build/bin/* /bin

ENTRYPOINT ["/bin/csi-rclone"]
