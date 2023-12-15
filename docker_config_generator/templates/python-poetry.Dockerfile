FROM python:3.11.5-bullseye as python-base

# https://python-poetry.org/docs#ci-recommendations
ARG POETRY_VERSION="1.4.2"
ENV POETRY_VERSION=${POETRY_VERSION}
ENV POETRY_HOME=/opt/poetry
ENV POETRY_VENV=/opt/poetry-venv

# Tell Poetry where to place its cache and virtual environment
ENV POETRY_CACHE_DIR=/opt/.cache

# Create stage for Poetry installation
FROM python-base as poetry-base

# Creating a virtual environment just for poetry and install it with pip
RUN python3 -m venv $POETRY_VENV \
    && $POETRY_VENV/bin/pip install -U pip setuptools \
    && $POETRY_VENV/bin/pip install poetry==${POETRY_VERSION}

# Create a new stage from the base python image
FROM python-base as final

# Copy Poetry to final image
COPY --from=poetry-base ${POETRY_VENV} ${POETRY_VENV}

# Add Poetry to PATH
ENV PATH="${PATH}:${POETRY_VENV}/bin"

# Build Args
ARG SETUP_COMMAND
ARG START_COMMAND

# Setup Workdir
WORKDIR /app

# Copy Project
COPY . .

# Install Dependencies
RUN ${SETUP_COMMAND}

# Setup entrypoint
RUN echo ${START_COMMAND} >> /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Run app
CMD ["sh", "-c", "/app/entrypoint.sh"]