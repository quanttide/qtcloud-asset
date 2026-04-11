# Flutter Web 应用
FROM ghcr.io/cirruslabs/flutter:3.24.5 AS builder

WORKDIR /app
COPY src/studio/ .

RUN flutter build web --release

# Nginx 托管
FROM nginx:alpine

COPY --from=builder /app/build/web /usr/share/nginx/html
COPY manifests/docker/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
