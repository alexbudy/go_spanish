package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"go_spanish/internal/store"
)

type profileSetting int

const (
	settingTTS profileSetting = iota
	settingNumQuestions
	settingNumAnswerOptions
	settingSaveCancel
)

type saveCancelSelection int

const (
	saveSelected saveCancelSelection = iota
	cancelSelected
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

	saveCancelSelection saveCancelSelection
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
		psc.selectedSetting = settingSaveCancel

	case settingNumQuestions:
		psc.selectedSetting = settingTTS

	case settingNumAnswerOptions:
		psc.selectedSetting = settingNumQuestions

	case settingSaveCancel:
		psc.selectedSetting = settingNumAnswerOptions
	}
}

func (psc *profileSettingsConfig) down() {
	switch psc.selectedSetting {
	case settingTTS:
		psc.selectedSetting = settingNumQuestions

	case settingNumQuestions:
		psc.selectedSetting = settingNumAnswerOptions

	case settingNumAnswerOptions:
		psc.selectedSetting = settingSaveCancel

	case settingSaveCancel:
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

	b.WriteString("\n\n")

	if psc.selectedSetting == settingSaveCancel {
		if psc.saveCancelSelection == saveSelected {
			b.WriteString(selectedStyle.Render("> SAVE "))
			b.WriteString("   CANCEL")
		} else {
			b.WriteString("  SAVE  ")
			b.WriteString(selectedStyle.Render("> CANCEL"))
		}
	} else {
		b.WriteString("  SAVE    CANCEL")
	}

	return b.String()
}

func (m *Model) buildProfileSettingsMenu() {
	profile, err := m.store.GetProfile(m.ctx, m.profile)
	if err != nil {
		fmt.Errorf("store: check some_new_column: %w", err)
	}
	m.profileSettings = newProfileSettings("Profile settings for "+m.profile, m.profile,
		profile.EnableSpeech, profile.DefaultNumQuestions, profile.DefaultNumAnswers)
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
		} else if m.profileSettings.selectedSetting == settingSaveCancel {
			m.profileSettings.saveCancelSelection = saveSelected
		}
	case "right", "l":
		if m.profileSettings.selectedSetting == settingNumQuestions {
			m.profileSettings.increaseQuestions()
		} else if m.profileSettings.selectedSetting == settingTTS {
			m.profileSettings.toggleEnableTTS()
		} else if m.profileSettings.selectedSetting == settingNumAnswerOptions {
			m.profileSettings.increaseNumAnswers()
		} else if m.profileSettings.selectedSetting == settingSaveCancel {
			m.profileSettings.saveCancelSelection = cancelSelected
		}
	case "tab":
		if m.profileSettings.selectedSetting == settingTTS {
			m.profileSettings.toggleEnableTTS()
		}
	case "enter":
		if m.profileSettings.selectedSetting == settingSaveCancel {
			if m.profileSettings.saveCancelSelection == saveSelected {
				err := m.store.UpdateProfile(m.ctx,
					store.Profile{
						Name: m.profile, EnableSpeech: m.profileSettings.enableTTS,
						DefaultNumQuestions: m.profileSettings.defaultNumQuestions,
						DefaultNumAnswers:   m.profileSettings.defaultNumAnswerOptions,
					})
				if err != nil {
					fmt.Errorf("store: check some_new_column: %w", err)
					return m, nil
				}

				m.updateProfileSuccess = "Profile settings updated successfully"
				m.screen = screenDirectionSelect
			} else {
				m.screen = screenDirectionSelect
			}
		}
	case "esc":
		m.updateProfileSuccess = ""
		m.screen = screenProfileSelect
	}

	return m, nil
}
