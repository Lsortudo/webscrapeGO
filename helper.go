package main

import "strings"

// Função auxiliar para criar nomes de arquivo seguros
func sanitizeFilename(name string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == ' ':
			return '_'
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'):
			return r
		default:
			return -1
		}
	}, name)
}
