FROM migrate/migrate
COPY migrations /migrations
ENTRYPOINT ["sh", "-c", "migrate -path /migrations -database $DATABASE_URL up"]
