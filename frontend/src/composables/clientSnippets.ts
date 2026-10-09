// Monta os comandos e dados de configuração de cada tipo de cliente.
// Tudo é calculado no navegador a partir de URL, hostname e token (que só existe na tela).

export type ClientConfig = { origin: string; hostname: string; token: string }

/** Valida o modelo sem confundir o comentário <# do PowerShell com HTML. */
export function isScriptTemplate(template: string, kind: 'sh' | 'ps1'): boolean {
  const start = template.trimStart()
  if (start.startsWith('<') && !(kind === 'ps1' && start.startsWith('<#'))) return false
  return ['__HOMEALIAS_URL__', '__HOMEALIAS_TOKEN__', '__HOMEALIAS_HOSTNAME__']
    .every((placeholder) => template.includes(placeholder))
}

/** Host sem esquema, como roteadores pedem no campo "Servidor". */
export function serverHost(origin: string): string {
  return origin.replace(/^https?:\/\//, '')
}

/** Preenche os placeholders de um script baixado de /client/*. */
export function configureScript(template: string, c: ClientConfig): string {
  return template
    .split('__HOMEALIAS_URL__').join(c.origin)
    .split('__HOMEALIAS_TOKEN__').join(c.token)
    .split('__HOMEALIAS_HOSTNAME__').join(c.hostname)
}

export const curlCommand = (c: ClientConfig) =>
  `curl -fsS -H "Authorization: Bearer ${c.token}" "${c.origin}/update?hostname=${c.hostname}"`

export const cronLine = (path = '/opt/homealias/homealias-update.sh') => `*/5 * * * * ${path} >/dev/null 2>&1`

// Docker sem imagem própria: a imagem oficial curlimages/curl repete a chamada a cada 5 minutos.
const LOOP = (c: ClientConfig) =>
  `while true; do curl -fsS -A homealias-docker/1.0 -H "Authorization: Bearer $HA_TOKEN" "${c.origin}/update?hostname=${c.hostname}"; echo; sleep 300; done`

export const dockerRun = (c: ClientConfig) =>
  `docker run -d --name homealias-ddns --restart unless-stopped \\\n  -e HA_TOKEN=${c.token} \\\n  curlimages/curl:latest \\\n  sh -c '${LOOP(c)}'`

export const dockerCompose = (c: ClientConfig) =>
  `services:
  homealias-ddns:
    image: curlimages/curl:latest
    restart: unless-stopped
    environment:
      HA_TOKEN: "${c.token}"
    command:
      - sh
      - -c
      - '${LOOP(c).split('$HA_TOKEN').join('$$HA_TOKEN')}'`

export type RouterField = { label: string; value: string; hint?: string; wide?: boolean; secret?: boolean }

/** Campos que a maioria dos roteadores/firewalls pede num serviço "DynDNS personalizado". */
export function routerFields(c: ClientConfig): RouterField[] {
  const host = serverHost(c.origin)
  return [
    { label: 'Serviço / Provedor', value: 'Personalizado (Custom DynDNS)', hint: 'Se houver "dyndns2" ou "DynDNS", também serve.' },
    { label: 'Servidor / Host de atualização', value: host, hint: 'Sem https://. Alguns roteadores pedem a URL inteira; use a de baixo.' },
    { label: 'Porta', value: '443', hint: 'Ative HTTPS/SSL se houver a opção.' },
    { label: 'Caminho', value: '/nic/update' },
    { label: 'Nome do host / Domínio', value: c.hostname },
    { label: 'Usuário', value: 'homealias', hint: 'Qualquer texto; alguns roteadores não aceitam vazio.' },
    { label: 'Senha', value: c.token, hint: 'É o token deste host.', secret: true },
    { label: 'URL completa (se o roteador pedir)', value: `${c.origin}/nic/update?hostname=${c.hostname}`, wide: true },
  ]
}

export const routerTest = (c: ClientConfig) => `curl -u homealias:${c.token} "${c.origin}/nic/update?hostname=${c.hostname}"`

export const openwrtConfig = (c: ClientConfig) =>
  `config service 'homealias'
  option enabled '1'
  option service_name 'custom'
  option lookup_host '${c.hostname}'
  option domain '${c.hostname}'
  option username 'homealias'
  option password '${c.token}'
  option use_https '1'
  option update_url 'https://[USERNAME]:[PASSWORD]@${serverHost(c.origin)}/nic/update?hostname=[DOMAIN]&myip=[IP]'
  option interface 'wan'
  option ip_source 'network'
  option ip_network 'wan'`

export const edgeosConfig = (c: ClientConfig) =>
  `set service dns dynamic interface eth0 service custom-homealias host-name ${c.hostname}
set service dns dynamic interface eth0 service custom-homealias login homealias
set service dns dynamic interface eth0 service custom-homealias password ${c.token}
set service dns dynamic interface eth0 service custom-homealias protocol dyndns2
set service dns dynamic interface eth0 service custom-homealias server ${serverHost(c.origin)}
commit; save`

export const windowsInstall = () => `powershell -ExecutionPolicy Bypass -File .\\homealias-update.ps1 -Install`
