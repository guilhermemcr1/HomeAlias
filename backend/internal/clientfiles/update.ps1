<#
HomeAlias - atualiza o IP do seu host no Windows.

Uso:
  .\homealias-update.ps1                  Atualiza agora (se o arquivo veio configurado do painel)
  .\homealias-update.ps1 -Install         Cria uma tarefa agendada (a cada 5 minutos) e atualiza agora
  .\homealias-update.ps1 -Uninstall       Remove a tarefa agendada
  .\homealias-update.ps1 -Url https://ddns.exemplo.com -Token SEU_TOKEN -Hostname casa.exemplo.com

Se o PowerShell bloquear scripts:
  powershell -ExecutionPolicy Bypass -File .\homealias-update.ps1 -Install
#>
param(
  [string]$Url = "__HOMEALIAS_URL__",
  [string]$Token = "__HOMEALIAS_TOKEN__",
  [string]$Hostname = "__HOMEALIAS_HOSTNAME__",
  [int]$EveryMinutes = 5,
  [switch]$Install,
  [switch]$Uninstall,
  [switch]$Ipv4 = ("__HOMEALIAS_IPV4__" -ne "0"),
  [switch]$Ipv6 = ("__HOMEALIAS_IPV6__" -eq "1"),
  [ValidateSet("", "4", "6", "4,6")]
  [string]$Families = ""
)

$TaskName = "HomeAlias DDNS"

if ($Uninstall) {
  Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false -ErrorAction SilentlyContinue
  Write-Output "Tarefa '$TaskName' removida."
  exit 0
}

foreach ($v in @($Url, $Token, $Hostname)) {
  if ([string]::IsNullOrWhiteSpace($v) -or $v.StartsWith("__")) {
    Write-Error "Configure -Url, -Token e -Hostname (ou baixe o script ja configurado no painel)."
    exit 2
  }
}

if ($Families) {
  $Ipv4 = $Families.Split(',') -contains '4'
  $Ipv6 = $Families.Split(',') -contains '6'
}
if (-not $Ipv4 -and -not $Ipv6) {
  Write-Error "Habilite IPv4 e/ou IPv6 para este host."
  exit 2
}
$curl = Get-Command curl.exe -ErrorAction SilentlyContinue
if (-not $curl) {
  Write-Error "curl.exe nao encontrado. Instale o cURL ou use uma versao atual do Windows."
  exit 2
}

if ($Install) {
  $script = $PSCommandPath
  $args = "-NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -File `"$script`""
  $taskFamilies = @()
  if ($Ipv4) { $taskFamilies += '4' }
  if ($Ipv6) { $taskFamilies += '6' }
  $args += " -Families $($taskFamilies -join ',')"
  $action = New-ScheduledTaskAction -Execute "powershell.exe" -Argument $args
  $trigger = New-ScheduledTaskTrigger -Once -At (Get-Date).AddMinutes(1) -RepetitionInterval (New-TimeSpan -Minutes $EveryMinutes)
  Register-ScheduledTask -TaskName $TaskName -Action $action -Trigger $trigger -Force | Out-Null
  Write-Output "Tarefa '$TaskName' criada: roda a cada $EveryMinutes minutos."
}

$uri = "$($Url.TrimEnd('/'))/update"
$ipFamilies = @()
if ($Ipv4) { $ipFamilies += 4 }
if ($Ipv6) { $ipFamilies += 6 }
$failed = $false
foreach ($family in $ipFamilies) {
  try {
    # Cabecalho por stdin: o token nao aparece na URL ou na lista de processos.
    $config = 'header = "Authorization: Bearer {0}"' -f $Token
    $output = @($config | & $curl.Source "-$family" -sS --max-time 20 -K - -A "homealias-windows/1.0" --get --data-urlencode "hostname=$Hostname" --write-out "`n%{http_code}" $uri)
    if ($LASTEXITCODE -ne 0) { throw "Falha na conexao IPv$family (curl: $LASTEXITCODE)." }
    $status = 0
    if ($output.Count -lt 2 -or -not [int]::TryParse($output[-1], [ref]$status)) { throw "Resposta HTTP invalida." }
    $body = $output[0..($output.Count - 2)] -join "`n"
    $resp = $body | ConvertFrom-Json -ErrorAction Stop
    if ($status -lt 200 -or $status -ge 300 -or $resp.success -ne $true) { throw "HTTP ${status}: $($resp.message)" }
    Write-Output ($resp | ConvertTo-Json -Compress)
  } catch {
    Write-Error "IPv${family}: $($_.Exception.Message)" -ErrorAction Continue
    $failed = $true
  }
}
if ($failed) { exit 1 }
