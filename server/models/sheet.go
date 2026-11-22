package models

import (
	errors "errors"
	strings "strings"
	time "time"

	db "qrcodeorganizer/server/db"
)

const maxSheetCodes = 300

var ErrSheetSize = errors.New( "a print run is 1-300 labels" )

// LabelSheet is one print run: codes minted together, waiting to be stuck on
// containers. Scanning one of them offers to set up the container it is on.
type LabelSheet struct {
	ID          uint64    `json:"id"`
	HouseholdID uint64    `json:"household_id"`
	Layout      string    `json:"layout"`
	Palette     string    `json:"palette"`
	Codes       []string  `json:"codes"`
	// Names are optional names printed under the codes, in the same order.
	Names     []string  `json:"names,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy uint64    `json:"created_by"`
}

func CreateSheet( store *db.Store , household_id uint64 , user_id uint64 , layout string , palette string , count int , names []string ) ( sheet *LabelSheet , err error ) {
	if count < 1 || count > maxSheetCodes {
		err = ErrSheetSize
		return
	}
	cleaned := []string{}
	for index , name := range names {
		if index >= count { break }
		name = strings.TrimSpace( name )
		if len( []rune( name ) ) > householdNameMax { name = string( []rune( name )[ :householdNameMax ] ) }
		cleaned = append( cleaned , name )
	}
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		next , tx_err := tx.NextSequence( db.BucketSheets )
		if tx_err != nil { return }
		codes , tx_err := mintCodesTx( tx , count , CodeEntry{ HouseholdID: household_id , SheetID: next } )
		if tx_err != nil { return }
		sheet = &LabelSheet{ ID: next , HouseholdID: household_id , Layout: layout , Palette: palette , Codes: codes , Names: cleaned , CreatedAt: time.Now().UTC() , CreatedBy: user_id }
		tx_err = tx.Put( db.BucketSheets , idKey( next ) , sheet )
		return
	} )
	if err != nil { sheet = nil }
	return
}

func GetSheet( store *db.Store , sheet_id uint64 ) ( sheet *LabelSheet , err error ) {
	sheet = &LabelSheet{}
	if err = store.Get( db.BucketSheets , idKey( sheet_id ) , sheet ); err != nil { sheet = nil }
	return
}

func ListSheets( store *db.Store , household_id uint64 ) ( sheets []*LabelSheet , err error ) {
	sheets = []*LabelSheet{}
	err = store.ForEach( db.BucketSheets , func() any { return &LabelSheet{} } , func( _ []byte , item any ) bool {
		sheet := item.( *LabelSheet )
		if sheet.HouseholdID == household_id { sheets = append( sheets , sheet ) }
		return true
	} )
	return
}

// SheetUse is how many of a sheet's codes are on containers by now.
func SheetUse( store *db.Store , sheet *LabelSheet ) ( used int , err error ) {
	for _ , code := range sheet.Codes {
		entry := &CodeEntry{}
		if get_err := store.Get( db.BucketCodes , mintedKey( code ) , entry ); get_err == nil && entry.LabelID != 0 { used += 1 }
	}
	return
}

// DeleteSheet forgets a print run. Codes already on containers stay theirs;
// the rest stop working, so stickers thrown away cannot be scanned later.
func DeleteSheet( store *db.Store , household_id uint64 , sheet_id uint64 ) ( err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		sheet := &LabelSheet{}
		if tx_err = tx.Get( db.BucketSheets , idKey( sheet_id ) , sheet ); tx_err != nil { return }
		if sheet.HouseholdID != household_id {
			tx_err = db.ErrNotFound
			return
		}
		for _ , code := range sheet.Codes {
			entry := &CodeEntry{}
			if get_err := tx.Get( db.BucketCodes , mintedKey( code ) , entry ); get_err != nil { continue }
			if entry.LabelID == 0 {
				if tx_err = tx.Delete( db.BucketCodes , mintedKey( code ) ); tx_err != nil { return }
			} else {
				entry.SheetID = 0
				if tx_err = tx.Put( db.BucketCodes , mintedKey( code ) , entry ); tx_err != nil { return }
			}
		}
		tx_err = tx.Delete( db.BucketSheets , idKey( sheet_id ) )
		return
	} )
	return
}
