# build stage
FROM golang:alpine AS build-env
RUN apk --no-cache add build-base git gcc
ADD . /src
RUN cd /src && go build -o goapp .

# final stage
FROM ubuntu:22.04
WORKDIR /app
COPY --from=build-env /src/goapp /app/
RUN mkdir /data
ENTRYPOINT ./goapp