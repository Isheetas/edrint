package main



import (
	"fmt"
	"math"
	"time"

	"github.com/sharat910/edrint/common"
	"github.com/sharat910/edrint/events"
	"github.com/sharat910/edrint/telemetry"
)

type JitterComputer struct {
	telemetry.BaseFlowTelemetry
	FirstPacketTS time.Time
	PrevIdx int
	PrevPacketTS 	time.Time
	DelayList		[]uint
	AvgJitter		float64
	IntervalMS		int
	ListJitter		[]float64

}

func NewJitterComputer(intervalMS int) telemetry.TeleGen {
	return func() telemetry.Telemetry {
		//var t JitterComputer
		//t.intervalMS = intervalMS

		var init []float64
		init = append(init, 0)
		return &JitterComputer{
			IntervalMS: intervalMS,
			ListJitter: init,
		}
	}
}

func (l *JitterComputer) Name() string {
	return "Jitter_computer"
}


// media detection
// content type field
func (l *JitterComputer) OnFlowPacket(p common.Packet) {
	if (p.IsOutbound == false) {

		if l.FirstPacketTS.IsZero() {
			l.FirstPacketTS = p.Timestamp
			l.PrevPacketTS = p.Timestamp
			return
		} 

		idx, _ := telemetry.GetIndex(l.FirstPacketTS, p.Timestamp, l.IntervalMS)

		if idx == l.PrevIdx {
			// append to delay
			var delay = uint(p.Timestamp.Sub(l.PrevPacketTS)/time.Millisecond)
			l.DelayList = append(l.DelayList, delay)
			//l.setStddev()
			//fmt.Println("delay", l.delayList)
			//fmt.Println(sd)
			
		} else if idx > l.PrevIdx {
			
			for l.PrevIdx < idx {
				var sd = l.computeStdDev(l.DelayList)
				//l.delayList = []
				fmt.Println("Std dev", l.DelayList, sd)
				l.ListJitter = append(l.ListJitter, sd)
				l.DelayList = nil
				l.PrevIdx +=1 
				// compute std and append to jitter array
				// reset delay array l.delaylist = nil
				// l.prevIdx +=1
			}
		}
		//fmt.Println("jitter delay", p.Timestamp, l.prevPacketTS)
		
		l.PrevPacketTS = p.Timestamp
		

	}
}

func (l *JitterComputer) computeStdDev(delayList []uint) float64 {
	var sum uint = 0
	var mean, sd float64

	if (len(delayList) == 0){
		return 0
	}
	for i:=0 ; i < len(delayList); i++ {
		sum = sum + delayList[i]
	}

	mean = float64(sum)/float64(len(delayList))
	for i:=0; i < len(delayList); i++  {
		sd += math.Pow(float64(delayList[i]) - mean, 2)
	}

	sd = math.Sqrt(sd/float64(len(delayList)))

	//find the last packet of second

	return sd
}

func (l *JitterComputer) Pubs() []events.Topic {
	return []events.Topic{"zoom_jitter"}
}

func (l *JitterComputer) Teardown() {
	fmt.Println("Teardown jitter")
	fmt.Println(l.FirstPacketTS)
	fmt.Println(l.ListJitter)
	fmt.Println(l.DelayList)


	/*l.Publish("zoom_jitter", struct {
		Header     common.FiveTuple
		listJitter []float64
		firstPacketTS time.Time
		
	}{
		l.GetHeader(),
		l.listJitter,
		l.firstPacketTS,
	})*/
	l.Publish("zoom_jitter", struct {
		Header     	common.FiveTuple
		IntervalMS 	int
		ListJitter	[]float64	
		
	}{
		l.GetHeader(),
		l.IntervalMS,
		l.ListJitter,
	})
}
