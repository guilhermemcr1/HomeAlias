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
