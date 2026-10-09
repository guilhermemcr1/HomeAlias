# Hardening da interface — 08/10/2026

Aplicação de Impeccable harden sobre as telas do painel e de autenticação, preservando a identidade visual existente e as alterações já presentes no workspace.

## Ajustes

- Login com validação local de e-mail e senha obrigatória, erros associados aos campos, foco no primeiro campo inválido e preservação dos valores em falhas.
- Foco em campos inválidos nos formulários compartilhados e no convite; campos desabilitados durante operações em modais e proteção nos handlers contra envios repetidos.
- Modais com nomes acessíveis, abertura sincronizada com a montagem, rolagem limitada à janela e bloqueio de cancelamento durante confirmação em andamento.
- Erros persistentes com nova tentativa em painel, conexões, hosts, tokens, histórico, alertas, usuários e auditoria. Falhas de carregamento não aparecem como cadastros vazios.
- Timeout de 30 segundos nas APIs, cancelamento, mensagens para falhas de rede/servidor, proteção contra respostas HTML e JSON inválido, e orientação para conferir operações que podem ter terminado após um timeout. Não há repetição automática de gravações.
- Validação do tamanho total do DNS e seleção de hosts disponíveis; prévia correta para a raiz `@`; datas inválidas não interrompem as listas.
- Navegação, ações, modais, cabeçalhos de código e campos do roteador adaptados a telas pequenas e texto ampliado; teclados de e-mail, controles de toque de pelo menos 44 px no mobile e respeito a movimento reduzido.
- Falhas de clipboard deixam de indicar sucesso; recuperação do foco e limpeza dos temporizadores. Scripts de cliente têm erros visíveis, botão de nova tentativa, timeout e cancelamento ao sair da tela.
- Logout com falha mantém o estado de autenticação local para permitir nova tentativa.

## Validação executada

- Frontend: 23 testes unitários aprovados, incluindo limites Unicode/DNS, API, CSRF, sessão expirada, timeout, cancelamento e datas inválidas.
- Backend: `go test ./internal/... ./tests/unit` aprovado.
- Build: `npm --prefix frontend run build -- --outDir /tmp/homealias-harden-verified-dist` aprovado, incluindo verificação de tipos.
- Navegador: todas as rotas em 320, 375, 390, 768 e 1280 px; dados longos, Unicode, IPv6, vazio, falhas de carregamento e recuperação. A rodada de cinco larguras aprovou 318 verificações.
- Verificações adicionais em 375 e 1280 px: edição de perfil/usuários, redefinição de senha, geração de token, quatro tipos de cliente, recuperação dos scripts, tema escuro e ampliação de texto em 200%. As correções finais de ampliação foram verificadas por largura.
- Detector Impeccable executado uma vez: identificou uma curva de animação com overshoot, substituída por desaceleração suave.
- `git diff --check` aprovado.

Os testes de navegador usam APIs simuladas e não alteram os cadastros reais. Não houve teste em aparelho físico, Safari/iOS, leitor de tela ou serviços reais de Cloudflare/e-mail/Telegram nesta rodada.

## Repetição dos testes

Veja `frontend/e2e/README.md`. Em `frontend`, com o servidor iniciado, execute `npm run test:e2e`. Os testes e a dependência Playwright foram adicionados ao projeto.

## Pendências do ambiente

- O build padrão não consegue limpar `frontend/dist/assets` por falta de permissão, inclusive fora do sandbox. Foi verificado com saída alternativa em `/tmp`; a propriedade/permissão da pasta existente permanece pendente.
- `npm audit` apontou 13 vulnerabilidades na árvore de ferramentas: 5 moderadas, 6 altas e 2 críticas. Entre os pacotes afetados estão Vitest, tinypool, Vite e Tailwind. A resolução completa indicada pelo npm inclui migrações de versão principal de Vitest e Tailwind; elas não foram aplicadas nesta revisão de interface.

## Nova rodada após as edições de DNS e conexão

Esta rodada usa Impeccable harden e a skill webapp-testing com as suítes Playwright do repositório. As APIs são simuladas: nenhuma credencial, registro DNS ou cadastro de produção é alterado.

Correções aplicadas:

- Navbar permite quebra de linha e altura conforme o conteúdo, corrigindo o overflow com ampliação de texto em 200%. O menu de usuário respeita a largura disponível.
- Removida a largura mínima fixa do body, que podia produzir overflow em 320 px quando o navegador reservava espaço para a barra de rolagem. Os testes agora comparam a largura de conteúdo com a área efetivamente disponível.
- Erros do grupo A/AAAA associados aos checkboxes e anunciados via ARIA; envio inválido recupera foco no primeiro checkbox.
- Mensagens de conflito HTTP 409 explicam o motivo retornado pela API. Respostas HTML/JSON brutas e mensagens excessivamente longas ficam fora dos erros de validação exibidos.
- TTL ausente aparece como “não informado”, sem mostrar `undefined`; a edição trata opções ausentes e orienta a recarregar quando a zona não pode ser identificada.
- Modais fechados marcados como ocultos para tecnologias assistivas. Títulos de edição, redefinição de senha e exclusão têm texto válido mesmo sem um recurso selecionado.

Cobertura adicional:

- Edição de conexão sem substituir a credencial; troca do token, falha e repetição, sem exibição da credencial original.
- Edição de hostname, validação da zona, conflitos e preservação dos campos após falha.
- Proxy/TTL, ausência de A/AAAA, recuperação de foco e reenvio.
- Cinco abas de cliente, incluindo cURL; A/AAAA/dual, downloads e navegação por teclado.
- Matriz de erros para quatro formulários em 320/375/1280 px: lentidão, múltiplos envios, bloqueio de cancelamento durante gravação, rede offline, respostas inválidas, permissões, conflitos, rate limit e erros de servidor. Sessão expirada remove o painel e redireciona ao login.

Os testes não substituem validação em aparelho físico, Safari/iOS, leitor de tela, banco real dedicado ou Cloudflare real. A revisão continua sendo de robustez funcional da interface, sem atestar ausência de vulnerabilidades. O diagnóstico de dependências acima pertence à primeira rodada e não foi repetido nesta etapa.

Resultados confirmados nesta rodada: 38 testes frontend, testes Go de domínio/API/unidade/contratos e build aprovados. A execução geral do navegador aprovou 318 verificações em cinco larguras, 50 verificações adicionais e os fluxos de famílias de IP, DNS e edição de recursos. A matriz adicional aprovou 909 verificações em 320/375/1280 px. As capturas foram inspecionadas; a última correção de dados ausentes e títulos dos modais está coberta pela verificação adicional em 320 px.

## Revisão de formulários e linguagem

- Campos passam a validar quando o usuário sai deles, sem marcar os demais campos antes da interação. A correção remove o erro; enviar o formulário continua validando todos os campos e focando o primeiro inválido.
- Dicas de preenchimento permanecem visíveis junto aos erros, com IDs separados e associação por `aria-describedby`. Campos inválidos recebem borda e foco de erro.
- Controle compartilhado para mostrar/ocultar senhas e o token digitado da Cloudflare, com nomes acessíveis por campo. Ao cancelar ou concluir alterações de senha, os valores são apagados; reabrir um modal também restaura o estado oculto e limpa o feedback anterior.
- Validação de senhas no navegador inclui as regras de repetição e senhas comuns do backend. Um teste confere a lista do servidor para detectar divergências. O convite exige confirmação de senha e rejeita códigos incompletos antes do envio.
- Seleções de conexão/domínio precisam corresponder aos itens disponíveis. Consultas prévias de endereço são canceladas ao editar ou fechar o formulário; uma resposta antiga não pode exibir um conflito para o novo nome. Conflitos de criação continuam visíveis se uma nova consulta falhar.
- Linguagem: domínio em vez de zona, endereço completo em vez de hostname nos campos, dispositivo em instruções e tempo de cache com explicação do TTL. Permissões Cloudflare e nomes de protocolos permanecem exatos onde são necessários para configurar o serviço.
- Histórico e auditoria apresentam resultados e ações em português; códigos e mensagens técnicas ficam em informações complementares. A página Alertas informa que o envio automático por regras ainda não está disponível.
- O filtro de histórico agora quebra em linha no mobile para acomodar rótulos de resultado mais claros.

A rodada usa APIs simuladas e cobre cinco larguras na suíte geral, falhas 400/403/404/409/429/500/502/503/504, resposta inválida, conexão indisponível e sessão expirada. `form-feedback.mjs` adiciona cenários de validação ao sair do campo, preservação das dicas, confirmação de senha, cancelamento de consultas antigas e linguagem em 320, 375 e 1280 px. Não foram feitas chamadas reais à Cloudflare, SMTP ou Telegram.

## Organização dos modais

O componente compartilhado admite três larguras: padrão para ações simples, ampla para token DDNS, conexão, perfil, usuários e canais, e maior para criação/edição de host. Os campos relacionados se organizam em duas colunas quando o espaço disponível permite; no celular e com texto ampliado, voltam a uma coluna. A rolagem interna mantém os controles acessíveis em telas baixas. `modal-layout.mjs` cobre seis modais nas larguras 320, 375, 768 e 1280 px, com APIs simuladas.

## Exclusão de conexões

A aba Conexões permite excluir uma conexão após confirmação. A operação exige sessão e CSRF, respeita o proprietário (ou acesso de administrador) e retorna 409 se existirem hosts vinculados. A chave estrangeira do banco impede a exclusão também quando um host é criado simultaneamente. A conta e os registros na Cloudflare não são alterados. Sucessos são registrados na auditoria; falhas mantêm o diálogo aberto para correção ou nova tentativa. O navegador cobre cancelamento, erros e envios repetidos com APIs simuladas; os testes Go do handler cobrem autorização, erros de banco, conflito e auditoria. A suíte de integração cobre a restrição no banco real quando `TEST_DATABASE_DSN` está configurado para um banco dedicado de testes.
