package main



import (

	"time"
	"fmt"

	"github.com/sharat910/edrint/common"
	"github.com/sharat910/edrint/events"
	"github.com/sharat910/edrint/telemetry"
)

type MBPSComputer struct {
	telemetry.BaseFlowTelemetry
	FirstPacketTS 	time.Time
	PrevIdx int
	PrevPacketTS 	time.Time
	Count			int
	IntervalMS		int
	ListMBPS			[]int
	Media			int

}

func NewMBPSComputer(intervalMS int) telemetry.TeleGen {
	return func() telemetry.Telemetry {
		//var t JitterComputer
		//t.intervalMS = intervalMS

		var init []int
		init = append(init, 0)
		return &MBPSComputer{
			IntervalMS: intervalMS,
			ListMBPS: init,
			Media: 0,
		}
	}
}

func (l *MBPSComputer) Name() string {
	return "MBPS_computer"
}


// media detection
// content type field
func (l *MBPSComputer) OnFlowPacket(p common.Packet) {
	if (p.IsOutbound == false) {

		if l.FirstPacketTS.IsZero() {
			l.FirstPacketTS = p.Timestamp
			l.PrevPacketTS = p.Timestamp
			return
		} 

		idx, _ := telemetry.GetIndex(l.FirstPacketTS, p.Timestamp, l.IntervalMS)

		if idx == l.PrevIdx {
			// append to delay

			l.Count++


			//l.setStddev()
			//fmt.Println("delay", l.delayList)
			//fmt.Println(sd)
			
		} else if idx > l.PrevIdx {
			
			for l.PrevIdx < idx {
				l.ListMBPS = append(l.ListMBPS, l.Count)
				l.Count = 0
				l.PrevIdx +=1 

			}
		}
		//fmt.Println("jitter delay", p.Timestamp, l.prevPacketTS)
		
		l.PrevPacketTS = p.Timestamp
		l.DetectMedia()
		

	}
}

func (l *MBPSComputer) DetectMedia() {
	// moving average -> 5 secs
	var wind = 5

	var wind_arr []int


	if (len(l.ListMBPS) < wind) {
		wind_arr = l.ListMBPS
	} else {
		for i:=0; i < wind; i++ {
			wind_arr = append(wind_arr, l.ListMBPS[len(l.ListMBPS) - i- 1])
		}
	}

	avg := l.Sum(l.ListMBPS) / len(l.ListMBPS)

	/*
		1: audio
		2: video
		3: content
	*/

	if (avg > 45 && avg < 55) {
		l.Media = 1					
	} else if (avg > 100) {
		l.Media = 2
	}




}

func (l *MBPSComputer) Sum(array []int) int {
	var sum = 0;
	for i:=0; i < len(array); i++ {
		sum += array[i]
	}
	return sum

}


func (l *MBPSComputer) Pubs() []events.Topic {
	return []events.Topic{"zoom_MBPS"}
}

func (l *MBPSComputer) Teardown() {

	fmt.Println("lenght of MBPS, ", len(l.ListMBPS))

	l.Publish("zoom_MBPS", struct {
		Header	common.FiveTuple
		ListMBPS	[]int	
		Media 	int
		
	}{
		l.GetHeader(),
		l.ListMBPS,
		l.Media,
	})
}
