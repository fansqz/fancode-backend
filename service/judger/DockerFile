FROM centos:latest

RUN yum update -y && \
    yum install -y java-11-openjdk python3 "@Development Tools" && \
    yum clean all

CMD ["/bin/bash"]