FROM alpine
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /
COPY backend /
EXPOSE 8080
CMD ["/backend"]
