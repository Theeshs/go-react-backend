Deployment linux
CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build -o gomovies ./cmd/api  


RUN
go run ./cmd/api 

Dump db
pg_dump --no-owner -h localhost -p 5432 -U postgres -movies > movies.sql  