package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/surfaceyu/edge-tts-go/edgeTTS"
)

func readSample() []string {
	b, _ := os.ReadFile("examples/vn/sample.txt")
	return strings.Split(string(b), "\n")
}

func main() {
	args := edgeTTS.Args{
		Voice:      "",
		Rate:       "+20%",
		Pitch:      "+50Hz",
		Volume:     "+200%",
		WriteMedia: "./sample1.mp3",
	}
	start := time.Now()
	tts := edgeTTS.NewTTS(args)
	// edgeTTS.PrintVoices("vi-VN")

	ct := readSample()

	for _, v := range ct {
		if v == "" {
			continue
		}
		speaker, text := parseSpeak(v)
		tts.AddTextWithVoice(text, speaker)
	}

	tts.Speak()
	fmt.Printf("time: %s", time.Since(start))
}

var voicer = map[string]string{
	"f": "vi-VN-HoaiMyNeural",
	"m": "vi-VN-NamMinhNeural",
}

func parseSpeak(str string) (string, string) {
	startIndex := strings.Index(str, "[[")
	endIndex := strings.Index(str, "]]")
	if startIndex == -1 || endIndex == -1 {
		return "", str
	}
	speaker := voicer[str[startIndex+2:endIndex]]
	annotation := str[endIndex+2:]

	return speaker, annotation
}
