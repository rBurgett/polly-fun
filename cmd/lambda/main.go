package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/rBurgett/polly-fun/cmd/lambda/internal/polly"
)

func main() {
	ctx := context.Background()

	log.Print("Starting Polly Fun!")

	client, err := polly.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	languages := []string{"en-US", "en-GB", "en-AU", "en-IN", "en-NZ", "en-ZA", "en-IE"}

	voices, err := polly.ListVoices(ctx, client, languages...)
	if err != nil {
		log.Fatal(err)
	}

	var v types.Voice

	for i := range voices {
		if voices[i].Id == "Matthew" {
			v = voices[i]
		}
		fmt.Println(fmt.Sprintf("%v", voices[i]))
	}

	if v.Id == "" {
		log.Fatal("No voice found")
	}

	text := "Hello, Isaac! This is your computer."

	err = polly.Synthesize(ctx, client, string(v.Id), string(v.SupportedEngines[0]), text, "text", "mp3", "/home/ryan/Downloads/test.mp3", "22050")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("All done!")
}
