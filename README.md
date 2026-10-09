# HomeAlias

DDNS self-hosted sobre a Cloudflare. Mantém um hostname (ex.: `casa.seudominio.com`) apontando para o IP atual da sua rede, com painel web, histórico de atualizações e alertas.

**Como funciona:** Conexão Cloudflare → Host → Token DDNS → o cliente (roteador, Docker, Windows, Linux) envia a atualização → o painel mostra o estado (online, atenção, offline).

## Sumário

- [Recursos](#recursos)
- [Início rápido](#início-rápido)
- [Primeiro acesso](#primeiro-acesso)
- [Atualizando o IP (clientes)](#atualizando-o-ip-clientes)
- [Configuração](#configuração)
- [Segurança](#segurança)
- [Manutenção](#manutenção)
- [Desenvolvimento](#desenvolvimento)
- [Estrutura do repositório](#estrutura-do-repositório)
- [Documentação](#documentação)

## Recursos

- Painel em pt-BR, tema "terminal" (JetBrains Mono em itálico; escuro/claro), com conexões, hosts, tokens, histórico, alertas e auditoria.
- Usuários por convite ou senha inicial (sem cadastro aberto), com edição de perfil, troca de senha e redefinição pelo admin.
- Endpoints de atualização compatíveis com DuckDNS (`/update`) e protocolo DDNS (`/nic/update`).
- Clientes prontos nas abas da tela do token: **Script Linux** (com cron), **Docker** sem imagem própria, **Roteador (DDNS)**, **Windows** (script `.ps1` já configurado, com tarefa agendada) e **cURL** para atualização manual.
- Canais de e-mail (SMTP) e Telegram com envio de teste. E-mails usam um [template HTML responsivo](backend/internal/alert/templates/email.html) com a identidade do HomeAlias e uma alternativa em texto simples. O envio automático por regras ainda está pendente; veja a [revisão de implementação](docs/IMPLEMENTATION_REVIEW.md).
- Hosts com escolha de proxy Cloudflare ou só DNS, configuração de A/AAAA e TTL na criação e edição.
- Edição de conexões com substituição do token Cloudflare e edição de hostname dentro da zona atual, mantendo os vínculos dos hosts e tokens DDNS.
- Segredos protegidos: token da Cloudflare cifrado (AES-GCM); token DDNS guardado só como hash e exibido uma única vez.

**Stack:** Go 1.23 (chi, sqlx) + MariaDB · Vue 3, Vite, TypeScript, Tailwind + DaisyUI · Docker Compose + Cloudflare Tunnel.

## Início rápido

**Pré-requisitos:** Docker com Compose, MariaDB já instalado no host, `openssl` e um Cloudflare Tunnel (`cloudflared`).

O Compose sobe só o backend (com o build do frontend); o banco é o MariaDB do host.

### 1. Banco de dados

Ajuste a senha em `deploy/sql/create-db.sql.example`, rode como root e importe o esquema:

```bash
sudo mariadb < deploy/sql/create-db.sql.example
sudo mariadb homealias < deploy/sql/schema.sql
```

O usuário do app só tem `SELECT/INSERT/UPDATE/DELETE`. Mantenha o MariaDB em `bind-address=127.0.0.1`: o backend roda com `network_mode: host` e conecta em `127.0.0.1:3306`, sem expor o banco.

### 2. Configuração

```bash
cp deploy/.env.example deploy/.env
openssl rand -hex 32   # rode uma vez para cada chave secreta
```

Preencha o `deploy/.env` (chaves, admin e `DATABASE_DSN` com a senha do banco). O backend se recusa a iniciar com as chaves de exemplo. Todas as variáveis estão em [Configuração](#configuração).

### 3. Subir

```bash
docker compose -f deploy/docker-compose.yml --env-file deploy/.env up -d --build
```

Os containers usam nomes que identificam o projeto e sua função:

- `homealias-api`: API e servidor do painel web.
- `homealias-frontend-build`: gera os arquivos do painel; encerrar com código `0` indica sucesso.

Após atualizar o Compose, o próximo `up` recria os containers com esses nomes e mantém o volume do frontend.

As redes são atribuídas explicitamente: a API usa `network_mode: host` para acessar o MariaDB local, e o build usa a rede bridge `homealias-build-network`. Os dois containers usam DNS `1.1.1.1` e `1.0.0.1`; personalize com `HOMEALIAS_DNS_PRIMARY` e `HOMEALIAS_DNS_SECONDARY` em `deploy/.env` se precisar resolver nomes da sua rede local.

Antes de aplicar alterações no deploy, valide o Compose:

```bash
docker compose -f deploy/docker-compose.yml --env-file deploy/.env config --quiet
```

O backend escuta só em `127.0.0.1:8080`. Aponte o cloudflared para `http://localhost:8080` (veja [`deploy/tunnel/README.md`](deploy/tunnel/README.md)). **Não** abra a porta 8080 no firewall nem no roteador.

### 4. Entrar

Acesse o painel com `HOMEALIAS_ADMIN_EMAIL` / `HOMEALIAS_ADMIN_PASSWORD` (o admin só é criado se não existir nenhum) e **troque a senha**.

## Primeiro acesso

No painel, siga a ordem: **Conexões** → **Hosts** → **Tokens**.

1. **Conexão:** cadastre o token da Cloudflare.
   Cloudflare → Meu perfil → Tokens de API → criar com **Zone · Zone · Read** e **Zone · DNS · Edit**, limitado às zonas necessárias. Cole apenas o token (não a Global API Key).
2. **Host:** informe o hostname e escolha A/AAAA, **Só DNS (sem proxy)** ou **Com proxy Cloudflare** e TTL. Com proxy, o TTL é automático. Depois, use **Editar DNS** para alterar essas opções. O registro novo é criado na primeira atualização DDNS; editar registros existentes mantém o conteúdo e aplica as opções escolhidas. Desativar um tipo interrompe suas atualizações, sem apagar o registro existente.
3. **Token:** gere o token DDNS (exibido uma única vez) e escolha o cliente na própria tela.

## Atualizando o IP (clientes)

Para trocar o token Cloudflare, abra **Conexões → Editar conexão**. Deixe o campo do token vazio para manter o atual ou cole o novo token para substituí-lo. A nova credencial é validada e precisa listar todas as zonas dos hosts vinculados; a validação não faz uma escrita DNS de teste. Mantenha as permissões **Zone · Zone · Read** e **Zone · DNS · Edit**. A conexão e os hosts mantêm seus IDs e vínculos.

Para renomear, abra **Hosts → Editar DNS → Hostname** e informe o nome completo na mesma zona. O app rejeita nomes já cadastrados ou com A/AAAA/CNAME existente na Cloudflare. Quando há um IP conhecido ou um registro antigo, cria o novo DNS com esse conteúdo; caso contrário, aguarda a primeira atualização DDNS. O DNS antigo é preservado. Os tokens continuam válidos para o host, mas os clientes precisam usar o novo hostname. Edite os scripts/roteadores existentes; se perder o token, gere outro, pois o segredo não pode ser recuperado no painel.

O IP vem do cabeçalho `CF-Connecting-IP`; parâmetros `ip=`/`myip=` são ignorados. Depois de gerar o token, o painel mostra cada cliente já preenchido para o seu host.

Os comandos e scripts gerados seguem os registros habilitados: **A usa IPv4 (`-4`)**, **AAAA usa IPv6 (`-6`)** e um host com ambos faz duas chamadas separadas. A conexão IPv6 não revela o IPv4 público do cliente, nem o contrário. Se a rede não oferece uma família habilitada, o cliente informa a falha e ainda tenta a outra. Ao alterar os tipos de registro do host, gere novamente as instruções ou ajuste as opções dos scripts existentes.

| Cliente | Como usar |
|---|---|
| **Script Linux** | Baixe `homealias-update.sh` e agende no cron (`*/5 * * * *`). O arquivo vem com as famílias do host; `HOMEALIAS_IPV4` e `HOMEALIAS_IPV6` aceitam `0` ou `1` para ajuste manual. |
| **Docker** | Container com `curlimages/curl` (abaixo), sem imagem própria. |
| **Roteador (DDNS)** | Servidor `SEU_HOST`, porta 443 com HTTPS, caminho `/nic/update`, usuário qualquer, senha = token. O painel lista cada campo com botão de copiar e exemplos para OpenWrt e EdgeOS. |
| **Windows** | Baixe `homealias-update.ps1` e rode `-Install` para criar a tarefa agendada ([`windows/README.md`](windows/README.md)). |
| **cURL** | Chamada manual (abaixo). |

**Docker:**

```bash
docker run -d --name homealias-ddns --restart unless-stopped \
  -e HA_TOKEN=SEU_TOKEN curlimages/curl:latest \
  sh -c 'while true; do curl -4 --fail-with-body -sS --max-time 20 -A homealias-docker/1.0 -H "Authorization: Bearer $HA_TOKEN" "https://SEU_HOST/update?hostname=casa.exemplo.com"; echo; sleep 300; done'
```

**cURL:**

```bash
curl -4 --fail-with-body -sS --max-time 20 -H "Authorization: Bearer SEU_TOKEN" "https://SEU_HOST/update?hostname=casa.exemplo.com"
```

## Configuração

Variáveis lidas pelo backend (`deploy/.env`):

| Variável | Obrigatória | Descrição |
|---|---|---|
| `HOMEALIAS_ENCRYPTION_KEY` | sim | 32 caracteres ASCII, 64 hex ou base64 de 32 bytes. Cifra os tokens da Cloudflare. **Perdê-la invalida as conexões salvas.** |
| `HOMEALIAS_SESSION_KEY` | sim | Mínimo de 32 caracteres. |
| `HOMEALIAS_ADMIN_EMAIL` / `HOMEALIAS_ADMIN_PASSWORD` | sim | Conta admin criada no primeiro start. |
| `DATABASE_DSN` | sim | DSN do MariaDB do host, ex.: `homealias:SENHA@tcp(127.0.0.1:3306)/homealias`. |
| `TRUST_CLOUDFLARE_HEADERS` | não (`true`) | Confia em `CF-Connecting-IP`. Use `true` só atrás do Tunnel. |
| `COOKIE_SECURE` | não (`true`) | Exige HTTPS para cookies; HTTPS direto ou confiável via Cloudflare sempre usa `Secure`. |
| `HOMEALIAS_BIND_IP` / `HOMEALIAS_PORT` | não (`127.0.0.1` / `8080`) | Interface e porta publicadas pelo Compose. |
| `SESSION_IDLE_MINUTES` | não (`60`) | Expiração da sessão por inatividade. |
| `STATUS_WARNING_AFTER_SEC` / `STATUS_OFFLINE_AFTER_SEC` | não (`900` / `3600`) | Limites de status dos hosts. |
| `RETENTION_DAYS` | não (`90`) | Retenção de histórico e auditoria. |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM` | não | Alertas por e-mail. |
| `TELEGRAM_BOT_TOKEN` | não | Alertas por Telegram. |

## Segurança

- **Senhas:** mínimo de 12 caracteres, sem senhas comuns/repetitivas; Argon2id (t=3, 64 MiB), com rehash automático de hashes antigos no login.
- **Rate limits** (por IP real via `CF-Connecting-IP`): login 10/min por IP e 8 por 15 min por conta; API do painel 240/min; aceite de convite 10/min; `/update` 120/min por IP e 60/min por token. Excesso responde `429` com `Retry-After`.
- **Cookies** `Secure` + `HttpOnly` + `SameSite=Lax` por padrão (`COOKIE_SECURE=true`), CSRF double-submit, CSP e demais cabeçalhos de segurança.
- **Segredos:** tokens DDNS guardados como SHA-256; token da Cloudflare cifrado com AES-GCM. Logs de acesso sem query string (o token do `/update` não é registrado).
- **Validação** de entrada no frontend **e** no backend (`backend/internal/validate`, espelhada em `frontend/src/composables/validation.ts`).
- **Escopo por dono:** usuários comuns só enxergam os próprios recursos; rotas de administração exigem perfil admin.
- **Recomendado na Cloudflare:** Cloudflare Access (Zero Trust) na frente do painel, deixando `/update` e `/nic/update` fora da política de Access, com uma regra de rate limit na WAF.

### Acesso pela LAN em HTTP (apenas desenvolvimento)

No `deploy/.env`: `HOMEALIAS_BIND_IP=0.0.0.0`, `COOKIE_SECURE=false` e **`TRUST_CLOUDFLARE_HEADERS=false`** (com `true`, qualquer pessoa da rede forja `CF-Connecting-IP` e burla os rate limits). Volte aos valores de produção antes de publicar. Neste modo o `/update` não funciona, pois depende do IP confiável da Cloudflare.

## Manutenção

Para atualizar a instalação:

```bash
git pull
docker compose -f deploy/docker-compose.yml --env-file deploy/.env up -d --build
```

Se vierem novas migrations em `backend/migrations/`, aplique os `*.up.sql` novos no MariaDB **antes** de subir (ou regenere `deploy/sql/schema.sql` em um banco novo).

## Desenvolvimento

```bash
# backend (os testes de API usam um banco dedicado: veja backend/tests/run-api-tests.sh)
cd backend && go vet ./... && go test ./...

# frontend (proxy de /api para localhost:8080)
cd frontend && npm ci && npm run dev
npm test            # vitest (validação e placeholders dos campos)
npx vue-tsc -b      # checagem de tipos
```

Convenções do painel: formulários de criação e edição abrem em modal (`FormModal`), todo campo de texto tem placeholder de exemplo (verificado por teste).

## Estrutura do repositório

```text
backend/    API, regras de domínio, migrations e testes (Go)
frontend/   Painel web (Vue)
agent/      Cliente em Go opcional (o painel recomenda Docker com `curlimages/curl`)
windows/    Guia do cliente Windows (o script fica em backend/internal/clientfiles)
deploy/     docker-compose.yml, .env.example, esquema SQL (`sql/`) e guia do Tunnel
docs/       Notas de robustez da interface
```

## Documentação

- [`docs/HARDENING.md`](docs/HARDENING.md): robustez dos formulários e telas do painel.

- [`docs/IMPLEMENTATION_REVIEW.md`](docs/IMPLEMENTATION_REVIEW.md): recursos verificados, correções e lacunas de implementação.
