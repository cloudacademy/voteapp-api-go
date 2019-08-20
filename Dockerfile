FROM alpine:latest

MAINTAINER Jeremy Cook <jeremy.cook@cloudacademy.com>

COPY api .

EXPOSE 8080

CMD ["./api"]
