package main



import (

	"time"
	"fmt"
	"math"

	"github.com/sharat910/edrint/common"
	"github.com/sharat910/edrint/events"
	"github.com/sharat910/edrint/telemetry"
)

/*
Media:
1: Audio
2: Video
3: Content
*/

type PPSComputer struct {
	telemetry.BaseFlowTelemetry
	FirstPacketTS 	time.Time
	PrevIdx int
	PrevPacketTS 	time.Time
	Count			int
	IntervalMS		int
	ListPPS			[]int
	ListMBPS		[]int
	CountMBPS		int
	Media			int
	MaxPps			int
	MaxLen 			int
	PacketLenList	[]int

	Audio			int
	Video			int
	Content			int
	Inactive		int
	Unknown			int


}

func NewPPSComputer(intervalMS int) telemetry.TeleGen {
	return func() telemetry.Telemetry {
		//var t JitterComputer
		//t.intervalMS = intervalMS

		var init []int
		init = append(init, 0)
		return &PPSComputer{
			IntervalMS: intervalMS,
			ListPPS: init,
			Media: 0,
			MaxLen: 0,
		}
	}
}

func (l *PPSComputer) Name() string {
	return "PPS_computer"
}


// media detection
// content type field
func (l *PPSComputer) OnFlowPacket(p common.Packet) {
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

			l.CountMBPS = l.CountMBPS + (len(p.Payload) * 8)

			if (len(l.PacketLenList) > 5){
				l.PacketLenList = l.PacketLenList[1:]
				l.PacketLenList = append(l.PacketLenList, len(p.Payload))
			}


			if (l.MaxLen < len(p.Payload)) {
				l.MaxLen = len(p.Payload)
			}

			//l.setStddev()
			//fmt.Println("delay", l.delayList)
			//fmt.Println(sd)
			
		} else if idx > l.PrevIdx {

			l.DetectMedia(l.MaxLen, len(p.Payload), l.Count, l.computeStdDev(l.ListPPS))
			l.SetMedia()
			

			
			for l.PrevIdx < idx {
				l.ListPPS = append(l.ListPPS, l.Count)
				l.ListMBPS = append(l.ListMBPS, l.CountMBPS)
				l.Count = 0
				l.CountMBPS = 0
				l.PrevIdx +=1 

			}
		}
		//fmt.Println("jitter delay", p.Timestamp, l.prevPacketTS)
		
		l.PrevPacketTS = p.Timestamp
		

	}
}


func (l* PPSComputer) DetectMedia(max_len int, curr_len int, curr_pps int, std_pps float64) {


		//fmt.Println("Max len: ", max_len, "Curr len: ", curr_len, "Curr pps: ", curr_pps, std_pps, l.Media)
		//fmt.Println("audio: ", l.Audio, "video: ", l.Video, "content: ", l.Content, "unknown: ", l.Unknown, "inactive: ", l.Inactive)


		if (max_len < 700 && curr_pps < 70 && curr_pps > 20) {

			l.Audio+=1 
		}

		if (max_len > 800 && curr_pps > 3 && curr_pps < 30 && std_pps < 10){
			//fmt.Println("Video", "Max len: ", max_len, "Curr len: ", curr_len, "Curr pps: ", curr_pps, std_pps, l.Media)

			l.Video+=1 
		}

		if (max_len > 800 && std_pps < 50 && curr_pps > 3){
			//fmt.Println("Video","Max len: ", max_len, "Curr len: ", curr_len, "Curr pps: ", curr_pps, std_pps, l.Media)

			l.Video+=1
		}

		if (max_len > 800 && std_pps > 50){
			//fmt.Println("Content", "Max len: ", max_len, "Curr len: ", curr_len, "Curr pps: ", curr_pps, std_pps, l.Media)

			l.Content+=1
		}

		if (curr_pps < 5 && curr_len < 120){
			l.Inactive+=1
		}


} 


func (l *PPSComputer) SetMedia() {
	max := 0
	if (l.Audio > max) {
		max = l.Audio
		l.Media = 1
	} 
	
	if (l.Video > max) {
		max = l.Video
		l.Media = 2
	} 
	
	if (l.Content > max) {
		max = l.Content
		l.Media = 3
	}

	if (l.Unknown > max) {
		max = l.Unknown
		l.Media = 4
	}

	if (l.Inactive > max) {
		max = l.Inactive
		l.Media = 5
	}
}


func (l *PPSComputer) computeStdDev(ppsList []int) float64 {
	var sum uint = 0
	var mean, sd float64

	if (len(ppsList) == 0){
		return 0
	}
	for i:=0 ; i < len(ppsList); i++ {
		sum = sum + uint(ppsList[i])
	}

	mean = float64(sum)/float64(len(ppsList))
	for i:=0; i < len(ppsList); i++  {
		sd += math.Pow(float64(ppsList[i]) - mean, 2)
	}

	sd = math.Sqrt(sd/float64(len(ppsList)))

	//find the last packet of second

	return sd
}

func (l *PPSComputer) Sum(array []int) int {
	var sum = 0;
	for i:=0; i < len(array); i++ {
		sum += array[i]
	}
	return sum

}


func (l *PPSComputer) Pubs() []events.Topic {
	return []events.Topic{"zoom_pps"}
}

func (l *PPSComputer) Teardown() {

	fmt.Println("lenght of pps, ", len(l.ListPPS))

	l.Publish("zoom_pps", struct {
		Header	common.FiveTuple
		ListPPS	[]int	
		Media 	int
		ListMBPS []int
		
	}{
		l.GetHeader(),
		l.ListPPS,
		l.Media,
		l.ListMBPS,
	})
}
