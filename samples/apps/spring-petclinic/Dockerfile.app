# ---
FROM --platform=amd64 maven:3.8.4-openjdk-11 as builder

ARG PETCLINIC_REPO=https://github.com/spring-projects/spring-petclinic.git
ARG PETCLINIC_COMMIT=a7439c74ea718c4f59fe6c7c643c4afe59d7e718
ENV PETCLINIC_DIR=/petclinic

RUN git clone "${PETCLINIC_REPO}" ${PETCLINIC_DIR} && \
    git --git-dir=${PETCLINIC_DIR}/.git reset --hard "${PETCLINIC_COMMIT}"

COPY patch /patch

RUN for i in /patch/*.patch; do echo " -> Applying $i"; git -C ${PETCLINIC_DIR} apply --verbose $i; done

WORKDIR ${PETCLINIC_DIR}

RUN mvn package -DskipTests -Dmaven.artifact.threads=4

# ---
FROM registry.redhat.io/ubi8/openjdk-11-runtime as runtime

COPY --from=builder /petclinic/target/spring-petclinic-*.jar /tmp/petclinic.jar

EXPOSE 8080

CMD ["java", "-Dorg.springframework.cloud.bindings.boot.enable=true", "-jar", "/tmp/petclinic.jar"]
