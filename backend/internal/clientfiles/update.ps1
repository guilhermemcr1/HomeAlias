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
  [switch]$Ipv6
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

if ($Install) {
  $script = $PSCommandPath
  $args = "-NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -File `"$script`""
  if ($Ipv6) { $args += " -Ipv6" }
  $action = New-ScheduledTaskAction -Execute "powershell.exe" -Argument $args
  $trigger = New-ScheduledTaskTrigger -Once -At (Get-Date).AddMinutes(1) -RepetitionInterval (New-TimeSpan -Minutes $EveryMinutes)
  Register-ScheduledTask -TaskName $TaskName -Action $action -Trigger $trigger -Force | Out-Null
  Write-Output "Tarefa '$TaskName' criada: roda a cada $EveryMinutes minutos."
}

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
$uri = "$($Url.TrimEnd('/'))/update?hostname=$([uri]::EscapeDataString($Hostname))"
# O token vai no cabecalho Authorization (nao na URL), fora de logs de proxy.
$headers = @{ "User-Agent" = "homealias-windows/1.0"; "Authorization" = "Bearer $Token" }
try {
  $resp = Invoke-RestMethod -Uri $uri -Headers $headers -Method Get -TimeoutSec 20
  Write-Output ($resp | ConvertTo-Json -Compress)
} catch {
  Write-Error $_.Exception.Message
  exit 1
}
