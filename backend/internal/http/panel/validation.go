package panel

import (
	"net/http"
	"strings"

	"github.com/homealias/homealias/backend/internal/validate"
)

// maxBody limita o corpo JSON dos formulários do painel.
const maxBody = 16 << 10

func limitBody(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
}

// cloudflareMessage traduz erros comuns da API da Cloudflare para algo acionável.
func cloudflareMessage(err error) string {
	msg := err.Error()
	low := strings.ToLower(msg)
	switch {
	case strings.Contains(low, "invalid request headers"), strings.Contains(low, "invalid api token"), strings.Contains(low, "invalid access token"), strings.Contains(low, "unable to authenticate"):
		return "A Cloudflare não aceitou esse token. Confira se copiou o Token de API inteiro (não a Global API Key nem o ID da conta) e se ele não foi revogado."
	case strings.Contains(low, "authentication error"), strings.Contains(low, "forbidden"):
		return "O token é válido, mas não tem permissão. Ele precisa de Zone · Zone · Read e Zone · DNS · Edit."
	}
	return msg
}

// reject responde 400 com a mensagem de validação (pt-BR) e informa se tratou o erro.
func reject(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if ve, ok := err.(*validate.Error); ok {
		http.Error(w, ve.Msg, http.StatusBadRequest)
		return true
	}
	http.Error(w, "bad request", http.StatusBadRequest)
	return true
}
