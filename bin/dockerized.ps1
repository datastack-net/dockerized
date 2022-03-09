docker-compose --env-file=%ENV_FILE_COMBINED% -f %DC_FILE% run --rm -w /host -v "%CD%:/host" %*
