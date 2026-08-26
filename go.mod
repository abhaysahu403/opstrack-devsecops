module github.com/example/opstrack

go 1.22

require github.com/jackc/pgx/v5 v5.9.2

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace golang.org/x/crypto => github.com/golang/crypto v0.21.0

replace golang.org/x/text => github.com/golang/text v0.14.0

replace golang.org/x/sync => github.com/golang/sync v0.6.0

replace gopkg.in/yaml.v3 => github.com/go-yaml/yaml v0.0.0-20200313102051-9f266ea9e77c

replace gopkg.in/check.v1 => github.com/go-check/check v0.0.0-20161208181325-20d25e280405
