# Multi-stage build for Svelte 5 + TypeScript + TailwindCSS frontend
FROM node:20-alpine AS build

WORKDIR /app

# Install dependencies inside Docker (never on host)
COPY package.json ./
RUN npm install --legacy-peer-deps

# Build static assets
COPY . .
RUN npm run build

# Production Nginx runtime
FROM nginx:alpine

COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
