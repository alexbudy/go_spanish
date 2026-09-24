package tui

import (
	"os/exec"

	"go_spanish/internal/store"
)

// Text-to-speech components
func speak(word string, locale store.Locale) error {
	if locale == store.EsToEn {
		return speakSpanishWord(word)
	}
	return speakEnglishWord(word)
}

func speakSpanishWord(text string) error {
	// Sabina - Mexican Spanish
	// Helena and Laura - Spain Spanish
	script := `
		Add-Type -AssemblyName System.Speech

		$s = New-Object System.Speech.Synthesis.SpeechSynthesizer

		$preferred = @(
			"Microsoft Sabina Desktop",
			"Microsoft Helena Desktop",
			"Microsoft Laura Desktop"
		)

		foreach ($name in $preferred) {
			$voice = $s.GetInstalledVoices() |
				Where-Object { $_.VoiceInfo.Name -eq $name } |
				Select-Object -First 1

			if ($voice) {
				$s.SelectVoice($name)
				break
			}
		}

		$s.Speak("` + text + `")
		`

	cmd := exec.Command(
		"powershell",
		"-Command",
		script,
		text,
	)

	return cmd.Run()
}

func speakEnglishWord(text string) error {
	script := `
		Add-Type -AssemblyName System.Speech

		$s = New-Object System.Speech.Synthesis.SpeechSynthesizer

		$preferred = @(
			"Microsoft David Desktop",
			"Microsoft Zira Desktop"
		)

		foreach ($name in $preferred) {
			$voice = $s.GetInstalledVoices() |
				Where-Object { $_.VoiceInfo.Name -eq $name } |
				Select-Object -First 1

			if ($voice) {
				$s.SelectVoice($name)
				break
			}
		}

		$s.Speak("` + text + `")
		`

	cmd := exec.Command(
		"powershell",
		"-Command",
		script,
		text,
	)

	return cmd.Run()
}
