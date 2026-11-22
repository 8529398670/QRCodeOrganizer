package ai

import (
	context "context"
	json "encoding/json"
	io "io"
	http "net/http"
	strings "strings"
	testing "testing"

	catalog "qrcodeorganizer/server/ai/catalog"
)

// fakeService answers each request with the next canned reply and keeps the
// bodies it was sent.
type fakeService struct {
	replies []string
	status  []int
	sent    []map[string]any
	urls    []string
}

func ( fake *fakeService ) RoundTrip( request *http.Request ) ( *http.Response , error ) {
	body := map[string]any{}
	if request.Body != nil {
		raw , _ := io.ReadAll( request.Body )
		json.Unmarshal( raw , &body )
	}
	fake.sent , fake.urls = append( fake.sent , body ) , append( fake.urls , request.URL.String() )
	index := len( fake.sent ) - 1
	return &http.Response{ StatusCode: fake.status[ index ] , Header: http.Header{ "Content-Type": { "application/json" } } ,
		Body: io.NopCloser( strings.NewReader( fake.replies[ index ] ) ) , Request: request } , nil
}

func useFake( t *testing.T , fake *fakeService ) {
	old := httpClient.Transport
	httpClient.Transport = fake
	t.Cleanup( func() { httpClient.Transport = old } )
}

const goodAnswer = `{"choices":[{"finish_reason":"stop","message":{"content":"{\"summary\":\"Tools\",\"items\":[{\"name\":\"Hammer\",\"description\":\"\",\"quantity\":2,\"brand\":\"\",\"category\":\"tools\",\"tags\":[\"hammer\"],\"box\":[0.1,0.1,0.2,0.2]}]}"}}]}`

func photoRequest( provider string ) ( request Request ) {
	request = Request{ Provider: provider , APIKey: "k" , Model: "some-model" , Effort: "low" , Photo: []byte{ 0xFF , 0xD8 , 0xFF } , Kind: "image/jpeg" , Hint: "Garage" }
	return
}

func TestAModelThatTakesNoEffortIsAskedAgainWithout( t *testing.T ) {
	fake := &fakeService{
		status:  []int{ 400 , 200 },
		replies: []string{ `{"error":{"message":"Unsupported parameter: 'reasoning_effort' is not supported with this model."}}` , goodAnswer },
	}
	useFake( t , fake )
	result , err := Extract( context.Background() , photoRequest( catalog.OpenAI ) )
	if err != nil { t.Fatal( err ) }
	if len( fake.sent ) != 2 || fake.sent[ 0 ][ "reasoning_effort" ] != "low" || fake.sent[ 1 ][ "reasoning_effort" ] != nil {
		t.Fatalf( "want low then no effort, sent %v" , fake.sent )
	}
	if fake.sent[ 0 ][ "max_completion_tokens" ] == nil || fake.sent[ 0 ][ "response_format" ] == nil { t.Fatalf( "OpenAI request missing fields: %v" , fake.sent[ 0 ] ) }
	if fake.urls[ 0 ] != "https://api.openai.com/v1/chat/completions" { t.Fatalf( "url %s" , fake.urls[ 0 ] ) }
	if result.Summary != "Tools" || len( result.Items ) != 1 || result.Items[ 0 ].Quantity != 2 || result.Items[ 0 ].Tags[ 0 ] != "tools" {
		t.Fatalf( "result %+v" , result )
	}
}

func TestAServiceWithoutSchemasGetsTheSchemaInThePrompt( t *testing.T ) {
	fake := &fakeService{
		status:  []int{ 400 , 200 },
		replies: []string{ `[{"error":{"message":"response_format json_schema is not supported"}}]` , goodAnswer },
	}
	useFake( t , fake )
	if _ , err := Extract( context.Background() , photoRequest( catalog.OpenRouter ) ); err != nil { t.Fatal( err ) }
	if fake.sent[ 1 ][ "response_format" ] != nil { t.Fatal( "format sent again" ) }
	if fake.sent[ 0 ][ "reasoning" ] == nil || fake.sent[ 0 ][ "max_tokens" ] == nil { t.Fatalf( "OpenRouter spellings missing: %v" , fake.sent[ 0 ] ) }
	encoded , _ := json.Marshal( fake.sent[ 1 ][ "messages" ] )
	if strings.Contains( string( encoded ) , "JSON Schema" ) == false { t.Fatal( "schema not in the prompt" ) }
}

func TestARefusalSaysWhy( t *testing.T ) {
	fake := &fakeService{ status: []int{ 401 } , replies: []string{ `{"error":{"message":"Incorrect API key provided"}}` } }
	useFake( t , fake )
	_ , err := Extract( context.Background() , photoRequest( catalog.XAI ) )
	if err == nil || err.Error() != "the xAI (Grok) API answered 401: Incorrect API key provided" { t.Fatalf( "got %v" , err ) }
}

func TestNoModelIsAnError( t *testing.T ) {
	request := photoRequest( catalog.Gemini )
	request.Model = ""
	if _ , err := Extract( context.Background() , request ); err != ErrNoModel { t.Fatalf( "got %v" , err ) }
}

func TestParseForgivesAFence( t *testing.T ) {
	result , err := parse( "Here you go:\n```json\n{\"summary\":\"x\",\"items\":[{\"name\":\" Cup \",\"quantity\":0}]}\n```" )
	if err != nil || result.Items[ 0 ].Name != "Cup" || result.Items[ 0 ].Quantity != 1 { t.Fatalf( "%+v %v" , result , err ) }
	if _ , err = parse( "no json here" ); err != ErrUnreadable { t.Fatalf( "got %v" , err ) }
}

func TestTheOpenAIListKeepsPhotoReaders( t *testing.T ) {
	fake := &fakeService{ status: []int{ 200 } , replies: []string{ `{"data":[{"id":"gpt-5-mini"},{"id":"gpt-4o"},{"id":"gpt-3.5-turbo"},{"id":"text-embedding-3-small"},{"id":"gpt-4o-realtime-preview"},{"id":"o3"},{"id":"gpt-image-1"},{"id":"dall-e-3"},{"id":"gpt-5-codex"}]}` } }
	useFake( t , fake )
	found , err := ListModels( context.Background() , catalog.OpenAI , "k" )
	if err != nil { t.Fatal( err ) }
	ids := []string{}
	for _ , model := range found { ids = append( ids , model.ID ) }
	if strings.Join( ids , "," ) != "gpt-4o,gpt-5-mini,o3" { t.Fatalf( "kept %v" , ids ) }
	if found[ 0 ].Efforts == nil || len( found[ 0 ].Efforts ) != 0 { t.Fatal( "gpt-4o should take no effort" ) }
}

func TestPickModel( t *testing.T ) {
	openai , _ := catalog.Lookup( catalog.OpenAI )
	list := []catalog.ModelInfo{ { ID: "gpt-4o" } , { ID: "gpt-5-mini-2025-08-07" } , { ID: "gpt-5-mini" } , { ID: "gpt-5.1-mini" } }
	if got := openai.PickModel( list ); got != "gpt-5.1-mini" { t.Fatalf( "picked %s" , got ) }
	gemini , _ := catalog.Lookup( catalog.Gemini )
	if got := gemini.PickModel( []catalog.ModelInfo{ { ID: "gemini-2.5-pro" } , { ID: "gemini-2.5-flash" } , { ID: "gemini-2.5-flash-lite" } } ); got != "gemini-2.5-flash" { t.Fatalf( "picked %s" , got ) }
}

func TestKeysMatchTheirService( t *testing.T ) {
	openai , _ := catalog.Lookup( catalog.OpenAI )
	anthropic , _ := catalog.Lookup( catalog.Anthropic )
	if openai.CheckKey( "sk-proj-abcdefghijklmnopqrstuvwxyz" ) != nil { t.Fatal( "good OpenAI key refused" ) }
	if openai.CheckKey( "sk-ant-abcdefghijklmnopqrstuvwxyz" ) == nil { t.Fatal( "Anthropic key taken as OpenAI's" ) }
	if anthropic.CheckKey( "sk-proj-abcdefghijklmnopqrstuvwxyz" ) == nil { t.Fatal( "OpenAI key taken as Anthropic's" ) }
	if anthropic.CheckKey( "sk-ant-abc def ghijklmnopqrstuvwxyz" ) == nil { t.Fatal( "space taken" ) }
}

func TestOpenRouterPrefersThePlainModel( t *testing.T ) {
	router , _ := catalog.Lookup( catalog.OpenRouter )
	list := []catalog.ModelInfo{ { ID: "google/gemini-3.8-flash:free" } , { ID: "google/gemini-3.8-flash" } , { ID: "openai/gpt-5-mini" } }
	if got := router.PickModel( list ); got != "google/gemini-3.8-flash" { t.Fatalf( "picked %s" , got ) }
}
