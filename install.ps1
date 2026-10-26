param(
    [string]$Dir,
    [int]$Port,
    [string]$Image,
    [string]$PublicUrl,
    [string]$GitHubOAuthClientId,
    [string]$GitHubOAuthClientSecret,
    [string]$GitHubOAuthScope,
    [switch]$Yes,
    [switch]$NoStart,
    [switch]$NoInstallDocker,
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

# Avec une copie du depot, utiliser le script de cette meme version.
if (-not $env:NODALIBRE_INSTALL_SCRIPT_URL -and $PSScriptRoot) {
    $localScript = Join-Path $PSScriptRoot "scripts/install.ps1"
    if (Test-Path -LiteralPath $localScript -PathType Leaf) {
        & $localScript @PSBoundParameters
        return
    }
}

$scriptUrl = if ($env:NODALIBRE_INSTALL_SCRIPT_URL) {
    $env:NODALIBRE_INSTALL_SCRIPT_URL
} else {
    "https://raw.githubusercontent.com/AnARCHIS12/Libre-claude/main/scripts/install.ps1"
}
# Conserver le fonctionnement de la commande historique irm ... | iex,
# sans imposer l'execution d'un fichier temporaire sous Windows.
$installer = Invoke-RestMethod -Uri $scriptUrl
& ([scriptblock]::Create([string]$installer)) @PSBoundParameters
