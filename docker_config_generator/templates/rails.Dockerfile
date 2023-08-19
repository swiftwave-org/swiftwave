# Install node dependencies
FROM node:18.11-slim AS node_deps
WORKDIR /app

# Install dependencies based on the preferred package manager
COPY package.json yarn.lock* package-lock.json* pnpm-lock.yaml* ./
RUN \
  if [ -f yarn.lock ]; then yarn --frozen-lockfile; \
  elif [ -f package-lock.json ]; then npm ci; \
  elif [ -f pnpm-lock.yaml ]; then yarn global add pnpm && pnpm i --frozen-lockfile; \
  else echo "Lockfile not found." && mkdir node_modules; \
  fi

# Install ruby dependencies
FROM ruby:3.2.1

RUN mkdir /app
WORKDIR /app

# install dependencies
RUN apt-get update -qq && apt-get install -y imagemagick shared-mime-info libvips && apt-get clean

RUN curl -sL https://deb.nodesource.com/setup_14.x | bash \
 && apt-get update && apt-get install -y nodejs && rm -rf /var/lib/apt/lists/* \
 && curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add - \
 && echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list \
 && apt-get update && apt-get install -y yarn && rm -rf /var/lib/apt/lists/* \
 && apt-get update && apt-get -y install cmake && rm -rf /var/lib/apt/lists/*

# COPY Gemfile
COPY Gemfile /app/Gemfile
COPY Gemfile.lock /app/Gemfile.lock

# Install bundler
RUN gem install bundler
RUN bundle install  --without production

# COPY node dependencies
COPY --from=node_deps /app/node_modules ./node_modules

# Copy the main application.
COPY . .

ARG RAILS_ENV=production
ARG RACK_ENV=production
ARG START_COMMAND="bundle exec rails server -p 3000 -b '0.0.0.0'"

# Setup entrypoint
RUN echo "${START_COMMAND}" > /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Create non root user
RUN adduser --disabled-password --gecos '' user
RUN chown -R user:user /app

# Switch to non-root user
USER user

# Run app
CMD ["sh", "-c", "/app/entrypoint.sh"]