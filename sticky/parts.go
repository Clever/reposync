package sticky

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/fatih/color"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type Part interface {
	register(line Line)
	Render() string
}

type textPart struct {
	text      string
	sprint    func(a ...interface{}) string
	width     int
	listeners []Line
}

type TextPart interface {
	Text(text string, parts ...interface{}) TextPart
	Color(attrs ...color.Attribute) TextPart
	Width(width int) TextPart
	Part
}

func NewTextPart(text string, params ...interface{}) TextPart {
	part := &textPart{
		text:   "",
		width:  0,
		sprint: color.New().SprintFunc(),
	}

	part.Text(text, params...)

	return part
}

func (p *textPart) triggerListeners() {
	for _, listener := range p.listeners {
		listener.render()
	}
}

func (p *textPart) register(line Line) {
	p.listeners = append(p.listeners, line)
}

func (p *textPart) Text(text string, params ...interface{}) TextPart {
	p.text = fmt.Sprintf(text, params...)

	p.triggerListeners()

	return p
}

func (p *textPart) Color(attrs ...color.Attribute) TextPart {
	p.sprint = color.New(attrs...).SprintFunc()

	p.triggerListeners()

	return p
}

// If Width is set to zero or less, the part will grow and shrink with Text
func (p *textPart) Width(width int) TextPart {
	p.width = width

	p.triggerListeners()

	return p
}

func (p *textPart) Render() string {
	var displayText string

	if p.width <= 0 {
		displayText = p.text
	} else if len(p.text) < p.width {
		tmpl := fmt.Sprintf("%% -%ds", p.width)
		displayText = fmt.Sprintf(tmpl, p.text)
	} else {
		displayText = p.text[:p.width]
	}

	return p.sprint(displayText)
}

type spinner []rune

var spinners = []spinner{
	spinner("←↖↑↗→↘↓↙"),
	spinner("▁▃▄▅▆▇█▇▆▅▄▃▁"),
	spinner("▖▘▝▗"),
	spinner("┤┘┴└├┌┬┐"),
	spinner("◢◣◤◥"),
	spinner("◰◳◲◱"),
	spinner("◴◷◶◵"),
	spinner("◐◓◑◒"),
	spinner(".oO@*"),
	spinner("|/-\\"),
	spinner("◡◠⊙⊚⊛⊝⊜◠"),
	spinner("⣾⣽⣻⢿⡿⣟⣯⣷"),
	spinner("⠁⠂⠄⡀⢀⠠⠐⠈"),
	spinner("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"),
	spinner("▉▊▋▌▍▎▏▎▍▌▋▊▉"),
	spinner("■□▪▫"),
	spinner("←↑→↓"),
	spinner("╫╪"),
	spinner("⇐⇖⇑⇗⇒⇘⇓⇙"),
	spinner("⠁⠁⠉⠙⠚⠒⠂⠂⠒⠲⠴⠤⠄⠄⠤⠠⠠⠤⠦⠖⠒⠐⠐⠒⠓⠋⠉⠈⠈"),
	spinner("⠈⠉⠋⠓⠒⠐⠐⠒⠖⠦⠤⠠⠠⠤⠦⠖⠒⠐⠐⠒⠓⠋⠉⠈"),
	spinner("⠁⠉⠙⠚⠒⠂⠂⠒⠲⠴⠤⠄⠄⠤⠴⠲⠒⠂⠂⠒⠚⠙⠉⠁"),
	spinner("⠋⠙⠚⠒⠂⠂⠒⠲⠴⠦⠖⠒⠐⠐⠒⠓⠋"),
	spinner("ｦｧｨｩｪｫｬｭｮｯｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉﾊﾋﾌﾍﾎﾏﾐﾑﾒﾓﾔﾕﾖﾗﾘﾙﾚﾛﾜﾝ"),
	spinner("▁▂▃▄▅▆▇█▉▊▋▌▍▎▏▏▎▍▌▋▊▉█▇▆▅▄▃▂▁"),
	spinner(".oO°Oo."),
	spinner("-+x*"),
	spinner("v<^>"),
}

const (
	Pending int = iota
	Active
	Success
	Fail
)

type statusPart struct {
	status    int
	spinner   spinner
	frame     int
	ticker    *time.Ticker
	listeners []Line
}

type StatusPart interface {
	Pending() StatusPart
	Active() StatusPart
	Success() StatusPart
	Fail() StatusPart
	Part
}

func NewStatusPart() StatusPart {
	return &statusPart{
		status:  Pending,
		spinner: spinners[rand.Intn(len(spinners))],
	}
}

func (p *statusPart) triggerListeners() {
	for _, listener := range p.listeners {
		listener.render()
	}
}

func (p *statusPart) register(line Line) {
	p.listeners = append(p.listeners, line)
}

func (p *statusPart) stopTicker() {
	if p.ticker != nil {
		p.ticker.Stop()
		p.ticker = nil
	}
}

func (p *statusPart) Active() StatusPart {
	if p.status == Active {
		return p
	}

	p.status = Active
	p.frame = -1
	p.ticker = time.NewTicker(time.Millisecond * 300)

	go func() {
		for _ = range p.ticker.C {
			p.triggerListeners()
		}
	}()

	p.triggerListeners()

	return p
}
func (p *statusPart) Pending() StatusPart {
	p.stopTicker()
	p.status = Pending

	p.triggerListeners()

	return p
}
func (p *statusPart) Success() StatusPart {
	p.stopTicker()
	p.status = Success

	p.triggerListeners()

	return p
}
func (p *statusPart) Fail() StatusPart {
	p.stopTicker()
	p.status = Fail

	p.triggerListeners()

	return p
}

func (p *statusPart) Render() string {
	if p.status == Pending {
		return color.New(color.FgYellow).SprintFunc()("?")
	}
	if p.status == Active {
		p.frame += 1
		return string(p.spinner[p.frame%len(p.spinner)])
	}
	if p.status == Success {
		return color.New(color.FgGreen).SprintFunc()("✓")
	}
	if p.status == Fail {
		return color.New(color.FgRed).SprintFunc()("✘")
	}

	log.Fatal("Unknown status")
	return ""
}
