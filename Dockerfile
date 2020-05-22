FROM ubuntu
COPY ./bin/paper-soccer-judge /app/judge
WORKDIR /app
CMD ["bash", "-c", "./judge"]