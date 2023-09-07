FROM python:3.10 as python-base

# https://python-poetry.org/docs#ci-recommendations
ARG POETRY_VERSION=1.4.2
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

ARG PORT="80"
ARG START_COMMAND="poetry run flask --app main run --host=0.0.0.0"
# Copy Poetry to app image
COPY --from=poetry-base ${POETRY_VENV} ${POETRY_VENV}

# Add Poetry to PATH
ENV PATH="${PATH}:${POETRY_VENV}/bin"

WORKDIR /app

# Copy Dependencies
COPY poetry.lock pyproject.toml ./

# [OPTIONAL] Validate the project is properly configured
RUN poetry check

# Install Dependencies
RUN poetry install --no-interaction --no-cache --without dev

# Create user
RUN useradd -m -s /usr/sbin/nologin user && chown -R user:user /app

# Copy Application
COPY . /app

# Run Application
ENV PORT=${PORT}
EXPOSE ${PORT}

# Setup entrypoint
RUN echo ${START_COMMAND} >> /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Switch to non-root user
USER user

# Run app
CMD ["sh", "-c", "/app/entrypoint.sh"]