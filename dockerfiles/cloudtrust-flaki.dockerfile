FROM cloudtrust-baseimage:f27

ARG flaki_service_git_tag
ARG flaki_service_release
ARG jaeger_release
ARG config_env
ARG config_git_tag
ARG config_repo

# Get dependencies and put jaeger collector where we expect it to be
RUN dnf -y install wget && \
    dnf clean all

RUN groupadd flaki && \
    useradd -m -s /sbin/nologin -g flaki flaki && \
    install -d -v -m755 /etc/flaki/ -o flaki -g flaki && \
    groupadd agent && \
    useradd -m -s /sbin/nologin -g agent agent && \
    install -d -v -m755 /etc/agent/ -o agent -g agent

# Get jaeger agent
WORKDIR /cloudtrust
RUN wget ${jaeger_release} && \
    tar -xzf v1.2.0.tar.gz && \
    mv -v v1.2.0 jaeger && \
    install -v -m0755 jaeger/agent-linux /etc/agent/agent && \
    rm v1.2.0.tar.gz && \
    rm -rf jaeger/

# Get flaki-service
WORKDIR /cloudtrust
RUN wget ${flaki_service_release} && \
    tar -xzf v1.0.tar.gz && \
    mv -v v1.0 flaki-service && \
    install -v -m0755 flaki-service/flakid /etc/flaki/flakid && \
    rm v1.0.tar.gz && \
    rm -rf flaki-service/

WORKDIR /cloudtrust
RUN git clone git@github.com:cloudtrust/flaki-service.git && \
    git clone ${config_repo} ./config

WORKDIR /cloudtrust/flaki-service
RUN git checkout ${flaki_service_git_tag} && \
    install -v -m0644 deploy/etc/security/limits.d/* /etc/security/limits.d/ && \
# monit
    install -v -m0644 deploy/etc/monit.d/* /etc/monit.d/ && \
# jaeger-agent
    install -v -o agent -g agent -m 644 deploy/etc/systemd/system/agent.service /etc/systemd/system/agent.service && \
    install -v -o root -g root -m 644 -d /etc/systemd/system/agent.service.d && \
    install -v -o root -g root -m 644 deploy/etc/systemd/system/agent.service.d/limit.conf /etc/systemd/system/agent.service.d/limit.conf && \
# flaki-service
    install -v -o flaki -g flaki -m 644 deploy/etc/systemd/system/flaki.service /etc/systemd/system/flaki.service && \
    install -v -o root -g root -m 644 -d /etc/systemd/system/flaki.service.d && \
    install -v -o root -g root -m 644 deploy/etc/systemd/system/flaki.service.d/limit.conf /etc/systemd/system/flaki.service.d/limit.conf 
    
WORKDIR /cloudtrust/config
RUN git checkout ${config_git_tag} && \
    install -v -m0755 -o agent -g agent deploy/${config_env}/etc/jaeger-agent/agent.yml /etc/agent/ && \
    install -v -m0755 -o flaki -g flaki deploy/${config_env}/etc/flaki/flakid.yml /etc/flaki/ 

# enable services
RUN systemctl enable flaki.service && \
    systemctl enable agent.service && \
    systemctl enable monit.service
