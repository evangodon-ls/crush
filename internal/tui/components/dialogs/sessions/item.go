package sessions

import (
	"image/color"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/session"
	"github.com/charmbracelet/crush/internal/tui/components/core/layout"
	"github.com/charmbracelet/crush/internal/tui/styles"
	"github.com/charmbracelet/crush/internal/tui/util"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"
)

type SessionItem interface {
	util.Model
	layout.Focusable
	layout.Sizeable
	FilterValue() string
	MatchIndexes([]int)
	Value() session.Session
	ID() string
}

type sessionItem struct {
	width        int
	session      session.Session
	focus        bool
	matchIndexes []int
	timeAgo      string
	isActive     bool
}

func NewSessionItem(sess session.Session, isActive bool) SessionItem {
	timeAgo := util.FormatTimeAgo(sess.UpdatedAt)
	if sess.UpdatedAt == 0 || sess.UpdatedAt == sess.CreatedAt {
		timeAgo = util.FormatTimeAgo(sess.CreatedAt)
	}

	return &sessionItem{
		session:  sess,
		timeAgo:  timeAgo,
		isActive: isActive,
	}
}

func (s *sessionItem) Init() tea.Cmd {
	return nil
}

func (s *sessionItem) Update(tea.Msg) (tea.Model, tea.Cmd) {
	return s, nil
}

func (s *sessionItem) View() string {
	t := styles.CurrentTheme()

	itemStyle := t.S().Base.Padding(0, 1).Width(s.width)
	innerWidth := s.width - 2

	var timeText string
	if s.isActive {
		timeText = "Active now"
	} else {
		timeText = s.timeAgo
	}

	timeStyle := t.S().Muted
	spacerStyle := lipgloss.NewStyle()
	if s.focus {
		timeStyle = t.S().TextSelected
		itemStyle = itemStyle.Background(t.Primary)
		spacerStyle = spacerStyle.Background(t.Primary)
	}

	timeWidth := ansi.StringWidth(timeText)
	availableTitleWidth := innerWidth - timeWidth - 2

	if availableTitleWidth < 10 {
		availableTitleWidth = innerWidth
		timeText = ""
		timeWidth = 0
	}

	titleStyle := t.S().Text.Width(availableTitleWidth)
	titleMatchStyle := t.S().Text.Underline(true)

	if s.focus {
		titleStyle = t.S().TextSelected.Width(availableTitleWidth)
		titleMatchStyle = t.S().TextSelected.Underline(true)
	}

	title := s.session.Title
	truncatedTitle := s.truncateWithMatches(title, availableTitleWidth)

	titleText := titleStyle.Render(truncatedTitle)
	if len(s.matchIndexes) > 0 {
		var ranges []lipgloss.Range
		for _, rng := range matchedRanges(s.matchIndexes) {
			start, stop := bytePosToVisibleCharPos(truncatedTitle, rng)
			ranges = append(ranges, lipgloss.NewRange(start, stop+1, titleMatchStyle))
		}
		titleText = lipgloss.StyleRanges(titleText, ranges...)
	}

	var content string
	if timeText != "" {
		spacer := spacerStyle.Width(innerWidth - availableTitleWidth - timeWidth).Render("")
		timeRendered := timeStyle.Render(timeText)
		content = lipgloss.JoinHorizontal(lipgloss.Left, titleText, spacer, timeRendered)
	} else {
		content = titleText
	}

	return itemStyle.Render(content)
}

func (s *sessionItem) Blur() tea.Cmd {
	s.focus = false
	return nil
}

func (s *sessionItem) Focus() tea.Cmd {
	s.focus = true
	return nil
}

func (s *sessionItem) GetSize() (int, int) {
	return s.width, 1
}

func (s *sessionItem) IsFocused() bool {
	return s.focus
}

func (s *sessionItem) SetSize(width int, height int) tea.Cmd {
	s.width = width
	return nil
}

func (s *sessionItem) MatchIndexes(indexes []int) {
	s.matchIndexes = indexes
}

func (s *sessionItem) FilterValue() string {
	return s.session.Title
}

func (s *sessionItem) Value() session.Session {
	return s.session
}

func (s *sessionItem) ID() string {
	return s.session.ID
}

func (s *sessionItem) truncateWithMatches(text string, width int) string {
	if width <= 0 {
		return ""
	}

	textLen := ansi.StringWidth(text)
	if textLen <= width {
		return text
	}

	if len(s.matchIndexes) == 0 {
		return ansi.Truncate(text, width, "…")
	}

	lastMatchPos := s.matchIndexes[len(s.matchIndexes)-1]

	lastMatchVisualPos := 0
	bytePos := 0
	gr := uniseg.NewGraphemes(text)
	for bytePos < lastMatchPos && gr.Next() {
		bytePos += len(gr.Str())
		lastMatchVisualPos += max(1, gr.Width())
	}

	ellipsisWidth := 1
	availableWidth := width - ellipsisWidth

	if lastMatchVisualPos < availableWidth {
		return ansi.Truncate(text, width, "…")
	}

	startVisualPos := max(0, lastMatchVisualPos-availableWidth+1)

	startBytePos := 0
	currentVisualPos := 0
	gr = uniseg.NewGraphemes(text)
	for currentVisualPos < startVisualPos && gr.Next() {
		startBytePos += len(gr.Str())
		currentVisualPos += max(1, gr.Width())
	}

	truncatedText := text[startBytePos:]
	truncatedText = ansi.Truncate(truncatedText, availableWidth, "")
	truncatedText = "…" + truncatedText
	return truncatedText
}

func matchedRanges(in []int) [][2]int {
	if len(in) == 0 {
		return [][2]int{}
	}
	current := [2]int{in[0], in[0]}
	if len(in) == 1 {
		return [][2]int{current}
	}
	var out [][2]int
	for i := 1; i < len(in); i++ {
		if in[i] == current[1]+1 {
			current[1] = in[i]
		} else {
			out = append(out, current)
			current = [2]int{in[i], in[i]}
		}
	}
	out = append(out, current)
	return out
}

func bytePosToVisibleCharPos(str string, rng [2]int) (int, int) {
	bytePos, byteStart, byteStop := 0, rng[0], rng[1]
	pos, start, stop := 0, 0, 0
	gr := uniseg.NewGraphemes(str)
	for byteStart > bytePos {
		if !gr.Next() {
			break
		}
		bytePos += len(gr.Str())
		pos += max(1, gr.Width())
	}
	start = pos
	for byteStop > bytePos {
		if !gr.Next() {
			break
		}
		bytePos += len(gr.Str())
		pos += max(1, gr.Width())
	}
	stop = pos
	return start, stop
}

type SessionItemOption func(*sessionItem)

func WithSessionBackgroundColor(c color.Color) SessionItemOption {
	return func(s *sessionItem) {
	}
}
