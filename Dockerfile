FROM openjdk:11-jre as with_java

FROM golang:1.20.7
COPY --from=with_java /usr/local/openjdk-11 /usr/local/openjdk-11
ENV JAVA_HOME=/usr/local/openjdk-11
ENV JAVA_VERSION=11.0.16
ENV PATH="$PATH:/usr/local/openjdk-11/bin"
