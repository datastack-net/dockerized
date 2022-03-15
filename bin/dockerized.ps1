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
      docker run `
        --rm `
        --entrypoint=go `
        -e "GOOS=windows" `
        -v "${DOCKERIZED_ROOT}:/src" `
        -v "${DOCKERIZED_ROOT}/build:/build" `
        -v "${DOCKERIZED_ROOT}/.cache:/go/pkg" `
        -w /src `
        "golang:1.17.8" `
        build -o /build/ lib/dockerized.go

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
