FROM eclipse-temurin:17-jdk-jammy AS builder

ARG BUILD_COMMAND="./gradlew clean build"

WORKDIR /opt/app
COPY gradle/ /opt/app/gradle/
COPY gradlew build.gradle ./
COPY settings.gradle ./

RUN chmod +x /opt/app/gradlew
RUN /opt/app/gradlew dependencies
RUN /opt/app/gradlew --refresh-dependencies

COPY ./src ./src

RUN ${BUILD_COMMAND}
 
FROM eclipse-temurin:17-jre-alpine AS runner

ARG PORT=8080
ENV JAR_FILE="app.jar"
WORKDIR /opt/app
RUN adduser -D user --shell /usr/sbin/nologin \
    && chown -R user:user /opt/app


EXPOSE ${PORT}
ENV PORT=${PORT}

COPY --from=builder /opt/app/build/libs/*.jar /opt/app/${JAR_FILE}
USER user
CMD java -jar /opt/app/${JAR_FILE}