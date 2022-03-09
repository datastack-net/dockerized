$DOCKERIZED_ENV_FILE_NAME = "dockerized.env"
$DOCKERIZED_ROOT = (get-item $PSScriptRoot).parent.FullName
$DOCKERIZED_COMPOSE_FILE = "${DOCKERIZED_ROOT}\docker-compose.yml"
$DOCKERIZED_ENV_FILE = "${DOCKERIZED_ROOT}\${DOCKERIZED_ENV_FILE_NAME}"

$HOST_HOME = "${HOME}"
$HOST_PWD = "${PWD}"
$SERVICE_ARGS = ($args | % { $_.replace('\', '/') })

function DotEnv
{
    param(
        [Parameter(Mandatory=$true)]
        $file
    )
    $lines = (Get-Content $file).Split("\n")
    foreach ($line in $lines)
    {
        if (! $line.StartsWith("#"))
        {
            $parts = $line.Split("=")
            if ($parts.Length -eq 2)
            {
                $key = $parts[0].Trim()
                $value = $parts[1].Trim().Replace("\r", "")
                Set-Item "env:$key" $value
            }
        }
    }
}

# Find file in parent directories, starting with the given directory
function FindUp
{
    param(
        [Parameter(Mandatory=$true)]
        $fileName,
        [Parameter(Mandatory=$true)]
        $startDirectory
    )
    $path = $startDirectory
    while ($path -ne "")
    {
        if (Test-Path -PathType Leaf $path -Path $fileName)
        {
            return $path
        }
        $path = (Get-Item $path).ParentPath
    }
}

function LoadEnvironment
{
    $envFile = FindUp $DOCKERIZED_ENV_FILE_NAME $PWD
    $envFiles = @()
    $DIR = Get-Item -Path $PWD
    while ($DIR.FullName -ne $DIR.Root.FullName)
    {
        $DOCKERIZED_ENV = "${DIR}\dockerized.env"
        if (Test-Path "$DOCKERIZED_ENV")
        {
            DotEnv $DOCKERIZED_ENV
            return
        }
        $DIR = Get-Item -Path $DIR.Parent.FullName
    }
}

LoadEnvironment

docker-compose `
    --env-file $DOCKERIZED_ENV_FILE `
    -f $DOCKERIZED_COMPOSE_FILE `
    run --rm `
    -v "${PWD}:/host" `
    -w /host `
    ${SERVICE_ARGS}
