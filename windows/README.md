# Cliente Windows

O script de atualização é entregue pelo painel já configurado, então não há nada para editar:

1. Em **Tokens**, gere o token e abra a aba **Windows**.
2. Clique em **Baixar homealias-update.ps1** (o arquivo já traz servidor, host e token).
3. No PowerShell, na pasta do arquivo:

```powershell
powershell -ExecutionPolicy Bypass -File .\homealias-update.ps1 -Install
```

Isso cria a tarefa agendada **HomeAlias DDNS** (a cada 5 minutos) e atualiza o IP na hora. Para remover: `.\homealias-update.ps1 -Uninstall`.

O modelo do script fica em `backend/internal/clientfiles/update.ps1` e é servido em `/client/update.ps1`
(o painel preenche URL, token e hostname no navegador; o token nunca passa pelo servidor nesse download).
Também dá para rodar com parâmetros: `-Url https://seu-host -Token SEU_TOKEN -Hostname casa.exemplo.com [-Ipv6]`.
