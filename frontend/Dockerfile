FROM node:24-bookworm-slim AS dependencies
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci

FROM dependencies AS development
COPY frontend/ ./
EXPOSE 3000
CMD ["npm", "run", "dev"]

FROM dependencies AS build
COPY frontend/ ./
RUN npm run build

FROM node:24-bookworm-slim AS production
ENV NODE_ENV=production PORT=3000
WORKDIR /app
COPY --from=build --chown=node:node /app/dist ./dist
COPY --from=build --chown=node:node /app/node_modules ./node_modules
COPY --from=build --chown=node:node /app/package.json /app/server.mjs ./
USER node
EXPOSE 3000
CMD ["node", "server.mjs"]
