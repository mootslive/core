# backend

```sh
migrate create -ext sql -dir db/migrations -seq create_initial_tables
sqlc generate
migrate -database "postgres://mootslive:mootslive@localhost:5432/mootslive" -path db/migrations up
```

Helpful resources:

- <https://www.sqlstyle.guide/>
