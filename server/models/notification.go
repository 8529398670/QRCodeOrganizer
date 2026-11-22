package models

import (
	time "time"

	db "qrcodeorganizer/server/db"
)

// There is no email. A reminder is a notification here (the bell in the
// navbar), a webhook event, and a line in the calendar feed.
const (
	NoticeExpiring = "expiring"
	NoticeExpired  = "expired"
	NoticeLowStock = "low_stock"
	NoticeOverdue  = "overdue"

	notificationKeep = 60 * 24 * time.Hour
)

type Notification struct {
	ID          uint64    `json:"id"`
	UserID      uint64    `json:"user_id"`
	HouseholdID uint64    `json:"household_id"`
	ItemID      uint64    `json:"item_id"`
	LabelID     uint64    `json:"label_id"`
	Kind        string    `json:"kind"`
	ItemName    string    `json:"item_name"`
	// Due is the date it is about: the expiry or the loan's due date.
	Due string `json:"due,omitempty"`
	// Key stops the same reminder going out twice.
	Key       string    `json:"key"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

// Notify stores reminders for everyone listed, skipping any whose key they
// already have. Answers how many were new.
func Notify( store *db.Store , user_ids []uint64 , template Notification ) ( added int , err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		have := map[uint64]bool{}
		tx_err = tx.ForEach( db.BucketNotifications , func() any { return &Notification{} } , func( _ []byte , item any ) bool {
			found := item.( *Notification )
			if found.Key == template.Key { have[ found.UserID ] = true }
			return true
		} )
		if tx_err != nil { return }
		for _ , user_id := range user_ids {
			if have[ user_id ] { continue }
			next , next_err := tx.NextSequence( db.BucketNotifications )
			if next_err != nil {
				tx_err = next_err
				return
			}
			notice := template
			notice.ID = next
			notice.UserID = user_id
			notice.CreatedAt = time.Now().UTC()
			if tx_err = tx.Put( db.BucketNotifications , idKey( next ) , notice ); tx_err != nil { return }
			added += 1
		}
		return
	} )
	return
}

// ListNotifications is one person's, newest first.
func ListNotifications( store *db.Store , user_id uint64 ) ( notices []*Notification , err error ) {
	notices = []*Notification{}
	err = store.ForEach( db.BucketNotifications , func() any { return &Notification{} } , func( _ []byte , item any ) bool {
		notice := item.( *Notification )
		if notice.UserID == user_id { notices = append( [] *Notification{ notice } , notices... ) }
		return true
	} )
	return
}

// MarkRead marks one (or, id 0, all) of a person's notifications read.
func MarkRead( store *db.Store , user_id uint64 , notice_id uint64 ) ( changed int , err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		marked := []*Notification{}
		tx_err = tx.ForEach( db.BucketNotifications , func() any { return &Notification{} } , func( _ []byte , item any ) bool {
			notice := item.( *Notification )
			if notice.UserID == user_id && notice.Read == false && ( notice_id == 0 || notice.ID == notice_id ) {
				notice.Read = true
				marked = append( marked , notice )
			}
			return true
		} )
		if tx_err != nil { return }
		for _ , notice := range marked {
			if tx_err = tx.Put( db.BucketNotifications , idKey( notice.ID ) , notice ); tx_err != nil { return }
		}
		changed = len( marked )
		return
	} )
	return
}

// PruneNotifications drops read ones older than two months.
func PruneNotifications( store *db.Store ) ( removed int , err error ) {
	removed , err = store.DeleteWhere( db.BucketNotifications , func() any { return &Notification{} } , func( _ []byte , item any ) bool {
		notice := item.( *Notification )
		return notice.Read && time.Since( notice.CreatedAt ) > notificationKeep
	} )
	return
}
