package routes

import (
	time "time"

	fiber "github.com/gofiber/fiber/v3"

	apidoc "qrcodeorganizer/server/apidoc"
	logs "qrcodeorganizer/server/logs"
	models "qrcodeorganizer/server/models"
	security "qrcodeorganizer/server/security"
)

// Whose AI key reads whose photos. Anyone may add their own key on their
// Account screen, for the service, model and effort they want, and pay for
// their own photos. Without one, whether they may use the server's key (the
// owner's bill) is for the admins to say: a default for everyone, and per
// person "follow the default", "may use it" or "must bring their own". The
// owner can always use the server's key, since it is theirs.

type myAIReply struct {
	aiStatusReply
	OwnKeySet      bool            `json:"own_key_set" jsonschema_description:"You have a key of your own set."`
	Last4          string          `json:"last4" jsonschema_description:"Your key's last four characters, to tell keys apart."`
	SetAt          *time.Time      `json:"set_at"`
	OwnProvider    string          `json:"own_provider" jsonschema_description:"The service your own key is for (or will be, once you add one)."`
	OwnModel       string          `json:"own_model" jsonschema_description:"The model you picked for your own key; empty is the default."`
	OwnEffort      string          `json:"own_effort" jsonschema_description:"The effort you picked for your own key; empty is the default."`
	SharedAllowed  bool            `json:"shared_allowed" jsonschema_description:"You may use the server's key when you have none of your own."`
	ServerProvider string          `json:"server_provider" jsonschema_description:"The service the server's key is for."`
	ServerModel    string          `json:"server_model" jsonschema_description:"The model the server's key reads photos with: the default for your own key on the same service."`
	ServerEffort   string          `json:"server_effort"`
	Providers      []aiProviderRow `json:"providers"`
	ModelsError    string          `json:"models_error,omitempty" jsonschema_description:"Why the model list could not be loaded, when a save tried."`
}

func ( handlers *Handlers ) myAIReply( user_id uint64 ) ( reply *myAIReply ) {
	own := models.GetUserAI( handlers.Store , user_id )
	owner_id , _ := models.OwnerID( handlers.Store )
	server_provider , server_model , server_effort := handlers.serverChoice()
	reply = &myAIReply{
		aiStatusReply:  *handlers.aiStatusFor( user_id ),
		OwnKeySet:      own.KeySet(),
		Last4:          own.Last4,
		SetAt:          own.SetAt,
		OwnProvider:    own.Service(),
		OwnModel:       own.Model,
		OwnEffort:      own.Effort,
		SharedAllowed:  user_id == owner_id || models.GetAIPolicy( handlers.Store ).MayUseShared( user_id ),
		ServerProvider: server_provider,
		ServerModel:    server_model,
		ServerEffort:   server_effort,
		Providers:      handlers.aiProviders( handlers.ownDefault ),
	}
	return
}

// @route("GET /api/account/ai", summary="Your photo-reading settings",
//     tag="Photo extraction",
//     description="Whether you have your own AI key (never the key itself) and for which service, the model "
//         "and effort you picked, whether you may use the server's key, and whose key your photos would go on now.",
//     mcp="my_ai_settings", read_only=True,
// )
func ( handlers *Handlers ) myAISettings( c fiber.Ctx , _ *apidoc.None ) ( reply *myAIReply , err error ) {
	reply = handlers.myAIReply( security.UserFrom( c ).ID )
	return
}

type setMyAIRequest struct {
	Provider  *string `json:"provider,omitempty" jsonschema:"enum=anthropic,enum=openai,enum=gemini,enum=xai,enum=openrouter" jsonschema_description:"The service your key is for. Changing it needs a key for the new one." error:"provider must be anthropic, openai, gemini, xai or openrouter"`
	APIKey    *string `json:"api_key,omitempty" jsonschema:"maxLength=300" jsonschema_description:"Your own key for that service. Leave out to keep the one you have." error:"api_key must be at most 300 characters"`
	RemoveKey bool    `json:"remove_key,omitempty" jsonschema_description:"Take your own key away."`
	Model     *string `json:"model,omitempty" jsonschema:"maxLength=128" jsonschema_description:"The model for your own key; empty is the default." error:"model must be at most 128 characters"`
	Effort    *string `json:"effort,omitempty" jsonschema:"maxLength=20" jsonschema_description:"One of the service's efforts; empty is the default." error:"effort must be at most 20 characters"`
}

// A key is a credential of your own, like a password, so only a browser
// session sets it: an API key cannot change whose bill photos go on.
//
// @route("POST /api/account/ai", summary="Set your own key, service, model and effort",
//     tag="Photo extraction", access=SESSION_ONLY, errors=(400,),
//     description="Store your own AI key for one service, sealed and never shown again, or remove it; and "
//         "pick the model and effort your own key uses. Only what you send changes. A new key also loads "
//         "that service's model list.",
// )
func ( handlers *Handlers ) setMyAI( c fiber.Ctx , in *setMyAIRequest ) ( reply *myAIReply , err error ) {
	user_id := security.UserFrom( c ).ID
	change := models.AIChange{ Provider: in.Provider , APIKey: in.APIKey , RemoveKey: in.RemoveKey , Model: in.Model , Effort: in.Effort }
	settings , set_err := models.SetUserAI( handlers.Store , handlers.cfg().SecretKey , user_id , change )
	if err = inventoryRefusal( set_err , "" ); err != nil { return }
	models_error := ""
	if in.RemoveKey {
		handlers.audit( c , logs.KindAdmin , "own ai key removed" , fiber.Map{} )
	} else if in.APIKey != nil && *in.APIKey != "" {
		handlers.audit( c , logs.KindAdmin , "own ai key set" , fiber.Map{ "provider": settings.Service() } )
		if refresh_err := handlers.refreshModels( settings.Service() , *in.APIKey ); refresh_err != nil { models_error = refresh_err.Error() }
	}
	reply = handlers.myAIReply( user_id )
	reply.ModelsError = models_error
	return
}

// @route("POST /api/account/ai/models", summary="Load a service's model list",
//     tag="Photo extraction", access=SESSION_ONLY, errors=(400,),
//     description="Ask a service which of its models can read photos, and which efforts they take, with the "
//         "key you typed (not stored) or your saved one, and keep the list for the screens.",
// )
func ( handlers *Handlers ) refreshMyModels( c fiber.Ctx , in *refreshModelsRequest ) ( reply *myAIReply , err error ) {
	user_id := security.UserFrom( c ).ID
	cfg := handlers.cfg()
	api_key , err := refreshKey( in.Provider , in.APIKey , models.GetUserAI( handlers.Store , user_id ) , func() ( string , error ) { return models.UserAPIKey( handlers.Store , cfg.SecretKey , user_id ) } )
	if err != nil { return }
	if refresh_err := handlers.refreshModels( in.Provider , api_key ); refresh_err != nil {
		err = refuse( fiber.StatusBadRequest , refresh_err.Error() )
		return
	}
	reply = handlers.myAIReply( user_id )
	return
}

type aiAccessRow struct {
	UserID      uint64 `json:"user_id"`
	DisplayName string `json:"display_name"`
	Access      string `json:"access" jsonschema:"enum=default,enum=shared,enum=own" jsonschema_description:"default follows default_shared; shared may use the server's key; own must bring their own."`
	MayShare    bool   `json:"may_share" jsonschema_description:"What it comes to: they may use the server's key."`
	OwnKeySet   bool   `json:"own_key_set" jsonschema_description:"They have a key of their own, which they always use first."`
	Owner       bool   `json:"owner" jsonschema_description:"The owner, who may always use the server's key."`
}

func ( aiAccessRow ) APIName() ( name string ) { name = "AIAccess"; return }

type aiAccessReply struct {
	DefaultShared bool          `json:"default_shared" jsonschema_description:"Everyone without their own key may use the server's, unless their row says otherwise."`
	ServerKeySet  bool          `json:"server_key_set"`
	People        []aiAccessRow `json:"people"`
	Paged
}

func ( handlers *Handlers ) aiAccessReply( page Page ) ( reply *aiAccessReply , err error ) {
	users , err := models.ListUsers( handlers.Store )
	if err != nil { return }
	policy := models.GetAIPolicy( handlers.Store )
	owner_id , _ := models.OwnerID( handlers.Store )
	server_set := models.ServerAI( handlers.Store ).KeySet()
	rows := []aiAccessRow{}
	for _ , user := range users {
		if user.Disabled() { continue }
		access := policy.People[ user.ID ]
		if access == "" { access = models.AIAccessDefault }
		rows = append( rows , aiAccessRow{
			UserID: user.ID , DisplayName: user.DisplayName , Access: access ,
			MayShare: user.ID == owner_id || policy.MayUseShared( user.ID ) ,
			OwnKeySet: models.GetUserAI( handlers.Store , user.ID ).KeySet() ,
			Owner: user.ID == owner_id ,
		} )
	}
	reply = &aiAccessReply{ DefaultShared: policy.DefaultShared , ServerKeySet: server_set }
	reply.People , reply.Paged = cutPage( rows , page )
	return
}

// @route("GET /api/admin/ai-access", summary="Who may use the server's AI key",
//     tag="Photo extraction", role=ADMIN,
//     description="The default for everyone, and each person's choice: follow the default, may use the "
//         "server's key, or must bring their own. Someone with their own key always uses it.",
//     mcp="admin_list_ai_access", read_only=True,
// )
func ( handlers *Handlers ) listAIAccess( c fiber.Ctx , in *Page ) ( reply *aiAccessReply , err error ) {
	reply , err = handlers.aiAccessReply( *in )
	return
}

type aiDefaultRequest struct {
	DefaultShared bool `json:"default_shared" jsonschema_description:"true lets everyone without their own key use the server's; false makes everyone bring their own, unless their row says otherwise." error:"default_shared must be true or false"`
}

// @route("POST /api/admin/ai-access", summary="Set the default for the server's AI key",
//     tag="Photo extraction", role=ADMIN,
//     description="Whether people without their own key may use the server's by default.",
//     mcp="admin_set_ai_default",
// )
func ( handlers *Handlers ) setAIDefault( c fiber.Ctx , in *aiDefaultRequest ) ( reply *aiAccessReply , err error ) {
	if _ , err = models.SetAIDefault( handlers.Store , in.DefaultShared ); err != nil { return }
	handlers.audit( c , logs.KindAdmin , "ai key default set" , fiber.Map{ "default_shared": in.DefaultShared } )
	reply , err = handlers.aiAccessReply( Page{} )
	return
}

type aiAccessRequest struct {
	userIDPath
	Access string `json:"access" jsonschema:"enum=default,enum=shared,enum=own" jsonschema_description:"default follows the default; shared may use the server's key; own must bring their own." error:"access must be default, shared or own"`
}

// @route("POST /api/admin/ai-access/{user_id}", summary="Set one person's use of the server's AI key",
//     bulk=True,
//     tag="Photo extraction", role=ADMIN, errors=(404,),
//     description="Let one person use the server's key, make them bring their own, or have them follow the "
//         "default.",
//     mcp="admin_set_ai_access",
// )
func ( handlers *Handlers ) setAIAccess( c fiber.Ctx , in *aiAccessRequest ) ( reply *aiAccessReply , err error ) {
	_ , set_err := models.SetAIAccess( handlers.Store , in.UserID , in.Access )
	if set_err == models.ErrNoSuchUser {
		err = refuse( fiber.StatusNotFound , "no such user" )
		return
	}
	if err = inventoryRefusal( set_err , "" ); err != nil { return }
	handlers.audit( c , logs.KindAdmin , "ai key access set" , fiber.Map{ "user_id": in.UserID , "access": in.Access } )
	reply , err = handlers.aiAccessReply( Page{} )
	return
}
