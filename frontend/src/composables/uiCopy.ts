const results: Record<string, string> = {
  good: 'IP atualizado', updated: 'IP atualizado', nochg: 'IP mantido', unchanged: 'IP mantido',
  ok: 'Concluído', success: 'Concluído', error: 'Falha na atualização', rejected: 'Atualização recusada',
  badauth: 'Token inválido', nohost: 'Host não encontrado', '911': 'Serviço indisponível',
}
export const resultLabel = (code: string) => results[code.toLowerCase()] ?? 'Resultado não identificado'

const actions: Record<string, string> = {
  login: 'Acesso ao painel', login_failed: 'Tentativa de acesso recusada',
  host_create: 'Host criado', host_update: 'Host atualizado', host_delete: 'Host removido',
  token_create: 'Token gerado', token_revoke: 'Token revogado',
  connection_create: 'Conexão criada', connection_update: 'Conexão atualizada',
  connection_delete: 'Conexão excluída',
  profile_update: 'Dados da conta alterados', password_change: 'Senha alterada',
  user_invite: 'Convite criado', user_create: 'Usuário criado', user_update: 'Usuário atualizado',
  user_disable: 'Usuário desativado', user_enable: 'Usuário reativado', user_password_reset: 'Senha redefinida',
}
export const actionLabel = (code: string) => actions[code] ?? 'Outra ação'
export const resourceLabel = (code: string) => ({ host: 'Host', ddns_token: 'Token DDNS', connection: 'Conexão', user: 'Usuário' }[code] ?? 'Outro item')

const summaries: Record<string, string> = {
  'login ok': 'Entrada no painel confirmada.', 'login failed': 'E-mail ou senha não aceitos.',
  'token emitted': 'Token de atualização gerado.', 'token revoked': 'Token de atualização revogado.',
  'host updated': 'Configuração do host alterada.', 'host deleted': 'Host removido.',
  'connection created': 'Conta da Cloudflare conectada.', 'connection updated': 'Conexão com a Cloudflare alterada.',
  'connection deleted': 'Conexão com a Cloudflare excluída.',
  'profile updated': 'Nome ou e-mail da conta alterado.', 'password changed': 'Senha da conta alterada.',
  'invite created': 'Link de convite gerado.', 'user created with initial password': 'Conta criada com uma senha inicial.',
  'user disabled': 'Acesso da conta desativado.', 'user enabled': 'Acesso da conta reativado.',
  'user updated': 'Dados ou permissões da conta alterados.', 'password reset by admin': 'Senha redefinida pelo administrador.',
}
export const auditSummary = (text: string) => summaries[text] ?? (text.startsWith('host ') ? `Endereço: ${text.slice(5)}` : text)

export function updateError(reason: string): string {
  if (reason === 'hostname not linked') return 'Este endereço não está autorizado pelo token. Confira o endereço e o token do dispositivo.'
  if (reason === 'type not enabled') return 'Este tipo de IP não está ativado no host. Confira se o dispositivo envia IPv4 ou IPv6.'
  return 'Não foi possível atualizar o endereço. Confira o dispositivo e a conexão com a Cloudflare.'
}
