@ECHO OFF
SET DOCKERIZED_ROOT="%~dp0%\.."
PowerShell -NoProfile -ExecutionPolicy Bypass -Command "& '%DOCKERIZED_ROOT%\bin\dockerized.ps1' %*"
