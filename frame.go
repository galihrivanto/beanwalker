package main

import (
	"time"

	"github.com/kr/beanstalk"
	"github.com/nsf/termbox-go"
)

const (
	titleLine  = "Beanstalkd Stats Monitor"
	statusLine = "[TAB] Rotate Focus  [ESC] Exit  [\u2190 \u2192 \u2191 \u2193] Scroll"

	infoColor = termbox.ColorYellow | termbox.AttrBold
)

type mainFrame struct {
	c               *beanstalk.Conn
	statEvt         chan struct{}
	tubesStatsGrid  *ScrollableGrid
	systemStatsGrid *ScrollableGrid
	controls        []Control
	focusIndex      int
}

func (m *mainFrame) getSystemStats() [][]string {
	// list tubes
	stats, err := m.c.Stats()
	if err != nil {
		return nil
	}

	data := [][]string{}

	row := []string{}
	// get headers as reference
	for _, col := range m.systemStatsGrid.Columns {
		value, _ := stats[col.Name]
		row = append(row, value)
	}

	data = append(data, row)

	return data
}

func (m *mainFrame) getTubeStats() [][]string {
	// list tubes
	tubes, err := m.c.ListTubes()
	if err != nil {
		return nil
	}

	data := [][]string{}

	for _, tubeName := range tubes {
		tube := &beanstalk.Tube{m.c, tubeName}
		stats, err := tube.Stats()
		if err != nil {
			return nil
		}
		row := []string{}

		// get headers as reference
		for _, col := range m.tubesStatsGrid.Columns {
			value, _ := stats[col.Name]
			row = append(row, value)
		}
		data = append(data, row)
	}

	return data
}

func (m *mainFrame) pollStats() {
	m.statEvt = make(chan struct{})
	go func() {

		defer close(m.statEvt)

		for {
			<-time.After(time.Second)
			sysData := m.getSystemStats()
			if sysData != nil {
				m.systemStatsGrid.UpdateData(sysData)
			}

			tubeData := m.getTubeStats()
			if tubeData != nil {
				m.tubesStatsGrid.UpdateData(tubeData)
			}

			m.statEvt <- struct{}{}
		}
	}()
}

func (m *mainFrame) onPaint() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	w, h := termbox.Size()
	m.systemStatsGrid.Resize(2, 3, w-4, 5)
	m.tubesStatsGrid.Resize(2, 8, w-4, h-10)

	writeText(3, 1, infoColor, termbox.ColorDefault, titleLine)
	for i := 1; i < w; i++ {
		termbox.SetCell(i, 2, '\u2500', termbox.ColorDefault, termbox.ColorDefault)
	}
	writeText(3, h-1, infoColor, termbox.ColorDefault, statusLine)
}

func (m *mainFrame) refresh() {
	m.onPaint()
	termbox.Flush()
}

func (m *mainFrame) connect() {
	c, err := beanstalk.Dial("tcp", "127.0.0.1:11300")
	if err != nil {
		return
	}
	m.c = c
}

func (m *mainFrame) disconnect() {
	if m.c != nil {
		m.c.Close()
	}
}

func (m *mainFrame) dispatchEvent(ev termbox.Event) bool {
	for _, c := range m.controls {
		if c.Focused() && c.HandleEvent(ev) {
			c.Refresh()
			return true
		}
	}
	return false
}

func (m *mainFrame) rotateFocus() {
	for _, c := range m.controls {
		c.SetFocus(false)
	}
	m.focusIndex++
	if m.focusIndex > len(m.controls)-1 {
		m.focusIndex = 0
	}
	m.controls[m.focusIndex].SetFocus(true)
}

func (m *mainFrame) startLoop() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	m.connect()
	defer m.disconnect()

	termbox.SetInputMode(termbox.InputEsc)
	m.refresh()

	evt := make(chan termbox.Event)
	defer close(evt)

	go func() {
		for {
			evt <- termbox.PollEvent()
		}
	}()

	m.pollStats()

	m.refresh()

mainloop:
	for {
		select {
		case ev := <-evt:

			if m.dispatchEvent(ev) {
				m.refresh()
			}

			switch ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyEsc:
					break mainloop

				case termbox.KeyTab:
					m.rotateFocus()
					m.refresh()

				}
			case termbox.EventError:
				panic(ev.Err)

			case termbox.EventResize:
				m.refresh()
			}

		case <-m.statEvt:
			termbox.Flush()
		}
	}
}

func (m *mainFrame) show() {
	if m.controls == nil {
		m.focusIndex = 0
		m.controls = []Control{}

		// system stats
		m.systemStatsGrid = &ScrollableGrid{
			X:       1,
			Y:       2,
			Title:   "System Stats",
			visible: true,
			Columns: []GridColumn{
				{"hostname", AlignLeft, 20},
				{"current-jobs-urgent", AlignLeft, 20},
				{"current-jobs-ready", AlignRight, 21},
				{"current-jobs-reserved", AlignRight, 21},
				{"current-jobs-delayed", AlignRight, 21},
				{"current-jobs-buried", AlignRight, 21},
				{"current-jobs-urgent", AlignRight, 20},
				{"current-jobs-ready", AlignRight, 20},
				{"current-jobs-reserved", AlignRight, 20},
				{"current-jobs-delayed", AlignRight, 20},
				{"current-jobs-buried", AlignRight, 20},
				{"cmd-put", AlignRight, 9},
				{"cmd-peek", AlignRight, 10},
				{"cmd-peek-ready", AlignRight, 16},
				{"cmd-peek-delayed", AlignRight, 18},
				{"cmd-peek-buried", AlignRight, 17},
				{"cmd-reserve", AlignRight, 13},
				{"cmd-use", AlignRight, 9},
				{"cmd-watch", AlignRight, 11},
				{"cmd-ignore", AlignRight, 12},
				{"cmd-delete", AlignRight, 12},
				{"cmd-release", AlignRight, 13},
				{"cmd-bury", AlignRight, 10},
				{"cmd-kick", AlignRight, 10},
				{"cmd-stats", AlignRight, 11},
				{"cmd-stats-job", AlignRight, 15},
				{"cmd-stats-tube", AlignRight, 16},
				{"cmd-list-tubes", AlignRight, 16},
				{"cmd-list-tube-used", AlignRight, 20},
				{"cmd-list-tubes-watched", AlignRight, 24},
				{"cmd-pause-tube", AlignRight, 16},
				{"job-timeouts", AlignRight, 14},
				{"total-jobs", AlignRight, 11},
				{"max-job-size", AlignRight, 13},
				{"current-tubes", AlignRight, 14},
				{"current-connections", AlignRight, 21},
				{"current-producers", AlignRight, 19},
				{"current-workers", AlignRight, 17},
				{"current-waiting", AlignRight, 17},
				{"total-connections", AlignRight, 19},
				{"pid", AlignRight, 10},
				{"version", AlignRight, 10},
				{"rusage-utime", AlignRight, 14},
				{"rusage-stime", AlignRight, 14},
				{"uptime", AlignRight, 10},
				{"binlog-oldest-index", AlignRight, 21},
				{"binlog-current-index", AlignRight, 22},
				{"binlog-max-size", AlignRight, 17},
				{"binlog-records-written", AlignRight, 24},
				{"binlog-records-migrated", AlignRight, 25},
				{"id", AlignRight, 20},
			},
		}
		m.systemStatsGrid.reset()
		m.controls = append(m.controls, m.systemStatsGrid)

		m.tubesStatsGrid = &ScrollableGrid{
			X:         1,
			Y:         2,
			VScroller: true,
			Title:     "Tubes Stats",
			visible:   true,
			Columns: []GridColumn{
				{"name", AlignLeft, 20},
				{"current-jobs-urgent", AlignRight, 21},
				{"current-jobs-ready", AlignRight, 21},
				{"current-jobs-reserved", AlignRight, 21},
				{"current-jobs-delayed", AlignRight, 21},
				{"current-jobs-buried", AlignRight, 21},
				{"total-jobs", AlignRight, 12},
				{"current-using", AlignRight, 15},
				{"current-waiting", AlignRight, 17},
				{"current-watching", AlignRight, 18},
				{"pause", AlignRight, 7},
				{"cmd-delete", AlignRight, 11},
				{"cmd-pause-tube", AlignRight, 16},
				{"pause-time-left", AlignRight, 17},
			},
		}
		m.tubesStatsGrid.reset()
		m.controls = append(m.controls, m.tubesStatsGrid)

		m.tubesStatsGrid.SetCustomDrawFunc(func(index int, col, value string) (termbox.Attribute, termbox.Attribute) {
			fg := FGColor
			bg := BGColor
			if index%2 == 0 {
				// striped rows
			}

			switch col {
			case "current-jobs-delayed":
				fg = termbox.ColorYellow

			case "current-jobs-buried":
				fg = termbox.ColorRed

			case "current-jobs-ready":
				if value != "0" {
					fg = termbox.ColorGreen
				}

			}

			return fg, bg
		})

		if len(m.controls) > 0 {
			m.controls[m.focusIndex].SetFocus(true)
		}

		// testing
		/*
			m.tubesStatsGrid.UpdateData([][]string{
				[]string{"tube 1", "120", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 2", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 3", "0", "230", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "330"},
				[]string{"tube 4", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "111", "0"},
				[]string{"tube 5", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 6", "120", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 7", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 8", "0", "230", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "330"},
				[]string{"tube 9", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "111", "0"},
				[]string{"tube 10", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 11", "120", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 12", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 13", "0", "230", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "330"},
				[]string{"tube 14", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "111", "0"},
				[]string{"tube 15", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 16", "120", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 17", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 18", "0", "230", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "330"},
				[]string{"tube 19", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "111", "0"},
				[]string{"tube 20", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 21", "120", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 22", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 23", "0", "230", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "330"},
				[]string{"tube 24", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "111", "0"},
				[]string{"tube 25", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 26", "120", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 27", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 28", "0", "230", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "330"},
				[]string{"tube 29", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "111", "0"},
				[]string{"tube 30", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 31", "120", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 32", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 33", "0", "230", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "330"},
				[]string{"tube 34", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "111", "0"},
				[]string{"tube 35", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 36", "120", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 37", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
				[]string{"tube 38", "0", "230", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "330"},
				[]string{"tube 39", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "111", "0"},
				[]string{"tube 40", "0", "0", "1", "0", "1", "1", "0", "0", "1", "0", "1", "1", "0"},
			})
		*/
	}
	m.startLoop()
}