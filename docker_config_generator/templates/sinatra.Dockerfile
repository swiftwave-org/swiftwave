FROM --platform=$BUILDPLATFORM ruby:latest as builder

ARG BUILDPLATFORM 
WORKDIR /usr/src/app

COPY Gemfile* ./
ENV BUNDLE_DEPLOYMENT=true 
# ENV BUNDLE_JOBS=4 
ENV BUNDLE_WITHOUT=development:test

RUN bundle install \
   && rm -rf vendor/bundle/ruby/3.1.0/cache/* 

COPY . .

FROM --platform=$BUILDPLATFORM ruby:slim as app

COPY --from=builder /usr/src/app /usr/src/app

ARG PORT=4567
ARG SERVER="thin"

ENV ENV=production
ENV BUNDLE_PATH='vendor/bundle'
ENV BUNDLE_DEPLOYMENT=true 
ENV BUNDLE_WITHOUT="development:test"
ENV PORT=${PORT}
WORKDIR /usr/src/app
RUN adduser -D user --shell /usr/sbin/nologin \
    && chown user:user /usr/src/app

RUN gem install rake

EXPOSE ${PORT}
USER user
CMD bundle exec rackup \
    --host 0.0.0.0 \
    --port ${PORT} \
    --env ${ENV} \
    --server ${SERVER}
