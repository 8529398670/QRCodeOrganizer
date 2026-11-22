// Package ai is Smart Photo Extraction: one top-down photo of what is going
// into a container, and Claude lists the things in it -- names, a few words
// each, how many, brand, tags, and where each one is in the photo, so the page
// can crop a picture of it. A person checks the list before anything is saved.
//
// The key is handed in by the caller, read from the database, never from the
// environment: the client is built with option.WithoutEnvironmentDefaults, so
// no ANTHROPIC_* variable can change where photos go or whose key pays.
package ai

import (
	context "context"
	base64 "encoding/base64"
	json "encoding/json"
	errors "errors"
	fmt "fmt"
	strings "strings"
	time "time"

	anthropic "github.com/anthropics/anthropic-sdk-go/v2"
	option "github.com/anthropics/anthropic-sdk-go/v2/option"

	models "qrcodeorganizer/server/models"
)

const (
	requestTimeout = 120 * time.Second
	maxItems       = 60
)

var (
	ErrRefused   = errors.New( "Claude would not describe that photo" )
	ErrTruncated = errors.New( "the photo held more than could be listed at once: try a photo of fewer things" )
	ErrUnreadable = errors.New( "Claude's answer could not be read" )
)

// Result is what Claude made of one photo.
type Result struct {
	Summary string                `json:"summary"`
	Items   []models.ProposedItem `json:"items"`
}

// schema is the shape Claude has to answer in (structured outputs).
var schema = map[string]any{
	"type": "object",
	"additionalProperties": false,
	"required": []string{ "summary" , "items" },
	"properties": map[string]any{
		"summary": map[string]any{ "type": "string" , "description": "A few words for the whole collection, usable as the container's name, such as \"Camping cookware\"." },
		"items": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"additionalProperties": false,
				"required": []string{ "name" , "description" , "quantity" , "brand" , "category" , "tags" , "box" },
				"properties": map[string]any{
					"name":        map[string]any{ "type": "string" , "description": "What it is, as someone would search for it: \"Cordless drill\", not \"Black tool\"." },
					"description": map[string]any{ "type": "string" , "description": "One short sentence: color, size, model, condition -- what tells it apart." },
					"quantity":    map[string]any{ "type": "number" , "description": "How many of this exact thing are visible." },
					"brand":       map[string]any{ "type": "string" , "description": "The brand if it can be read on the item, else empty." },
					"category":    map[string]any{ "type": "string" , "description": "One broad category: tools, kitchen, electronics, clothing, documents, toys, decorations, sports, crafts, office, bathroom, medicine, food, cables, or other." },
					"tags":        map[string]any{ "type": "array" , "items": map[string]any{ "type": "string" } , "description": "Two to five lower-case words someone might search for." },
					"box":         map[string]any{ "type": "array" , "items": map[string]any{ "type": "number" } , "description": "Where it is in the photo: [x, y, width, height], each a fraction of the photo from 0 to 1, x and y the top-left corner." },
				},
			},
		},
	},
}

const instructions = "You catalog the contents of storage containers for a home inventory app. " +
	"The photo is taken from above, of things laid out before they go into a bin or box. " +
	"List every separate physical item you can see, once each; group identical items into one entry with " +
	"their quantity. Ignore the table, floor, background, hands, and the container itself. " +
	"Name things the way their owner would search for them later. Read brands and model numbers only when " +
	"they are legible -- never guess one. If the photo shows no items, answer with an empty list."

// Extract asks Claude what is in a photo. kind is its MIME type (image/jpeg,
// image/png, image/webp or image/gif).
func Extract( ctx context.Context , api_key string , model string , effort string , photo []byte , kind string , hint string ) ( result *Result , err error ) {
	client := anthropic.NewClient(
		option.WithoutEnvironmentDefaults(),
		option.WithAPIKey( api_key ),
		option.WithRequestTimeout( requestTimeout ),
		option.WithMaxRetries( 2 ),
	)
	prompt := "List the items in this photo."
	if hint = strings.TrimSpace( hint ); hint != "" {
		prompt += " The owner calls this container \"" + hint + "\"."
	}
	params := anthropic.BetaMessageNewParams{
		Model:     anthropic.Model( model ),
		MaxTokens: 16000,
		System:    []anthropic.BetaTextBlockParam{ { Text: instructions } },
		Messages: []anthropic.BetaMessageParam{
			anthropic.NewBetaUserMessage(
				anthropic.NewBetaImageBlock( anthropic.BetaBase64ImageSourceParam{
					Data:      base64.StdEncoding.EncodeToString( photo ),
					MediaType: anthropic.BetaBase64ImageSourceMediaType( kind ),
				} ),
				anthropic.NewBetaTextBlock( prompt ),
			),
		},
		OutputConfig: anthropic.BetaOutputConfigParam{
			Effort: anthropic.BetaOutputConfigEffort( effort ),
			Format: anthropic.BetaJSONOutputFormatParam{ Schema: schema },
		},
	}
	// A photo a safety check declines is answered by another model rather
	// than coming back empty. Opus and Sonnet take this; Haiku has no
	// server-side fallback, so it is left off there.
	if strings.HasPrefix( model , "claude-haiku" ) == false {
		params.Fallbacks = anthropic.BetaFallbacksParamOfDefault()
		params.Betas = []anthropic.AnthropicBeta{ anthropic.AnthropicBetaServerSideFallback2026_07_01 }
	}
	call_ctx , cancel := context.WithTimeout( ctx , requestTimeout * 3 )
	defer cancel()
	response , err := client.Beta.Messages.New( call_ctx , params )
	if err != nil {
		var api_err *anthropic.Error
		if errors.As( err , &api_err ) {
			err = fmt.Errorf( "the Claude API answered %d" , api_err.StatusCode )
		}
		return
	}
	switch response.StopReason {
	case anthropic.BetaStopReasonRefusal:
		err = ErrRefused
		return
	case anthropic.BetaStopReasonMaxTokens:
		err = ErrTruncated
		return
	}
	text := strings.Builder{}
	for _ , block := range response.Content {
		if variant , ok := block.AsAny().( anthropic.BetaTextBlock ); ok { text.WriteString( variant.Text ) }
	}
	result = &Result{}
	if json.Unmarshal( []byte( text.String() ) , result ) != nil {
		result = nil
		err = ErrUnreadable
		return
	}
	result.Items = clean( result.Items )
	return
}

// clean keeps what Claude said within the limits the inventory has.
func clean( found []models.ProposedItem ) ( items []models.ProposedItem ) {
	items = []models.ProposedItem{}
	for _ , item := range found {
		item.Name = trim( item.Name , 80 )
		if item.Name == "" { continue }
		item.Description = trim( item.Description , 500 )
		item.Brand = trim( item.Brand , 80 )
		item.Category = strings.ToLower( trim( item.Category , 40 ) )
		if item.Quantity < 1 || item.Quantity > 10000 { item.Quantity = 1 }
		tags := []string{}
		for _ , tag := range item.Tags {
			tag = strings.ToLower( trim( tag , 40 ) )
			if tag != "" && len( tags ) < 8 { tags = append( tags , tag ) }
		}
		if item.Category != "" && item.Category != "other" { tags = append( []string{ item.Category } , tags... ) }
		item.Tags = tags
		if len( item.Box ) != 4 {
			item.Box = nil
		} else {
			for index , value := range item.Box { item.Box[ index ] = min( max( value , 0 ) , 1 ) }
		}
		items = append( items , item )
		if len( items ) >= maxItems { break }
	}
	return
}

func trim( value string , most int ) ( trimmed string ) {
	trimmed = strings.TrimSpace( value )
	if runes := []rune( trimmed ); len( runes ) > most { trimmed = string( runes[ :most ] ) }
	return
}
