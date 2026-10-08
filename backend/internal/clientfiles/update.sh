#!/bin/sh
# HomeAlias - atualiza o IP do seu host (Linux, macOS, NAS, Raspberry Pi).
#
# Uso:
#   ./homealias-update.sh                      (se foi baixado ja configurado pelo painel)
#   HOMEALIAS_URL=https://ddns.exemplo.com HOMEALIAS_TOKEN=... HOMEALIAS_HOSTNAME=casa.exemplo.com ./homealias-update.sh
#
# Opcoes (variaveis de ambiente):
#   HOMEALIAS_IPV6=1   tambem atualiza o registro AAAA (requer IPv6 na rede)
#
# Agende com o cron (a cada 5 minutos):
#   */5 * * * * /caminho/homealias-update.sh >/dev/null 2>&1

HA_URL="${HOMEALIAS_URL:-__HOMEALIAS_URL__}"
HA_TOKEN="${HOMEALIAS_TOKEN:-__HOMEALIAS_TOKEN__}"
HA_HOST="${HOMEALIAS_HOSTNAME:-__HOMEALIAS_HOSTNAME__}"
HA_IPV6="${HOMEALIAS_IPV6:-0}"

for v in "$HA_URL" "$HA_TOKEN" "$HA_HOST"; do
  case "$v" in
    ""|__*)
      echo "homealias: configure HOMEALIAS_URL, HOMEALIAS_TOKEN e HOMEALIAS_HOSTNAME (ou baixe o script ja configurado no painel)." >&2
      exit 2
      ;;
  esac
done

HA_URL="${HA_URL%/}"
UA="homealias-shell/1.0"

update() { # $1 = -4 ou -6
  if command -v curl >/dev/null 2>&1; then
    # O token vai por stdin (-K -): nao aparece na lista de processos.
    printf 'header = "Authorization: Bearer %s"\n' "$HA_TOKEN" |
      curl "$1" -fsS --max-time 20 -K - -A "$UA" --get --data-urlencode "hostname=$HA_HOST" "$HA_URL/update"
  elif command -v wget >/dev/null 2>&1; then
    wget "$1" -qO- --timeout=20 --user-agent="$UA" --header="Authorization: Bearer $HA_TOKEN" "$HA_URL/update?hostname=$HA_HOST"
  else
    echo "homealias: instale curl ou wget." >&2
    return 127
  fi
}

update -4
rc=$?
echo
if [ "$HA_IPV6" = "1" ]; then
  update -6 || echo "homealias: IPv6 indisponivel nesta rede (ignorado)." >&2
  echo
fi
exit $rc
