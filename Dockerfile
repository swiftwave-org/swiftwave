# build stage
FROM golang:1.21-rc-bookworm AS build-env
ENV CGO_ENABLED=1
ADD . /src
RUN cd /src && go build -o goapp .

# final stage
FROM ubuntu:22.04
RUN mkdir /app
RUN mkdir /data
WORKDIR /app
COPY --from=build-env /src/goapp /app/goapp
RUN apt-get update && apt-get install -y ca-certificates
ENTRYPOINT /app/goapp