FROM  maven:3.3.1

RUN groupadd admin && useradd -g admin -s /bin/bash -c Admin admin  && mkdir -p /home/admin && chown admin /home/admin
USER admin
RUN cd /home/admin && git clone https://github.com/alibaba/RocketMQ.git
WORKDIR /home/admin/RocketMQ
RUN mvn -Dmaven.test.skip=true clean package install assembly:assembly -U
RUN ln -s target/alibaba-rocketmq-broker/alibaba-rocketmq devenv
ADD broker.properties /etc/rocketmq/broker.properties
ADD logback_broker.xml /home/admin/RocketMQ/devenv/conf/logback_broker.xml
ADD play.sh /home/admin/RocketMQ/devenv/bin/play.sh
ADD runserver.sh /home/admin/RocketMQ/devenv/bin/runserver.sh
ADD runbroker.sh /home/admin/RocketMQ/devenv/bin/runbroker.sh
ADD mqbroker /home/admin/RocketMQ/devenv/bin/mqbroker
WORKDIR /home/admin/RocketMQ/devenv/bin
USER root
RUN ["chattr", "-i", "play.sh"]
RUN ["chmod", "+x", "mqnamesrv"]
RUN ["chmod", "+x", "mqbroker"]
RUN ["chmod", "+x", "play.sh"]
ENV JAVA_HOME /usr/lib/jvm/java-7-openjdk-amd64
CMD ["/home/admin/RocketMQ/devenv/bin/play.sh"]
EXPOSE 9876
EXPOSE 10911
