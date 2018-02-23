FROM cloudtrust-baseimage:f27

ARG flaki_service_git_tag
ARG flaki_service_url

WORKDIR /cloudtrust
RUN git clone git@github.com:cloudtrust/flaki-service.git
ADD ./agent-linux /cloudtrust/agent

# Get Flaki service
#RUN wget flaki_service_url
ADD ./flakid /cloudtrust/flakid

WORKDIR /cloudtrust/flaki-service
RUN git checkout ${flaki_service_git_tag} && \
    groupadd flaki && \
    useradd -m -s /sbin/nologin -g flaki flaki && \
    groupadd agent && \
    useradd -m -s /sbin/nologin -g agent agent && \
    install -v -m0644 deploy/etc/security/limits.d/* /etc/security/limits.d/ && \
# monit
    install -v -m0644 deploy/etc/monit.d/* /etc/monit.d/ && \    
# flaki
    install -d -v -m755 /etc/flaki/ -o flaki -g flaki && \
    install -v -m0755 deploy/etc/flaki/* /etc/flaki/ && \
    install -v -m0755 /cloudtrust/flakid /etc/flaki/ && \
    chown flaki:flaki /etc/flaki/flakid && \
    install -v -o flaki -g flaki -m 644 deploy/etc/systemd/system/flaki.service /etc/systemd/system/flaki.service && \
    install -v -o root -g root -m 644 -d /etc/systemd/system/flaki.service.d && \
    install -v -o root -g root -m 644 deploy/etc/systemd/system/flaki.service.d/limit.conf /etc/systemd/system/flaki.service.d/limit.conf && \
# jaeger agent
    install -d -v -m755 /etc/agent/ -o agent -g agent && \
    install -v -m0755 deploy/etc/agent/* /etc/agent/ && \
    install -v -m0755 /cloudtrust/agent /etc/agent/ && \
    chown agent:agent /etc/agent/agent && \
    install -v -o agent -g agent -m 644 deploy/etc/systemd/system/agent.service /etc/systemd/system/agent.service && \
    install -v -o root -g root -m 644 -d /etc/systemd/system/agent.service.d && \
    install -v -o root -g root -m 644 deploy/etc/systemd/system/agent.service.d/limit.conf /etc/systemd/system/agent.service.d/limit.conf && \
# enable services
    systemctl enable flaki.service && \
    systemctl enable agent.service && \
    systemctl enable monit.service

EXPOSE 5555 8888
