package models

// The AI keys for Smart Photo Extraction: the server's, which the owner sets
// on the Settings screen, and each person's own, set on their Account screen.
// config.yaml holds no secrets and is read-only to the app inside the
// container, so keys live here instead. Each is sealed on its own with a
// subkey of secret.key (like a TOTP secret), so it stays secret even with
// server.encrypt_at_rest off, and rekey.go carries it across a rotation.
// Nothing ever hands a key back: the screens only learn whether one is set.
//
// A key belongs to one service (catalog.Providers): Anthropic, OpenAI,
// Gemini, xAI or OpenRouter. Records from before there was a choice have no
// provider, which means Anthropic.

import (
	errors "errors"
	fmt "fmt"
	strconv "strconv"
	strings "strings"
	time "time"

	catalog "qrcodeorganizer/server/ai/catalog"
	db "qrcodeorganizer/server/db"
	encryption "qrcodeorganizer/server/encryption"
)

var aiSecretKey = []byte( "ai" )

var (
	ErrBadAPIKey   = catalog.ErrBadKey
	ErrBadModel    = catalog.ErrBadModel
	ErrBadEffort   = catalog.ErrBadEffort
	ErrBadProvider = catalog.ErrBadProvider
	ErrNeedKey     = errors.New( "a key belongs to one service: enter a key for the service you picked" )
	ErrBadAccess   = errors.New( "access must be default, shared or own" )
)

// AISettings is one key and the choices that go with it: the server's, or
// one person's. The sealed key keeps its old field name, anthropic_key, so
// keys stored before other services existed still open, and rekey.go finds
// them all under one name.
type AISettings struct {
	Key      string     `json:"anthropic_key,omitempty"`
	Provider string     `json:"provider,omitempty"`
	Model    string     `json:"model,omitempty"`
	Effort   string     `json:"effort,omitempty"`
	SetAt    *time.Time `json:"set_at,omitempty"`
	SetBy    uint64     `json:"set_by,omitempty"`
	// Last4 is shown so keys can be told apart.
	Last4 string `json:"last4,omitempty"`
}

// KeySet is whether a key is stored.
func ( settings *AISettings ) KeySet() ( set bool ) {
	set = settings.Key != ""
	return
}

// Service is the provider id, "" read as Anthropic.
func ( settings *AISettings ) Service() ( provider string ) {
	provider = catalog.Normal( settings.Provider )
	return
}

// AIChange is what one save changes. A nil field is left as it is.
type AIChange struct {
	Provider  *string
	APIKey    *string
	RemoveKey bool
	Model     *string
	Effort    *string
}

// apply changes settings in place. Picking another service needs a key for
// it, unless there is no key to begin with; the old model and effort belong
// to the old service and are dropped, unless this save names new ones.
func ( settings *AISettings ) apply( secret_key string , user_id uint64 , change AIChange ) ( err error ) {
	provider , _ := catalog.Lookup( settings.Provider )
	new_key := ""
	if change.APIKey != nil { new_key = strings.TrimSpace( *change.APIKey ) }
	if change.Provider != nil && catalog.Normal( *change.Provider ) != provider.ID {
		next , ok := catalog.Lookup( *change.Provider )
		if ok == false {
			err = ErrBadProvider
			return
		}
		if settings.KeySet() && new_key == "" && change.RemoveKey == false {
			err = fmt.Errorf( "%w (the one saved is for %s)" , ErrNeedKey , provider.Name )
			return
		}
		provider = next
		settings.Model , settings.Effort = "" , ""
	}
	if change.Model != nil {
		if err = catalog.CheckModel( strings.TrimSpace( *change.Model ) ); err != nil { return }
		settings.Model = strings.TrimSpace( *change.Model )
	}
	if change.Effort != nil {
		if err = provider.CheckEffort( *change.Effort ); err != nil { return }
		settings.Effort = *change.Effort
	}
	if change.RemoveKey {
		settings.Key , settings.Last4 , settings.SetAt , settings.SetBy = "" , "" , nil , 0
	} else if new_key != "" {
		if err = provider.CheckKey( new_key ); err != nil { return }
		sealed , seal_err := encryption.ChaChaEncryptString( encryption.SubKey( secret_key , encryption.PurposeAI ) , new_key )
		if seal_err != nil {
			err = seal_err
			return
		}
		now := time.Now().UTC()
		settings.Key , settings.Last4 , settings.SetAt , settings.SetBy = sealed , new_key[ len( new_key ) - 4: ] , &now , user_id
	}
	settings.Provider = provider.ID
	return
}

func ( settings *AISettings ) open( secret_key string ) ( api_key string , err error ) {
	if settings.Key == "" { return }
	api_key , err = encryption.ChaChaDecryptBase64String( encryption.SubKey( secret_key , encryption.PurposeAI ) , settings.Key )
	return
}

// ---- The server's key ----------------------------------------------------

// ServerAI is the server's key and choices (the key still sealed).
func ServerAI( store *db.Store ) ( settings *AISettings ) {
	settings = &AISettings{}
	if store.Get( db.BucketAppSecrets , aiSecretKey , settings ) != nil { settings = &AISettings{} }
	return
}

// ServerAPIKey is the server's key, opened, or "" when none is set.
func ServerAPIKey( store *db.Store , secret_key string ) ( api_key string , err error ) {
	api_key , err = ServerAI( store ).open( secret_key )
	return
}

// SetServerAI changes the server's key and choices.
func SetServerAI( store *db.Store , secret_key string , user_id uint64 , change AIChange ) ( settings *AISettings , err error ) {
	settings = ServerAI( store )
	if err = settings.apply( secret_key , user_id , change ); err != nil { return }
	err = store.Put( db.BucketAppSecrets , aiSecretKey , settings )
	return
}

// ---- Each person's own key, and who may use the server's ----------------
//
// Anyone can add their own key on their Account screen, for the service,
// model and effort they want; their photos then go on their own bill.
// Whether someone without one may use the server's key (the owner's bill) is
// the admins' call: a default for everyone, and a choice per person. Own keys
// are in the same bucket under "user:<id>", in the same anthropic_key field,
// so rekey.go carries them across a rotation too.

const (
	AIAccessDefault = "default"
	AIAccessShared  = "shared"
	AIAccessOwn     = "own"
)

var aiPolicyKey = []byte( "ai_policy" )

// AIPolicy is who may use the server's key.
type AIPolicy struct {
	// DefaultShared lets everyone without their own key use the server's,
	// unless their own entry says otherwise.
	DefaultShared bool              `json:"default_shared"`
	People        map[uint64]string `json:"people,omitempty"`
}

func userAIKey( user_id uint64 ) ( key []byte ) {
	key = []byte( "user:" + strconv.FormatUint( user_id , 10 ) )
	return
}

func GetUserAI( store *db.Store , user_id uint64 ) ( settings *AISettings ) {
	settings = &AISettings{}
	if store.Get( db.BucketAppSecrets , userAIKey( user_id ) , settings ) != nil { settings = &AISettings{} }
	return
}

// UserAPIKey is the person's own key, opened, or "" without one.
func UserAPIKey( store *db.Store , secret_key string , user_id uint64 ) ( api_key string , err error ) {
	api_key , err = GetUserAI( store , user_id ).open( secret_key )
	return
}

// SetUserAI changes a person's own key and choices. An empty model or effort
// goes back to the default.
func SetUserAI( store *db.Store , secret_key string , user_id uint64 , change AIChange ) ( settings *AISettings , err error ) {
	settings = GetUserAI( store , user_id )
	if err = settings.apply( secret_key , user_id , change ); err != nil { return }
	err = store.Put( db.BucketAppSecrets , userAIKey( user_id ) , settings )
	return
}

// ---- Each service's model list ------------------------------------------
//
// Refreshed from the service with whichever key asked, and kept so the
// screens can offer the list without asking again. Model names are not
// secret.

type AIModelList struct {
	Models      []catalog.ModelInfo `json:"models"`
	RefreshedAt *time.Time          `json:"refreshed_at,omitempty"`
}

func aiModelsKey( provider string ) ( key []byte ) {
	key = []byte( "models:" + catalog.Normal( provider ) )
	return
}

// AIModels is a service's last refreshed list, or what the app knows of it
// before anyone has refreshed.
func AIModels( store *db.Store , provider string ) ( list *AIModelList ) {
	list = &AIModelList{}
	if store.Get( db.BucketAppSecrets , aiModelsKey( provider ) , list ) != nil || len( list.Models ) == 0 {
		list = &AIModelList{ Models: catalog.BuiltIn( provider ) }
	}
	return
}

func SaveAIModels( store *db.Store , provider string , found []catalog.ModelInfo ) ( list *AIModelList , err error ) {
	now := time.Now().UTC()
	list = &AIModelList{ Models: found , RefreshedAt: &now }
	err = store.Put( db.BucketAppSecrets , aiModelsKey( provider ) , list )
	return
}

func GetAIPolicy( store *db.Store ) ( policy *AIPolicy ) {
	policy = &AIPolicy{}
	if store.Get( db.BucketAppSecrets , aiPolicyKey , policy ) != nil { policy = &AIPolicy{} }
	if policy.People == nil { policy.People = map[uint64]string{} }
	return
}

// MayUseShared is whether a person may have their photos read with the
// server's key.
func ( policy *AIPolicy ) MayUseShared( user_id uint64 ) ( allowed bool ) {
	switch policy.People[ user_id ] {
	case AIAccessShared:
		allowed = true
	case AIAccessOwn:
		allowed = false
	default:
		allowed = policy.DefaultShared
	}
	return
}

// SetAIDefault and SetAIAccess are the admins' two switches.
func SetAIDefault( store *db.Store , shared bool ) ( policy *AIPolicy , err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		policy = &AIPolicy{}
		if tx.Get( db.BucketAppSecrets , aiPolicyKey , policy ) != nil { policy = &AIPolicy{} }
		policy.DefaultShared = shared
		tx_err = tx.Put( db.BucketAppSecrets , aiPolicyKey , policy )
		return
	} )
	return
}

func SetAIAccess( store *db.Store , user_id uint64 , access string ) ( policy *AIPolicy , err error ) {
	if access != AIAccessDefault && access != AIAccessShared && access != AIAccessOwn {
		err = ErrBadAccess
		return
	}
	if _ , get_err := GetUser( store , user_id ); get_err != nil {
		err = ErrNoSuchUser
		return
	}
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		policy = &AIPolicy{}
		if tx.Get( db.BucketAppSecrets , aiPolicyKey , policy ) != nil { policy = &AIPolicy{} }
		if policy.People == nil { policy.People = map[uint64]string{} }
		if access == AIAccessDefault {
			delete( policy.People , user_id )
		} else {
			policy.People[ user_id ] = access
		}
		tx_err = tx.Put( db.BucketAppSecrets , aiPolicyKey , policy )
		return
	} )
	return
}
