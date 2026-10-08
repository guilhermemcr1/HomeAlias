# Cloudflare Tunnel

O backend escuta apenas em `127.0.0.1:8080` (`network_mode: host`). Rode o cloudflared no próprio host
(ou em um contêiner com `network_mode: host`) e aponte para localhost:

```yaml
ingress:
  - hostname: homealias.exemplo.com
    service: http://localhost:8080
  - service: http_status:404
```

Requisitos em `deploy/.env`: `TRUST_CLOUDFLARE_HEADERS=true`, `COOKIE_SECURE=true`, `HOMEALIAS_BIND_IP=127.0.0.1`.
Com `TRUST_CLOUDFLARE_HEADERS=true` o backend confia no `CF-Connecting-IP`; por isso a porta **nunca** pode ficar acessível
fora do Tunnel (sem port-forward, sem `0.0.0.0`).

Recomendações na Cloudflare:

- **Access (Zero Trust)** protegendo o painel (`/` e `/api/*`), com login por e-mail/SSO. Crie uma política *Bypass* apenas para `/update` e `/nic/update`, usados pelos clientes DDNS (autenticados por token).
- **Rate limiting (WAF)** em `/update*`, `/nic/update` e `/api/auth/*` como segunda camada.
- Token da API da Cloudflare do HomeAlias com **Zone:Read** e **DNS:Edit** apenas nas zonas necessárias.
