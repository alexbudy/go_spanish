package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
)

const linebreak = "     ---------" // visual line break

// choiceItem is a single selectable entry in a choiceList.
type choiceItem struct {
	label      string
	value      string
	selectable bool // is this item selectable or not
}

func newChoiceItem(label, value string) choiceItem {
	return choiceItem{label: label, value: value, selectable: true}
}

func newSeparatorItem() choiceItem {
	return choiceItem{
		label:      linebreak,
		selectable: false,
	}
}

// choiceList is a small carousel-style menu: up/down (with wraparound) moves
// the cursor, enter selects. It intentionally avoids bubbles/list, which is
// built for large, filterable lists rather than a handful of fixed choices.
type choiceList struct {
	title  string
	items  []choiceItem
	cursor int

	renameIndex int // index of item we are renaming (or -1 if not renaming)
	renameInput *textinput.Model
}

func newChoiceList(title string, items []choiceItem) choiceList {
	return choiceList{title: title, items: items}
}

func (c *choiceList) up() {
	if len(c.items) == 0 {
		return
	}

	for {
		c.cursor--

		if c.cursor < 0 {
			c.cursor = len(c.items) - 1
		}

		if c.items[c.cursor].selectable {
			return
		}
	}
}

func (c *choiceList) down() {
	if len(c.items) == 0 {
		return
	}
	for {
		c.cursor = (c.cursor + 1) % len(c.items)
		if c.items[c.cursor].selectable {
			return
		}
	}
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

	number := 1
	for i, item := range c.items {
		if item.selectable == false { // new line entries should be skipped
			b.WriteString(item.label + "\n")
			continue
		}

		itemIdxToDisplay := strconv.Itoa(number) + ". "
		number++
		if i == c.cursor {
			if i == c.renameIndex && c.renameInput != nil {
				b.WriteString(selectedStyle.Render("> " + itemIdxToDisplay))
				b.WriteString(c.renameInput.View())
			} else {
				b.WriteString(cursorStyle.Render("> "))
				b.WriteString(selectedStyle.Render(itemIdxToDisplay + item.label))
			}
		} else {
			b.WriteString("  " + itemIdxToDisplay)
			b.WriteString(item.label)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func labeledItems(labels ...string) []choiceItem {
	items := make([]choiceItem, len(labels))
	for i, l := range labels {
		items[i] = newChoiceItem(l, l)
	}
	return items
}
