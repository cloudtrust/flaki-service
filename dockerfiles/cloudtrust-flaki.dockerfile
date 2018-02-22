FROM cloudtrust-baseimage:f27

ARG flaki_service_git_tag
ARG flaki_service_url

WORKDIR /cloudtrust
RUN git clone git@github.com:cloudtrust/flaki-service.git

# Get Flaki service
RUN wget flaki_service_url && \

WORKDIR /cloudtrust/flaki-service
RUN git checkout ${flaki_service_git_tag} && \
    groupadd flaki && \
    useradd -m -s /sbin/nologin -g flaki flaki && \
    install -v -m0644 deploy/etc/security/limits.d/* /etc/security/limits.d/ && \
# monit
    install -v -m0644 deploy/etc/monit.d/* /etc/monit.d/ && \    
# flaki
    install -d -v -m755 /etc/flaki/ -o flaki -g flaki && \
    install -v -m0755 deploy/etc/flaki/* /etc/flaki/ && \
    install -v -m0755 /cloudtrust/flaki/flakiService /etc/flaki/ && \
    chown flaki:flaki /etc/flaki/flakiService && \
    install -v -o flaki -g flaki -m 644 deploy/etc/systemd/system/flaki.service /etc/systemd/system/flaki.service && \
    install -v -o root -g root -m 644 -d /etc/systemd/system/flaki.service.d && \
    install -v -o root -g root -m 644 deploy/etc/systemd/system/flaki.service.d/limit.conf /etc/systemd/system/flaki.service.d/limit.conf && \
# enable services
    systemctl enable flaki.service && \
    systemctl enable jaegeragent.service && \
    systemctl enable monit.service

EXPOSE 80
