package tui

import (
	"os/exec"

	"go_spanish/internal/store"
)

// Text-to-speech components
func speak(word string, locale store.Locale) error {
	voice := "David" // Microsoft English voice
	if locale == store.EsToEn {
		voice = "Sabina" // Microsoft Spanish voice
	}

	cmd := exec.Command(
		"powershell",
		"-Command",
		`Add-Type -AssemblyName System.Speech; $s = New-Object System.Speech.Synthesis.SpeechSynthesizer; $s.SelectVoice("Microsoft `+voice+` Desktop"); $s.Speak("`+word+`");`,
	)

	err := cmd.Run()

	return err
}
