FROM golang:latest AS builder

WORKDIR /workspace
COPY . /workspace/

RUN make

FROM mcr.microsoft.com/cstsectools/codeql-container:latest
WORKDIR /workspace
COPY --from=builder  /workspace/data/Yi /workspace/Yi

ENTRYPOINT ["tail", "-f", "/dev/null"]


