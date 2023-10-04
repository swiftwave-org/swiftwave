# Runtime stage
FROM nginx:stable-bullseye

COPY . /usr/share/nginx/html

# Modify nginx file permissions
RUN chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    chown -R nginx:nginx /etc/nginx/conf.d
RUN touch /var/run/nginx.pid && \
    chown -R nginx:nginx /var/run/nginx.pid

EXPOSE 80
ENV PORT 80
CMD ["nginx", "-g", "daemon off;"]