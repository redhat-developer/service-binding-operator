# ---
FROM --platform=amd64 maven:3.8.4-openjdk-11 as builder

ARG APP_REPO=https://github.com/thombergs/code-examples
ARG APP_COMMIT=351804a083d3fced44437b912b7fd8f61d9de85a
ENV APP_DIR=/app

RUN git clone "${APP_REPO}" ${APP_DIR} && \
    git --git-dir=${APP_DIR}/.git reset --hard "${APP_COMMIT}"

COPY patch /patch

RUN for i in /patch/*.patch; do echo " -> Applying $i"; git -C ${APP_DIR} apply --verbose $i; done

WORKDIR ${APP_DIR}/spring-boot/spring-boot-elasticsearch

RUN mvn package -DskipTests -Dmaven.artifact.threads=4

# ---
FROM registry.access.redhat.com/ubi8/openjdk-11-runtime as runtime

COPY --from=builder /app/spring-boot/spring-boot-elasticsearch/target/productsearchapp-*.jar /tmp/app.jar

EXPOSE 8080

CMD ["java", "-jar", "/tmp/app.jar"]
