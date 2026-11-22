package ai

// Anthropic (Claude), through the official SDK. The client is built with
// option.WithoutEnvironmentDefaults, so no ANTHROPIC_* variable can change
// where photos go or whose key pays.

import (
	context "context"
	base64 "encoding/base64"
	errors "errors"
	fmt "fmt"
	log "log"
	strings "strings"

	anthropic "github.com/anthropics/anthropic-sdk-go/v2"
	option "github.com/anthropics/anthropic-sdk-go/v2/option"

	catalog "qrcodeorganizer/server/ai/catalog"
)

func claudeClient( api_key string ) ( client anthropic.Client ) {
	client = anthropic.NewClient(
		option.WithoutEnvironmentDefaults(),
		option.WithAPIKey( api_key ),
		option.WithRequestTimeout( requestTimeout ),
		option.WithMaxRetries( 2 ),
	)
	return
}

// fallbackModels take server-side fallbacks: a photo a safety check declines
// is answered by another model rather than coming back empty. Other models
// (Haiku, anything older) reject the parameter.
var fallbackModels = []string{ "claude-fable-5-2" , "claude-fable-5-1" , "claude-mythos-5-2" , "claude-mythos-5-1" , "claude-opus-5-5" , "claude-opus-5" , "claude-sonnet-5-5" }

func takesFallbacks( model string ) ( ok bool ) {
	for _ , known := range fallbackModels {
		if known == model { ok = true }
	}
	return
}

func readWithClaude( ctx context.Context , request Request ) ( text string , err error ) {
	client := claudeClient( request.APIKey )
	effort , format , fallbacks := request.Effort , true , takesFallbacks( request.Model )
	// A model that does not take one of the options says so with a 400; the
	// photo is sent again without it rather than failing.
	for attempt := 0 ; attempt < 4 ; attempt++ {
		ask := prompt( request.Hint )
		if format == false { ask += schemaPrompt() }
		params := anthropic.BetaMessageNewParams{
			Model:     anthropic.Model( request.Model ),
			MaxTokens: maxTokens,
			System:    []anthropic.BetaTextBlockParam{ { Text: instructions } },
			Messages: []anthropic.BetaMessageParam{
				anthropic.NewBetaUserMessage(
					anthropic.NewBetaImageBlock( anthropic.BetaBase64ImageSourceParam{
						Data:      base64.StdEncoding.EncodeToString( request.Photo ),
						MediaType: anthropic.BetaBase64ImageSourceMediaType( request.Kind ),
					} ),
					anthropic.NewBetaTextBlock( ask ),
				),
			},
			OutputConfig: anthropic.BetaOutputConfigParam{ Effort: anthropic.BetaOutputConfigEffort( effort ) },
		}
		if format { params.OutputConfig.Format = anthropic.BetaJSONOutputFormatParam{ Schema: schema } }
		if fallbacks {
			params.Fallbacks = anthropic.BetaFallbacksParamOfDefault()
			params.Betas = []anthropic.AnthropicBeta{ anthropic.AnthropicBetaServerSideFallback2026_07_01 }
		}
		response , call_err := client.Beta.Messages.New( ctx , params )
		if call_err == nil {
			switch response.StopReason {
			case anthropic.BetaStopReasonRefusal:
				err = ErrRefused
				return
			case anthropic.BetaStopReasonMaxTokens:
				err = ErrTruncated
				return
			}
			out := strings.Builder{}
			for _ , block := range response.Content {
				if variant , ok := block.AsAny().( anthropic.BetaTextBlock ); ok { out.WriteString( variant.Text ) }
			}
			text = out.String()
			return
		}
		var api_err *anthropic.Error
		if errors.As( call_err , &api_err ) == false {
			err = call_err
			return
		}
		log.Printf( "[ai] Claude API %d %s model %s request %s: %s" , api_err.StatusCode , api_err.Type() , request.Model , api_err.RequestID , api_err.Message )
		if api_err.StatusCode == 400 {
			switch {
			case fallbacks && mentions( api_err.Message , "fallback" ):
				fallbacks = false
				continue
			case effort != "" && mentions( api_err.Message , "effort" ):
				effort = ""
				continue
			case format && mentions( api_err.Message , "output_config.format" , "structured output" , "json_schema" , "output format" ):
				format = false
				continue
			}
		}
		err = apiError( "Claude" , api_err.StatusCode , api_err.Message )
		return
	}
	err = apiError( "Claude" , 400 , "the model took none of the ways this app asks" )
	return
}

// apiError is a service's refusal in words a person can act on. Its message
// never holds the key or the photo.
func apiError( service string , status int , message string ) ( err error ) {
	if message = trim( message , 300 ); message != "" {
		err = fmt.Errorf( "the %s API answered %d: %s" , service , status , message )
	} else {
		err = fmt.Errorf( "the %s API answered %d" , service , status )
	}
	return
}

// listClaude is Anthropic's model list: every model still served that reads
// images, with the effort levels each one takes, newest first.
func listClaude( ctx context.Context , api_key string ) ( found []catalog.ModelInfo , err error ) {
	client := claudeClient( api_key )
	pager := client.Models.ListAutoPaging( ctx , anthropic.ModelListParams{} )
	found = []catalog.ModelInfo{}
	for pager.Next() {
		model := pager.Current()
		if model.Lifecycle == "retired" || model.Capabilities.ImageInput.Supported == false { continue }
		efforts := []string{}
		levels := model.Capabilities.Effort
		if levels.Supported {
			for _ , level := range []struct{ name string ; ok bool }{
				{ "low" , levels.Low.Supported } , { "medium" , levels.Medium.Supported } , { "high" , levels.High.Supported } ,
				{ "xhigh" , levels.Xhigh.Supported } , { "max" , levels.Max.Supported } ,
			} {
				if level.ok { efforts = append( efforts , level.name ) }
			}
		}
		found = append( found , catalog.ModelInfo{ ID: model.ID , Name: model.DisplayName , Efforts: efforts } )
	}
	if err = pager.Err(); err != nil {
		var api_err *anthropic.Error
		if errors.As( err , &api_err ) { err = apiError( "Claude" , api_err.StatusCode , api_err.Message ) }
	}
	return
}
