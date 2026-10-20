FROM node:24-bookworm-slim AS dependencies
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci

FROM dependencies AS development
COPY frontend/ ./
EXPOSE 3000
CMD ["npm", "run", "dev"]

FROM dependencies AS build
ARG PUBLIC_WAYFORPAY_PAY_HOST=
ENV PUBLIC_WAYFORPAY_PAY_HOST=$PUBLIC_WAYFORPAY_PAY_HOST
COPY frontend/ ./
RUN npm run build

# The production image ships runtime dependencies only, so a compromised
# container cannot reach Playwright, MSW, Oxlint or the rest of the toolchain.
FROM dependencies AS runtime-dependencies
RUN npm ci --omit=dev

FROM node:24-bookworm-slim AS production
ENV NODE_ENV=production PORT=3000
WORKDIR /app
COPY --from=build --chown=node:node /app/dist ./dist
COPY --from=runtime-dependencies --chown=node:node /app/node_modules ./node_modules
COPY --from=build --chown=node:node /app/package.json /app/server.mjs /app/server-log.mjs ./
USER node
EXPOSE 3000
CMD ["node", "server.mjs"]
