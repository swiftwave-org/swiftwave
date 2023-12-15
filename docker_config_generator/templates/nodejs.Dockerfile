FROM node:18.18.0-bullseye
WORKDIR /app

# Copy source code
COPY . .

# Install dependencies
ARG SETUP_COMMAND="npm install"
RUN ${SETUP_COMMAND}

# Start the app
ARG START_COMMAND="npm run start"
RUN echo "${START_COMMAND}" > /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["sh", "-c", "/app/entrypoint.sh"]