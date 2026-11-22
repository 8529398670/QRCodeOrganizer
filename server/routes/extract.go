package routes

import (
	context "context"
	errors "errors"
	io "io"
	log "log"
	time "time"

	fiber "github.com/gofiber/fiber/v3"

	ai "qrcodeorganizer/server/ai"
	catalog "qrcodeorganizer/server/ai/catalog"
	apidoc "qrcodeorganizer/server/apidoc"
	logs "qrcodeorganizer/server/logs"
	models "qrcodeorganizer/server/models"
	security "qrcodeorganizer/server/security"
)

// Smart Photo Extraction. The page uploads a top-down photo, then asks for a
// draft here. Reading a photo takes the AI model a while, longer than a
// request may stay open, so the draft is made in the background and the page
// asks after it until it is ready. Nothing reaches the inventory until a
// person ticks what is right and accepts it.
//
// The photo goes to whichever service the key belongs to (catalog.Providers).
// The server's key is set by the owner on Settings; anyone can add their own
// on Account.

const ExtractionWorking = "working"

// aiPlan is which service, key, model and effort one person's photos are
// read with: their own key with their own choices, or the server's key with
// the server's choices when an admin lets them, or none.
type aiPlan struct {
	provider string
	apiKey   string
	model    string
	effort   string
	using    string
}

// providerDefault is what an unpicked model and effort mean for a service:
// config.yaml's ai.model and ai.effort for Anthropic, and for the others the
// best match in the refreshed list, at the model's own effort.
func ( handlers *Handlers ) providerDefault( provider string ) ( model string , effort string ) {
	cfg := handlers.cfg()
	if catalog.Normal( provider ) == catalog.Anthropic {
		model , effort = cfg.AI.Model , cfg.AI.Effort
		return
	}
	service , _ := catalog.Lookup( provider )
	model = service.PickModel( models.AIModels( handlers.Store , provider ).Models )
	return
}

// serverChoice is the service, model and effort the server's key reads
// photos with.
func ( handlers *Handlers ) serverChoice() ( provider string , model string , effort string ) {
	server := models.ServerAI( handlers.Store )
	provider , model , effort = server.Service() , server.Model , server.Effort
	default_model , default_effort := handlers.providerDefault( provider )
	if model == "" { model = default_model }
	if effort == "" { effort = default_effort }
	return
}

// ownDefault is what an unpicked model and effort mean for a person's own
// key: the server's choice when it is the same service, else the service's
// default.
func ( handlers *Handlers ) ownDefault( provider string ) ( model string , effort string ) {
	server_provider , server_model , server_effort := handlers.serverChoice()
	if catalog.Normal( provider ) == server_provider {
		model , effort = server_model , server_effort
		return
	}
	model , effort = handlers.providerDefault( provider )
	return
}

func ( handlers *Handlers ) ownChoice( own *models.AISettings ) ( provider string , model string , effort string ) {
	provider , model , effort = own.Service() , own.Model , own.Effort
	default_model , default_effort := handlers.ownDefault( provider )
	if model == "" { model = default_model }
	if effort == "" { effort = default_effort }
	return
}

func ( handlers *Handlers ) aiPlanFor( user_id uint64 , open_key bool ) ( plan aiPlan , err error ) {
	cfg := handlers.cfg()
	plan.using = "none"
	if cfg.AI.Enabled == false { return }
	own := models.GetUserAI( handlers.Store , user_id )
	if own.KeySet() {
		plan.using = "own"
		plan.provider , plan.model , plan.effort = handlers.ownChoice( own )
		if open_key { plan.apiKey , err = models.UserAPIKey( handlers.Store , cfg.SecretKey , user_id ) }
		return
	}
	owner_id , _ := models.OwnerID( handlers.Store )
	if models.ServerAI( handlers.Store ).KeySet() && ( user_id == owner_id || models.GetAIPolicy( handlers.Store ).MayUseShared( user_id ) ) {
		plan.using = "shared"
		plan.provider , plan.model , plan.effort = handlers.serverChoice()
		if open_key { plan.apiKey , err = models.ServerAPIKey( handlers.Store , cfg.SecretKey ) }
	}
	return
}

type aiStatusReply struct {
	Ready        bool   `json:"ready" jsonschema_description:"Your photos can be read: switched on, and there is a key you may use."`
	Enabled      bool   `json:"enabled" jsonschema_description:"ai.enabled in the settings."`
	KeySet       bool   `json:"key_set" jsonschema_description:"The server has a key of its own."`
	Using        string `json:"using" jsonschema:"enum=own,enum=shared,enum=none" jsonschema_description:"Whose key your photos would go on: your own, the server's, or none you may use."`
	Provider     string `json:"provider" jsonschema_description:"The service your photos would go to: anthropic, openai, gemini, xai or openrouter."`
	ProviderName string `json:"provider_name"`
	Model        string `json:"model" jsonschema_description:"The model your photos would be read with; empty until one is picked or a model list is loaded."`
	Effort       string `json:"effort" jsonschema_description:"Empty is the model's own."`
	DailyLimit   int    `json:"daily_limit" jsonschema_description:"Photos one household may send a day with the server's key."`
}

func providerName( id string ) ( name string ) {
	if service , ok := catalog.Lookup( id ); ok { name = service.Name }
	return
}

func ( handlers *Handlers ) aiStatusFor( user_id uint64 ) ( reply *aiStatusReply ) {
	cfg := handlers.cfg()
	plan , _ := handlers.aiPlanFor( user_id , false )
	reply = &aiStatusReply{ Enabled: cfg.AI.Enabled , KeySet: models.ServerAI( handlers.Store ).KeySet() , Using: plan.using ,
		Provider: plan.provider , ProviderName: providerName( plan.provider ) , Model: plan.model , Effort: plan.effort , DailyLimit: cfg.AI.DailyLimitPerHousehold }
	reply.Ready = plan.using != "none" && plan.model != ""
	return
}

// @route("GET /api/ai", summary="Whether Smart Photo Extraction is ready for you",
//     tag="Photo extraction",
//     description="Whether your photos can be read, whose key they would go on (your own or the server's), "
//         "and which service and model would read them.",
//     mcp="ai_status", read_only=True,
// )
func ( handlers *Handlers ) getAIStatus( c fiber.Ctx , _ *apidoc.None ) ( reply *aiStatusReply , err error ) {
	reply = handlers.aiStatusFor( security.UserFrom( c ).ID )
	return
}

// ---- The services and their model lists --------------------------------

type aiProviderRow struct {
	ID          string              `json:"id" jsonschema_description:"anthropic, openai, gemini, xai or openrouter."`
	Name        string              `json:"name"`
	KeyHint     string              `json:"key_hint" jsonschema_description:"How its keys start."`
	KeysURL     string              `json:"keys_url" jsonschema_description:"Where to get a key."`
	Efforts     []string            `json:"efforts" jsonschema_description:"The effort levels it takes, least first."`
	Models      []catalog.ModelInfo `json:"models" jsonschema_description:"Its models that read photos, from the last refresh. A model's efforts are null when the service does not say, empty when it takes none."`
	RefreshedAt *time.Time          `json:"refreshed_at" jsonschema_description:"When the list was last loaded from the service; null is the app's built-in list."`
	// What "default" means here, for this screen.
	DefaultModel  string `json:"default_model" jsonschema_description:"The model an unpicked model means; empty when there is none yet (load the list)."`
	DefaultEffort string `json:"default_effort" jsonschema_description:"The effort an unpicked effort means; empty is the model's own."`
}

func ( aiProviderRow ) APIName() ( name string ) { name = "AIProvider"; return }

func ( handlers *Handlers ) aiProviders( defaults func( provider string ) ( string , string ) ) ( rows []aiProviderRow ) {
	rows = []aiProviderRow{}
	for _ , service := range catalog.Providers {
		list := models.AIModels( handlers.Store , service.ID )
		default_model , default_effort := defaults( service.ID )
		rows = append( rows , aiProviderRow{ ID: service.ID , Name: service.Name , KeyHint: service.KeyHint , KeysURL: service.KeysURL ,
			Efforts: service.Efforts , Models: list.Models , RefreshedAt: list.RefreshedAt , DefaultModel: default_model , DefaultEffort: default_effort } )
	}
	return
}

var errNoModels = errors.New( "that key's service lists no models that can read photos" )

// refreshModels loads a service's model list with a key and keeps it.
func ( handlers *Handlers ) refreshModels( provider string , api_key string ) ( err error ) {
	found , err := ai.ListModels( context.Background() , provider , api_key )
	if err != nil { return }
	if len( found ) == 0 {
		err = errNoModels
		return
	}
	_ , err = models.SaveAIModels( handlers.Store , provider , found )
	return
}

// refreshKey is the key a refresh uses: the one typed in, checked against
// the service, else the stored one when it is for that service.
func refreshKey( provider string , typed *string , stored *models.AISettings , open func() ( string , error ) ) ( api_key string , err error ) {
	service , ok := catalog.Lookup( provider )
	if ok == false {
		err = refuse( fiber.StatusBadRequest , catalog.ErrBadProvider.Error() )
		return
	}
	if typed != nil && *typed != "" {
		if check_err := service.CheckKey( *typed ); check_err != nil {
			err = refuse( fiber.StatusBadRequest , check_err.Error() )
			return
		}
		api_key = *typed
		return
	}
	if stored.KeySet() && stored.Service() == service.ID {
		api_key , err = open()
		return
	}
	err = refuse( fiber.StatusBadRequest , "enter a key for " + service.Name + " to load its models" )
	return
}

// ---- The server's key (the owner's Settings card) -------------------------

type aiKeyReply struct {
	aiStatusReply
	Last4        string          `json:"last4" jsonschema_description:"The key's last four characters, to tell keys apart."`
	SetAt        *time.Time      `json:"set_at"`
	ChosenModel  string          `json:"chosen_model" jsonschema_description:"The model picked for the server's key; empty is the default (model above)."`
	ChosenEffort string          `json:"chosen_effort" jsonschema_description:"The effort picked; empty is the default (effort above)."`
	AppModel     string          `json:"app_model" jsonschema_description:"config.yaml's ai.model, the default when the server's key is Anthropic's."`
	AppEffort    string          `json:"app_effort"`
	Providers    []aiProviderRow `json:"providers"`
	ModelsError  string          `json:"models_error,omitempty" jsonschema_description:"Why the model list could not be loaded, when a save or refresh tried."`
}

func ( handlers *Handlers ) aiKeyReply() ( reply *aiKeyReply ) {
	cfg := handlers.cfg()
	server := models.ServerAI( handlers.Store )
	provider , model , effort := handlers.serverChoice()
	reply = &aiKeyReply{
		aiStatusReply: aiStatusReply{ Enabled: cfg.AI.Enabled , KeySet: server.KeySet() , Provider: provider , ProviderName: providerName( provider ) ,
			Model: model , Effort: effort , DailyLimit: cfg.AI.DailyLimitPerHousehold , Ready: cfg.AI.Enabled && server.KeySet() && model != "" , Using: "shared" },
		Last4: server.Last4 , SetAt: server.SetAt , ChosenModel: server.Model , ChosenEffort: server.Effort ,
		AppModel: cfg.AI.Model , AppEffort: cfg.AI.Effort , Providers: handlers.aiProviders( handlers.providerDefault ),
	}
	if server.KeySet() == false { reply.Using = "none" }
	return
}

// @route("GET /api/owner/ai", summary="The server's AI key, as far as it can be shown",
//     tag="Photo extraction", role=OWNER,
//     description="Which service the server's key is for, its last four characters and when it was set, the "
//         "model and effort it reads photos with, and every service with its model list. The key itself is "
//         "never shown again.",
// )
func ( handlers *Handlers ) getAIKey( c fiber.Ctx , _ *apidoc.None ) ( reply *aiKeyReply , err error ) {
	reply = handlers.aiKeyReply()
	return
}

type setAIKeyRequest struct {
	Provider  *string `json:"provider,omitempty" jsonschema:"enum=anthropic,enum=openai,enum=gemini,enum=xai,enum=openrouter" jsonschema_description:"The service the key is for. Changing it needs a key for the new one." error:"provider must be anthropic, openai, gemini, xai or openrouter"`
	APIKey    *string `json:"api_key,omitempty" jsonschema:"maxLength=300" jsonschema_description:"A key for that service. Leave out to keep the one there is." error:"api_key must be at most 300 characters"`
	RemoveKey bool    `json:"remove_key,omitempty" jsonschema_description:"Take the server's key away."`
	Model     *string `json:"model,omitempty" jsonschema:"maxLength=128" jsonschema_description:"The model; empty is the default." error:"model must be at most 128 characters"`
	Effort    *string `json:"effort,omitempty" jsonschema:"maxLength=20" jsonschema_description:"One of the service's efforts; empty is the default." error:"effort must be at most 20 characters"`
}

func ( in *setAIKeyRequest ) change() ( change models.AIChange ) {
	change = models.AIChange{ Provider: in.Provider , APIKey: in.APIKey , RemoveKey: in.RemoveKey , Model: in.Model , Effort: in.Effort }
	return
}

// @route("POST /api/owner/ai", summary="Set the server's AI key, service, model and effort",
//     tag="Photo extraction", role=OWNER, errors=(400,),
//     description="Store the key Smart Photo Extraction uses, sealed in the database, for one service, or "
//         "remove it; and pick the model and effort. Only what you send changes. A new key also loads that "
//         "service's model list.",
// )
func ( handlers *Handlers ) setAIKey( c fiber.Ctx , in *setAIKeyRequest ) ( reply *aiKeyReply , err error ) {
	cfg := handlers.cfg()
	settings , set_err := models.SetServerAI( handlers.Store , cfg.SecretKey , security.UserFrom( c ).ID , in.change() )
	if err = inventoryRefusal( set_err , "" ); err != nil { return }
	models_error := ""
	if in.RemoveKey {
		handlers.audit( c , logs.KindAdmin , "server ai key removed" , fiber.Map{} )
	} else if in.APIKey != nil && *in.APIKey != "" {
		handlers.audit( c , logs.KindAdmin , "server ai key set" , fiber.Map{ "provider": settings.Service() } )
		if refresh_err := handlers.refreshModels( settings.Service() , *in.APIKey ); refresh_err != nil { models_error = refresh_err.Error() }
	}
	reply = handlers.aiKeyReply()
	reply.ModelsError = models_error
	return
}

type refreshModelsRequest struct {
	Provider string  `json:"provider" jsonschema:"enum=anthropic,enum=openai,enum=gemini,enum=xai,enum=openrouter" error:"provider must be anthropic, openai, gemini, xai or openrouter"`
	APIKey   *string `json:"api_key,omitempty" jsonschema:"maxLength=300" jsonschema_description:"A key to ask with, not stored. Leave out to use the saved key, when it is for this service." error:"api_key must be at most 300 characters"`
}

// @route("POST /api/owner/ai/models", summary="Load a service's model list for the server's key",
//     tag="Photo extraction", role=OWNER, errors=(400,),
//     description="Ask a service which of its models can read photos, and which efforts they take, and keep "
//         "the list for the screens.",
// )
func ( handlers *Handlers ) refreshServerModels( c fiber.Ctx , in *refreshModelsRequest ) ( reply *aiKeyReply , err error ) {
	cfg := handlers.cfg()
	api_key , err := refreshKey( in.Provider , in.APIKey , models.ServerAI( handlers.Store ) , func() ( string , error ) { return models.ServerAPIKey( handlers.Store , cfg.SecretKey ) } )
	if err != nil { return }
	if refresh_err := handlers.refreshModels( in.Provider , api_key ); refresh_err != nil {
		err = refuse( fiber.StatusBadRequest , refresh_err.Error() )
		return
	}
	reply = handlers.aiKeyReply()
	return
}

type extractRequest struct {
	labelPath
	FileID string `json:"file_id" jsonschema:"minLength=1" jsonschema_description:"The photo, already uploaded (files_start_upload): JPEG, PNG, WebP or GIF, taken from above." error:"file_id is required"`
}

type extractionRow struct {
	ID        uint64                `json:"id"`
	LabelID   uint64                `json:"label_id"`
	FileID    string                `json:"file_id"`
	Status    string                `json:"status" jsonschema:"enum=working,enum=ready,enum=accepted,enum=failed"`
	Summary   string                `json:"summary" jsonschema_description:"A name for the whole lot, to suggest for the container."`
	Items     []models.ProposedItem `json:"items" jsonschema_description:"What was found. box is where, as fractions of the photo: x, y, width, height."`
	Error     string                `json:"error"`
	CreatedAt time.Time             `json:"created_at"`
}

func ( extractionRow ) APIName() ( name string ) { name = "PhotoDraft"; return }

func newExtractionRow( extraction *models.Extraction ) ( row *extractionRow ) {
	row = &extractionRow{ ID: extraction.ID , LabelID: extraction.LabelID , FileID: extraction.FileID , Status: extraction.Status ,
		Summary: extraction.Summary , Items: extraction.Items , Error: extraction.Error , CreatedAt: extraction.CreatedAt }
	if row.Items == nil { row.Items = []models.ProposedItem{} }
	return
}

// @route("POST /api/households/{household_id}/labels/{label_id}/extract", summary="Read a photo of a container's contents",
//     tag="Photo extraction", status=202, errors=(400, 403, 404, 409, 429),
//     description="Ask Claude to list the items in an uploaded top-down photo, with your own API key if you "
//         "set one, else the server's if an admin lets you. Answers at once with a draft "
//         "whose status is working; ask get_extraction until it is ready, then accept_extraction with the "
//         "items that are right.",
//     mcp="extract_photo",
// )
func ( handlers *Handlers ) extractPhoto( c fiber.Ctx , in *extractRequest ) ( reply *extractionRow , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	user_id := security.UserFrom( c ).ID
	plan , err := handlers.aiPlanFor( user_id , true )
	if err != nil { return }
	if plan.using == "none" || plan.apiKey == "" {
		err = refuse( fiber.StatusConflict , "Smart Photo Extraction is not set up for you: add your own AI key on your Account screen, or ask an admin to let you use the server's" )
		return
	}
	if plan.model == "" {
		err = refuse( fiber.StatusConflict , "no model is picked for " + providerName( plan.provider ) + ": load its model list and pick one on the Account screen (your own key) or Settings (the server's)" )
		return
	}
	label , get_err := models.GetLabel( handlers.Store , in.LabelID )
	if get_err != nil || label.HouseholdID != household.ID || label.InTrash() {
		err = refuse( fiber.StatusNotFound , "no such container" )
		return
	}
	used , err := models.ExtractionsToday( handlers.Store , household.ID )
	if err != nil { return }
	if plan.using == "shared" && used >= handlers.cfg().AI.DailyLimitPerHousehold {
		err = refuse( fiber.StatusTooManyRequests , "this household has read as many photos as it may today" )
		return
	}
	photos := []string{ in.FileID }
	if err = handlers.checkPhotos( c , household , &photos ); err != nil { return }
	file , err := models.GetFile( handlers.Store , in.FileID )
	if err != nil { return }
	cfg := handlers.cfg()
	if file.Size > int64( cfg.AI.MaxPhotoBytes ) {
		err = refuse( fiber.StatusBadRequest , "that photo is too big to send: take it again at a smaller size" )
		return
	}
	reader , err := models.OpenFile( cfg , file )
	if err != nil { return }
	photo , read_err := io.ReadAll( reader )
	reader.Close()
	if read_err != nil {
		err = read_err
		return
	}
	kind := imageType( photo )
	if kind == "" {
		err = refuse( fiber.StatusBadRequest , "that file is not a JPEG, PNG, WebP or GIF photo" )
		return
	}
	extraction := &models.Extraction{ HouseholdID: household.ID , LabelID: label.ID , FileID: file.ID , Status: ExtractionWorking ,
		CreatedAt: time.Now().UTC() , CreatedBy: user_id , SharedKey: plan.using == "shared" }
	if err = models.SaveExtraction( handlers.Store , extraction ); err != nil { return }
	handlers.audit( c , logs.KindAdmin , "photo sent to ai" , fiber.Map{ "household_id": household.ID , "label_id": label.ID , "file_id": file.ID , "key": plan.using , "provider": plan.provider , "model": plan.model } )
	go handlers.runExtraction( *extraction , ai.Request{ Provider: plan.provider , APIKey: plan.apiKey , Model: plan.model , Effort: plan.effort , Photo: photo , Kind: kind , Hint: label.Name } )
	reply = newExtractionRow( extraction )
	return
}

// runExtraction reads the photo and stores what came back, or why not.
func ( handlers *Handlers ) runExtraction( extraction models.Extraction , request ai.Request ) {
	defer func() {
		if failure := recover(); failure != nil {
			log.Printf( "[ai] extraction %d panicked: %v" , extraction.ID , failure )
			extraction.Status , extraction.Error = models.ExtractionFailed , "something went wrong reading the photo"
			models.SaveExtraction( handlers.Store , &extraction )
		}
	}()
	result , err := ai.Extract( context.Background() , request )
	if err != nil {
		log.Printf( "[ai] extraction %d failed: %v" , extraction.ID , err )
		extraction.Status , extraction.Error = models.ExtractionFailed , err.Error()
	} else {
		extraction.Status , extraction.Summary , extraction.Items = models.ExtractionReady , result.Summary , result.Items
	}
	if save_err := models.SaveExtraction( handlers.Store , &extraction ); save_err != nil {
		log.Printf( "[ai] could not save extraction %d: %v" , extraction.ID , save_err )
	}
}

type extractionPath struct {
	householdPath
	ExtractionID uint64 `json:"extraction_id" path:"extraction_id" jsonschema:"minimum=1" jsonschema_description:"The draft's id, from extract_photo." error:"invalid draft id"`
}

func ( handlers *Handlers ) extraction( c fiber.Ctx , in *extractionPath , need string ) ( household *models.Household , extraction *models.Extraction , err error ) {
	household , _ , err = handlers.household( c , in.HouseholdID , need )
	if err != nil { return }
	extraction , get_err := models.GetExtraction( handlers.Store , in.ExtractionID )
	if get_err != nil || extraction.HouseholdID != household.ID {
		err = refuse( fiber.StatusNotFound , "no such draft" )
	}
	return
}

// @route("GET /api/households/{household_id}/extractions/{extraction_id}", summary="A photo draft",
//     tag="Photo extraction", errors=(404,),
//     description="What Claude found in a photo so far: status working, then ready (with items) or failed "
//         "(with error).",
//     mcp="get_extraction", read_only=True,
// )
func ( handlers *Handlers ) getExtraction( c fiber.Ctx , in *extractionPath ) ( reply *extractionRow , err error ) {
	_ , extraction , err := handlers.extraction( c , in , needRead )
	if err != nil { return }
	reply = newExtractionRow( extraction )
	return
}

type acceptExtractionRequest struct {
	extractionPath
	Items []itemFields `json:"items" jsonschema:"minItems=1,maxItems=100" jsonschema_description:"The items to keep, as they should be saved -- the draft's, edited as needed, with photos cropped from it if you like." error:"send 1-100 items"`
}

// @route("POST /api/households/{household_id}/extractions/{extraction_id}/accept", summary="Save a photo draft's items",
//     tag="Photo extraction", status=201, errors=(400, 403, 404, 409),
//     description="Put the items you kept from a draft into its container, and mark the draft done.",
//     mcp="accept_extraction",
// )
func ( handlers *Handlers ) acceptExtraction( c fiber.Ctx , in *acceptExtractionRequest ) ( reply *createItemsReply , err error ) {
	household , extraction , err := handlers.extraction( c , &in.extractionPath , needEdit )
	if err != nil { return }
	if extraction.Status != models.ExtractionReady {
		err = refuse( fiber.StatusConflict , "that draft is not ready, or was already saved" )
		return
	}
	created , err := handlers.createItems( c , &createItemsRequest{
		labelPath: labelPath{ householdPath: householdPath{ HouseholdID: household.ID } , LabelID: extraction.LabelID },
		Items:     in.Items,
	} )
	if err != nil { return }
	extraction.Status = models.ExtractionAccepted
	if err = models.SaveExtraction( handlers.Store , extraction ); err != nil { return }
	reply = created
	return
}
