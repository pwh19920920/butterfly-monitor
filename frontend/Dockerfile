FROM node:22-alpine AS builder
WORKDIR /usr/src/app
COPY package.json package-lock.json ./
RUN npm ci --legacy-peer-deps
COPY ./ ./
RUN npm run build

FROM nginx:1.27-alpine
WORKDIR /usr/share/nginx/html/
RUN mkdir -p /etc/nginx/templates
COPY ./default.conf.template /etc/nginx/templates
COPY --from=builder /usr/src/app/dist /usr/share/nginx/html/
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
