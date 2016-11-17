FROM  maven:3.3.1

RUN groupadd admin && useradd -g admin -s /bin/bash -c Admin admin  && mkdir -p /home/admin && chown admin /home/admin
USER admin
RUN cd /home/admin && git clone https://github.com/alibaba/RocketMQ.git
WORKDIR /home/admin/RocketMQ
RUN mvn -Dmaven.test.skip=true clean package install assembly:assembly -U
RUN ln -s target/alibaba-rocketmq-broker/alibaba-rocketmq devenv
ADD broker.properties /etc/rocketmq/broker.properties
ADD play.sh /home/admin/RocketMQ/devenv/bin/play.sh

RUN chmod 777 /home/admin/RocketMQ/devenv/bin/play.sh
RUN chmod 777 /home/admin/RocketMQ/devenv/bin/mqnamesrv
RUN chmod 777 /home/admin/RocketMQ/devenv/bin/mqbroker

WORKDIR /home/admin/RocketMQ/devenv/bin
ENTRYPOINT ["/home/admin/RocketMQ/devenv/bin/play.sh"]
EXPOSE 9876
EXPOSE 10911