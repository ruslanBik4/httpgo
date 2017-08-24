FROM ubuntu:14.04
CMD go golang
RUN echo 'Hi, I am in your container' \
        >/usr/share/nginx/html/index.html
EXPOSE 80