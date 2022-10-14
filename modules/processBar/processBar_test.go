package processBar

import (
	"testing"
	"time"
)

func TestProcessBar(t *testing.T) {
	total := float64(100000)
	pb := NewBar(total)

	pb.Start()
	for {
		select {
		case <-time.Tick(time.Second):
			pb.Process(9999)
			pb.Render()
			if pb.current == pb.total {
				break
			}
		}
	}
}
