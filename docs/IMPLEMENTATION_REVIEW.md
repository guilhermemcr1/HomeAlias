# Revisão da implementação — 08/10/2026

## Base e limite da comparação

Foram encontrados README.md, PRODUCT.md, DESIGN.md, docs/HARDENING.md, os guias de deploy/Windows e a documentação dos testes. Não foi encontrado no workspace um documento original detalhado de requisitos ou plano de implementação. Esta revisão compara os recursos anunciados nesses arquivos com o código; as lacunas adicionais da API são classificadas como implementação parcial, e não como requisitos confirmados do documento ausente.

Esta é uma revisão funcional, não uma nova auditoria de segurança. Não houve acesso à conta Cloudflare, alteração de DNS real ou envio de notificações reais.

## Proxy DNS: lacuna confirmada e corrigida

A migration, o modelo de host, a API e o provider Cloudflare já tinham `proxied`. O painel não oferecia o campo nem o enviava ao criar um host, resultando em `false`. Também não havia interface de edição DNS.

Agora há escolha explícita **Só DNS (sem proxy)** / **Com proxy Cloudflare**, tanto na criação quanto em **Editar DNS**. A lista mostra o modo e o TTL. A/AAAA e TTL também podem ser editados. O padrão continua sem proxy, sem alteração automática nos hosts já cadastrados.

Com proxy, o backend normaliza o TTL para `1` (automático), e o painel desabilita o controle de TTL. Isso segue a [regra de TTL da Cloudflare](https://developers.cloudflare.com/dns/manage-dns-records/reference/ttl/). O proxy é voltado ao tráfego HTTP/HTTPS; **Só DNS** resolve diretamente o IP de origem, conforme a [documentação de proxy](https://developers.cloudflare.com/dns/proxy-status/).

O backend agora rejeita edição sem A/AAAA habilitados, preserva limites de status não enviados e propaga erros da Cloudflare. Ao editar um registro adotado que ainda não recebeu DDNS, consulta e preserva o conteúdo existente ao mudar proxy/TTL. Uma sincronização sem registro existente nem IP conhecido pede a primeira atualização DDNS, em vez de indicar sucesso sem aplicar nada.

Referências: `frontend/src/components/HostDnsOptions.vue`, `frontend/src/pages/HostsPage.vue`, `backend/internal/http/panel/hosts.go` e `backend/internal/dns/cloudflare/client.go`.

### Comportamentos que precisam ficar claros

- Cadastrar um host novo salva a configuração. O registro DNS novo é criado na primeira atualização DDNS; não é criado com um IP fictício.
- A escolha de proxy vale para os tipos habilitados e gerenciados pelo host. Desativar A/AAAA não apaga o registro existente na Cloudflare.
- A edição mantém o endereço conhecido ou existente. Alterar A/AAAA exige ajustar ou baixar novamente os scripts dos clientes.
- MariaDB e Cloudflare não participam de uma transação distribuída. Se uma família for aplicada e outra falhar, o endpoint informa falha e não salva a nova configuração local; pode haver alteração parcial remota. Repetir a edição reaplica as opções. Se o banco falhar após a Cloudflare aceitar, também é necessário conferir e repetir a operação. Ainda não existe fila de reconciliação.

## Lacunas restantes

### Edição de conexão e hostname: implementada

- **Conexões → Editar conexão** permite mudar o nome e substituir o token de API. Campo vazio conserva a credencial; a original nunca é preenchida no formulário. A substituição valida o token, confere acesso às zonas vinculadas, cifra o novo segredo e renova zonas/status antes de responder. Uma falha de validação conserva a credencial anterior. Editar apenas o nome não regrava o segredo. A validação testa listagem de zonas, não a permissão de escrita mediante alteração DNS real.
- **Hosts → Editar DNS** permite renomear o hostname na zona atual. Valida formato e comprimento, rejeita conflitos no app e A/AAAA/CNAME existentes na Cloudflare e conserva os vínculos pelo ID do host. Copia para o novo nome os IPs conhecidos ou o conteúdo do DNS antigo, mantendo proxy/TTL. O registro antigo não é removido. Sem conteúdo conhecido, a criação aguarda atualização DDNS.
- Scripts e roteadores devem ser ajustados para o nome novo; o token permanece válido, mas a atualização com o hostname antigo passa a ser rejeitada. O painel não recupera tokens DDNS já emitidos.
- A possibilidade de alteração parcial remota descrita acima também se aplica à renomeação; se a operação falhar após criar um registro novo, confira o DNS antes de repetir, pois a verificação de conflito impede sobrescrever um destino já existente.
- Testes Go cobrem as zonas exigidas pelo token, sigilo da resposta, cópia dos IPs/opções e rejeição de destino ocupado. `edit-resources.mjs` cobre campo de token vazio, substituição, falha/repetição, sigilo no painel, edição de hostname, conflitos e limite à zona atual em 375/1280 px. Build e 37 testes frontend aprovados. Não houve teste na Cloudflare real ou em banco dedicado nesta rodada.

| Prioridade | Situação comprovada | O que falta | Evidência |
|---|---|---|---|
| Alta | README anunciava alertas por e-mail/Telegram sem diferenciar envio de teste e automático. O worker busca regras `no_contact`, descarta o resultado e só registra contagem de tokens próximos da expiração. | Avaliar gatilhos, escopos e estados; enviar aos canais; respeitar intervalos; registrar entregas, falhas e recuperação. Atualizado o README para não anunciar automação pronta. | `backend/internal/alert/worker.go`, `backend/internal/http/panel/alerts.go` |
| Alta | A API cadastra regras, mas a tela de Alertas apenas lista regras e cria/testa canais. | Formulário de criação de regras e definição de gatilho, parâmetros, escopo e canais. Rotas de edição, desativação e exclusão também não existem. | `frontend/src/pages/AlertsPage.vue`, `backend/internal/http/router.go` |
| Alta | A exclusão de host ignora erros ao apagar DNS e ao excluir vínculos/host no banco, podendo responder 204 apesar de falha. | Propagar falhas, evitar remoção local quando o usuário pediu exclusão DNS e a Cloudflare falhou, e definir recuperação de operações parciais. | `backend/internal/http/panel/hosts.go`, método `Delete` |
| Média | Existe API de limites globais de status, mas o painel não tem configuração. Os valores gravados em `instance_settings` não são carregados no cálculo da lista de hosts, que usa os defaults do ambiente. | Conectar os defaults persistidos ao cálculo de status, validar valores e ordem dos limites e disponibilizar configuração global/por host. | `backend/internal/http/panel/settings.go`, `hosts.go`, `backend/internal/domain/status.go`, `backend/internal/http/router.go` |
| Média | A API aceita expiração e revogação de tokens; o painel apenas emite um token ligado a um host. Não existe rota de listagem de tokens. | Lista de metadados sem revelar segredo, expiração no formulário e ações de revogação/estado; decidir se o painel também deverá suportar tokens de múltiplos hosts. | `frontend/src/pages/TokenEmitPage.vue`, `backend/internal/http/panel/tokens.go`, `backend/internal/http/router.go` |
| Média | O agente Go opcional faz duas chamadas com o mesmo cliente e transporte padrão; detectar duas famílias nas interfaces não garante chamadas IPv4 e IPv6 separadas. | Forçar transporte `tcp4`/`tcp6`, selecionar as famílias do host e relatar falhas por família. Os scripts Linux/Windows e comandos Docker foram ajustados; o agente opcional não. | `agent/main.go` |
| Média | A Cloudflare lista somente a primeira página de até 50 zonas. | Paginação para contas com mais de 50 zonas. | `backend/internal/dns/cloudflare/client.go`, `ValidateCredentials` |

## Verificações desta alteração

- `go test ./internal/... ./tests/unit ./tests/contract`: aprovado. Testes novos verificam proxy/TTL nas duas famílias, preservação do conteúdo adotado, falha parcial e ausência de registro antes da primeira atualização.
- Frontend: 37 testes aprovados; checagem TypeScript e build aprovados com saída em `/tmp/homealias-proxy-build`.
- `dns-options.mjs`: criação/edição, proxy ligado/desligado, TTL, bloqueio sem A/AAAA e erro/repetição aprovados em 375 e 1280 px; capturas inspecionadas.
- Detector Impeccable sobre os componentes alterados: sem ocorrências.
- O build usa saída alternativa e loader `runner`, pois o diretório de saída/cache padrão tem restrições de permissão no ambiente.
- Não executados nesta rodada: banco de integração dedicado, Cloudflare real, aparelho físico e PowerShell no Windows.

O checklist definitivo de aderência depende de localizar os documentos originais mencionados pelo usuário.

## Atualização da revisão de interface

O overflow da navbar com texto ampliado em 200% foi corrigido: a navegação pode quebrar linhas e o cabeçalho acompanha a altura do conteúdo. A nova rodada também associa erros de A/AAAA aos controles, melhora mensagens de conflito e verifica a largura útil do documento em telas estreitas. A cobertura foi ampliada com uma matriz de erros de formulários; detalhes em [HARDENING.md](HARDENING.md). As outras lacunas funcionais da tabela acima permanecem pendentes.
