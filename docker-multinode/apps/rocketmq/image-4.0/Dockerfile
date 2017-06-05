FROM java:8

# Rocketmq version
ENV ROCKETMQ_VERSION 4.0.0-incubating

ENV BASE_DIR /opt
# Rocketmq home
ENV ROCKETMQ_HOME  ${BASE_DIR}/rocketmq-${ROCKETMQ_VERSION}
ENV JAVA_OPT " -Duser.home=${BASE_DIR}"


WORKDIR  ${ROCKETMQ_HOME}

RUN mkdir -p \
		${BASE_DIR}/logs \
		${BASE_DIR}/store

RUN curl https://dist.apache.org/repos/dist/release/incubator/rocketmq/${ROCKETMQ_VERSION}/rocketmq-all-${ROCKETMQ_VERSION}-bin-release.zip -o rocketmq.zip \
          && unzip rocketmq.zip \
          && mv apache-rocketmq-all/* . \
          && rmdir apache-rocketmq-all  \
          && rm rocketmq.zip


RUN chmod +x bin/mqbroker
RUN chmod +x bin/mqnamesrv

CMD sh bin/mqbroker


EXPOSE  9876 10909 10911
VOLUME /opt/logs \
		/opt/store

