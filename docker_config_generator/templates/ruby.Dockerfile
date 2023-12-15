# Install ruby dependencies
FROM ruby:3.2.2-bullseye

# Build args
ARG NODE_MAJOR
ARG INSTALL_COMMAND
ARG START_COMMAND

RUN mkdir /app
WORKDIR /app

# Install dependencies
RUN apt update -yqq && apt install -yqq ca-certificates cmake curl gnupg imagemagick shared-mime-info libvips

# Install nodejs
RUN mkdir -p /etc/apt/keyrings \ 
 && curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg \
 && echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_$NODE_MAJOR.x nodistro main" | tee /etc/apt/sources.list.d/nodesource.list \
 && apt update && apt install nodejs -yqq \
 && npm install -g yarn

# Copy the main application.
COPY . .

# Install dependencies
RUN ${INSTALL_COMMAND} 

# Clean up
RUN apt clean && rm -rf /var/lib/apt/lists/*

# Setup entrypoint
RUN echo "${START_COMMAND}" > /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Run app
CMD ["sh", "-c", "/app/entrypoint.sh"]
