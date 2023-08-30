go mod init github.com/iuliancarnaru/gobank

docker run -p 5432:5432 --name some-postgres -e POSTGRES_PASSWORD=gobank -d postgres

INFO: always soft delete in production (use a flag field ex. status: "active" / "deleted") and filter when query
