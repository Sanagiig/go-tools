package processBar

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

type ProcessBar struct {
	mu         sync.Mutex
	symbol     string
	percent    float64
	current    float64
	total      float64
	startTime  time.Time
	blockCount int
	isDone     bool
}

func NewBar(total float64) *ProcessBar {
	return &ProcessBar{
		mu:         sync.Mutex{},
		symbol:     "█",
		percent:    0.0,
		current:    0,
		total:      total,
		blockCount: 50,
	}
}

func (p *ProcessBar) Process(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 数据合法性校验
	cur := math.Max(p.current+float64(n), 0)
	p.current = math.Min(cur, p.total)
	p.Render()
}

func (p *ProcessBar) Start() {
	p.startTime = time.Now()
	p.Render()
}

func (p *ProcessBar) Render() {
	if p.isDone {
		return
	}

	p.percent = float64(p.current/p.total) * 100
	barCount := int(math.Round(p.current / p.total * float64(p.blockCount)))
	barStr := strings.Repeat(p.symbol, barCount) + strings.Repeat(" ", p.blockCount-barCount)

	fmt.Printf("\r当前进度： [%s]  %0.2f%% ( %v/%v ) ", barStr, p.percent, p.current, p.total)
	p.done()
}

func (p *ProcessBar) done() {
	if p.current < p.total {
		return
	}

	endTime := time.Now()
	p.isDone = true
	fmt.Printf("\n本次任务从 %v 到 %v ，共耗时 %s", getFomatTimeStr(p.startTime), getFomatTimeStr(endTime), getDurationStr(p.startTime, endTime))
}

func getDurationStr(start, end time.Time) string {
	var h, m, s int64
	res := bytes.NewBuffer([]byte{})
	H := int64(time.Hour)
	M := int64(time.Minute)
	S := int64(time.Second)
	duration := end.UnixNano() - start.UnixNano()

	switch {
	case duration > H:
		h = duration / H
		duration -= h * H
		res.WriteString(fmt.Sprintf("%v 小时  ", h))
		fallthrough
	case duration > M:
		m = duration / M
		duration -= m * M
		res.WriteString(fmt.Sprintf("%v 分  ", m))
		fallthrough
	case duration > S:
		s = duration / S
		res.WriteString(fmt.Sprintf("%v 秒  ", s))
	default:
		res.WriteString("0秒")
	}

	return res.String()
}

func getFomatTimeStr(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
