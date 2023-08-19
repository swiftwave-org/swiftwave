FROM eclipse-temurin:17-jdk-jammy AS builder

ARG BUILD_COMMAND="./gradlew build"

WORKDIR /opt/app
COPY gradle/ gradle
COPY gradlew build.gradle 
COPY settings.gradle ./

RUN /opt/app/gradlew dependencies

COPY ./src ./src

RUN ${BUILD_COMMAND}
 
FROM eclipse-temurin:17-jre-alpine AS runner

ARG PORT=8080
ARG JAR_FILE="app.jar"
RUN adduser -D user --shell /usr/sbin/nologin \
    && chown -R user:user /opt/app

WORKDIR /opt/app

EXPOSE ${PORT}
ENV PORT=${PORT}

COPY --from=builder /opt/app/target/*.jar /opt/app/${JAR_FILE}
USER user
CMD java -jar /opt/app/${JAR_FILE}
