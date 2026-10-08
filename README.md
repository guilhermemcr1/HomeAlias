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
- Endpoints de atualização compatíveis com DuckDNS (`/update`) e DynDNS (`/nic/update`).
- Clientes prontos na tela do token: **Script Linux** (com cron), **Docker** sem imagem própria, **Roteador (DynDNS)** e **Windows** (script `.ps1` já configurado, com tarefa agendada).
- Alertas por e-mail (SMTP) e Telegram.
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

O backend escuta só em `127.0.0.1:8080`. Aponte o cloudflared para `http://localhost:8080` (veja [`deploy/tunnel/README.md`](deploy/tunnel/README.md)). **Não** abra a porta 8080 no firewall nem no roteador.

### 4. Entrar

Acesse o painel com `HOMEALIAS_ADMIN_EMAIL` / `HOMEALIAS_ADMIN_PASSWORD` (o admin só é criado se não existir nenhum) e **troque a senha**.

## Primeiro acesso

No painel, siga a ordem: **Conexões** → **Hosts** → **Tokens**.

1. **Conexão:** cadastre o token da Cloudflare.
   Cloudflare → Meu perfil → Tokens de API → criar com **Zone · Zone · Read** e **Zone · DNS · Edit**, limitado às zonas necessárias. Cole apenas o token (não a Global API Key).
2. **Host:** informe o hostname que será mantido atualizado.
3. **Token:** gere o token DDNS (exibido uma única vez) e escolha o cliente na própria tela.

## Atualizando o IP (clientes)

O IP vem do cabeçalho `CF-Connecting-IP`; parâmetros `ip=`/`myip=` são ignorados. Depois de gerar o token, o painel mostra cada cliente já preenchido para o seu host.

| Cliente | Como usar |
|---|---|
| **Script Linux** | Baixe `homealias-update.sh` e agende no cron (`*/5 * * * *`). Opcional: `HOMEALIAS_IPV6=1` para o registro AAAA. |
| **Docker** | Container com `curlimages/curl` (abaixo), sem imagem própria. |
| **Roteador (DynDNS)** | Servidor `SEU_HOST`, porta 443 com HTTPS, caminho `/nic/update`, usuário qualquer, senha = token. O painel lista cada campo com botão de copiar e exemplos para OpenWrt e EdgeOS. |
| **Windows** | Baixe `homealias-update.ps1` e rode `-Install` para criar a tarefa agendada ([`windows/README.md`](windows/README.md)). |
| **cURL** | Chamada manual (abaixo). |

**Docker:**

```bash
docker run -d --name homealias-ddns --restart unless-stopped \
  -e HA_TOKEN=SEU_TOKEN curlimages/curl:latest \
  sh -c 'while true; do curl -fsS -A homealias-docker/1.0 -H "Authorization: Bearer $HA_TOKEN" "https://SEU_HOST/update?hostname=casa.exemplo.com"; echo; sleep 300; done'
```

**cURL:**

```bash
curl -fsS -H "Authorization: Bearer SEU_TOKEN" "https://SEU_HOST/update?hostname=casa.exemplo.com"
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
