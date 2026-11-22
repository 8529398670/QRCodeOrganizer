package ai

// OpenAI, Google (Gemini), xAI (Grok) and OpenRouter all take the same
// request: OpenAI's chat completions, with the photo as a data: URL and the
// answer held to the schema by response_format. Each has its own address and
// a few spellings of its own; the model lists differ more, so each has its
// own reader below.

import (
	bytes "bytes"
	context "context"
	base64 "encoding/base64"
	json "encoding/json"
	errors "errors"
	fmt "fmt"
	io "io"
	log "log"
	http "net/http"
	sort "sort"
	strings "strings"
	time "time"

	catalog "qrcodeorganizer/server/ai/catalog"
)

type chatService struct {
	url string
	// maxField is the name of the output cap: OpenAI's reasoning models
	// refuse max_tokens, Gemini's and OpenRouter's spell it the old way.
	maxField string
	// effortObject sends effort as reasoning: {effort} (OpenRouter) rather
	// than reasoning_effort.
	effortObject bool
}

var chatServices = map[string]chatService{
	catalog.OpenAI:     { url: "https://api.openai.com/v1/chat/completions" , maxField: "max_completion_tokens" },
	catalog.XAI:        { url: "https://api.x.ai/v1/chat/completions" , maxField: "max_completion_tokens" },
	catalog.Gemini:     { url: "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions" , maxField: "max_tokens" },
	catalog.OpenRouter: { url: "https://openrouter.ai/api/v1/chat/completions" , maxField: "max_tokens" , effortObject: true },
}

var httpClient = &http.Client{ Timeout: requestTimeout }

// serviceError is a service's answer when it was not a 2xx.
type serviceError struct {
	status  int
	message string
}

func ( problem *serviceError ) Error() ( text string ) {
	text = fmt.Sprintf( "%d: %s" , problem.status , problem.message )
	return
}

// send makes one request, trying again on the answers that say "later"
// (429, 5xx) and on a dropped connection.
func send( ctx context.Context , method string , url string , header http.Header , body []byte ) ( reply []byte , err error ) {
	waits := []time.Duration{ 2 * time.Second , 6 * time.Second }
	for attempt := 0 ; ; attempt++ {
		var reader io.Reader
		if body != nil { reader = bytes.NewReader( body ) }
		request , build_err := http.NewRequestWithContext( ctx , method , url , reader )
		if build_err != nil {
			err = build_err
			return
		}
		request.Header = header.Clone()
		if body != nil { request.Header.Set( "Content-Type" , "application/json" ) }
		request.Header.Set( "Accept" , "application/json" )
		response , call_err := httpClient.Do( request )
		retry := false
		if call_err != nil {
			err , retry = call_err , ctx.Err() == nil
		} else {
			reply , err = io.ReadAll( io.LimitReader( response.Body , 32 << 20 ) )
			response.Body.Close()
			if err == nil && ( response.StatusCode < 200 || response.StatusCode > 299 ) {
				err = &serviceError{ status: response.StatusCode , message: errorMessage( reply ) }
				retry = response.StatusCode == 429 || response.StatusCode >= 500
			}
		}
		if err == nil || retry == false || attempt >= len( waits ) { return }
		select {
		case <-ctx.Done():
			return
		case <-time.After( waits[ attempt ] ):
		}
	}
}

// errorMessage digs the explanation out of an error answer, in the shapes
// these services use: {"error": {"message"}}, [{"error": ...}], {"error": "..."}.
func errorMessage( body []byte ) ( message string ) {
	type detail struct {
		Message string `json:"message"`
	}
	var one struct {
		Error   json.RawMessage `json:"error"`
		Message string          `json:"message"`
	}
	var many []struct {
		Error detail `json:"error"`
	}
	if json.Unmarshal( body , &one ) == nil {
		inner := detail{}
		if json.Unmarshal( one.Error , &inner ) == nil && inner.Message != "" {
			message = inner.Message
		} else if text := ""; json.Unmarshal( one.Error , &text ) == nil && text != "" {
			message = text
		} else {
			message = one.Message
		}
	} else if json.Unmarshal( body , &many ) == nil && len( many ) > 0 {
		message = many[ 0 ].Error.Message
	}
	if message == "" { message = trim( string( body ) , 200 ) }
	return
}

func bearer( api_key string ) ( header http.Header ) {
	header = http.Header{}
	header.Set( "Authorization" , "Bearer " + api_key )
	return
}

func readWithChat( ctx context.Context , provider catalog.Provider , request Request ) ( text string , err error ) {
	service := chatServices[ provider.ID ]
	image := "data:" + request.Kind + ";base64," + base64.StdEncoding.EncodeToString( request.Photo )
	effort , format , most := request.Effort , true , maxTokens
	// A model that does not take one of the options says so with a 400; the
	// photo is sent again without it rather than failing.
	for attempt := 0 ; attempt < 5 ; attempt++ {
		ask := prompt( request.Hint )
		if format == false { ask += schemaPrompt() }
		body := map[string]any{
			"model": request.Model,
			"messages": []any{
				map[string]any{ "role": "system" , "content": instructions },
				map[string]any{ "role": "user" , "content": []any{
					map[string]any{ "type": "image_url" , "image_url": map[string]any{ "url": image } },
					map[string]any{ "type": "text" , "text": ask },
				} },
			},
			service.maxField: most,
		}
		if format {
			body[ "response_format" ] = map[string]any{ "type": "json_schema" , "json_schema": map[string]any{ "name": "photo_items" , "strict": true , "schema": schema } }
		}
		if effort != "" {
			if service.effortObject {
				body[ "reasoning" ] = map[string]any{ "effort": effort }
			} else {
				body[ "reasoning_effort" ] = effort
			}
		}
		encoded , _ := json.Marshal( body )
		reply , call_err := send( ctx , http.MethodPost , service.url , bearer( request.APIKey ) , encoded )
		var problem *serviceError
		if errors.As( call_err , &problem ) {
			log.Printf( "[ai] %s API %d model %s: %s" , provider.Name , problem.status , request.Model , problem.message )
			if problem.status == 400 || problem.status == 422 {
				switch {
				case effort != "" && mentions( problem.message , "reasoning" , "effort" ):
					effort = ""
					continue
				case format && mentions( problem.message , "response_format" , "json_schema" , "structured" , "schema" ):
					format = false
					continue
				case most > 4096 && mentions( problem.message , "max_tokens" , "max_completion_tokens" , "maximum" , "too large" ):
					most = 4096
					continue
				}
			}
			err = apiError( provider.Name , problem.status , problem.message )
			return
		}
		if call_err != nil {
			err = call_err
			return
		}
		text , err = chatText( reply )
		return
	}
	err = apiError( provider.Name , 400 , "the model took none of the ways this app asks" )
	return
}

// chatText is the answer's text, or why there is none.
func chatText( reply []byte ) ( text string , err error ) {
	var answer struct {
		Choices []struct {
			FinishReason string `json:"finish_reason"`
			Message      struct {
				Content json.RawMessage `json:"content"`
				Refusal string          `json:"refusal"`
			} `json:"message"`
		} `json:"choices"`
	}
	if json.Unmarshal( reply , &answer ) != nil || len( answer.Choices ) == 0 {
		err = ErrUnreadable
		return
	}
	choice := answer.Choices[ 0 ]
	if choice.Message.Refusal != "" || choice.FinishReason == "content_filter" {
		err = ErrRefused
		return
	}
	if choice.FinishReason == "length" {
		err = ErrTruncated
		return
	}
	// Content is a string, or (some services) a list of text parts.
	if json.Unmarshal( choice.Message.Content , &text ) != nil {
		var parts []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		json.Unmarshal( choice.Message.Content , &parts )
		for _ , part := range parts { text += part.Text }
	}
	if strings.TrimSpace( text ) == "" { err = ErrUnreadable }
	return
}

// ---- Model lists -------------------------------------------------------

// ListModels asks a service which of its models can read a photo.
func ListModels( ctx context.Context , provider_id string , api_key string ) ( found []catalog.ModelInfo , err error ) {
	provider , ok := catalog.Lookup( provider_id )
	if ok == false {
		err = catalog.ErrBadProvider
		return
	}
	list_ctx , cancel := context.WithTimeout( ctx , 30 * time.Second )
	defer cancel()
	switch provider.ID {
	case catalog.Anthropic:
		found , err = listClaude( list_ctx , api_key )
	case catalog.OpenAI:
		found , err = listOpenAI( list_ctx , api_key )
	case catalog.Gemini:
		found , err = listGemini( list_ctx , api_key )
	case catalog.XAI:
		found , err = listXAI( list_ctx , api_key )
	case catalog.OpenRouter:
		found , err = listOpenRouter( list_ctx , api_key )
	}
	var problem *serviceError
	if errors.As( err , &problem ) { err = apiError( provider.Name , problem.status , problem.message ) }
	if len( found ) > 500 { found = found[ :500 ] }
	return
}

func sortByID( found []catalog.ModelInfo ) {
	sort.Slice( found , func( a , b int ) bool { return found[ a ].ID < found[ b ].ID } )
}

func hasAny( value string , words ...string ) ( found bool ) {
	for _ , word := range words {
		if strings.Contains( value , word ) { found = true }
	}
	return
}

func contains( list []string , value string ) ( found bool ) {
	for _ , item := range list {
		if item == value { found = true }
	}
	return
}

// listOpenAI keeps the chat models that read images. OpenAI's list says
// nothing about either, so it goes by name: the GPT-4o/4.1/5 and o-series
// families, without the audio, realtime, search, image-making, Codex and Pro
// models, which chat completions does not serve.
func listOpenAI( ctx context.Context , api_key string ) ( found []catalog.ModelInfo , err error ) {
	reply , err := send( ctx , http.MethodGet , "https://api.openai.com/v1/models" , bearer( api_key ) , nil )
	if err != nil { return }
	var list struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err = json.Unmarshal( reply , &list ); err != nil { return }
	found = []catalog.ModelInfo{}
	for _ , model := range list.Data {
		id := model.ID
		family := strings.HasPrefix( id , "gpt-4o" ) || strings.HasPrefix( id , "gpt-4.1" ) || strings.HasPrefix( id , "gpt-4-turbo" ) ||
			strings.HasPrefix( id , "chatgpt-" ) || ( strings.HasPrefix( id , "gpt-" ) && id >= "gpt-5" ) ||
			( len( id ) > 1 && id[ 0 ] == 'o' && id[ 1 ] >= '1' && id[ 1 ] <= '9' )
		if family == false || hasAny( id , "audio" , "realtime" , "tts" , "transcribe" , "search" , "image" , "embedding" ,
			"moderation" , "instruct" , "codex" , "deep-research" , "-pro" , "o1-mini" , "o1-preview" , "computer-use" ) { continue }
		info := catalog.ModelInfo{ ID: id , Name: id }
		// The GPT-4 family does not reason, so it takes no effort.
		if strings.HasPrefix( id , "gpt-4" ) || strings.HasPrefix( id , "chatgpt-" ) || strings.HasSuffix( id , "-chat-latest" ) { info.Efforts = []string{} }
		found = append( found , info )
	}
	sortByID( found )
	return
}

// listGemini uses Gemini's own list (the OpenAI-style one leaves out what a
// model can do): the Gemini models that write text, without the embedding,
// speech, live and image-making ones.
func listGemini( ctx context.Context , api_key string ) ( found []catalog.ModelInfo , err error ) {
	header := http.Header{}
	header.Set( "x-goog-api-key" , api_key )
	found = []catalog.ModelInfo{}
	page := ""
	for pages := 0 ; pages < 5 ; pages++ {
		url := "https://generativelanguage.googleapis.com/v1beta/models?pageSize=1000"
		if page != "" { url += "&pageToken=" + page }
		reply , call_err := send( ctx , http.MethodGet , url , header , nil )
		if call_err != nil {
			err = call_err
			return
		}
		var list struct {
			Models []struct {
				Name        string   `json:"name"`
				DisplayName string   `json:"displayName"`
				Methods     []string `json:"supportedGenerationMethods"`
				Thinking    bool     `json:"thinking"`
			} `json:"models"`
			Next string `json:"nextPageToken"`
		}
		if err = json.Unmarshal( reply , &list ); err != nil { return }
		for _ , model := range list.Models {
			id := strings.TrimPrefix( model.Name , "models/" )
			if strings.HasPrefix( id , "gemini-" ) == false || contains( model.Methods , "generateContent" ) == false ||
				hasAny( id , "embedding" , "tts" , "live" , "native-audio" , "-image" , "aqa" , "robotics" , "computer-use" ) { continue }
			info := catalog.ModelInfo{ ID: id , Name: model.DisplayName }
			if info.Name == "" { info.Name = id }
			if model.Thinking == false { info.Efforts = []string{} }
			found = append( found , info )
		}
		if page = list.Next; page == "" { break }
	}
	sortByID( found )
	return
}

// listXAI keeps the Grok models that take images in.
func listXAI( ctx context.Context , api_key string ) ( found []catalog.ModelInfo , err error ) {
	reply , err := send( ctx , http.MethodGet , "https://api.x.ai/v1/language-models" , bearer( api_key ) , nil )
	if err != nil { return }
	var list struct {
		Models []struct {
			ID     string   `json:"id"`
			Inputs []string `json:"input_modalities"`
		} `json:"models"`
	}
	if err = json.Unmarshal( reply , &list ); err != nil { return }
	found = []catalog.ModelInfo{}
	for _ , model := range list.Models {
		if contains( model.Inputs , "image" ) == false { continue }
		found = append( found , catalog.ModelInfo{ ID: model.ID , Name: model.ID } )
	}
	sortByID( found )
	return
}

// listOpenRouter keeps the models that take images in and can be held to a
// schema, and knows which of them reason.
func listOpenRouter( ctx context.Context , api_key string ) ( found []catalog.ModelInfo , err error ) {
	reply , err := send( ctx , http.MethodGet , "https://openrouter.ai/api/v1/models" , bearer( api_key ) , nil )
	if err != nil { return }
	var list struct {
		Data []struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Architecture struct {
				Inputs []string `json:"input_modalities"`
			} `json:"architecture"`
			Parameters []string `json:"supported_parameters"`
		} `json:"data"`
	}
	if err = json.Unmarshal( reply , &list ); err != nil { return }
	found = []catalog.ModelInfo{}
	for _ , model := range list.Data {
		// A :batch variant answers later, not while the page waits.
		if contains( model.Architecture.Inputs , "image" ) == false || strings.HasSuffix( model.ID , ":batch" ) { continue }
		if contains( model.Parameters , "structured_outputs" ) == false && contains( model.Parameters , "response_format" ) == false { continue }
		info := catalog.ModelInfo{ ID: model.ID , Name: model.Name }
		if contains( model.Parameters , "reasoning" ) == false { info.Efforts = []string{} }
		found = append( found , info )
	}
	sortByID( found )
	return
}
