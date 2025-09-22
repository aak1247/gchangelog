FROM alpine:latest
LABEL authors="aak1247"

RUN apk add git

ENTRYPOINT ["top", "-b"]