FROM python:3.11-buster as builder
 
RUN apt-get update && apt-get install -y git

ARG POETRY_VERSION=1.4.2

RUN pip install poetry==${POETRY_VERSION}
 
ENV POETRY_NO_INTERACTION=1 \
POETRY_VIRTUALENVS_IN_PROJECT=1 \
POETRY_VIRTUALENVS_CREATE=1 \
POETRY_CACHE_DIR=/tmp/poetry_cache
 
WORKDIR /app
 
COPY pyproject.toml poetry.lock ./
 
RUN poetry install --without dev --no-root && rm -rf $POETRY_CACHE_DIR
 
# The runtime image, used to just run the code provided its virtual environment
FROM python:3.11-slim-buster as runner

ARG PORT="80"
ARG START_COMMAND="poetry run streamlit run main.py --server.port ${PORT}"
ENV VIRTUAL_ENV=/app/.venv \
PATH="/app/.venv/bin:$PATH"
ENV PORT=${PORT}
EXPOSE ${PORT}

RUN adduser -D user --shell /usr/sbin/nologin \
    && chown -R user:user /app
WORKDIR /app

COPY --from=builder ${VIRTUAL_ENV} ${VIRTUAL_ENV}
 
COPY . .

RUN echo ${START_COMMAND} >> /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

USER user
CMD ["sh", "-c", "/app/entrypoint.sh"]