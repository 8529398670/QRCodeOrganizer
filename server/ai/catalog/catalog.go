// Package catalog is the AI services Smart Photo Extraction can send a photo
// to, what their keys look like, and the effort levels each one takes. It has
// no dependencies of its own, so the models package (which stores the keys)
// and the ai package (which calls the services) can both use it.
package catalog

import (
	errors "errors"
	fmt "fmt"
	regexp "regexp"
	sort "sort"
	strings "strings"
)

const (
	Anthropic  = "anthropic"
	OpenAI     = "openai"
	Gemini     = "gemini"
	XAI        = "xai"
	OpenRouter = "openrouter"
)

// Provider is one AI service.
type Provider struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// KeyHint is the start of a key, for the key field's placeholder.
	KeyHint string `json:"key_hint"`
	// KeysURL is where someone gets a key.
	KeysURL string `json:"keys_url"`
	// Efforts are the effort levels this service takes, least first. A model
	// can take fewer (ModelInfo.Efforts); empty means "the model's own".
	Efforts []string `json:"efforts"`

	keyPrefix  string
	notPrefix  []string
	preferred  []*regexp.Regexp
}

// ModelInfo is one model a service offers that can read a photo.
type ModelInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Efforts is what this model takes when the service says so: nil is not
	// known (the service's list is offered), empty takes none.
	Efforts []string `json:"efforts"`
}

func patterns( list ...string ) ( compiled []*regexp.Regexp ) {
	for _ , pattern := range list { compiled = append( compiled , regexp.MustCompile( pattern ) ) }
	return
}

// Providers is every service, in the order the screens offer them.
var Providers = []Provider{
	{ ID: Anthropic , Name: "Anthropic (Claude)" , KeyHint: "sk-ant-..." , KeysURL: "https://platform.claude.com/settings/keys" ,
		Efforts: []string{ "low" , "medium" , "high" , "xhigh" , "max" } , keyPrefix: "sk-ant-" ,
		preferred: patterns( `^claude-opus-\d` , `^claude-sonnet-\d` , `^claude-` ) },
	{ ID: OpenAI , Name: "OpenAI (ChatGPT)" , KeyHint: "sk-..." , KeysURL: "https://platform.openai.com/api-keys" ,
		Efforts: []string{ "none" , "minimal" , "low" , "medium" , "high" , "xhigh" } , keyPrefix: "sk-" , notPrefix: []string{ "sk-ant-" , "sk-or-" } ,
		preferred: patterns( `^gpt-\d+(\.\d+)?-mini$` , `^gpt-\d+(\.\d+)?$` , `^gpt-\d` , `^o\d` ) },
	{ ID: Gemini , Name: "Google (Gemini)" , KeyHint: "AIza..." , KeysURL: "https://aistudio.google.com/apikey" ,
		Efforts: []string{ "none" , "minimal" , "low" , "medium" , "high" } ,
		preferred: patterns( `^gemini-flash-latest$` , `^gemini-[\d.]+-flash$` , `^gemini-.*flash` , `^gemini-` ) },
	{ ID: XAI , Name: "xAI (Grok)" , KeyHint: "xai-..." , KeysURL: "https://console.x.ai" ,
		Efforts: []string{ "low" , "high" } , keyPrefix: "xai-" ,
		preferred: patterns( `^grok-\d.*fast` , `^grok-\d` , `^grok-` ) },
	{ ID: OpenRouter , Name: "OpenRouter (many models)" , KeyHint: "sk-or-..." , KeysURL: "https://openrouter.ai/settings/keys" ,
		Efforts: []string{ "none" , "minimal" , "low" , "medium" , "high" , "xhigh" } , keyPrefix: "sk-or-" ,
		preferred: patterns( `^google/gemini-.*flash` , `^openai/gpt-.*mini` , `^anthropic/claude-sonnet` ) },
}

// Lookup finds a service by id; "" is Anthropic, which every key stored
// before there was a choice belongs to.
func Lookup( id string ) ( provider Provider , ok bool ) {
	if id == "" { id = Anthropic }
	for _ , known := range Providers {
		if known.ID == id { provider , ok = known , true }
	}
	return
}

// Normal is a provider id as stored, "" meaning Anthropic.
func Normal( id string ) ( normal string ) {
	normal = id
	if normal == "" { normal = Anthropic }
	return
}

var (
	ErrBadProvider = errors.New( "provider must be anthropic, openai, gemini, xai or openrouter" )
	ErrBadKey      = errors.New( "that does not look like an API key" )
	ErrBadModel    = errors.New( "that is not a model name: letters, digits and . _ : / @ + - only, at most 128" )
	ErrBadEffort   = errors.New( "that effort is not one this service takes" )
)

// CheckKey catches a key pasted into the wrong service, or a stray space.
func ( provider Provider ) CheckKey( key string ) ( err error ) {
	if len( key ) < 20 || len( key ) > 300 || strings.ContainsAny( key , " \t\r\n" ) {
		err = ErrBadKey
		return
	}
	if provider.keyPrefix != "" && strings.HasPrefix( key , provider.keyPrefix ) == false {
		err = fmt.Errorf( "%w: keys for %s start %s" , ErrBadKey , provider.Name , provider.KeyHint )
		return
	}
	for _ , not := range provider.notPrefix {
		if strings.HasPrefix( key , not ) {
			err = fmt.Errorf( "%w: that key is for another service, not %s" , ErrBadKey , provider.Name )
			return
		}
	}
	return
}

var modelID = regexp.MustCompile( `^[A-Za-z0-9][A-Za-z0-9._:/@+\-]{0,127}$` )

func CheckModel( model string ) ( err error ) {
	if model != "" && modelID.MatchString( model ) == false { err = ErrBadModel }
	return
}

// CheckEffort is whether this service takes an effort ("" always: the
// model's own).
func ( provider Provider ) CheckEffort( effort string ) ( err error ) {
	if effort == "" { return }
	for _ , known := range provider.Efforts {
		if known == effort { return }
	}
	err = fmt.Errorf( "%w: %s takes %s" , ErrBadEffort , provider.Name , strings.Join( provider.Efforts , ", " ) )
	return
}

var dated = regexp.MustCompile( `preview|exp|latest-?\d|-\d{4}-?\d{2}-?\d{2}|-\d{3,}|:` )

// PickModel chooses a sensible model from a refreshed list when a person has
// not picked one: the first preference that matches, a plain name over a
// dated or preview one, the newest version first.
func ( provider Provider ) PickModel( models []ModelInfo ) ( model string ) {
	for _ , pattern := range provider.preferred {
		found := []string{}
		for _ , info := range models {
			if pattern.MatchString( info.ID ) { found = append( found , info.ID ) }
		}
		if len( found ) == 0 { continue }
		sort.Slice( found , func( a , b int ) bool {
			dated_a , dated_b := dated.MatchString( found[ a ] ) , dated.MatchString( found[ b ] )
			if dated_a != dated_b { return dated_b }
			return found[ a ] > found[ b ]
		} )
		model = found[ 0 ]
		return
	}
	if len( models ) > 0 { model = models[ 0 ].ID }
	return
}

// BuiltIn is what each service offers before anyone has refreshed its list:
// Anthropic's current models, and nothing for the rest, whose names change
// too often to write down.
func BuiltIn( id string ) ( models []ModelInfo ) {
	models = []ModelInfo{}
	if Normal( id ) == Anthropic {
		all := []string{ "low" , "medium" , "high" , "xhigh" , "max" }
		models = []ModelInfo{
			{ ID: "claude-opus-5-5" , Name: "Claude Opus 5.5" , Efforts: all },
			{ ID: "claude-sonnet-5-5" , Name: "Claude Sonnet 5.5" , Efforts: all },
			{ ID: "claude-haiku-5-5" , Name: "Claude Haiku 5.5" , Efforts: all },
		}
	}
	return
}
