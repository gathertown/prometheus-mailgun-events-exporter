FROM golang:1.17.8-alpine3.15 as builder
RUN apk update && apk add ca-certificates curl git make tzdata
RUN adduser -u 5003 --gecos '' --disabled-password --no-create-home gather
COPY . /app
WORKDIR /app
RUN make buildlinux

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bin/mailgun_events_exporter-linux /bin/mailgun_events_exporter
COPY --from=builder /etc/passwd /etc/passwd
USER gather
CMD ["mailgun_events_exporter"]
