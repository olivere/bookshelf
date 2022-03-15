FROM alpine
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /
COPY frontend /
EXPOSE 8080
CMD ["/frontend"]
