module termdbms

go 1.16

require (
	github.com/atotto/clipboard v0.1.2
	github.com/charmbracelet/bubbles v0.9.0
	github.com/charmbracelet/bubbletea v0.18.0
	github.com/charmbracelet/lipgloss v0.4.0
	github.com/hazelcast/hazelcast-go-client v1.1.1
	github.com/mattn/go-isatty v0.0.14-0.20210829144114-504425e14f74 // indirect
	github.com/mattn/go-runewidth v0.0.13
	github.com/muesli/reflow v0.3.0
	github.com/muesli/termenv v0.9.0
	github.com/sahilm/fuzzy v0.1.0
	modernc.org/sqlite v1.13.0
)

replace github.com/hazelcast/hazelcast-go-client v1.1.1 => github.com/yuce/hazelcast-go-client v1.1.2-0.20211126095718-c5e8451306b1
