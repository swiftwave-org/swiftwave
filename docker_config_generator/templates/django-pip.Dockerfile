# -- build stager --
FROM python:3.10-alpine3.18 AS build

# Args
ARG DEPENDENCY_FILE=requirements.txt

RUN apk add --update --virtual .build-deps \
    build-base \
    postgresql-dev \
    python3-dev \
    libpq

COPY requirements.txt /app/requirements.txt
RUN pip install gunicorn
RUN pip install -r /app/requirements.txt

# -- release stager --
FROM python:3.10-alpine3.18 AS final
RUN apk upgrade --no-cache
RUN apk add libpq

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# -- args
ARG START_COMMAND="python manage.py makemigrations && python manage.py migrate && gunicorn <project_name>.wsgi:application --bind 0.0.0.0:80"

# -- copy from build stage --
WORKDIR /app
COPY . /app

COPY --from=build /usr/local/lib/python3.10/site-packages/ /usr/local/lib/python3.10/site-packages/
COPY --from=build /usr/local/bin/ /usr/local/bin/


# -- app setup --
RUN adduser -D user --shell /usr/sbin/nologin \
    && chown -R user:user /app

# Setup entrypoint
RUN echo "${START_COMMAND}" > /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
USER user

# Run app
CMD ["sh", "-c", "/app/entrypoint.sh"]