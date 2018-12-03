FROM enmasseproject/java-base:8-10

ARG version
ARG commit
ENV VERSION=${version} COMMIT=${commit}

ADD target/tenant-service-${VERSION}.jar /tenant-service.jar

CMD ["/opt/run-java/launch_java.sh", "/tenant-service.jar"]
