package edgeTTS

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"

	"golang.org/x/term"
)

type EdgeTTS struct {
	communicator *Communicate
	texts        []*CommunicateTextTask
	outCome      io.WriteCloser
}

type Args struct {
	Text           string
	Voice          string
	Pitch          string
	Proxy          string
	Rate           string
	Volume         string
	WordsInCue     float64
	WriteMedia     string
	WriteSubtitles string
}

func isTerminal(file *os.File) bool {
	return term.IsTerminal(int(file.Fd()))
}

func PrintVoices(locale string) {
	voices, err := listVoices()
	if err != nil {
		log.Fatalf("Failed %v\n", err)
		return
	}
	sort.Slice(voices, func(i, j int) bool {
		return voices[i].ShortName < voices[j].ShortName
	})

	
	voiceMap := make(map[string]interface{})
	for _, voice := range voices {
		if locale != "" && voice.Locale != locale {
			continue
		}
		voiceMap[voice.ShortName] = voice.Gender
	}

	jsonData, err := json.Marshal(voiceMap)
	fmt.Println(string(jsonData))
}

func NewTTS(args Args) *EdgeTTS {
	if isTerminal(os.Stdin) && isTerminal(os.Stdout) && args.WriteMedia == "" {
		fmt.Fprintln(os.Stderr, "Warning: TTS output will be written to the terminal. Use --write-media to write to a file.")
		fmt.Fprintln(os.Stderr, "Press Ctrl+C to cancel the operation. Press Enter to continue.")
		_, _ = fmt.Scanln()
	}
	if _, err := os.Stat(args.WriteMedia); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(args.WriteMedia), 0755)
		if err != nil {
			log.Fatalf("Failed to create dir: %v\n", err)
			return nil
		}
	}
	tts := NewCommunicate().WithVoice(args.Voice).WithRate(args.Rate).WithVolume(args.Volume).WithPitch(args.Pitch)
	file, err := os.OpenFile(args.WriteMedia, os.O_WRONLY|os.O_APPEND|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to open file: %v\n", err)
		return nil
	}
	tts.openWs()
	return &EdgeTTS{
		communicator: tts,
		outCome:      file,
		texts:        []*CommunicateTextTask{},
	}
}

func (eTTS *EdgeTTS) task(text string, voice string, pitch string, rate string, volume string) *CommunicateTextTask {
	return &CommunicateTextTask{
		text: text,
		option: CommunicateTextOption{
			voice:  voice,
			pitch:  pitch,
			rate:   rate,
			volume: volume,
		},
	}
}

func (eTTS *EdgeTTS) AddTextDefault(text string) *EdgeTTS {
	eTTS.texts = append(eTTS.texts, eTTS.task(text, eTTS.communicator.option.voice, eTTS.communicator.option.pitch, eTTS.communicator.option.rate, eTTS.communicator.option.volume))
	return eTTS
}

func (eTTS *EdgeTTS) AddTextWithVoice(text string, voice string) *EdgeTTS {
	eTTS.texts = append(eTTS.texts, eTTS.task(text, voice, eTTS.communicator.option.pitch, eTTS.communicator.option.rate, eTTS.communicator.option.volume))
	return eTTS
}

func (eTTS *EdgeTTS) AddText(text string, voice string, pitch string, rate string, volume string) *EdgeTTS {
	eTTS.texts = append(eTTS.texts, eTTS.task(text, voice, pitch, rate, volume))
	return eTTS
}

func (eTTS *EdgeTTS) Speak() {
	defer eTTS.communicator.close()
	defer eTTS.outCome.Close()

	go eTTS.communicator.allocateTask(eTTS.texts)
	eTTS.communicator.createPool()
	for _, text := range eTTS.texts {
		if _, err := eTTS.outCome.Write(text.speechData); err != nil {
			log.Fatalln("Failed to write to file:", err)
		}
	}
}
