# Minimal Ubuntu for Linux acceptance; Go binaries are cross-compiled on the host.
ARG BASE=ubuntu:24.04
FROM ${BASE}
RUN apt-get update \
  && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends bash dbus tmux zsh \
  && rm -rf /var/lib/apt/lists/*
RUN useradd -m tester
USER tester
