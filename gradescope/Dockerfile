# Fedora Dockerfile artifacts?
# LABEL maintainer=Clement Verna <cverna@fedoraproject.org>
# ENV DISTTAG=f38container FGC=f38 FBR=f38
# ADD file ... in /
# CMD ["/bin/bash"]
FROM fedora:38

RUN /bin/sh -c mkdir /var/run/sshd
RUN /bin/sh -c dnf -y

# Used for init process to fix signaling in containers
# See https://github.com/Yelp/dumb-init
COPY /gradescope-base/dumb-init_1.2.1_amd64 /usr/bin/dumb-init
# buildkit

RUN /bin/sh -c chmod +x
RUN /bin/sh -c mkdir -p

# SSH Config
COPY /gradescope-base/sshd_config /etc/ssh/sshd_config
# buildkit

# Autograder (update + harness + run)
RUN /bin/sh -c mkdir /autograder
COPY /harness/update_harness.py /autograder/update_harness.py
# buildkit
COPY /harness/update_and_run_harness.sh /autograder/update_and_run_harness.sh
# buildkit
COPY /harness/harness.py /autograder/harness.py
# buildkit

# Scoring enivronment setup
RUN /bin/sh -c mkdir /gradescope
COPY /fedora/setup.sh /gradescope/setup.sh
# buildkit
RUN /bin/sh -c chmod +x

# Start SSH
COPY /fedora/start_sshd.sh /usr/local/sbin/start_sshd.sh
# buildkit
# To kill upon SSH session exit
COPY /gradescope-base/ssh_wrapper.sh /usr/local/sbin/ssh_wrapper.sh
# buildkit

COPY /gradescope-base/motd /etc/motd
# buildkit
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8
ENTRYPOINT ["/usr/bin/dumb-init" "--"]
CMD ["/autograder/update_and_run_harness.sh"]