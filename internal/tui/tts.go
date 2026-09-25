package tui

import (
	"os"
	"os/exec"
	"strings"

	"go_spanish/internal/store"
)

// Text-to-speech components
func speak(m Model, word string, locale store.Locale) error {
	if prof, err := m.store.GetProfile(m.ctx, m.profile); err == nil && !prof.EnableSpeech {
		return nil // skip speaking if speech not enabled
	}

	preferred := []string{ // english voices
		"Microsoft David Desktop",
		"Microsoft Zira Desktop",
	}

	if locale == store.EsToEn {
		// spanish voices
		preferred = []string{
			"Microsoft Sabina Desktop",
			"Microsoft Helena Desktop",
			"Microsoft Laura Desktop",
		}
	}

	return speakWithVoices(word, preferred)
}

func speakWithVoices(text string, preferred []string) error {
	script := `
		Add-Type -AssemblyName System.Speech

		$s = New-Object System.Speech.Synthesis.SpeechSynthesizer

		foreach ($name in $env:TTS_VOICES -split '\|') {
			$voice = $s.GetInstalledVoices() |
				Where-Object { $_.VoiceInfo.Name -eq $name } |
				Select-Object -First 1

			if ($voice) {
				$s.SelectVoice($name)
				break
			}
		}

		$s.Speak($env:TTS_TEXT)
		`

	cmd := exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-Command",
		script,
	)

	cmd.Env = append(
		os.Environ(),
		"TTS_TEXT="+text,
		"TTS_VOICES="+strings.Join(preferred, "|"),
	)

	return cmd.Run()
}
