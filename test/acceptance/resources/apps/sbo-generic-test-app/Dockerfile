FROM alpine:3.12

RUN apk update && \
    apk add --no-cache \
    lighttpd \
    bash \
    postgresql-client \
    mysql-client \
    vim \
    jq && \
    rm -rf /var/cache/apk/*

RUN chmod -R go+rwx /run /var

RUN ls /

COPY filepermissions.sh /bin/
COPY conf/* /etc/lighthttpd/
COPY start_http /bin/
COPY env.sh /bin/
COPY postgres-ready.sh /bin/
COPY mysql-ready.sh /bin/

EXPOSE 8080

ENTRYPOINT ["/bin/start_http"]
