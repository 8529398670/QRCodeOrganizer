package models

import (
	errors "errors"
	math "math"
	strconv "strconv"
	strings "strings"
	time "time"

	db "qrcodeorganizer/server/db"
)

const (
	itemDescriptionMax = 4000
	itemHistoryMax     = 100
	maxTags            = 30
	maxFieldValues     = 30
	DateLayout         = "2006-01-02"
)

var (
	ErrBadDate      = errors.New( "dates look like 2026-12-31" )
	ErrBadQuantity  = errors.New( "quantity must be 0 or more" )
	ErrNoSuchLabel  = errors.New( "no such container" )
)

// ItemEvent is one line of an item's history.
type ItemEvent struct {
	At     time.Time `json:"at"`
	UserID uint64    `json:"user_id"`
	// Kind is created, edited, moved, quantity, lent, returned, deleted or
	// restored.
	Kind   string    `json:"kind"`
	Detail string    `json:"detail,omitempty"`
}

// Item is one thing in a container.
type Item struct {
	ID          uint64   `json:"id"`
	HouseholdID uint64   `json:"household_id"`
	LabelID     uint64   `json:"label_id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Photos      []string `json:"photos"`
	Tags        []string `json:"tags"`
	Quantity    float64  `json:"quantity"`
	Unit        string   `json:"unit,omitempty"`
	// LowStock: a reminder goes out when Quantity drops to this or below.
	LowStock *float64 `json:"low_stock,omitempty"`
	// Expires is a date, YYYY-MM-DD. RemindDays overrides the household's lead.
	Expires    string            `json:"expires,omitempty"`
	RemindDays *int              `json:"remind_days,omitempty"`
	Fields     map[string]string `json:"fields"`
	Barcode    string            `json:"barcode,omitempty"`
	Brand      string            `json:"brand,omitempty"`
	Serial     string            `json:"serial,omitempty"`
	// For an insurance list: what it cost and when, and its receipts.
	PurchaseDate string   `json:"purchase_date,omitempty"`
	Value        float64  `json:"value,omitempty"`
	Receipts     []string `json:"receipts"`
	// Lent out: to whom, and when it should come back.
	LentTo    string      `json:"lent_to,omitempty"`
	LentDue   string      `json:"lent_due,omitempty"`
	History   []ItemEvent `json:"history"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	CreatedBy uint64      `json:"created_by"`
	Trash
}

// ItemChanges is an edit: only the fields that are set change.
type ItemChanges struct {
	Name         *string
	Description  *string
	Photos       *[]string
	Tags         *[]string
	Quantity     *float64
	Unit         *string
	LowStock     **float64
	Expires      *string
	RemindDays   **int
	Fields       *map[string]string
	Barcode      *string
	Brand        *string
	Serial       *string
	PurchaseDate *string
	Value        *float64
	Receipts     *[]string
	LentTo       *string
	LentDue      *string
}

func ( item *Item ) note( user_id uint64 , kind string , detail string ) {
	item.History = append( item.History , ItemEvent{ At: time.Now().UTC() , UserID: user_id , Kind: kind , Detail: detail } )
	if len( item.History ) > itemHistoryMax { item.History = item.History[ len( item.History ) - itemHistoryMax: ] }
}

func GetItem( store *db.Store , item_id uint64 ) ( item *Item , err error ) {
	item = &Item{}
	if err = store.Get( db.BucketItems , idKey( item_id ) , item ); err != nil { item = nil }
	return
}

// ListItems is every item in a household, deleted ones included.
func ListItems( store *db.Store , household_id uint64 ) ( items []*Item , err error ) {
	items = []*Item{}
	err = store.ForEach( db.BucketItems , func() any { return &Item{} } , func( _ []byte , found any ) bool {
		item := found.( *Item )
		if item.HouseholdID == household_id { items = append( items , item ) }
		return true
	} )
	return
}

func checkDate( value string ) ( cleaned string , err error ) {
	cleaned = strings.TrimSpace( value )
	if cleaned == "" { return }
	if _ , parse_err := time.Parse( DateLayout , cleaned ); parse_err != nil { err = ErrBadDate }
	return
}

func cleanText( value string , most int ) ( cleaned string , err error ) {
	cleaned = strings.TrimSpace( value )
	if len( cleaned ) > most { err = ErrTooLong }
	return
}

func goodNumber( value float64 ) ( result bool ) {
	result = math.IsNaN( value ) == false && math.IsInf( value , 0 ) == false && value >= 0 && value < 1e12
	return
}

func applyItemChanges( item *Item , changes ItemChanges , fields []FieldDef ) ( err error ) {
	if changes.Name != nil {
		if item.Name , err = cleanName( *changes.Name ); err != nil { return }
	}
	if changes.Description != nil {
		if item.Description , err = cleanText( *changes.Description , itemDescriptionMax ); err != nil { return }
	}
	if changes.Photos != nil {
		if item.Photos , err = cleanList( *changes.Photos , maxPhotos , 64 ); err != nil { return }
	}
	if changes.Receipts != nil {
		if item.Receipts , err = cleanList( *changes.Receipts , maxPhotos , 64 ); err != nil { return }
	}
	if changes.Tags != nil {
		tags := []string{}
		for _ , tag := range *changes.Tags { tags = append( tags , strings.ToLower( tag ) ) }
		if item.Tags , err = cleanList( tags , maxTags , 40 ); err != nil { return }
	}
	if changes.Quantity != nil {
		if goodNumber( *changes.Quantity ) == false {
			err = ErrBadQuantity
			return
		}
		item.Quantity = *changes.Quantity
	}
	if changes.Unit != nil {
		if item.Unit , err = cleanText( *changes.Unit , 20 ); err != nil { return }
	}
	if changes.LowStock != nil {
		if *changes.LowStock != nil && goodNumber( **changes.LowStock ) == false {
			err = ErrBadQuantity
			return
		}
		item.LowStock = *changes.LowStock
	}
	if changes.Expires != nil {
		if item.Expires , err = checkDate( *changes.Expires ); err != nil { return }
	}
	if changes.RemindDays != nil {
		if *changes.RemindDays != nil && ( **changes.RemindDays < 0 || **changes.RemindDays > 365 ) {
			err = errors.New( "remind_days must be 0-365" )
			return
		}
		item.RemindDays = *changes.RemindDays
	}
	if changes.Fields != nil {
		known := map[string]FieldDef{}
		for _ , field := range fields { known[ field.ID ] = field }
		values := map[string]string{}
		for id , value := range *changes.Fields {
			field , is_known := known[ id ]
			value = strings.TrimSpace( value )
			if is_known == false || value == "" { continue }
			if len( value ) > 500 {
				err = ErrTooLong
				return
			}
			if field.Type == "date" {
				if _ , err = checkDate( value ); err != nil { return }
			}
			values[ id ] = value
		}
		if len( values ) > maxFieldValues {
			err = ErrTooMany
			return
		}
		item.Fields = values
	}
	for _ , text := range []struct{ into *string ; from *string ; most int }{
		{ &item.Barcode , changes.Barcode , 200 },
		{ &item.Brand , changes.Brand , 80 },
		{ &item.Serial , changes.Serial , 120 },
		{ &item.LentTo , changes.LentTo , 80 },
	} {
		if text.from == nil { continue }
		if *text.into , err = cleanText( *text.from , text.most ); err != nil { return }
	}
	if changes.PurchaseDate != nil {
		if item.PurchaseDate , err = checkDate( *changes.PurchaseDate ); err != nil { return }
	}
	if changes.LentDue != nil {
		if item.LentDue , err = checkDate( *changes.LentDue ); err != nil { return }
	}
	if changes.Value != nil {
		if goodNumber( *changes.Value ) == false {
			err = errors.New( "value must be 0 or more" )
			return
		}
		item.Value = *changes.Value
	}
	return
}

func liveLabelTx( tx *db.Tx , household_id uint64 , label_id uint64 ) ( label *Label , err error ) {
	label = &Label{}
	if get_err := tx.Get( db.BucketLabels , idKey( label_id ) , label ); get_err != nil || label.HouseholdID != household_id || label.InTrash() {
		label = nil
		err = ErrNoSuchLabel
	}
	return
}

// CreateItems puts new items in a container in one transaction: by hand one
// at a time, or a whole photo's worth from Smart Photo Extraction.
func CreateItems( store *db.Store , household *Household , label_id uint64 , user_id uint64 , all []ItemChanges ) ( items []*Item , err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		items = []*Item{}
		if _ , tx_err = liveLabelTx( tx , household.ID , label_id ); tx_err != nil { return }
		for _ , changes := range all {
			next , next_err := tx.NextSequence( db.BucketItems )
			if next_err != nil {
				tx_err = next_err
				return
			}
			now := time.Now().UTC()
			item := &Item{ ID: next , HouseholdID: household.ID , LabelID: label_id , Quantity: 1 , CreatedAt: now , UpdatedAt: now , CreatedBy: user_id }
			if changes.Name == nil {
				tx_err = ErrBadName
				return
			}
			if tx_err = applyItemChanges( item , changes , household.Fields ); tx_err != nil { return }
			item.note( user_id , "created" , "" )
			if tx_err = tx.Put( db.BucketItems , idKey( next ) , item ); tx_err != nil { return }
			items = append( items , item )
		}
		return
	} )
	if err != nil { items = nil }
	return
}

// UpdateItem edits an item that is not in Trash, noting what kind of change it
// was in its history.
func UpdateItem( store *db.Store , household *Household , item_id uint64 , user_id uint64 , changes ItemChanges ) ( item *Item , before *Item , err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		item = &Item{}
		if tx_err = tx.Get( db.BucketItems , idKey( item_id ) , item ); tx_err != nil { return }
		if item.HouseholdID != household.ID || item.InTrash() {
			tx_err = db.ErrNotFound
			return
		}
		copied := *item
		before = &copied
		if tx_err = applyItemChanges( item , changes , household.Fields ); tx_err != nil { return }
		switch {
		case changes.LentTo != nil && before.LentTo == "" && item.LentTo != "":
			item.note( user_id , "lent" , item.LentTo )
		case changes.LentTo != nil && before.LentTo != "" && item.LentTo == "":
			item.note( user_id , "returned" , before.LentTo )
		case changes.Quantity != nil && before.Quantity != item.Quantity:
			item.note( user_id , "quantity" , formatAmount( before.Quantity ) + " → " + formatAmount( item.Quantity ) )
		default:
			item.note( user_id , "edited" , "" )
		}
		item.UpdatedAt = time.Now().UTC()
		tx_err = tx.Put( db.BucketItems , idKey( item_id ) , item )
		return
	} )
	if err != nil { item , before = nil , nil }
	return
}

// AdjustQuantity adds delta (which may be negative) to an item's count, never
// below zero, in one transaction so two people taking one each both count.
func AdjustQuantity( store *db.Store , household_id uint64 , item_id uint64 , user_id uint64 , delta float64 ) ( item *Item , err error ) {
	if math.IsNaN( delta ) || math.IsInf( delta , 0 ) || math.Abs( delta ) > 1e9 {
		err = ErrBadQuantity
		return
	}
	err = store.UpdateValue( db.BucketItems , idKey( item_id ) , func() any { return &Item{} } , func( found any ) ( mutate_err error ) {
		item = found.( *Item )
		if item.HouseholdID != household_id || item.InTrash() {
			mutate_err = db.ErrNotFound
			return
		}
		old := item.Quantity
		item.Quantity = math.Max( 0 , item.Quantity + delta )
		item.note( user_id , "quantity" , formatAmount( old ) + " → " + formatAmount( item.Quantity ) )
		item.UpdatedAt = time.Now().UTC()
		return
	} )
	if err != nil { item = nil }
	return
}

// MoveItems puts items in another container of the same household.
func MoveItems( store *db.Store , household_id uint64 , item_ids []uint64 , label_id uint64 , user_id uint64 ) ( moved []*Item , err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		moved = []*Item{}
		target , tx_err := liveLabelTx( tx , household_id , label_id )
		if tx_err != nil { return }
		for _ , item_id := range item_ids {
			item := &Item{}
			if get_err := tx.Get( db.BucketItems , idKey( item_id ) , item ); get_err != nil || item.HouseholdID != household_id || item.InTrash() {
				tx_err = db.ErrNotFound
				return
			}
			if item.LabelID == label_id { continue }
			from := item.LabelID
			item.LabelID = label_id
			item.note( user_id , "moved" , labelNameTx( tx , from ) + " → " + target.Name )
			item.UpdatedAt = time.Now().UTC()
			if tx_err = tx.Put( db.BucketItems , idKey( item_id ) , item ); tx_err != nil { return }
			moved = append( moved , item )
		}
		return
	} )
	if err != nil { moved = nil }
	return
}

// MoveContents moves everything live in one container into another.
func MoveContents( store *db.Store , household_id uint64 , from_id uint64 , to_id uint64 , user_id uint64 ) ( moved []*Item , err error ) {
	ids := []uint64{}
	items , err := ListItems( store , household_id )
	if err != nil { return }
	for _ , item := range items {
		if item.LabelID == from_id && item.InTrash() == false { ids = append( ids , item.ID ) }
	}
	moved , err = MoveItems( store , household_id , ids , to_id , user_id )
	return
}

func labelNameTx( tx *db.Tx , label_id uint64 ) ( name string ) {
	label := &Label{}
	if tx.Get( db.BucketLabels , idKey( label_id ) , label ) == nil { name = label.Name }
	return
}

// TrashItem moves an item in or out of Trash. Restoring one whose container is
// in Trash too is refused: put the container back first.
func TrashItem( store *db.Store , household_id uint64 , item_id uint64 , actor_id uint64 , trashed bool ) ( item *Item , err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		item = &Item{}
		if tx_err = tx.Get( db.BucketItems , idKey( item_id ) , item ); tx_err != nil { return }
		if item.HouseholdID != household_id {
			tx_err = db.ErrNotFound
			return
		}
		if trashed {
			if item.InTrash() == false { item.note( actor_id , "deleted" , "" ) }
			item.throwAway( actor_id )
		} else {
			if _ , tx_err = liveLabelTx( tx , household_id , item.LabelID ); tx_err != nil {
				tx_err = errors.New( "its container is in Trash: restore the container first" )
				return
			}
			if item.InTrash() { item.note( actor_id , "restored" , "" ) }
			item.bringBack()
		}
		tx_err = tx.Put( db.BucketItems , idKey( item_id ) , item )
		return
	} )
	if err != nil { item = nil }
	return
}

// PurgeTrash removes what has been in Trash longer than keep, for the hourly
// job: items, then containers (with their codes), then locations.
func PurgeTrash( store *db.Store , keep time.Duration ) ( removed int , err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		count , tx_err := tx.DeleteWhere( db.BucketItems , func() any { return &Item{} } , func( _ []byte , item any ) bool {
			return item.( *Item ).expired( keep )
		} )
		if tx_err != nil { return }
		removed += count
		labels := []*Label{}
		tx_err = tx.ForEach( db.BucketLabels , func() any { return &Label{} } , func( _ []byte , item any ) bool {
			if label := item.( *Label ); label.expired( keep ) { labels = append( labels , label ) }
			return true
		} )
		if tx_err != nil { return }
		for _ , label := range labels {
			if tx_err = purgeLabelTx( tx , label ); tx_err != nil { return }
		}
		removed += len( labels )
		count , tx_err = tx.DeleteWhere( db.BucketLocations , func() any { return &Location{} } , func( _ []byte , item any ) bool {
			return item.( *Location ).expired( keep )
		} )
		removed += count
		return
	} )
	return
}

func formatAmount( value float64 ) ( text string ) {
	text = strings.TrimRight( strings.TrimRight( strconv.FormatFloat( value , 'f' , 3 , 64 ) , "0" ) , "." )
	if text == "" { text = "0" }
	return
}
