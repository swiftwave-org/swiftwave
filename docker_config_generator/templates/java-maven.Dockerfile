FROM eclipse-temurin:17-jdk-jammy AS builder

# Build args
ARG BUILD_COMMAND

# Setup Workdir
WORKDIR /opt/app

# Copy source code
COPY . .

# Update file permissions
RUN chmod +x mvnw

# Install OS dependencies
RUN test -f AptFile && apt update -yqq && xargs -a AptFile apt install -yqq || true

# Run SetupCommand
RUN test -f SetupCommand && while read -r cmd; do eval "$cmd"; done < SetupCommand || true

# Download dependencies
RUN mvnw dependency:go-offline

# Build app
RUN ${BUILD_COMMAND}

## Runner
FROM eclipse-temurin:17-jre-alpine AS runner

# Build args
ARG OUTPUT_JAR_FILE
ARG START_COMMAND

# Setup Workdir
WORKDIR /opt/app

# Install OS dependencies
COPY AptFile* ./
RUN test -f AptFile && apt update -yqq && xargs -a AptFile apt install -yqq || true

# Run SetupCommand
COPY SetupCommand* ./
RUN test -f SetupCommand && while read -r cmd; do eval "$cmd"; done < SetupCommand || true

# Copy jar file
COPY --from=builder /opt/app/target/*.jar ./${OUTPUT_JAR_FILE}

# Create entrypoint
RUN echo ${START_COMMAND} > /home/entrypoint.sh
RUN chmod +x /home/entrypoint.sh

# Setup Entrypoint
ENTRYPOINT ["sh", "-c", "/home/entrypoint.sh"]