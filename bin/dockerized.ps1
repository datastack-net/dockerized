# CONSTANTS
$DOCKERIZED_ROOT = (get-item $PSScriptRoot).parent.FullName
$DOCKERIZED_COMPOSE_FILE = "${DOCKERIZED_ROOT}\docker-compose.yml"
$DOCKERIZED_BINARY = "${DOCKERIZED_ROOT}\build\dockerized.exe"

function Write-StdErr {
  param ([PSObject] $InputObject)
  $outFunc = if ($Host.Name -eq 'ConsoleHost') {
    [Console]::Error.WriteLine
  } else {
    $host.ui.WriteErrorLine
  }
  if ($InputObject) {
    [void] $outFunc.Invoke($InputObject.ToString())
  } else {
    [string[]] $lines = @()
    $Input | % { $lines += $_.ToString() }
    [void] $outFunc.Invoke($lines -join "`r`n")
  }
}

# region COMPILE DOCKERIZED
$DOCKERIZED_COMPILE=""
if ($args[0] -eq "--compile") {
    $DOCKERIZED_COMPILE = $true
    $_, $args = $args
}

if (($DOCKERIZED_COMPILE -eq $true) -Or !(Test-Path "$DOCKERIZED_BINARY"))
{
    Write-StdErr "Compiling dockerized..."
    docker-compose `
        -f "$DOCKERIZED_COMPOSE_FILE" `
        run --rm `
        -e "GOOS=windows" `
        _compile

    if ($LASTEXITCODE -ne 0) {
        Write-StdErr "Failed to compile dockerized."
        exit $LASTEXITCODE
    }

    if ($args.Length -eq 0)
    {
        Write-StdErr "Compiled dockerized."
        exit 0
    }
}
# endregion

# RUN DOCKERIZED:
& $DOCKERIZED_BINARY $args
