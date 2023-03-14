FROM openjdk:11-bullseye
ARG MAELSTROM_URL="https://github.com/jepsen-io/maelstrom/releases/download/v0.2.3/maelstrom.tar.bz2"
ARG MAELSTROM_OUT="maelstrom.tar.bz2"
WORKDIR /srv/app/
COPY . /srv/app/

RUN apt update -y \ 
	&& apt install -y graphviz gnuplot ruby-full
RUN curl -fsSLo ${MAELSTROM_OUT} ${MAELSTROM_URL} \
	&& tar -xjvf ${MAELSTROM_OUT}
CMD ["tail", "-f", "/dev/null"]