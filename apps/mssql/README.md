# Dockerized mssql (Microsoft SQL Server)

How to run Microsoft SQL Server within docker?

```shell
dockerized mssql
```

- The default password is `Dockerized1`
- Override the password with the environment variable `SA_PASSWORD`
  - `SA_PASSWORD=<password> dockerized mssql`
  - Or change it in your `dockerized.env` file
