package tui

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type profileSetting int

const (
	settingTTS profileSetting = iota
	settingNumQuestions
	settingNumAnswerOptions
)

type profileSettingsConfig struct {
	title                   string
	profile                 string // unique profile name
	enableTTS               bool   // allow text-to-speech for this profile
	defaultNumQuestions     int    // how many questions to ask for quiz, default value
	defaultNumAnswerOptions int    // how many answers to show for question, default

	selectedSetting profileSetting

	settingsUpdateMsg string // message to show when settings updated
	settingsUpdateErr string // message to show when settings update fails
	// TODO - show SAVE and CANCEL
}

func newProfileSettings(title string, profile string, allowTTS bool, defaultNumQuestions int, defaultNumAnswerOptions int) profileSettingsConfig {
	return profileSettingsConfig{
		title: title, profile: profile, enableTTS: allowTTS,
		defaultNumQuestions:     defaultNumQuestions,
		defaultNumAnswerOptions: defaultNumAnswerOptions,
		selectedSetting:         settingTTS,
	}
}

func (psc *profileSettingsConfig) toggleEnableTTS() {
	psc.enableTTS = !psc.enableTTS
}

func (psc *profileSettingsConfig) up() {
	switch psc.selectedSetting {
	case settingTTS:
		psc.selectedSetting = settingNumAnswerOptions

	case settingNumQuestions:
		psc.selectedSetting = settingTTS

	case settingNumAnswerOptions:
		psc.selectedSetting = settingNumQuestions
	}
}

func (psc *profileSettingsConfig) down() {
	switch psc.selectedSetting {
	case settingTTS:
		psc.selectedSetting = settingNumQuestions

	case settingNumQuestions:
		psc.selectedSetting = settingNumAnswerOptions

	case settingNumAnswerOptions:
		psc.selectedSetting = settingTTS
	}
}

func (psc *profileSettingsConfig) increaseQuestions() {
	if psc.defaultNumQuestions < 20 {
		psc.defaultNumQuestions++
	}
}

func (psc *profileSettingsConfig) decreaseQuestions() {
	if psc.defaultNumQuestions > 1 {
		psc.defaultNumQuestions--
	}
}

func (psc *profileSettingsConfig) increaseNumAnswers() {
	if psc.defaultNumAnswerOptions < 6 {
		psc.defaultNumAnswerOptions++
	}
}

func (psc *profileSettingsConfig) decreaseNumAnswers() {
	if psc.defaultNumAnswerOptions > 1 {
		psc.defaultNumAnswerOptions--
	}
}

func (psc profileSettingsConfig) view() string {
	var b strings.Builder

	b.WriteString(promptStyle.Render(psc.title))
	b.WriteString("\n\n")

	if psc.selectedSetting == settingTTS {
		b.WriteString(selectedStyle.Render("> Allow TTS: "))
	} else {
		b.WriteString("  Allow TTS: ")
	}
	if psc.enableTTS {
		b.WriteString(ttsSelectionStyle.Render("YES"))
	} else {
		b.WriteString(ttsSelectionStyle.Render("NO"))
	}

	b.WriteString("\n")

	if psc.selectedSetting == settingNumQuestions {
		b.WriteString(selectedStyle.Render("> Number of Questions: "))
	} else {
		b.WriteString("  Number of Questions: ")
	}
	b.WriteString(ttsSelectionStyle.Render(strconv.Itoa(psc.defaultNumQuestions)))

	b.WriteString("\n")

	if psc.selectedSetting == settingNumAnswerOptions {
		b.WriteString(selectedStyle.Render("> Number of Answers Choice: "))
	} else {
		b.WriteString("  Number of Answers Choice: ")
	}
	b.WriteString(ttsSelectionStyle.Render(strconv.Itoa(psc.defaultNumAnswerOptions)))

	return b.String()
}

func (m *Model) buildProfileSettingsMenu() {
	m.profileSettings = newProfileSettings("Profile settings for "+m.profile, m.profile, true, 10, 4)
}

func (m Model) updateProfileSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)

	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "up", "k":
		m.profileSettings.up()
	case "down", "j":
		m.profileSettings.down()
	case "left", "h":
		if m.profileSettings.selectedSetting == settingNumQuestions {
			m.profileSettings.decreaseQuestions()
		} else if m.profileSettings.selectedSetting == settingTTS {
			m.profileSettings.toggleEnableTTS()
		} else if m.profileSettings.selectedSetting == settingNumAnswerOptions {
			m.profileSettings.decreaseNumAnswers()
		}
	case "right", "l":
		if m.profileSettings.selectedSetting == settingNumQuestions {
			m.profileSettings.increaseQuestions()
		} else if m.profileSettings.selectedSetting == settingTTS {
			m.profileSettings.toggleEnableTTS()
		} else if m.profileSettings.selectedSetting == settingNumAnswerOptions {
			m.profileSettings.increaseNumAnswers()
		}
	case "tab":
		if m.profileSettings.selectedSetting == settingTTS {
			m.profileSettings.toggleEnableTTS()
		}
	case "esc":
		m.screen = screenProfileSelect
	}

	return m, nil
}
