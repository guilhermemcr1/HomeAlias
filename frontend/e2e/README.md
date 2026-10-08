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

`harden.mjs` cobre todas as rotas em 320, 375, 390, 768 e 1280 px, dados longos, erros de carregamento, recuperação, foco em campos inválidos, login offline e envios repetidos. `forms-and-tokens.mjs` cobre edição de conta/usuários, senhas, emissão de token, os quatro tipos de cliente, recuperação de scripts, estados vazios, tema escuro e texto ampliado em 200%.

Para uma largura específica, use `E2E_WIDTHS=375` (ou uma lista separada por vírgulas). As capturas ficam em `/tmp/homealias-*.png`. Os testes verificam o Chrome com viewports simulados; não substituem uma rodada em aparelhos reais com Safari/iOS ou Chrome/Android.
