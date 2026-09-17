package tui

import (
	"strconv"
	"strings"
)

// choiceItem is a single selectable entry in a choiceList.
type choiceItem struct {
	label string
	value string
}

// choiceList is a small carousel-style menu: up/down (with wraparound) moves
// the cursor, enter selects. It intentionally avoids bubbles/list, which is
// built for large, filterable lists rather than a handful of fixed choices.
type choiceList struct {
	title  string
	items  []choiceItem
	cursor int
}

func newChoiceList(title string, items []choiceItem) choiceList {
	return choiceList{title: title, items: items}
}

func (c *choiceList) up() {
	if len(c.items) == 0 {
		return
	}
	c.cursor--
	if c.cursor < 0 {
		c.cursor = len(c.items) - 1
	}
}

func (c *choiceList) down() {
	if len(c.items) == 0 {
		return
	}
	c.cursor = (c.cursor + 1) % len(c.items)
}

func (c choiceList) selected() choiceItem {
	return c.items[c.cursor]
}

// Pass optional profile name to include in the title, e.g. "Select a training direction for <profile>".
func (c choiceList) view(profile ...string) string {
	var b strings.Builder
	if c.title != "" {
		b.WriteString(promptStyle.Render(c.title))
		if len(profile) > 0 {
			b.WriteString(promptStyle.Render(" for "))
			b.WriteString(promptStyleProfile.Render(strings.Join(profile, "")))
		}

		b.WriteString("\n\n")
	}
	for i, item := range c.items {
		if i == c.cursor {
			b.WriteString(cursorStyle.Render("> "))

			lines := strings.Split(item.label, "\n")
			b.WriteString(selectedStyle.Render(strconv.Itoa(i+1) + ". " + lines[0]))

			// add secondary lines as part of selection, but with a different style so they don't compete with the primary line
			if len(lines) > 1 {
				for _, l := range lines[1:] {
					b.WriteString("\n")
					b.WriteString(selectedStyleSecondary.Render(l))
				}
			}

		} else {
			b.WriteString("  " + strconv.Itoa(i+1) + ". ")
			b.WriteString(item.label)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func labeledItems(labels ...string) []choiceItem {
	items := make([]choiceItem, len(labels))
	for i, l := range labels {
		items[i] = choiceItem{label: l, value: l}
	}
	return items
}
