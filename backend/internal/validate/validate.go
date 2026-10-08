// Package validate concentra as regras de entrada dos formulários do painel.
// O frontend espelha estas regras em src/composables/validation.ts; o backend
// nunca confia só no cliente.
package validate

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	MaxName     = 80
	MaxEmail    = 254
	MaxPassword = 128
	MinPassword = 12
	MinAPIToken = 20
	MaxAPIToken = 256
	MaxFQDN     = 253
)

var (
	dnsLabel   = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)
	telegramID = regexp.MustCompile(`^(-?[0-9]{5,20}|@[A-Za-z][A-Za-z0-9_]{4,31})$`)
	tokenChars = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	uuidRe     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// Error é um erro de validação com mensagem pronta para o usuário (pt-BR).
type Error struct{ Msg string }

func (e *Error) Error() string { return e.Msg }

func fail(msg string) error { return &Error{Msg: msg} }

func hasControl(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

// Name valida um rótulo livre (nome de conexão, token, canal, pessoa).
func Name(field, s string) (string, error) {
	s = strings.TrimSpace(s)
	switch {
	case s == "":
		return "", fail(field + " é obrigatório.")
	case utf8.RuneCountInString(s) > MaxName:
		return "", fail(field + " deve ter no máximo 80 caracteres.")
	case hasControl(s):
		return "", fail(field + " contém caracteres inválidos.")
	}
	return s, nil
}

// Email normaliza (trim + minúsculas) e valida um endereço simples, sem nome de exibição.
func Email(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "", fail("E-mail é obrigatório.")
	}
	if len(s) > MaxEmail {
		return "", fail("E-mail deve ter no máximo 254 caracteres.")
	}
	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address != s || !strings.Contains(s[strings.LastIndex(s, "@"):], ".") {
		return "", fail("Informe um e-mail válido, como nome@exemplo.com.")
	}
	return s, nil
}

// Password exige 8 a 128 caracteres.
func Password(s string) error {
	n := utf8.RuneCountInString(s)
	switch {
	case n < MinPassword:
		return fail("A senha precisa ter pelo menos 12 caracteres.")
	case n > MaxPassword:
		return fail("A senha deve ter no máximo 128 caracteres.")
	}
	if strings.TrimSpace(s) == "" {
		return fail("A senha não pode ser só espaços.")
	}
	if distinctRunes(s) < 5 {
		return fail("A senha é repetitiva demais. Use caracteres variados ou uma frase longa.")
	}
	if isCommonPassword(s) {
		return fail("Essa senha é muito comum. Escolha outra, de preferência uma frase longa.")
	}
	return nil
}

func distinctRunes(s string) int {
	seen := map[rune]struct{}{}
	for _, r := range s {
		seen[unicode.ToLower(r)] = struct{}{}
	}
	return len(seen)
}

// commonPasswords cobre bases típicas de dicionário (já com 12+ caracteres) e os
// valores de exemplo do projeto. A comparação ignora caixa e remove separadores.
var commonPasswords = map[string]struct{}{}

func init() {
	for _, p := range []string{
		"changemenow", "changemenow123", "homealias", "homealias123", "homealiasadmin",
		"password1234", "password12345", "passwordpassword", "senha1234567", "senhasenha1234",
		"123456789012", "1234567890123", "123456123456", "qwertyuiop12", "qwertyuiopas",
		"qwerty123456", "abcdefghijkl", "abc123abc123", "administrator", "administrador",
		"admin1234567", "adminadmin123", "letmein12345", "welcome12345", "iloveyou1234",
		"mudar123456", "trocar123456", "senhaforte123", "minhasenha123", "brasil123456",
	} {
		commonPasswords[p] = struct{}{}
	}
}

func isCommonPassword(s string) bool {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	_, bad := commonPasswords[b.String()]
	return bad
}

// APIToken valida o formato do token Cloudflare (sem espaços, tamanho plausível).
func APIToken(s string) (string, error) {
	s = strings.TrimSpace(s)
	// Tolera o que costuma vir junto ao copiar: "Bearer " e aspas.
	if len(s) > 7 && strings.EqualFold(s[:7], "bearer ") {
		s = strings.TrimSpace(s[7:])
	}
	s = strings.Trim(s, "\"'`")
	switch {
	case s == "":
		return "", fail("O token de API é obrigatório.")
	case len(s) < MinAPIToken || len(s) > MaxAPIToken:
		return "", fail("O token de API tem tamanho inválido. Copie-o inteiro da Cloudflare.")
	case strings.IndexFunc(s, unicode.IsSpace) >= 0 || hasControl(s):
		return "", fail("O token de API não pode conter espaços.")
	case !tokenChars.MatchString(s):
		return "", fail("O token de API tem caracteres inválidos. Cole só o token (letras, números, - e _), sem aspas.")
	}
	return s, nil
}

// HostName valida o subdomínio (um ou mais rótulos DNS) ou "@" para a raiz da zona.
func HostName(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "", fail("O subdomínio é obrigatório.")
	}
	if s == "@" {
		return s, nil
	}
	for _, label := range strings.Split(s, ".") {
		if !dnsLabel.MatchString(label) {
			return "", fail("Subdomínio inválido: use letras, números e hífen, sem começar ou terminar com hífen.")
		}
	}
	return s, nil
}

// ZoneName valida o nome de zona (domínio com ao menos dois rótulos).
func ZoneName(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	labels := strings.Split(s, ".")
	if s == "" || len(s) > MaxFQDN || len(labels) < 2 {
		return "", fail("Zona inválida.")
	}
	for _, label := range labels {
		if !dnsLabel.MatchString(label) {
			return "", fail("Zona inválida.")
		}
	}
	return s, nil
}

// FQDNLen confere o tamanho total do nome completo.
func FQDNLen(fqdn string) error {
	if len(fqdn) > MaxFQDN {
		return fail("O nome completo do host passa de 253 caracteres.")
	}
	return nil
}

// TTL aceita 1 (automático na Cloudflare) ou 60 a 86400 segundos; 0 vira 1.
func TTL(ttl int) (int, error) {
	if ttl == 0 {
		return 1, nil
	}
	if ttl == 1 || (ttl >= 60 && ttl <= 86400) {
		return ttl, nil
	}
	return 0, fail("TTL deve ser automático (1) ou entre 60 e 86400 segundos.")
}

// ID confere se o valor é um UUID.
func ID(field, s string) error {
	if !uuidRe.MatchString(s) {
		return fail(field + " inválido.")
	}
	return nil
}

// ChannelDestination valida o destino conforme o tipo do canal de alerta.
func ChannelDestination(kind, dest string) (string, error) {
	dest = strings.TrimSpace(dest)
	switch kind {
	case "email":
		return Email(dest)
	case "telegram":
		if !telegramID.MatchString(dest) {
			return "", fail("Informe o Chat ID numérico (ex.: 123456789) ou @canal.")
		}
		return dest, nil
	}
	return "", fail("Tipo de canal inválido.")
}
