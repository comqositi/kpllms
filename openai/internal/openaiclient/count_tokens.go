package openaiclient

import (
	"fmt"
	"log"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

const (
	_tokenApproximation = 4
)

const (
	_gpt35TurboContextSize   = 4096
	_gpt4VisionPreview       = 128000
	_gpt432KContextSize      = 32768
	_gpt4ContextSize         = 8192
	_textDavinci3ContextSize = 4097
	_textBabbage1ContextSize = 2048
	_textAda1ContextSize     = 2048
	_textCurie1ContextSize   = 2048
	_codeDavinci2ContextSize = 8000
	_codeCushman1ContextSize = 2048
	_textBisonContextSize    = 2048
	_chatBisonContextSize    = 2048
	_defaultContextSize      = 2048
)

// nolint:gochecknoglobals
var modelToContextSize = map[string]int{
	"gpt-3.5-turbo":        _gpt35TurboContextSize,
	"gpt-4-vision-preview": _gpt4VisionPreview,
	"gpt-4-32k":            _gpt432KContextSize,
	"gpt-4":                _gpt4ContextSize,
	"text-davinci-003":     _textDavinci3ContextSize,
	"text-curie-001":       _textCurie1ContextSize,
	"text-babbage-001":     _textBabbage1ContextSize,
	"text-ada-001":         _textBabbage1ContextSize,
	"code-davinci-002":     _codeDavinci2ContextSize,
	"code-cushman-001":     _codeCushman1ContextSize,
}

// ModelContextSize gets the max number of tokens for a language model. If the model
// name isn't recognized the default value 2048 is returned.
func GetModelContextSize(model string) int {
	contextSize, ok := modelToContextSize[model]
	if !ok {
		return _defaultContextSize
	}
	return contextSize
}

// CountTokens gets the number of tokens the text contains.
func CountTokens(model, text string) int {
	e, err := tiktoken.EncodingForModel(model)
	if err != nil {
		e, err = tiktoken.GetEncoding("gpt2")
		if err != nil {
			log.Printf("[WARN] Failed to calculate number of tokens for model, falling back to approximate count")
			return 0
		}
	}
	return len(e.Encode(text, nil, nil))
}

// CalculateMaxTokens calculates the max number of tokens that could be added to a text.
func CalculateMaxTokens(model, text string) int {
	return GetModelContextSize(model) - CountTokens(model, text)
}

func NumTokensFromMessages(messages []*ChatMessage, model string) (numTokens int) {
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		err = fmt.Errorf("encoding for model: %v", err)
		return 0
	}

	var tokensPerMessage, tokensPerName int
	switch model {
	case "gpt-3.5-turbo-0613",
		"gpt-3.5-turbo-16k-0613",
		"gpt-4-0314",
		"gpt-4-32k-0314",
		"gpt-4-0613",
		"gpt-4-32k-0613":
		tokensPerMessage = 3
		tokensPerName = 1
	case "gpt-3.5-turbo-0301":
		tokensPerMessage = 4 // every message follows <|start|>{role/name}\n{content}<|end|>\n
		tokensPerName = -1   // if there's a name, the role is omitted
	default:
		if strings.Contains(model, "gpt-3.5-turbo") {
			log.Println("warning: gpt-3.5-turbo may update over time. Returning num tokens assuming gpt-3.5-turbo-0613.")
			return NumTokensFromMessages(messages, "gpt-3.5-turbo-0613")
		} else if strings.Contains(model, "gpt-4") {
			log.Println("warning: gpt-4 may update over time. Returning num tokens assuming gpt-4-0613.")
			return NumTokensFromMessages(messages, "gpt-4-0613")
		} else {
			err = fmt.Errorf("num_tokens_from_messages() is not implemented for model %s. See https://github.com/openai/openai-python/blob/main/chatml.md for information on how messages are converted to tokens.", model)
			log.Println(err)
			return
		}
	}

	for _, message := range messages {
		content, ok := message.Content.(string)
		if !ok {
			continue
		}
		numTokens += tokensPerMessage
		numTokens += len(tkm.Encode(content, nil, nil))
		numTokens += len(tkm.Encode(message.Role, nil, nil))
		numTokens += len(tkm.Encode(message.Name, nil, nil))
		if message.Name != "" {
			numTokens += tokensPerName
		}
	}
	numTokens += 3 // every reply is primed with <|start|>assistant<|message|>
	return numTokens
}
