FROM nginx:alpine
RUN addgroup -g 1001 nginx && adduser -D -u 1001 -G nginx nginx
RUN mkdir -p /var/cache/nginx/api /var/cache/nginx/static
RUN chown -R nginx:nginx /var/cache/nginx
COPY nginx.conf /etc/nginx/nginx.conf
USER nginx
EXPOSE 80 443
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost/health || exit 1
CMD ["nginx", "-g", "daemon off;"]
