package routes

import (
	context "context"
	io "io"
	log "log"
	time "time"

	fiber "github.com/gofiber/fiber/v3"

	ai "qrcodeorganizer/server/ai"
	apidoc "qrcodeorganizer/server/apidoc"
	logs "qrcodeorganizer/server/logs"
	models "qrcodeorganizer/server/models"
	security "qrcodeorganizer/server/security"
)

// Smart Photo Extraction. The page uploads a top-down photo, then asks for a
// draft here. Reading a photo takes Claude a while, longer than a request may
// stay open, so the draft is made in the background and the page asks after
// it until it is ready. Nothing reaches the inventory until a person ticks
// what is right and accepts it.

const ExtractionWorking = "working"

// aiPlan is which key, model and effort one person's photos are read with:
// their own key with their own choices, or the server's key with the app's
// settings when an admin lets them, or none.
type aiPlan struct {
	apiKey string
	model  string
	effort string
	using  string
}

func ( handlers *Handlers ) aiPlanFor( user_id uint64 , open_key bool ) ( plan aiPlan , err error ) {
	cfg := handlers.cfg()
	plan.using = "none"
	if cfg.AI.Enabled == false { return }
	own := models.GetUserAI( handlers.Store , user_id )
	if own.AnthropicKey != "" {
		plan.using , plan.model , plan.effort = "own" , own.Model , own.Effort
		if plan.model == "" { plan.model = cfg.AI.Model }
		if plan.effort == "" { plan.effort = cfg.AI.Effort }
		if open_key { plan.apiKey , err = models.UserAPIKey( handlers.Store , cfg.SecretKey , user_id ) }
		return
	}
	server_set , _ , _ := models.AnthropicKeyInfo( handlers.Store )
	owner_id , _ := models.OwnerID( handlers.Store )
	if server_set && ( user_id == owner_id || models.GetAIPolicy( handlers.Store ).MayUseShared( user_id ) ) {
		plan.using , plan.model , plan.effort = "shared" , cfg.AI.Model , cfg.AI.Effort
		if open_key { plan.apiKey , err = models.AnthropicKey( handlers.Store , cfg.SecretKey ) }
	}
	return
}

type aiStatusReply struct {
	Ready      bool   `json:"ready" jsonschema_description:"Your photos can be read: switched on, and there is a key you may use."`
	Enabled    bool   `json:"enabled" jsonschema_description:"ai.enabled in the settings."`
	KeySet     bool   `json:"key_set" jsonschema_description:"The server has a key of its own."`
	Using      string `json:"using" jsonschema:"enum=own,enum=shared,enum=none" jsonschema_description:"Whose key your photos would go on: your own, the server's, or none you may use."`
	Model      string `json:"model" jsonschema_description:"The model your photos would be read with."`
	Effort     string `json:"effort"`
	DailyLimit int    `json:"daily_limit" jsonschema_description:"Photos one household may send a day with the server's key."`
}

func ( handlers *Handlers ) aiStatusFor( user_id uint64 ) ( reply *aiStatusReply ) {
	cfg := handlers.cfg()
	set , _ , _ := models.AnthropicKeyInfo( handlers.Store )
	plan , _ := handlers.aiPlanFor( user_id , false )
	reply = &aiStatusReply{ Enabled: cfg.AI.Enabled , KeySet: set , Using: plan.using , Model: plan.model , Effort: plan.effort , DailyLimit: cfg.AI.DailyLimitPerHousehold }
	reply.Ready = plan.using != "none"
	return
}

// @route("GET /api/ai", summary="Whether Smart Photo Extraction is ready for you",
//     tag="Photo extraction",
//     description="Whether your photos can be read, whose key they would go on (your own or the server's), "
//         "and with which model.",
//     mcp="ai_status", read_only=True,
// )
func ( handlers *Handlers ) getAIStatus( c fiber.Ctx , _ *apidoc.None ) ( reply *aiStatusReply , err error ) {
	reply = handlers.aiStatusFor( security.UserFrom( c ).ID )
	return
}

type aiKeyReply struct {
	aiStatusReply
	Last4 string     `json:"last4" jsonschema_description:"The key's last four characters, to tell keys apart."`
	SetAt *time.Time `json:"set_at"`
}

func ( handlers *Handlers ) aiKeyReply() ( reply *aiKeyReply ) {
	_ , last4 , set_at := models.AnthropicKeyInfo( handlers.Store )
	set , _ , _ := models.AnthropicKeyInfo( handlers.Store )
	cfg := handlers.cfg()
	reply = &aiKeyReply{ aiStatusReply: aiStatusReply{ Enabled: cfg.AI.Enabled , KeySet: set , Model: cfg.AI.Model , Effort: cfg.AI.Effort ,
		DailyLimit: cfg.AI.DailyLimitPerHousehold , Ready: cfg.AI.Enabled && set , Using: "shared" } , Last4: last4 , SetAt: set_at }
	if set == false { reply.Using = "none" }
	return
}

// @route("GET /api/owner/ai", summary="The Anthropic API key, as far as it can be shown",
//     tag="Photo extraction", role=OWNER,
//     description="Whether an API key is set, its last four characters and when it was set. The key itself "
//         "is never shown again.",
// )
func ( handlers *Handlers ) getAIKey( c fiber.Ctx , _ *apidoc.None ) ( reply *aiKeyReply , err error ) {
	reply = handlers.aiKeyReply()
	return
}

type setAIKeyRequest struct {
	APIKey string `json:"api_key" jsonschema:"maxLength=300" jsonschema_description:"An Anthropic API key (sk-ant-...), or empty to remove it." error:"api_key must be at most 300 characters"`
}

// @route("POST /api/owner/ai", summary="Set or remove the Anthropic API key",
//     tag="Photo extraction", role=OWNER, errors=(400,),
//     description="Store the API key Smart Photo Extraction uses, sealed in the database, or remove it with "
//         "an empty value. It is never shown again.",
// )
func ( handlers *Handlers ) setAIKey( c fiber.Ctx , in *setAIKeyRequest ) ( reply *aiKeyReply , err error ) {
	set_err := models.SetAnthropicKey( handlers.Store , handlers.cfg().SecretKey , in.APIKey , security.UserFrom( c ).ID )
	if err = inventoryRefusal( set_err , "" ); err != nil { return }
	action := "anthropic key set"
	if in.APIKey == "" { action = "anthropic key removed" }
	handlers.audit( c , logs.KindAdmin , action , fiber.Map{} )
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
		err = refuse( fiber.StatusConflict , "Smart Photo Extraction is not set up for you: add your own Anthropic API key on your Account screen, or ask an admin to let you use the server's" )
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
	handlers.audit( c , logs.KindAdmin , "photo sent to claude" , fiber.Map{ "household_id": household.ID , "label_id": label.ID , "file_id": file.ID , "key": plan.using , "model": plan.model } )
	go handlers.runExtraction( *extraction , plan.apiKey , plan.model , plan.effort , photo , kind , label.Name )
	reply = newExtractionRow( extraction )
	return
}

// runExtraction reads the photo and stores what came back, or why not.
func ( handlers *Handlers ) runExtraction( extraction models.Extraction , api_key string , model string , effort string , photo []byte , kind string , hint string ) {
	defer func() {
		if failure := recover(); failure != nil {
			log.Printf( "[ai] extraction %d panicked: %v" , extraction.ID , failure )
			extraction.Status , extraction.Error = models.ExtractionFailed , "something went wrong reading the photo"
			models.SaveExtraction( handlers.Store , &extraction )
		}
	}()
	result , err := ai.Extract( context.Background() , api_key , model , effort , photo , kind , hint )
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
