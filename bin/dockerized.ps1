# CONSTANTS
$DOCKERIZED_ROOT = (get-item $PSScriptRoot).parent.FullName
$DOCKERIZED_COMPOSE_FILE = "${DOCKERIZED_ROOT}\docker-compose.yml"
$DOCKERIZED_BINARY = "${DOCKERIZED_ROOT}\build\dockerized.exe"

# region COMPILE DOCKERIZED
$DOCKERIZED_COMPILE=""
if ($args[0] -eq "--compile") {
    $DOCKERIZED_COMPILE = $true
    $_, $args = $args
}

if (($DOCKERIZED_COMPILE -eq $true) -Or !(Test-Path "$DOCKERIZED_BINARY"))
{
    Write-Error "Compiling dockerized..."
    docker-compose `
        -f "$DOCKERIZED_COMPOSE_FILE" `
        run --rm `
        -e "GOOS=windows" `
        _compile

    if ($args.Length -eq 0)
    {
        exit 0
    }
}
# endregion

# RUN DOCKERIZED:
& $DOCKERIZED_BINARY $args
