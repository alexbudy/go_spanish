package tui

import (
	"os/exec"
)

// Text-to-speech components
func speakSpanishWord(word string) error {
	cmd := exec.Command(
		"powershell",
		"-Command",
		`Add-Type -AssemblyName System.Speech; $s = New-Object System.Speech.Synthesis.SpeechSynthesizer; $s.SelectVoice("Microsoft Sabina Desktop"); $s.Speak("`+word+`");`,
	)

	err := cmd.Run()

	return err
}
