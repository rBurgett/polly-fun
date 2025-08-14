package polly

import (
	"context"
	"io"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	"github.com/aws/aws-sdk-go-v2/service/polly/types"
)

func NewClient(ctx context.Context) (*polly.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return polly.NewFromConfig(cfg), nil
}

func ListVoices(ctx context.Context, cli *polly.Client, languages ...string) ([]types.Voice, error) {
	var voices []types.Voice
	for _, l := range languages {
		newVoices, err := listVoicesByLanguage(ctx, cli, l)
		if err != nil {
			return nil, err
		}

		voices = append(voices, newVoices...)
	}

	// Sort for stable output
	sort.Slice(voices, func(i, j int) bool {
		if voices[i].LanguageCode != voices[j].LanguageCode {
			return voices[i].LanguageCode < voices[j].LanguageCode
		}
		return aws.ToString(voices[i].Name) < aws.ToString(voices[j].Name)
	})

	return voices, nil
}

func listVoicesByLanguage(ctx context.Context, cli *polly.Client, language string) ([]types.Voice, error) {
	var voices []types.Voice
	var next *string
	for {
		out, err := cli.DescribeVoices(ctx, &polly.DescribeVoicesInput{
			LanguageCode: types.LanguageCode(language),
			NextToken:    next,
		})
		if err != nil {
			return nil, err
		}
		voices = append(voices, out.Voices...)
		if out.NextToken == nil || *out.NextToken == "" {
			break
		}
		next = out.NextToken
	}

	return voices, nil
}

func Synthesize(ctx context.Context, cli *polly.Client, voiceID, engine, text, textType, format, outfile string, sampleRate string) error {
	input := &polly.SynthesizeSpeechInput{
		OutputFormat: types.OutputFormat(format), // "mp3", "ogg_vorbis", or "pcm"
		Text:         aws.String(text),
		VoiceId:      types.VoiceId(voiceID),
	}
	if textType != "" {
		input.TextType = types.TextType(textType) // "text" or "ssml"
	}
	if engine != "" {
		input.Engine = types.Engine(engine) // "neural" or "standard"
	}
	if sampleRate != "" {
		input.SampleRate = aws.String(sampleRate) // e.g., "22050" for MP3, "24000" for Neural MP3
	}

	out, err := cli.SynthesizeSpeech(ctx, input)
	if err != nil {
		return err
	}
	defer out.AudioStream.Close()

	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, out.AudioStream)
	return err
}
