# Install dependencies only when needed
FROM node:lts-slim AS deps
WORKDIR /app
# Install dependencies based on the preferred package manager
COPY package.json yarn.lock* package-lock.json* pnpm-lock.yaml* ./
RUN \
  if [ -f yarn.lock ]; then yarn --frozen-lockfile; \
  elif [ -f package-lock.json ]; then npm ci; \
  elif [ -f pnpm-lock.yaml ]; then yarn global add pnpm && pnpm i --frozen-lockfile; \
  else echo "Lockfile not found." && exit 1; \
  fi

# Production image, copy all the files and run
FROM node:lts-slim AS runner

ARG PORT=80

WORKDIR /app
ENV NODE_ENV production
RUN addgroup --gid 1001 nodejs
RUN adduser --disabled-password --gecos "" --uid 1001 --ingroup nodejs expressjs
RUN chown expressjs:nodejs /app
COPY . .
COPY --from=deps /app/node_modules ./node_modules

USER expressjs
EXPOSE ${PORT}
ENV PORT ${PORT}
CMD ["npm", "run", "start"]