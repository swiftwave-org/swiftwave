FROM golang:1.20-alpine AS builder

ARG BUILD_COMMAND="go build -o app ."
ARG NAME="app"

ENV CGO_ENABLED=0
RUN apk update && apk --no-cache upgrade
RUN apk --no-cache add ca-certificates git
WORKDIR /build
# Fetch dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN ${BUILD_COMMAND}

FROM alpine:latest AS runner

ARG NAME="app"
RUN apk --no-cache upgrade
RUN mkdir /user  \
    && adduser -D user --shell /usr/sbin/nologin \
    && chown -R user:user /user
WORKDIR /user

COPY --from=builder /build/${NAME} .
EXPOSE ${PORT}
ENV PORT ${PORT}
USER user

RUN echo "/user/${NAME}" > /user/entrypoint.sh
RUN chmod +x /user/entrypoint.sh

ENTRYPOINT ["sh", "-c", "/user/entrypoint.sh"]