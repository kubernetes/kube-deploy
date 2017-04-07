# Jenkins JNLP slave Docker image

[`jenkinsci/jnlp-slave`](https://hub.docker.com/r/jenkinsci/jnlp-slave/)

A [Jenkins](https://jenkins-ci.org) slave using JNLP to establish connection.

See [Jenkins Distributed builds](https://wiki.jenkins-ci.org/display/JENKINS/Distributed+builds) for more info.

Make sure your ECS container agent is [updated](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-agent-update.html) before running. Older versions do not properly handle the entryPoint parameter. See the [entryPoint](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html#container_definitions) definition for more information.

## Configuration specifics

By default, JnlpProtocol3 is disabled due to the known stability and scalability issues.
You can enable this protocol on your own risk using the 
<code>JNLP_PROTOCOL_OPTS=-Dorg.jenkinsci.remoting.engine.JnlpProtocol3.disabled=false</code> property.

## Running

To run a Docker container

    docker run jenkinsci/jnlp-slave -url http://jenkins-server:port <secret> <slave name>

optional environment variables:

* `JENKINS_URL`: url for the Jenkins server, can be used as a replacement to `-url` option, or to set alternate jenkins URL
* `JENKINS_TUNNEL`: (`HOST:PORT`) connect to this slave host and port instead of Jenkins server, assuming this one do route TCP traffic to Jenkins master. Useful when when Jenkins runs behind a load balancer, reverse proxy, etc.

