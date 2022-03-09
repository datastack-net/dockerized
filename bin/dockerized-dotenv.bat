@ECHO OFF
SETLOCAL ENABLEDELAYEDEXPANSION
FOR /F "eol=# tokens=*" %%i in ('type %1') DO (
    @ECHO ON
    REM assign to variable
    set _VAR=%%i
    set _VAR=!_VAR:${=%%!
    set _VAR=!_VAR:}=%%!
    ECHO !_VAR!
    SET !_VAR!
    @ECHO OFF
)

echo %DEFAULT_BASE%
