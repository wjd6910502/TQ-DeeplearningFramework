FROM dev.oa.com/tqlab/tq-base:1.0

USER root

RUN mkdir -p /app/server
RUN mkdir -p /app/server/log

COPY . /app/server

WORKDIR /app/server

#RUN cd ..

ENV LD_LIBRARY_PATH /usr/local/lib
RUN apt install vim -y
RUN ./build_linux_amd64.sh


