FROM node:26-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json .npmrc ./
RUN npm ci
COPY . ./
COPY builder.sh /usr/local/bin/webapp-builder
RUN chmod +x /usr/local/bin/webapp-builder && mkdir /site && chown node /site
RUN chown -R node /app
USER node
ENV API_URL=http://api:8080 \
    SITE_DIR=/site \
    WEBAPP_INTERVAL=900 \
    WEBAPP_RETRY=60
VOLUME /site
ENTRYPOINT ["webapp-builder"]

FROM caddy:2-alpine AS server
COPY Caddyfile /etc/caddy/Caddyfile
EXPOSE 80
