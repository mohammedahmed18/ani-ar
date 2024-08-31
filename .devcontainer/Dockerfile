FROM ubuntu:22.04

ENV GO_VERSION=1.22.4

# Install necessary dependencies
RUN apt-get update && \
    apt-get install -y \
    wget \
    fuse \
    git \
    gcc \
    fontconfig \
    unzip \
    curl \
    ripgrep

# install golang
RUN wget -P /tmp "https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz"
RUN tar -C /usr/local -xzf "/tmp/go${GO_VERSION}.linux-amd64.tar.gz"
RUN rm "/tmp/go${GO_VERSION}.linux-amd64.tar.gz"
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Install Node.js
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs

# install yarn
Run npm i -g yarn

# install redis
RUN apt-get update && apt-get install -y redis-server 

#install nvim
RUN curl -LO https://github.com/neovim/neovim/releases/download/v0.10.1/nvim-linux64.tar.gz
RUN rm -rf /opt/nvim
RUN tar -C /opt -xzf nvim-linux64.tar.gz
ENV PATH="/opt/nvim-linux64/bin:${PATH}"
Run chmod +x /opt/nvim-linux64/bin/nvim 

# copy nvim configuration
RUN git clone https://github.com/NvChad/starter ~/.config/nvim

COPY nvim-config/lspconfig.lua /root/.config/nvim/lua/configs
COPY nvim-config/null-ls.lua /root/.config/nvim/lua/configs
COPY nvim-config/init-plugins.lua /root/.config/nvim/lua/plugins/init.lua

RUN wget https://github.com/ryanoasis/nerd-fonts/releases/download/v2.1.0/FiraCode.zip -O /tmp/FiraCode.zip && \
unzip /tmp/FiraCode.zip -d /usr/share/fonts && \
fc-cache -fv && \
rm /tmp/FiraCode.zip

# Set up environment variables
ENV PATH="/usr/local/bin:$PATH"
# Set the working directory
WORKDIR /workspace

# Install necessary Go tools
RUN go install golang.org/x/tools/gopls@latest
RUN go install github.com/air-verse/air@latest
RUN go install mvdan.cc/gofumpt@latest
RUN go install github.com/incu6us/goimports-reviser/v3@latest
RUN go install github.com/segmentio/golines@latest


# install oh-my-zsh
COPY zsh-in-docker.sh /usr/bin
RUN chmod +x /usr/bin/zsh-in-docker.sh
RUN sh /usr/bin/zsh-in-docker.sh -- \
    -t https://github.com/denysdovhan/spaceship-prompt \
    -a 'SPACESHIP_PROMPT_ADD_NEWLINE="false"' \
    -a 'SPACESHIP_PROMPT_SEPARATE_LINE="false"' \
    -p git \
    -p ssh-agent \
    -p https://github.com/zsh-users/zsh-autosuggestions \
    -p https://github.com/zsh-users/zsh-completions


# install oh my posh
RUN curl -s https://ohmyposh.dev/install.sh | bash -s -- -d /usr/local/bin

RUN echo 'eval "$(oh-my-posh init zsh --config ~/.cache/oh-my-posh/themes/1_shell.omp.json)"' >> ~/.zshrc


# Copy the project into the container
COPY . /workspace
