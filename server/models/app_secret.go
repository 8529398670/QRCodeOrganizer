package models

// The Anthropic API key for Smart Photo Extraction. The owner sets it from the
// Settings screen rather than in config.yaml, which holds no secrets and is
// read-only to the app inside the container. It is sealed on its own with a
// subkey of secret.key (like a TOTP secret), so it stays secret even with
// server.encrypt_at_rest off, and rekey.go carries it across a rotation.
// Nothing ever hands it back: the screen only learns whether one is set.

import (
	errors "errors"
	strconv "strconv"
	strings "strings"
	time "time"

	db "qrcodeorganizer/server/db"
	encryption "qrcodeorganizer/server/encryption"
)

var aiSecretKey = []byte( "ai" )

var ErrBadAPIKey = errors.New( "that does not look like an Anthropic API key" )

type AppSecrets struct {
	AnthropicKey string     `json:"anthropic_key,omitempty"`
	SetAt        *time.Time `json:"set_at,omitempty"`
	SetBy        uint64     `json:"set_by,omitempty"`
	// Last4 is shown on the Settings screen so the owner can tell keys apart.
	Last4 string `json:"last4,omitempty"`
}

// SetAnthropicKey stores the key, or (empty) removes it.
func SetAnthropicKey( store *db.Store , secret_key string , api_key string , user_id uint64 ) ( err error ) {
	api_key = strings.TrimSpace( api_key )
	if api_key == "" {
		err = store.Delete( db.BucketAppSecrets , aiSecretKey )
		return
	}
	if looksLikeAnthropicKey( api_key ) == false {
		err = ErrBadAPIKey
		return
	}
	sealed , err := encryption.ChaChaEncryptString( encryption.SubKey( secret_key , encryption.PurposeAI ) , api_key )
	if err != nil { return }
	now := time.Now().UTC()
	err = store.Put( db.BucketAppSecrets , aiSecretKey , AppSecrets{ AnthropicKey: sealed , SetAt: &now , SetBy: user_id , Last4: api_key[ len( api_key ) - 4: ] } )
	return
}

// AnthropicKey is the stored key, or "" when none is set.
func AnthropicKey( store *db.Store , secret_key string ) ( api_key string , err error ) {
	secrets := &AppSecrets{}
	if get_err := store.Get( db.BucketAppSecrets , aiSecretKey , secrets ); get_err != nil || secrets.AnthropicKey == "" { return }
	api_key , err = encryption.ChaChaDecryptBase64String( encryption.SubKey( secret_key , encryption.PurposeAI ) , secrets.AnthropicKey )
	return
}

// AnthropicKeyInfo is what the Settings screen may know about the key.
func AnthropicKeyInfo( store *db.Store ) ( set bool , last4 string , set_at *time.Time ) {
	secrets := &AppSecrets{}
	if store.Get( db.BucketAppSecrets , aiSecretKey , secrets ) != nil || secrets.AnthropicKey == "" { return }
	set , last4 , set_at = true , secrets.Last4 , secrets.SetAt
	return
}

// ---- Each person's own key, and who may use the server's ----------------
//
// Anyone can add their own Anthropic key on their Account screen, with the
// model and effort they want; their photos then go on their own bill. Whether
// someone without one may use the server's key (the owner's bill) is the
// admins' call: a default for everyone, and a choice per person. Own keys are
// in the same bucket under "user:<id>", in the same anthropic_key field, so
// rekey.go carries them across a rotation too.

const (
	AIAccessDefault = "default"
	AIAccessShared  = "shared"
	AIAccessOwn     = "own"
)

// The models a person may pick for their own key, cheapest first.
var AIModels = []string{ "claude-haiku-5-5" , "claude-sonnet-5-5" , "claude-opus-5-5" }

var aiPolicyKey = []byte( "ai_policy" )

var (
	ErrBadModel  = errors.New( "model must be claude-haiku-5-5, claude-sonnet-5-5 or claude-opus-5-5" )
	ErrBadEffort = errors.New( "effort must be low, medium or high" )
	ErrBadAccess = errors.New( "access must be default, shared or own" )
)

// UserAI is one person's own key and choices.
type UserAI struct {
	AnthropicKey string     `json:"anthropic_key,omitempty"`
	Last4        string     `json:"last4,omitempty"`
	SetAt        *time.Time `json:"set_at,omitempty"`
	Model        string     `json:"model,omitempty"`
	Effort       string     `json:"effort,omitempty"`
}

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

func GetUserAI( store *db.Store , user_id uint64 ) ( settings *UserAI ) {
	settings = &UserAI{}
	if store.Get( db.BucketAppSecrets , userAIKey( user_id ) , settings ) != nil { settings = &UserAI{} }
	return
}

// UserAPIKey is the person's own key, opened, or "" without one.
func UserAPIKey( store *db.Store , secret_key string , user_id uint64 ) ( api_key string , err error ) {
	settings := GetUserAI( store , user_id )
	if settings.AnthropicKey == "" { return }
	api_key , err = encryption.ChaChaDecryptBase64String( encryption.SubKey( secret_key , encryption.PurposeAI ) , settings.AnthropicKey )
	return
}

func validAIModel( model string ) ( ok bool ) {
	for _ , known := range AIModels {
		if known == model { ok = true }
	}
	return
}

// SetUserAI changes a person's own settings. A nil field is left as it is; an
// empty api_key with remove_key true takes their key away; an empty model or
// effort goes back to the app's.
func SetUserAI( store *db.Store , secret_key string , user_id uint64 , api_key *string , remove_key bool , model *string , effort *string ) ( settings *UserAI , err error ) {
	settings = GetUserAI( store , user_id )
	if model != nil {
		if *model != "" && validAIModel( *model ) == false {
			err = ErrBadModel
			return
		}
		settings.Model = *model
	}
	if effort != nil {
		if *effort != "" && *effort != "low" && *effort != "medium" && *effort != "high" {
			err = ErrBadEffort
			return
		}
		settings.Effort = *effort
	}
	if remove_key {
		settings.AnthropicKey , settings.Last4 , settings.SetAt = "" , "" , nil
	} else if api_key != nil && strings.TrimSpace( *api_key ) != "" {
		trimmed := strings.TrimSpace( *api_key )
		if looksLikeAnthropicKey( trimmed ) == false {
			err = ErrBadAPIKey
			return
		}
		sealed , seal_err := encryption.ChaChaEncryptString( encryption.SubKey( secret_key , encryption.PurposeAI ) , trimmed )
		if seal_err != nil {
			err = seal_err
			return
		}
		now := time.Now().UTC()
		settings.AnthropicKey , settings.Last4 , settings.SetAt = sealed , trimmed[ len( trimmed ) - 4: ] , &now
	}
	err = store.Put( db.BucketAppSecrets , userAIKey( user_id ) , settings )
	return
}

func looksLikeAnthropicKey( api_key string ) ( ok bool ) {
	ok = strings.HasPrefix( api_key , "sk-ant-" ) && len( api_key ) >= 20 && len( api_key ) <= 300 && strings.ContainsAny( api_key , " \t\r\n" ) == false
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
