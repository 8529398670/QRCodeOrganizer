// Package ai is Smart Photo Extraction: one top-down photo of what is going
// into a container, and an AI model lists the things in it -- names, a few
// words each, how many, brand, tags, and where each one is in the photo, so
// the page can crop a picture of it. A person checks the list before anything
// is saved.
//
// The photo can go to Anthropic (Claude), OpenAI, Google (Gemini), xAI (Grok)
// or OpenRouter (see the catalog package). The key is handed in by the
// caller, read from the database, never from the environment, so no variable
// can change where photos go or whose key pays.
package ai

import (
	context "context"
	encoding_json "encoding/json"
	errors "errors"
	strings "strings"
	time "time"

	catalog "qrcodeorganizer/server/ai/catalog"
	models "qrcodeorganizer/server/models"
)

const (
	requestTimeout = 120 * time.Second
	maxTokens      = 16000
	maxItems       = 60
)

var (
	ErrRefused    = errors.New( "the AI model would not describe that photo" )
	ErrTruncated  = errors.New( "the photo held more than could be listed at once: try a photo of fewer things" )
	ErrUnreadable = errors.New( "the AI model's answer could not be read" )
	ErrNoModel    = errors.New( "no model is picked for this key: choose one on the Account or Settings screen" )
)

// Request is one photo to read.
type Request struct {
	Provider string
	APIKey   string
	Model    string
	// Effort is "" for the model's own.
	Effort string
	Photo  []byte
	// Kind is the photo's MIME type: image/jpeg, image/png, image/webp or image/gif.
	Kind string
	// Hint is the container's name, if it has one.
	Hint string
}

// Result is what the model made of one photo.
type Result struct {
	Summary string                `json:"summary"`
	Items   []models.ProposedItem `json:"items"`
}

// schema is the shape the answer has to take (structured outputs).
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

func prompt( hint string ) ( text string ) {
	text = "List the items in this photo."
	if hint = strings.TrimSpace( hint ); hint != "" {
		text += " The owner calls this container \"" + hint + "\"."
	}
	return
}

// schemaPrompt is the instruction a model gets when its service will not
// hold it to the schema: answer with JSON of this shape and nothing else.
func schemaPrompt() ( text string ) {
	shape , _ := encoding_json.Marshal( schema )
	text = " Answer with one JSON object and nothing else -- no prose, no code fence -- matching this JSON Schema: " + string( shape )
	return
}

// Extract asks the chosen service what is in a photo.
func Extract( ctx context.Context , request Request ) ( result *Result , err error ) {
	provider , ok := catalog.Lookup( request.Provider )
	if ok == false {
		err = catalog.ErrBadProvider
		return
	}
	if strings.TrimSpace( request.Model ) == "" {
		err = ErrNoModel
		return
	}
	call_ctx , cancel := context.WithTimeout( ctx , requestTimeout * 4 )
	defer cancel()
	var text string
	if provider.ID == catalog.Anthropic {
		text , err = readWithClaude( call_ctx , request )
	} else {
		text , err = readWithChat( call_ctx , provider , request )
	}
	if err != nil { return }
	result , err = parse( text )
	return
}

// parse reads the answer, forgiving a model that wrapped it in a code fence
// or a sentence when its service could not hold it to the schema.
func parse( text string ) ( result *Result , err error ) {
	text = strings.TrimSpace( text )
	if start , end := strings.Index( text , "{" ) , strings.LastIndex( text , "}" ); start >= 0 && end > start {
		text = text[ start : end + 1 ]
	}
	result = &Result{}
	if encoding_json.Unmarshal( []byte( text ) , result ) != nil {
		result = nil
		err = ErrUnreadable
		return
	}
	result.Items = clean( result.Items )
	return
}

// clean keeps what the model said within the limits the inventory has.
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

// mentions is whether an error message talks about any of these words, which
// is how a refused option is told apart from a refused photo.
func mentions( message string , words ...string ) ( found bool ) {
	message = strings.ToLower( message )
	for _ , word := range words {
		if strings.Contains( message , word ) { found = true }
	}
	return
}
