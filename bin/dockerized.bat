@ECHO ON
SETLOCAL
SET HOME=%HOMEDRIVE%%HOMEPATH%
SET PWD=%CD%
SET DOCKERIZED_ROOT="%~dp0%\.."
SET DC_FILE="%DOCKERIZED_ROOT%\docker-compose.yml"
SET ENV_FILE="%DOCKERIZED_ROOT%\.env"
SET ENV_FILE_CONFIG="%DOCKERIZED_ROOT%\config.env"
SET ENV_FILE_COMBINED="%DOCKERIZED_ROOT%\temp\temp.env"
type %ENV_FILE% %ENV_FILE_CONFIG% > "%ENV_FILE_COMBINED%"

echo %ENV_FILE_COMBINED%
REM CALL %DOCKERIZED_ROOT%/bin/dockerized-dotenv.bat %ENV_FILE%
REM CALL %DOCKERIZED_ROOT%/bin/dockerized-dotenv.bat %ENV_FILE_CONFIG%
REM @ECHO ON
REM docker-compose --env-file=%ENV_FILE_COMBINED% -f %DC_FILE% run --rm -w /host -v "%CD%:/host" %*
PowerShell -NoProfile -ExecutionPolicy Bypass -File %DOCKERIZED_ROOT%\dockerized.ps1 %*
ENDLOCAL
