FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends git curl ca-certificates wget openssh-client \
      python3 python3-pip python3-venv sudo && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p -m 755 /etc/apt/keyrings && \
    wget -qO- https://cli.github.com/packages/githubcli-archive-keyring.gpg \
      > /etc/apt/keyrings/githubcli-archive-keyring.gpg && \
    chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
      > /etc/apt/sources.list.d/github-cli.list && \
    apt-get update && \
    apt-get install -y --no-install-recommends gh && \
    rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && \
    apt-get install -y --no-install-recommends nodejs && \
    rm -rf /var/lib/apt/lists/* && \
    corepack enable && \
    corepack prepare pnpm@latest --activate

RUN useradd -m -s /bin/bash claude && \
    echo 'claude ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

ENV PATH="/home/claude/.local/bin:${PATH}"

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
# Create base directory for host path symlinks (owned by claude)
RUN mkdir -p /Users && chown claude:claude /Users

WORKDIR /workspace
# Note: WORKDIR is overridden at runtime via --workdir to match host path

ENTRYPOINT ["/entrypoint.sh"]
CMD ["claude"]
