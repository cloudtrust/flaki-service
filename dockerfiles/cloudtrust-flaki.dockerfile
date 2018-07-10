FROM cloudtrust-baseimage:f27

ARG flaki_service_git_tag
ARG flaki_service_release
ARG jaeger_release
ARG config_git_tag
ARG config_repo

RUN groupadd flaki && \
    useradd -m -s /sbin/nologin -g flaki flaki && \
    install -d -v -m755 /opt/flaki -o root -g root && \
    install -d -v -m755 /etc/flaki -o flaki -g flaki && \
    groupadd agent && \
    useradd -m -s /sbin/nologin -g agent agent && \
    install -d -v -m755 /opt/agent -o root -g root && \
    install -d -v -m755 /etc/agent -o agent -g agent

WORKDIR /cloudtrust
RUN git clone git@github.com:cloudtrust/flaki-service.git && \
    git clone ${config_repo} ./config

WORKDIR /cloudtrust/flaki-service
RUN git checkout ${flaki_service_git_tag}

WORKDIR /cloudtrust/flaki-service
# Install regular stuff. Systemd, monit...
RUN install -v -m0644 deploy/etc/security/limits.d/* /etc/security/limits.d/ && \
    install -v -m0644 deploy/etc/monit.d/* /etc/monit.d/ 

##
##  FLAKI DAEMON
##  

WORKDIR /cloudtrust
RUN wget ${flaki_service_release} -O flaki.tar.gz && \
    mkdir flaki && \
    tar -xf flaki.tar.gz -C flaki --strip-components 1 && \
    install -v -m0755 flaki/flakid /opt/flaki/flakid && \
    rm flaki.tar.gz && \
    rm -rf flaki/

WORKDIR /cloudtrust/flaki-service
RUN install -v -o flaki -g flaki -m 644 deploy/etc/systemd/system/flaki.service /etc/systemd/system/flaki.service && \
    install -d -v -o root -g root -m 644 /etc/systemd/system/flaki.service.d && \
    install -v -o root -g root -m 644 deploy/etc/systemd/system/flaki.service.d/limit.conf /etc/systemd/system/flaki.service.d/limit.conf

##
##  JAEGER AGENT
##  

WORKDIR /cloudtrust
RUN wget ${jaeger_release} -O jaeger.tar.gz && \
    mkdir jaeger && \
    tar -xf jaeger.tar.gz -C jaeger --strip-components 1 && \
    install -v -m0755 jaeger/agent-linux /opt/agent/agent && \
    rm jaeger.tar.gz && \
    rm -rf jaeger/

WORKDIR /cloudtrust/flaki-service
RUN install -v -o agent -g agent -m 644 deploy/etc/systemd/system/agent.service /etc/systemd/system/agent.service && \
    install -d -v -o root -g root -m 644 /etc/systemd/system/agent.service.d && \
    install -v -o root -g root -m 644 deploy/etc/systemd/system/agent.service.d/limit.conf /etc/systemd/system/agent.service.d/limit.conf

##
##  CONFIG
##

WORKDIR /cloudtrust/config
RUN git checkout ${config_git_tag}

WORKDIR /cloudtrust/config
RUN install -v -m0755 -o agent -g agent deploy/etc/jaeger-agent/agent.yml /etc/agent/ && \
    install -v -m0755 -o flaki -g flaki deploy/etc/flaki/flakid.yml /etc/flaki/ 

# Enable services
RUN systemctl enable flaki.service && \
    systemctl enable agent.service && \
    systemctl enable monit.service
