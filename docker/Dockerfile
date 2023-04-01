FROM alpine:edge as builder

ADD https://raw.fastgit.org/njhallett/apk-fastest-mirror/main/apk-fastest-mirror.sh /
RUN sed -i 's/https/http/g' /apk-fastest-mirror.sh && sh /apk-fastest-mirror.sh -t 50 && apk add --no-cache --progress git make go
ARG GOPROXY
ADD . /build
RUN export GOPROXY=${GOPROXY} && cd /build && go mod download -x
RUN export GOPROXY=${GOPROXY} && cd /build && make
RUN export GOPROXY=${GOPROXY} && go install -ldflags '-s -w' github.com/rclone/rclone@latest && mv $(go env GOPATH)/bin/rclone /build/bin

FROM alpine:latest

ADD https://raw.fastgit.org/njhallett/apk-fastest-mirror/main/apk-fastest-mirror.sh /
RUN sed -i 's/https/http/g' /apk-fastest-mirror.sh && sh /apk-fastest-mirror.sh -t 50 && apk add --no-cache --progress fuse3
COPY --from=builder /build/bin/* /bin

ENTRYPOINT ["/bin/csi-rclone"]
