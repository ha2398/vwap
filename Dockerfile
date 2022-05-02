FROM golang:1.17

# Build project.
COPY . src
WORKDIR src
RUN make install

ENTRYPOINT [ "vwap" ]

