# CONSTANTS
$DOCKERIZED_ENV_FILE_NAME = "dockerized.env"
$DOCKERIZED_ROOT = (get-item $PSScriptRoot).parent.FullName
$DOCKERIZED_COMPOSE_FILE = "${DOCKERIZED_ROOT}\docker-compose.yml"
$DOCKERIZED_ENV_FILE = "${DOCKERIZED_ROOT}\${DOCKERIZED_ENV_FILE_NAME}"
$Env:HOME = "${HOME}"

# OPTIONS
$DOCKERIZED_OPT_VERBOSE = $false

# RUNTIME VARIABLES
$_PWD_PATH = Get-Item -Path $PWD
$HOST_PWD = "${PWD}"
$HOST_DIR_NAME = If ($_PWD_PATH.Root.FullName -eq $_PWD_PATH.FullName) {""} Else {$_PWD_PATH.Name}
$SERVICE_ARGS = ""

# PARSE OPTIONS
if ($args[0] -eq '-v')
{
    $DOCKERIZED_OPT_VERBOSE = $true
    $args[0] = ""
}

# convert windows paths to unix paths
$SERVICE_ARGS = ($args | % { $_.replace('\', '/') })

function DotEnv
{
    param(
        [Parameter(Mandatory = $true)]
        $file
    )

    # exit if file is not found
    if ($file -eq "" -or !(Test-Path $file))
    {
        return
    }

    if ($DOCKERIZED_OPT_VERBOSE)
    {
        Write-Host "Loading environment from $file" -ForegroundColor Green
    }

    $content = (Get-Content $file)

    if ($content.Length -eq 0)
    {
        return
    }

    $lines = $content.Split("\n")
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

function FindUp
{
    param($FILE, $DIR)
    $PATH = Get-Item -Path $DIR
    while ($PATH.FullName -ne $PATH.Root.FullName)
    {
        $TARGET_FILE = "${PATH}\${FILE}"
        if (Test-Path "$TARGET_FILE")
        {
            return $TARGET_FILE
        }
        $PATH = Get-Item -Path $PATH.Parent.FullName
    }
    return ""
}

function LoadEnvironment
{
    $envFile = FindUp $DOCKERIZED_ENV_FILE_NAME $PWD
    DotEnv "${HOME}\${DOCKERIZED_ENV_FILE_NAME}"
    DotEnv $envFile
}

LoadEnvironment

docker-compose `
    --env-file $DOCKERIZED_ENV_FILE `
    -f $DOCKERIZED_COMPOSE_FILE `
    run --rm `
    -e "HOST_HOME=$HOST_HOME" `
    -v "${PWD}:/host/${HOST_DIR_NAME}" `
    -w /host/${HOST_DIR_NAME} `
    ${SERVICE_ARGS}
