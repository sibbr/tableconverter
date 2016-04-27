# create image:
# docker build -t tableconverter-docker .
# run image
# docker run --rm --name tableconverter-image --publish 8080:8080 tableconverter-docker
FROM golang:onbuild
EXPOSE 8080
