FROM alpine:latest

RUN apk --no-cache add tzdata ca-certificates \
 && mkdir /data

COPY http-over-ssh /

ENV HOS_USER=root \
    HOS_TIMEOUT=10s \
    HOS_KEY_DIR=/data \
    HOS_LISTEN=:8080 \
    HOS_METRICS=1
EXPOSE 8080

WORKDIR /data
ENTRYPOINT ["/http-over-ssh"]

