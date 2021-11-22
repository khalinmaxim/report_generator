# This file is a template, and might need editing before it works on your project.
FROM golang:alpine AS builder

WORKDIR /usr/src/app

COPY . .
RUN go build -v

FROM alpine

LABEL maintainer="Vladimir Yavorskiy <vovka@krevedko.su>"
# We'll likely need to add SSL root certificates
RUN apk --no-cache add ca-certificates
# Для подстановки переменных окружения
RUN apk add --no-cache gettext bash postgresql-client

WORKDIR /app 
COPY --from=builder /usr/src/app/reporter .
COPY docker/reporter.cfg.template .


WORKDIR /usr/local/sbin
COPY docker/entrypoint.sh .
COPY sbin/*.sh .
RUN chmod 777 /app
ENTRYPOINT ["entrypoint.sh"]

