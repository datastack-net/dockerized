@ECHO OFF
SET DOCKERIZED_ROOT="%~dp0%\.."
REM SET HOME=%HOMEDRIVE%%HOMEPATH%
REM SET PWD=%CD%
REM SET DC_FILE="%DOCKERIZED_ROOT%\docker-compose.yml"
REM SET ENV_FILE="%DOCKERIZED_ROOT%\.env"
REM SET ENV_FILE_CONFIG="%DOCKERIZED_ROOT%\config.env"
REM SET ENV_FILE_COMBINED="%DOCKERIZED_ROOT%\temp\temp.env"
REM type %ENV_FILE% %ENV_FILE_CONFIG% > "%ENV_FILE_COMBINED%"

REM echo %ENV_FILE_COMBINED%
REM CALL %DOCKERIZED_ROOT%/bin/dockerized-dotenv.bat %ENV_FILE%
REM @ECHO ON
REM docker-compose --env-file=%ENV_FILE_COMBINED% -f %DC_FILE% run --rm -w /host -v "%CD%:/host" %*
REM CALL %DOCKERIZED_ROOT%/bin/dockerized-dotenv.bat %DOCKERIZED_ROOT%\dockerized.env
PowerShell -NoProfile -ExecutionPolicy Bypass -Command "& '%DOCKERIZED_ROOT%\bin\dockerized.ps1' %*"
