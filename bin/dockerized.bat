@ECHO OFF
SET _DOCKERIZED_PS=%~dp0%dockerized.ps1
PowerShell -NoProfile -ExecutionPolicy Bypass -Command "& '%_DOCKERIZED_PS%' %*"
