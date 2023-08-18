# -- build stager --
FROM python:3.10-alpine3.18 AS build

# Args
ARG DEPENDENCY_FILE=requirements.txt
ARG PEX_WRAPPER="pex_wrapper"

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# -- deps upgrades and installation --
RUN apk add -y --update --no-cache gcc python3 python3-dev py3-pip musl-dev linux-headers
RUN python3 -m ensurepip --upgrade && python3 -m pip install pex~=2.1.47

# -- build pex from deps --
RUN mkdir /source
COPY ${DEPENDENCY_FILE} /source/
RUN pex -r /source/${DEPENDENCY_FILE} -o /source/${PEX_WRAPPER}

# -- release stager --
FROM python:3.10-alpine3.18 AS final
RUN apk upgrade --no-cache

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

# -- args
ARG PORT=80
ARG START_COMMAND="python app.py"
ARG PEX_WRAPPER="pex_wrapper"

# -- copy from build stage --
WORKDIR /app
COPY . /app
COPY --from=build /source /app/

# -- app setup --
ENV PORT=${PORT}
EXPOSE ${PORT}

# Set alias
RUN echo "alias python=/app/pex_wrapper" > /app/entrypoint.sh
# Setup entrypoint
RUN echo "${START_COMMAND}" >> /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Run app
CMD ["sh", "-c", "/app/entrypoint.sh"]