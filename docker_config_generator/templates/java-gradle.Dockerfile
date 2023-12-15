FROM eclipse-temurin:19-jdk-jammy AS builder

# Build args
ARG DOWNLOAD_DEPENDENCY_COMMAND
ARG BUILD_COMMAND

# Setup Workdir
WORKDIR /opt/app

# Copy source code
COPY . .

# Update file permissions
RUN chmod +x ./gradlew

# Run gradle dependencies
RUN ${DOWNLOAD_DEPENDENCY_COMMAND}

# Build app
RUN ${BUILD_COMMAND}

## Runner
FROM eclipse-temurin:19-jre-alpine AS runner

# Build args
ARG OUTPUT_JAR_FILE
ARG START_COMMAND

# Setup Workdir
WORKDIR /opt/app

# Copy jar file
COPY --from=builder /opt/app/build/libs/*.jar ./${OUTPUT_JAR_FILE}

# Create entrypoint
RUN echo ${START_COMMAND} > /home/entrypoint.sh
RUN chmod +x /home/entrypoint.sh

# Setup Entrypoint
ENTRYPOINT ["sh", "-c", "/home/entrypoint.sh"]