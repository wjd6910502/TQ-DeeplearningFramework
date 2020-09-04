FROM hub.oa.com/turinglab/dmlpframework-base:1.7

RUN mkdir -p /app/dmlp-platform
COPY ./src /app/dmlp-platform
COPY ./tqservice /app/dmlp-platform
COPY ./test.sh /app/dmlp-platform
COPY ./upload_log.sh /app/dmlp-platform
COPY ./myhostfile /app/dmlp-platform
RUN chmod 0700 /app/dmlp-platform/tqservice

COPY ./start.sh /app/dmlp-platform
RUN chmod 0700 /app/dmlp-platform/start.sh
COPY ./startslave.sh /app/dmlp-platform
RUN chmod 0700 /app/dmlp-platform/startslave.sh

RUN mkdir -p /app/dmlp-platform/config
COPY ./config /app/dmlp-platform/config

RUN mkdir -p /app/dmlp-platform/log

# 添加ssh 双端口,修改系统配置文件
RUN mv /app/dmlp-platform/ssh/sshd_config /etc/ssh/sshd_config

RUN mv /app/dmlp-platform/ssh  ~/.ssh 
RUN chmod 0600 ~/.ssh/id_rsa

#RUN apt-get install net-tools
#RUN pip3 install boto3
WORKDIR /app/dmlp-platform
#RUN apt install vim -y

