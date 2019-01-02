FROM alpine:latest

RUN apk --no-cache add tzdata ca-certificates \
 && mkdir /data

COPY http-over-ssh /

ENV ZONEINFO    /zoneinfo.zip
ENV HOS_USER    root
ENV HOS_TIMEOUT 10s
ENV HOS_KEY_DIR /data
ENV HOS_LISTEN  :8080
ENV HOS_METRICS 1
EXPOSE 8080

WORKDIR /data
ENTRYPOINT ["/http-over-ssh"]

