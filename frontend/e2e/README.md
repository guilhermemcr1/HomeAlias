# Validação de formulários e mobile

As suítes usam Playwright com APIs simuladas: não criam usuários, hosts ou tokens reais.

Em `frontend`, instale as dependências e inicie o app:

```sh
npm ci
npm run dev
```

Em outro terminal, também em `frontend`:

```sh
npm run test:e2e
```

O navegador padrão é `/usr/bin/google-chrome`. Para outro executável ou porta:

```sh
E2E_CHROME_PATH=/caminho/chrome E2E_BASE_URL=http://127.0.0.1:5184 npm run test:e2e
```

`harden.mjs` cobre todas as rotas em 320, 375, 390, 768 e 1280 px, dados longos, erros de carregamento, recuperação, foco em campos inválidos, login offline e envios repetidos. `forms-and-tokens.mjs` cobre edição de conta/usuários, senhas, emissão de token, as cinco abas de cliente (incluindo cURL), recuperação de scripts, estados vazios, tema escuro e texto ampliado em 200%.

Para uma largura específica, use `E2E_WIDTHS=375` (ou uma lista separada por vírgulas). As capturas ficam em `/tmp/homealias-*.png`. Os testes verificam o Chrome com viewports simulados; não substituem uma rodada em aparelhos reais com Safari/iOS ou Chrome/Android.

`ip-families.mjs` valida comandos e downloads para A, AAAA e ambos em 375 e 1280 px. `dns-options.mjs` valida criação com proxy, edição para só DNS, TTL automático, bloqueio sem famílias e preservação dos campos após falha na API, nas mesmas larguras. Essas duas suítes usam larguras fixas.

`edit-resources.mjs` valida edição da conexão sem trocar o segredo, troca do token e falha/repetição, além de edição de hostname, conflito e validação de zona em 375/1280 px. Usa credenciais fictícias e não altera DNS real.

`form-failures.mjs` exercita edição de conexão/hostname, emissão de token e criação de canal em 320, 375 e 1280 px. Cobre requisições lentas, envios repetidos, bloqueio de cancelamento durante gravação, erros 400/403/404/409/429/500/502/503/504, JSON inválido e falha de rede, além do redirecionamento ao login após 401. Verifica preservação dos campos, reativação dos controles, mensagens e ausência de erros JavaScript. A verificação de overflow usa a largura disponível do documento, descontando a barra de rolagem.

`form-feedback.mjs` verifica os campos ao perder foco, correções sem novo envio, dicas e erros associados, mostrar/ocultar segredos, limpeza de senha ao cancelar, convite com confirmação de senha, consulta de endereço obsoleta e rótulos em português no histórico/auditoria. Usa 320, 375 e 1280 px e dados fictícios.

`modal-layout.mjs` verifica seis modais em 320, 375, 768 e 1280 px: largura maior no desktop, campos em duas colunas quando há espaço, ausência de overflow, texto ampliado em 200% e acesso aos botões em telas de pouca altura.

`delete-connection.mjs` valida exclusão em 320, 375 e 1280 px: confirmação/cancelamento, conflito com hosts, falta de permissão, item removido, falha do servidor e de rede, recuperação, envio único, cabeçalho CSRF, estado vazio e texto em 200%. As chamadas à API são simuladas.
