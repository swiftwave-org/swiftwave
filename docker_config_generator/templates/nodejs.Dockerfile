# Install dependencies only when needed
FROM node:lts-slim AS deps
WORKDIR /app
# Install dependencies based on the preferred package manager
COPY package.json yarn.lock* package.json* package-lock.json* pnpm-lock.yaml* ./
RUN \
  if [ -f yarn.lock ]; then yarn --frozen-lockfile; \
  elif [ -f package-lock.json ]; then npm ci; \
  elif [ -f pnpm-lock.yaml ]; then yarn global add pnpm && pnpm i --frozen-lockfile; \
  else echo "Lockfile not found. \nFallback : Running `npm install`" && npm install;  \
  fi

# Production image, copy all the files and run
FROM node:lts-slim AS runner

ARG PORT=80
ARG START_COMMAND="npm run start"

WORKDIR /app
ENV NODE_ENV production
RUN addgroup --gid 1001 nodejs
RUN adduser --disabled-password --gecos "" --uid 1001 --ingroup nodejs nodejs
RUN chown nodejs:nodejs /app
COPY . .
COPY --from=deps /app/node_modules ./node_modules

EXPOSE ${PORT}
ENV PORT ${PORT}

RUN echo "${START_COMMAND}" > /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
USER nodejs

ENTRYPOINT ["sh", "-c", "/app/entrypoint.sh"]