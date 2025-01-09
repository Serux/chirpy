#cd "$(dirname "$0")"
cd ../sql/schema
goose postgres "postgres://postgres:postgres@localhost:5432/chirpy" up
