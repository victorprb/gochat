FROM scratch
COPY builds/gochat /bin/gochat
EXPOSE 8080
ENTRYPOINT ["/bin/gochat"]
