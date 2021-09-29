package viewer

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"termdbms/database"
	"termdbms/tuiutil"
)

var (
	HeaderHeight       = 2
	FooterHeight       = 1
	MaxInputLength     int
	HeaderStyle        lipgloss.Style
	FooterStyle        lipgloss.Style
	HeaderDividerStyle lipgloss.Style
	InitialModel       *TuiModel
)

func (m *TuiModel) Data() *UIData {
	if m.QueryData != nil {
		return m.QueryData
	}

	return &m.DefaultData
}

func (m *TuiModel) Table() *TableState {
	if m.QueryResult != nil {
		return m.QueryResult
	}

	return &m.DefaultTable
}

func SetStyles() {
	HeaderStyle = lipgloss.NewStyle()
	FooterStyle = lipgloss.NewStyle()

	HeaderDividerStyle = lipgloss.NewStyle().
		Align(lipgloss.Center)

	if !tuiutil.Ascii {
		HeaderStyle = HeaderStyle.
			Foreground(lipgloss.Color(tuiutil.HeaderTopForeground()))

		FooterStyle = FooterStyle.
			Foreground(lipgloss.Color(tuiutil.FooterForeground()))

		HeaderDividerStyle = HeaderDividerStyle.
			Foreground(lipgloss.Color(tuiutil.HeaderBottom()))
	}
}

// INIT UPDATE AND RENDER

// Init currently doesn't do anything but necessary for interface adherence
func (m TuiModel) Init() tea.Cmd {
	SetStyles()

	return nil
}

// Update is where all commands and whatnot get processed
func (m TuiModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var (
		command  tea.Cmd
		commands []tea.Cmd
	)

	if !m.UI.FormatModeEnabled {
		m.Viewport, _ = m.Viewport.Update(message)
	}

	switch msg := message.(type) {
	case tea.MouseMsg:
		HandleMouseEvents(&m, &msg)
		m.SetViewSlices()
		break
	case tea.WindowSizeMsg:
		event := HandleWindowSizeEvents(&m, &msg)
		if event != nil {
			commands = append(commands, event)
		}
		break
	case tea.KeyMsg:
		str := msg.String()
		if m.UI.ShowClipboard {
			state := m.ClipboardList.FilterState()
			if (str == "q" || str == "esc" || str == "enter") && state == list.Unfiltered {
				switch str {
				case "enter":
					i, ok := m.ClipboardList.SelectedItem().(SQLSnippet)
					if ok {
						ExitToDefaultView(&m)
						CreatePopulatedBuffer(&m, nil, i.Query)
						m.UI.SQLEdit = true
					}
					break
				default:
					ExitToDefaultView(&m)
				}
				return m, nil
			}
			
			m.ClipboardList, command = m.ClipboardList.Update(msg)
			break
		}
		// when fullscreen selection viewing is in session, don't allow UI manipulation other than quit or exit
		s := msg.String()
		invalidRenderCommand := m.UI.RenderSelection &&
			s != "esc" &&
			s != "ctrl+c" &&
			s != "q" &&
			s != "p" &&
			s != "m" &&
			s != "n"
		if invalidRenderCommand {
			break
		}

		if s == "ctrl+c" || (s == "q" && (!m.UI.EditModeEnabled && !m.UI.FormatModeEnabled)) {
			return m, tea.Quit
		}

		event := HandleKeyboardEvents(&m, &msg)
		if event != nil {
			commands = append(commands, event)
		}
		if !m.UI.EditModeEnabled && m.Ready {
			m.SetViewSlices()
			if m.UI.FormatModeEnabled {
				MoveCursorWithinBounds(&m)
			}
		}

		break
	case error:
		return m, nil
	}

	if m.Viewport.HighPerformanceRendering {
		commands = append(commands, command)
	}

	return m, tea.Batch(commands...)
}

// View is where all rendering happens
func (m TuiModel) View() string {
	if !m.Ready || m.Viewport.Width == 0 {
		return "\n\tInitializing..."
	}

	// this ensures that all 3 parts can be worked on concurrently(ish)
	done := make(chan bool, 3)

	var footer, header, content string

	// body
	go func(c *string) {
		*c = AssembleTable(&m)
		done <- true
	}(&content)

	// header
	go func(h *string) {
		if m.UI.ShowClipboard {
			done <- true
			return
		}

		var (
			builder []string
		)

		style := m.GetBaseStyle()

		if !tuiutil.Ascii {
			// for column headers
			style = style.Foreground(lipgloss.Color(tuiutil.HeaderForeground())).
				BorderBackground(lipgloss.Color(tuiutil.HeaderBorderBackground())).
				Background(lipgloss.Color(tuiutil.HeaderBackground()))
		}
		headers := m.Data().TableHeadersSlice
		for i, d := range headers { // write all headers
			if m.UI.ExpandColumn != -1 && i != m.UI.ExpandColumn {
				continue
			}

			text := " " + TruncateIfApplicable(&m, d)
			builder = append(builder, style.
				Render(text))
		}

		{
			// schema name
			var headerTop string

			if m.UI.EditModeEnabled || m.UI.FormatModeEnabled {
				headerTop = m.TextInput.Model.View()
				if !m.TextInput.Model.Focused() {
					headerTop = HeaderStyle.Copy().Faint(true).Render(headerTop)
				}
			} else {
				headerTop = fmt.Sprintf(" %s (%d/%d) - %d record(s) + %d column(s)",
					m.GetSchemaName(),
					m.UI.CurrentTable,
					len(m.Data().TableHeaders), // look at how headers get rendered to get accurate record number
					len(m.GetColumnData()),
					len(m.GetHeaders())) // this will need to be refactored when filters get added
				headerTop = HeaderStyle.Render(headerTop)
			}

			headerMid := lipgloss.JoinHorizontal(lipgloss.Left, builder...)
			*h = lipgloss.JoinVertical(lipgloss.Left, headerTop, headerMid)
		}

		done <- true
	}(&header)

	// footer (shows row/col for now)
	go func(f *string) {
		if m.UI.ShowClipboard {
			done <- true
			return
		}
		var (
			row int
			col int
		)
		if !m.UI.FormatModeEnabled { // reason we flip is because it makes more sense to store things by column for data
			row = m.GetRow() + m.Viewport.YOffset
			col = m.GetColumn() + m.Scroll.ScrollXOffset
		} else { // but for format mode thats just a regular row/col situation
			row = m.Format.CursorX
			col = m.Format.CursorY + m.Viewport.YOffset
		}
		footer := fmt.Sprintf(" %d, %d + %d, %d ", m.MouseData.X, m.MouseData.Y, row, col)
		undoRedoInfo := fmt.Sprintf(" undo(%d) / redo(%d) ", len(m.UndoStack), len(m.RedoStack))
		switch m.Table().Database.(type) {
		case *database.SQLite:
			break
		default:
			undoRedoInfo = ""
			break
		}
		gapSize := m.Viewport.Width - lipgloss.Width(footer) - lipgloss.Width(undoRedoInfo) - 2
		queryResultsFlag := "├"
		if m.QueryData != nil || m.QueryResult != nil {
			queryResultsFlag = "*"
		}
		footer = FooterStyle.Render(undoRedoInfo) + queryResultsFlag + strings.Repeat("─", gapSize) + "┤" + FooterStyle.Render(footer)
		*f = footer

		done <- true
	}(&footer)

	// block until all 3 done
	<-done
	<-done
	<-done

	close(done) // close

	if m.UI.ShowClipboard {
		return content
	}

	return fmt.Sprintf("%s\n%s\n%s", header, content, footer) // render
}
