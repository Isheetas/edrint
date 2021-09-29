package main

import (
	"fmt"
	"path/filepath"
	"github.com/sharat910/edrint/telemetry"
	"github.com/sharat910/edrint/events"
	"github.com/rs/zerolog/log"
	"github.com/sharat910/edrint"
	"github.com/sharat910/edrint/processor"
	"github.com/spf13/viper"
)

func main() {

	SetupConfig()

	edrint.SetupLogging(viper.GetString("log.level"))

	manager := edrint.New()

	manager.RegisterProc(processor.NewFlowProcessor(2))
	fmt.Printf("Hello1\n")

	rules := GetClassificationRules()
	manager.RegisterProc(processor.NewHeaderClassifer(rules))

	// Add name of tf and the function (algorithm) calculating the metric
	teleManager := processor.NewTelemetryManager()
	teleManager.AddTFToClass("zoomtcp", telemetry.NewTCPRetransmit(1000))
	teleManager.AddTFToClass("zoomtcp", telemetry.NewTCPRTT())
	teleManager.AddTFToClass("zoomudp", NewLossComputer(1000))
	//teleManager.AddTFToClass("zoomudp", NewJitterComputer(1000))
	teleManager.AddTFToClass("zoomudp", NewPPSComputer(1000))
	//teleManager.AddTFToClass("amazonprime", telemetry.NewFlowSummary())
	//teleManager.AddTFToClass("amazonprime", telemetry.NewHTTPChunkDetector(100))




	manager.RegisterProc(teleManager)


	// Add process that will be computed, name used for event topic
	manager.RegisterProc(processor.NewDumper(fmt.Sprintf("./files/dumps/%s.json.log",
		filepath.Base(viper.GetString("packets.source"))),
		[]events.Topic{
			events.TELEMETRY_TCP_RETRANSMIT,
			events.FLOW_ATTACH_TELEMETRY,
			events.TELEMETRY_TCP_RTT,
			"zoom_loss",
			//"zoom_jitter",
			"zoom_pps",
		}))
	fmt.Printf("registered process\n")

	err := manager.InitProcessors()
	if err != nil {
		log.Fatal().Err(err).Msg("init error")
	}

	fmt.Printf("registered process nil\n")

	err = manager.Run(edrint.ParserConfig{
		CapMode:    edrint.PCAPFILE,
		CapSource:  viper.GetString("packets.source"),
		DirMode:    edrint.CLIENT_IP,
		DirMatches: viper.GetStringSlice("packets.direction.client_ips"),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("some error occurred")
	}

	fmt.Printf("ENd\n")

	
}

func GetClassificationRules() map[string]map[string]string {
	rules := make(map[string]map[string]string)
	config := viper.GetStringMap(fmt.Sprintf("processors.header_classifier.classes"))
	for class := range config {
		rules[class] = viper.GetStringMapString(fmt.Sprintf("processors.header_classifier.classes.%s", class))
	}
	return rules
}
