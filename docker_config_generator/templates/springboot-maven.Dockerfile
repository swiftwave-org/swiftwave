FROM eclipse-temurin:17-jdk-jammy AS builder

ARG BUILD_COMMAND="./mvnw clean install"

WORKDIR /opt/app
COPY .mvn/ .mvn
COPY mvnw pom.xml ./

RUN ./mvnw dependency:go-offline

COPY ./src ./src

RUN ./mvnw clean install
 
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
