FROM python:3.11.5-bullseye

# Build Args
ARG SETUP_COMMAND
ARG START_COMMAND

# Setup Workdir
WORKDIR /app

# Copy source code
COPY . .

# Install pipenv
RUN pip install pipenv

# Install dependencies
RUN ${SETUP_COMMAND}

# Setup entrypoint
RUN echo "${START_COMMAND}" >> /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Run app
CMD ["sh", "-c", "/app/entrypoint.sh"]