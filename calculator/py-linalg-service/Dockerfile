ARG VARIANT=3.9.13-slim-bullseye
FROM python:${VARIANT}

COPY entrypoint.sh entrypoint-dev.sh /

WORKDIR /workspace

RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && chown -R 42420:42420 /workspace \
    && python -m pip install --upgrade pip \
    && chmod +x /entrypoint*.sh

ENTRYPOINT "/entrypoint.sh"


# FROM pytorch/pytorch

# COPY entrypoint.sh entrypoint-dev.sh requirements.txt /

# WORKDIR /workspace

# RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
#     && chown -R 42420:42420 /workspace \
#     && pip install -r /requirements.txt \
#     && chmod +x /entrypoint*.sh

# ENV HOME=/workspace

# ENTRYPOINT "/entrypoint.sh"