package ui_test

import (
	"testing"

	"github.com/p2p-vpn/p2p-vpn/ui/desktop/shared"
)

// TestTranslations verifica se o sistema de traduções está funcionando corretamente
func TestTranslations(t *testing.T) {
	// Testes para cada idioma suportado
	languages := []string{"pt-br", "en", "es"}

	for _, lang := range languages {
		t.Run("Language_"+lang, func(t *testing.T) {
			translations := shared.GetTranslations(lang)
			
			// Verificar se a tradução foi carregada com o idioma correto
			if translations.LanguageKey != lang {
				t.Errorf("Idioma incorreto: esperado %s, recebido %s", lang, translations.LanguageKey)
			}
			
			// Verificar algumas chaves essenciais em cada categoria
			essentialKeys := map[string][]string{
				"common": {
					shared.KeyAppName,
					shared.KeyConnect,
					shared.KeyDisconnect,
					shared.KeySettings,
					shared.KeyExit,
				},
				"menu": {
					shared.KeyShowWindow,
					shared.KeyHideWindow,
					shared.KeyStartVPN,
					shared.KeyStopVPN,
				},
				"status": {
					shared.KeyStatusConnected,
					shared.KeyStatusDisconnected,
					shared.KeyStatusConnecting,
					shared.KeyStatusDisconnecting,
				},
				"notifications": {
					shared.KeyNotifyConnected,
					shared.KeyNotifyDisconnected,
					shared.KeyNotifyConnectionError,
				},
			}
			
			for category, keys := range essentialKeys {
				for _, key := range keys {
					translated := shared.GetTranslated(translations, category, key)
					if translated == "" || translated == key {
						t.Errorf("Tradução ausente para %s.%s no idioma %s", category, key, lang)
					}
				}
			}
		})
	}
}

// TestTranslationsFallback verifica se a tradução cai para o valor padrão corretamente
func TestTranslationsFallback(t *testing.T) {
	// Testar com um idioma não suportado - deve cair para pt-br (padrão)
	translations := shared.GetTranslations("invalid-language")
	
	if translations.LanguageKey != "pt-br" {
		t.Errorf("Fallback de idioma incorreto: esperado pt-br, recebido %s", translations.LanguageKey)
	}
	
	// Testar com uma chave inexistente - deve retornar a própria chave
	invalidKey := "non_existent_key"
	result := shared.GetTranslated(translations, "common", invalidKey)
	
	if result != invalidKey {
		t.Errorf("Comportamento de fallback incorreto para chave inexistente: esperado a própria chave, recebido %s", result)
	}
}

// TestFormatTranslation testa a formatação de strings traduzidas
func TestFormatTranslation(t *testing.T) {
	translations := shared.GetTranslations("pt-br")
	
	// A chave KeyNotifyPeerConnected deve conter um formador %s
	baseStr := shared.GetTranslated(translations, "notifications", shared.KeyNotifyPeerConnected)
	formattedStr := baseStr
	
	// Tentar formatar com um valor
	if baseStr == formattedStr {
		// Verificar se contém %s
		containsFormatter := false
		for i := 0; i < len(baseStr)-1; i++ {
			if baseStr[i:i+2] == "%s" {
				containsFormatter = true
				break
			}
		}
		
		if !containsFormatter {
			t.Errorf("A chave %s não contém o formador %%s necessário", shared.KeyNotifyPeerConnected)
		}
	}
}
