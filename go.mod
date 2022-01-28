module termdbms

go 1.16

require (
	github.com/atotto/clipboard v0.1.2
	github.com/charmbracelet/bubbles v0.9.0
	github.com/charmbracelet/bubbletea v0.18.0
	github.com/charmbracelet/lipgloss v0.4.0
	github.com/hazelcast/hazelcast-go-client v1.2.0
	github.com/mattn/go-isatty v0.0.14-0.20210829144114-504425e14f74 // indirect
	github.com/mattn/go-runewidth v0.0.13
	github.com/muesli/reflow v0.3.0
	github.com/muesli/termenv v0.9.0
	github.com/sahilm/fuzzy v0.1.0
	golang.org/x/sys v0.0.0-20210902050250-f475640dd07b // indirect
	golang.org/x/tools v0.0.0-20201124115921-2c860bdd6e78 // indirect
)

replace github.com/hazelcast/hazelcast-go-client v1.2.0 => github.com/yuce/hazelcast-go-client v1.1.2-0.20220124142245-1906eb58ac78
