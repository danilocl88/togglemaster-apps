package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var validFlagName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

// validateFlagName impede que input do cliente seja utilizado como
// path arbitrario na comunicacao entre os microservicos.
func validateFlagName(flagName string) error {
	if !validFlagName.MatchString(flagName) {
		return fmt.Errorf("flag_name invalido")
	}

	return nil
}

// validateInternalServiceURL valida a configuracao administrativa da URL
// de um servico interno. O cliente HTTP somente pode acessar o hostname
// esperado do microservico ou localhost para desenvolvimento.
func validateInternalServiceURL(
	rawURL string,
	serviceName string,
	expectedPort string,
) (string, error) {

	rawURL = strings.TrimSpace(rawURL)

	if rawURL == "" {
		return "", fmt.Errorf("URL vazia")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("URL invalida: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("scheme nao permitido: %s", parsed.Scheme)
	}

	if parsed.User != nil {
		return "", fmt.Errorf("userinfo nao permitido na URL")
	}

	if parsed.RawQuery != "" {
		return "", fmt.Errorf("query string nao permitida na URL base")
	}

	if parsed.Fragment != "" {
		return "", fmt.Errorf("fragment nao permitido na URL base")
	}

	if parsed.Path != "" && parsed.Path != "/" {
		return "", fmt.Errorf("path nao permitido na URL base")
	}

	host := strings.ToLower(parsed.Hostname())

	allowedHosts := map[string]bool{
		serviceName: true,
		serviceName + ".togglemaster.svc.cluster.local": true,
		"localhost": true,
		"127.0.0.1": true,
		"::1":       true,
	}

	if !allowedHosts[host] {
		return "", fmt.Errorf(
			"hostname nao permitido para %s: %s",
			serviceName,
			host,
		)
	}

	if parsed.Port() != expectedPort {
		return "", fmt.Errorf(
			"porta nao permitida para %s: %s",
			serviceName,
			parsed.Port(),
		)
	}

	parsed.Path = ""
	parsed.RawPath = ""

	return strings.TrimRight(parsed.String(), "/"), nil
}

// buildInternalServiceEndpoint adiciona somente um resource conhecido e
// um flag_name previamente validado a uma URL-base ja validada.
func buildInternalServiceEndpoint(
	baseURL string,
	resource string,
	flagName string,
) (string, error) {

	if err := validateFlagName(flagName); err != nil {
		return "", err
	}

	switch resource {
	case "flags", "rules":
	default:
		return "", fmt.Errorf("resource interno nao permitido")
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("URL base invalida: %w", err)
	}

	endpoint, err := url.JoinPath(
		baseURL,
		resource,
		flagName,
	)
	if err != nil {
		return "", fmt.Errorf("erro construindo endpoint: %w", err)
	}

	finalURL, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("endpoint invalido: %w", err)
	}

	// Defesa adicional: JoinPath jamais pode mudar scheme ou host.
	if !strings.EqualFold(finalURL.Scheme, base.Scheme) ||
		!strings.EqualFold(finalURL.Host, base.Host) {
		return "", fmt.Errorf("endpoint tentou alterar origem do servico")
	}

	return endpoint, nil
}
