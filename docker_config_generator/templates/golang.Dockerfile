FROM golang:1.20-alpine AS builder

# -- Args
ARG BUILD_COMMAND="go build -o app ."
ARG NAME="app"
ARG PORT="80"

# -- build env setup --
ENV CGO_ENABLED=0
RUN apk update && apk --no-cache upgrade
RUN apk --no-cache add ca-certificates git
WORKDIR /build

# -- Fetch dependencies --
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN ${BUILD_COMMAND}

# -- Runner stage --
FROM alpine:latest AS runner

# -- env setup --
ARG NAME="app"
RUN apk --no-cache upgrade
RUN mkdir /user  \
    && adduser -D user --shell /usr/sbin/nologin \
    && chown -R user:user /user
WORKDIR /user

COPY --from=builder /build/${NAME} .
EXPOSE ${PORT}
ENV PORT ${PORT}

RUN echo "/user/${NAME}" > /user/entrypoint.sh
RUN chmod +x /user/entrypoint.sh
USER user

ENTRYPOINT ["sh", "-c", "/user/entrypoint.sh"]