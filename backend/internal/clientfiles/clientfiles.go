// Package clientfiles serve os scripts de cliente (Shell e PowerShell) para download.
// Os arquivos trazem placeholders __HOMEALIAS_*__; o painel baixa o modelo e preenche
// URL, token e hostname no navegador, então o token nunca passa por este endpoint.
package clientfiles

import (
	"embed"
	"net/http"
)

//go:embed update.sh update.ps1
var files embed.FS

var served = map[string]struct{ src, name, ctype string }{
	"update.sh":  {"update.sh", "homealias-update.sh", "text/x-shellscript; charset=utf-8"},
	"update.ps1": {"update.ps1", "homealias-update.ps1", "text/plain; charset=utf-8"},
}

// Handler responde GET /client/{file}. Público: contém apenas o modelo, sem segredos.
func Handler(file func(*http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f, ok := served[file(r)]
		if !ok {
			http.NotFound(w, r)
			return
		}
		data, err := files.ReadFile(f.src)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", f.ctype)
		w.Header().Set("Content-Disposition", `attachment; filename="`+f.name+`"`)
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(data)
	}
}
