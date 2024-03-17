FROM alpine
ENV LANGUAGE="en"
COPY build/tg_bot .
RUN apk add --no-cache ca-certificates &&\
    chmod +x tg_bot
EXPOSE 80/tcp
CMD ["./tg_bot"]