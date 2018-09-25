FROM golang:1.11.0-alpine3.8 as builder

WORKDIR /develop
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod vendor -o etcd-snapper-b2 -ldflags '-w'

FROM alpine:3.8

RUN apk --update --no-cache --no-progress add tini ca-certificates

ENV ESB_WAIT_FOR_CHANGES_INTERVAL=5000 \
    ESB_B2_UPLOAD_RETRY_INTERVAL=5000

COPY --from=builder /develop/etcd-snapper-b2 /usr/local/bin/etcd-snapper-b2
RUN chmod +x /usr/local/bin/etcd-snapper-b2

ENTRYPOINT ["tini", "--"]
CMD ["etcd-snapper-b2", "/snapshot.db"]
