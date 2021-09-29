package main



import (
	"fmt"
	"time"
	"encoding/binary"
	"github.com/sharat910/edrint/common"
	"github.com/sharat910/edrint/events"
	"github.com/sharat910/edrint/telemetry"
)

type LossComputer struct {
	telemetry.BaseFlowTelemetry
	FirstPacketTS 	 	time.Time
	PrevIdx			 	int
	PrevPacketTS	 	time.Time
	UpSeqNumExpected 	int
	TotalLoss        	int
	Loss				int
	TotalPacket			int
	ListLoss 			[]float64	
	SeqNum 				[]int
	LostPacketSeq		[]int
	LostSeqMap			map[int]int
	IntervalMS			int
}

func NewLossComputer(intervalMS int) telemetry.TeleGen {
	return func() telemetry.Telemetry {
		var init []float64
		init = append(init, 0)
		return &LossComputer{
			IntervalMS: intervalMS,
			TotalLoss: 0,
			TotalPacket: 0,
			Loss: 0,
			UpSeqNumExpected: -1,
			ListLoss: init,


		}
	}
}

func (l *LossComputer) Name() string {
	return "loss_computer"
}


// media detection
// content type field
func (l *LossComputer) OnFlowPacket(p common.Packet) {
	if (p.IsOutbound == false) {

		if l.FirstPacketTS.IsZero(){
			fmt.Println("first packet")
			l.FirstPacketTS = p.Timestamp
			l.PrevPacketTS = p.Timestamp

			seq := l.ExtractSeqNum(p)

			if (seq != -1){
				l.UpSeqNumExpected = seq+1
			}

			return
		}

		idx, _ := telemetry.GetIndex(l.FirstPacketTS, p.Timestamp, l.IntervalMS)

		//check if packet is valid data
		seq := l.ExtractSeqNum(p)
		if (seq == -1){
			return
		}

		//fmt.Println("Got sequence")

		if idx == l.PrevIdx {
			l.TotalPacket++
			if (l.UpSeqNumExpected == -1){

				
				l.UpSeqNumExpected = seq+1
				return
			}

			if seq == l.UpSeqNumExpected {
				if (l.UpSeqNumExpected == 65535) {
					l.UpSeqNumExpected = 0
				} else {
					l.UpSeqNumExpected = seq+1
				}
				
			} else {
				/*if (l.ifSeqInLostPacket(seq) == true) {
					// remove from lost packet seq list, reduce number of lost packets
					fmt.Println("found sequence in lost packet")
					l.TotalLoss--
					l.LostSeqMap[seq] = 1
					delete(l.LostSeqMap, seq)
	
				} else {
					l.LostPacketSeq = append(l.LostPacketSeq, l.UpSeqNumExpected)
					fmt.Println(l.LostPacketSeq)
					l.UpSeqNumExpected = seq+1
					l.TotalLoss++
					fmt.Println(l.TotalLoss)
				}*/
				for (l.UpSeqNumExpected != seq){

					if (l.UpSeqNumExpected == 65535){
						l.UpSeqNumExpected = 0
					} else {
						l.UpSeqNumExpected += 1
					}
					l.Loss +=1 
					
				}

				
				if (l.UpSeqNumExpected == 65535) {
					l.UpSeqNumExpected = 0
				} else {
					l.UpSeqNumExpected += 1

				}


				//fmt.Println("Seq: ", seq)
				
			} 
			
		} else if idx > l.PrevIdx {
			//fmt.Println("Loss: ", l.Loss, l.TotalPacket)
			for l.PrevIdx < idx {
				ploss := float64(0)

				if (l.Loss > 0 && l.TotalPacket > 0){
					ploss = ((float64(l.Loss)/ float64((l.Loss + l.TotalPacket)))) * 100
					//fmt.Println(ploss)
				} else {
					ploss = float64(0)
				}

				l.ListLoss = append(l.ListLoss, ploss)
				//fmt.Println(l.ListLoss)
				
				l.Loss = 0
				l.TotalPacket = 0
				l.PrevIdx += 1
			}
		}

			
		
	}
}

// use map instead of list, 
// interval packet lost
func (l *LossComputer) ifSeqInLostPacket(seq int) bool {
	
	_, found := l.LostSeqMap[seq]
	return found
}

func (l *LossComputer) ExtractSeqNum(p common.Packet) int {

	if (p.Payload[0] == 05) {
		var seq []byte = p.Payload[1:3]
		data := int(binary.BigEndian.Uint16(seq))
		return data

	} else {
		return -1
	}
	
}



// name used to register proc in main.go
func (l *LossComputer) Pubs() []events.Topic {
	return []events.Topic{"zoom_loss"}
}

// variabled dumped after porcessing
func (l *LossComputer) Teardown() {
	l.Publish("zoom_loss", struct {
		Header    common.FiveTuple
		TotalLoss int
		ListLoss  []float64
		
	}{
		l.GetHeader(),			//these are the attributes that are dumped in the log files in cmd/edrint/files/dumps
		l.TotalLoss,
		l.ListLoss,
	})
}
