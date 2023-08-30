go mod init github.com/iuliancarnaru/gobank

docker run -p 5432:5432 --name some-postgres -e POSTGRES_PASSWORD=gobank -d postgres
