# dockerized postgres

You can use dockerized postgres to query your database, and manage backups, without locally installing postgres.

The commands work as usual, except:

- Use `host.docker.internal` instead of `localhost`.
- You can only access files in the current working directory (e.g. when dumping).

## psql

```shell
dockerized psql --host "host.docker.internal" --username <username>
```

## pg_dumpall

```shell
  dockerized pg_dumpall \
    --host "host.docker.internal" \
    --file "backup.sql" \
    --username root \
    --no-password \
    --quote-all-identifiers \
    --verbose
```