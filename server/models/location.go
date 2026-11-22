package models

import (
	errors "errors"
	time "time"

	db "qrcodeorganizer/server/db"
)

var ErrLocationLoop = errors.New( "a location cannot sit inside itself" )

// Location is a place containers sit: House › Garage › Shelf 2.
type Location struct {
	ID          uint64    `json:"id"`
	HouseholdID uint64    `json:"household_id"`
	Name        string    `json:"name"`
	ParentID    uint64    `json:"parent_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Trash
}

func GetLocation( store *db.Store , location_id uint64 ) ( location *Location , err error ) {
	location = &Location{}
	if err = store.Get( db.BucketLocations , idKey( location_id ) , location ); err != nil { location = nil }
	return
}

// ListLocations is every location in a household, deleted ones included.
func ListLocations( store *db.Store , household_id uint64 ) ( locations []*Location , err error ) {
	locations = []*Location{}
	err = store.ForEach( db.BucketLocations , func() any { return &Location{} } , func( _ []byte , item any ) bool {
		location := item.( *Location )
		if location.HouseholdID == household_id { locations = append( locations , location ) }
		return true
	} )
	return
}

// checkParent refuses a parent in another household, a deleted one, or one
// that would make a loop.
func checkLocationParent( tx *db.Tx , household_id uint64 , location_id uint64 , parent_id uint64 ) ( err error ) {
	seen := map[uint64]bool{}
	for next := parent_id ; next != 0 ; {
		if next == location_id || seen[ next ] {
			err = ErrLocationLoop
			return
		}
		seen[ next ] = true
		parent := &Location{}
		if get_err := tx.Get( db.BucketLocations , idKey( next ) , parent ); get_err != nil || parent.HouseholdID != household_id || parent.InTrash() {
			err = db.ErrNotFound
			return
		}
		next = parent.ParentID
	}
	return
}

func CreateLocation( store *db.Store , household_id uint64 , name string , parent_id uint64 ) ( location *Location , err error ) {
	cleaned , err := cleanName( name )
	if err != nil { return }
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		if tx_err = checkLocationParent( tx , household_id , 0 , parent_id ); tx_err != nil { return }
		next , tx_err := tx.NextSequence( db.BucketLocations )
		if tx_err != nil { return }
		location = &Location{ ID: next , HouseholdID: household_id , Name: cleaned , ParentID: parent_id , CreatedAt: time.Now().UTC() }
		tx_err = tx.Put( db.BucketLocations , idKey( next ) , location )
		return
	} )
	if err != nil { location = nil }
	return
}

// UpdateLocation renames a location and moves it under another (0 for top).
func UpdateLocation( store *db.Store , household_id uint64 , location_id uint64 , name *string , parent_id *uint64 ) ( location *Location , err error ) {
	var cleaned string
	if name != nil {
		if cleaned , err = cleanName( *name ); err != nil { return }
	}
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		location = &Location{}
		if tx_err = tx.Get( db.BucketLocations , idKey( location_id ) , location ); tx_err != nil { return }
		if location.HouseholdID != household_id || location.InTrash() {
			tx_err = db.ErrNotFound
			return
		}
		if name != nil { location.Name = cleaned }
		if parent_id != nil {
			if tx_err = checkLocationParent( tx , household_id , location_id , *parent_id ); tx_err != nil { return }
			location.ParentID = *parent_id
		}
		tx_err = tx.Put( db.BucketLocations , idKey( location_id ) , location )
		return
	} )
	if err != nil { location = nil }
	return
}

// TrashLocation and RestoreLocation move a location in and out of Trash.
func TrashLocation( store *db.Store , household_id uint64 , location_id uint64 , actor_id uint64 , trashed bool ) ( location *Location , err error ) {
	err = store.UpdateValue( db.BucketLocations , idKey( location_id ) , func() any { return &Location{} } , func( item any ) ( mutate_err error ) {
		location = item.( *Location )
		if location.HouseholdID != household_id {
			mutate_err = db.ErrNotFound
			return
		}
		if trashed {
			location.throwAway( actor_id )
		} else {
			location.bringBack()
		}
		return
	} )
	if err != nil { location = nil }
	return
}
