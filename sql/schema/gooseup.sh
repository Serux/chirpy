cd "$(dirname "$0")"
goose postgres "postgres://postgres:postgres@localhost:5432/chirpy" up
