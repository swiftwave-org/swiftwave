FROM docker:dind

# Switch to root user
USER root

# Install packages
RUN apk update --no-cache && apk add --no-cache gcc g++ make sqlite iproute2 net-tools bind-tools bash apache2-utils sudo shadow redis

ARG NEWUSER=user
# Create non-root user
RUN addgroup --gid 1000 ${NEWUSER}
RUN adduser -D -u 1000 -G ${NEWUSER} ${NEWUSER}
RUN echo "${NEWUSER}:${NEWUSER}" | chpasswd
RUN echo "${NEWUSER} ALL=(ALL) ALL" > /etc/sudoers.d/${NEWUSER} && chmod 0440 /etc/sudoers.d/${NEWUSER}
RUN mkdir -p /home/${NEWUSER}
RUN chown -R ${NEWUSER}:${NEWUSER} /home/${NEWUSER}

# Create a docker group
RUN addgroup docker
RUN addgroup ${NEWUSER} docker

# Download and install Go
RUN wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
RUN rm go1.21.0.linux-amd64.tar.gz
RUN echo "export PATH=\$PATH:/usr/local/go/bin" >> /home/${NEWUSER}/.bashrc


# Set go cache directory
RUN mkdir -p /home/${NEWUSER}/go/pkg/mod/
RUN chown -R ${NEWUSER}:${NEWUSER} /home/${NEWUSER}/go/pkg/mod/

# Download and install ttyd
RUN wget https://github.com/tsl0922/ttyd/releases/download/1.7.3/ttyd.x86_64 -O /usr/local/bin/ttyd
RUN chmod +x /usr/local/bin/ttyd
RUN chown ${NEWUSER}:${NEWUSER} /usr/local/bin/ttyd

# Switch to non-root user
USER root