# InfiniteDB

A scalable database

## Environment variables

| Variable             | Description                    | Default              |
|----------------------|--------------------------------|----------------------|
| DATABASE_PATH        | Path to database files         | /var/lib/infinitedb/ |
| AUTHENTICATION       | Enables authentication         | true                 |
| PORT                 | Database listen port           | 8080                 |
| REQUEST_LOGGING      | Prints request logs to console | false                |
| CACHE_SIZE           | Size of in-memory object cache | 1000                 |
| TLS                  | Enables TLS                    | false                |
| TLS_CERT             | Path to TLS Cert               |                      |
| TLS_KEY              | Path to TLS Key                |                      |
| WEBSOCKET_READ_LIMIT |                                | 1024000000           |