FROM golang:latest AS builder

WORKDIR /workspace
COPY . /workspace/

RUN make

FROM mcr.microsoft.com/cstsectools/codeql-container:latest

USER root
RUN usermod -u 1001 codeql
RUN mkdir /usr/local/codeql-home/codeql-repo/python/ql/src/Security/MyQL/
RUN chown codeql:codeql /usr/local/codeql-home/codeql-repo/python/ql/src/Security/MyQL/

WORKDIR /workspace
COPY --from=builder /workspace/data/Yi /workspace/Yi
RUN chown -R codeql:codeql /workspace

USER codeql

ENTRYPOINT ["tail", "-f", "/dev/null"]


