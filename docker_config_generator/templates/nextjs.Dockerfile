FROM node:18.11-slim AS builder
WORKDIR /app

# Copy source code
COPY . .

# Install dependencies
ARG SETUP_COMMAND="npm install"
RUN ${SETUP_COMMAND}

# Prepare script
RUN touch ./modify.js
RUN echo 'let data = require("./next.config.js");' >> ./modify.js
RUN echo 'data.output = "standalone";' >> ./modify.js
RUN echo 'require("fs").writeFileSync("./next.config.js", `module.exports = ${JSON.stringify(data, null, 4)}`);'  >> ./modify.js
RUN node ./modify.js
RUN rm ./modify.js

# Build nextjs app
ARG BUILD_COMMAND="npm run build"
RUN ${BUILD_COMMAND}

# Production image, copy all the files and run next
FROM node:18.11-slim AS runner

WORKDIR /app

RUN addgroup --gid 1001 nodejs
RUN adduser --disabled-password --gecos "" --uid 1001 --ingroup nodejs nextjs
COPY --from=builder /app/next.config.js ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./        
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 80
ENV PORT 80

CMD ["node", "server.js"]