package models

// A calendar feed: an address a phone's calendar subscribes to, listing the
// expiry dates and loan due dates of the households its owner is in. Calendar
// apps cannot sign in, so the address itself is the credential, shaped
// <id>.<secret> like the other ways in, and only the secret's hash is kept.
// Making a new one, or switching it off, ends the old address.

import (
	strings "strings"
	time "time"

	db "qrcodeorganizer/server/db"
	encryption "qrcodeorganizer/server/encryption"
)

type FeedToken struct {
	ID         string    `json:"id"`
	UserID     uint64    `json:"user_id"`
	SecretHash string    `json:"secret_hash"`
	CreatedAt  time.Time `json:"created_at"`
}

// IssueFeedToken replaces the user's feed address with a new one.
func IssueFeedToken( store *db.Store , user_id uint64 ) ( credential string , err error ) {
	id := encryption.GenerateURLToken( 9 )
	secret := encryption.GenerateURLToken( 24 )
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		if _ , tx_err = tx.DeleteWhere( db.BucketFeedTokens , func() any { return &FeedToken{} } , func( _ []byte , item any ) bool {
			return item.( *FeedToken ).UserID == user_id
		} ); tx_err != nil { return }
		tx_err = tx.Put( db.BucketFeedTokens , []byte( id ) , FeedToken{ ID: id , UserID: user_id , SecretHash: encryption.Sha256Hex( secret ) , CreatedAt: time.Now().UTC() } )
		return
	} )
	if err == nil { credential = id + "." + secret }
	return
}

// RevokeFeedToken switches the user's feed address off.
func RevokeFeedToken( store *db.Store , user_id uint64 ) ( removed int , err error ) {
	removed , err = store.DeleteWhere( db.BucketFeedTokens , func() any { return &FeedToken{} } , func( _ []byte , item any ) bool {
		return item.( *FeedToken ).UserID == user_id
	} )
	return
}

// HasFeedToken says whether the user has a feed address switched on.
func HasFeedToken( store *db.Store , user_id uint64 ) ( has bool , err error ) {
	err = store.ForEach( db.BucketFeedTokens , func() any { return &FeedToken{} } , func( _ []byte , item any ) bool {
		if item.( *FeedToken ).UserID == user_id { has = true }
		return has == false
	} )
	return
}

// ResolveFeedToken is the account a feed address belongs to, or 0.
func ResolveFeedToken( store *db.Store , credential string ) ( user_id uint64 ) {
	id , secret , found := strings.Cut( credential , "." )
	if found == false || id == "" || secret == "" { return }
	token := &FeedToken{}
	if store.Get( db.BucketFeedTokens , []byte( id ) , token ) != nil { return }
	if encryption.ConstantTimeEqual( token.SecretHash , encryption.Sha256Hex( secret ) ) == false { return }
	user , err := GetUser( store , token.UserID )
	if err != nil || user.Disabled() { return }
	user_id = user.ID
	return
}
