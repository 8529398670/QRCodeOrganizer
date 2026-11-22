package models

import (
	errors "errors"
	strings "strings"
	time "time"

	db "qrcodeorganizer/server/db"
)

const (
	MoveNone      = ""
	MovePacked    = "packed"
	MoveInTransit = "in_transit"
	MoveUnpacked  = "unpacked"

	notesMax     = 4000
	maxPhotos    = 30
	maxClaimed   = 20
	maxFlags     = 10
)

var (
	ErrLabelLoop    = errors.New( "a container cannot sit inside itself" )
	ErrBadMove      = errors.New( "moving status must be packed, in_transit, unpacked or empty" )
	ErrTooMany      = errors.New( "too many of those on one record" )
	ErrTooLong      = errors.New( "that text is too long" )
)

// Label is a container with a SmartLabel on it: a bin, box, drawer or shelf.
type Label struct {
	ID          uint64 `json:"id"`
	HouseholdID uint64 `json:"household_id"`
	// Code is the container's own code (see code.go).
	Code string `json:"code"`
	// Claimed are other codes that also open it: a bought sticker, a barcode.
	Claimed       []string `json:"claimed"`
	Name          string   `json:"name"`
	Color         string   `json:"color,omitempty"`
	ContainerType string   `json:"container_type,omitempty"`
	LocationID    uint64   `json:"location_id,omitempty"`
	// ParentID is the container this one sits in: a box on a shelf unit.
	ParentID    uint64    `json:"parent_id,omitempty"`
	Photos      []string  `json:"photos"`
	Notes       string    `json:"notes,omitempty"`
	MoveStatus  string    `json:"move_status,omitempty"`
	Flags       []string  `json:"flags"`
	Destination string    `json:"destination,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   uint64    `json:"created_by"`
	Trash
}

// LabelChanges is an edit: only the fields that are set change.
type LabelChanges struct {
	Name          *string
	Color         *string
	ContainerType *string
	LocationID    *uint64
	ParentID      *uint64
	Photos        *[]string
	Notes         *string
	MoveStatus    *string
	Flags         *[]string
	Destination   *string
}

func GetLabel( store *db.Store , label_id uint64 ) ( label *Label , err error ) {
	label = &Label{}
	if err = store.Get( db.BucketLabels , idKey( label_id ) , label ); err != nil { label = nil }
	return
}

// ListLabels is every container in a household, deleted ones included.
func ListLabels( store *db.Store , household_id uint64 ) ( labels []*Label , err error ) {
	labels = []*Label{}
	err = store.ForEach( db.BucketLabels , func() any { return &Label{} } , func( _ []byte , item any ) bool {
		label := item.( *Label )
		if label.HouseholdID == household_id { labels = append( labels , label ) }
		return true
	} )
	return
}

func cleanList( values []string , most int , each_max int ) ( cleaned []string , err error ) {
	cleaned = []string{}
	seen := map[string]bool{}
	for _ , value := range values {
		value = strings.TrimSpace( value )
		if value == "" || seen[ value ] { continue }
		if len( value ) > each_max {
			err = ErrTooLong
			return
		}
		seen[ value ] = true
		cleaned = append( cleaned , value )
	}
	if len( cleaned ) > most { err = ErrTooMany }
	return
}

func applyLabelChanges( tx *db.Tx , label *Label , changes LabelChanges ) ( err error ) {
	if changes.Name != nil {
		if label.Name , err = cleanName( *changes.Name ); err != nil { return }
	}
	if changes.Color != nil {
		color := strings.ToLower( strings.TrimSpace( *changes.Color ) )
		if color != "" && ValidColor( color ) == false {
			err = errors.New( "color must look like #1e88e5" )
			return
		}
		label.Color = color
	}
	if changes.ContainerType != nil {
		kind := strings.TrimSpace( *changes.ContainerType )
		if len( kind ) > 40 {
			err = ErrTooLong
			return
		}
		label.ContainerType = kind
	}
	if changes.LocationID != nil {
		if *changes.LocationID != 0 {
			location := &Location{}
			if get_err := tx.Get( db.BucketLocations , idKey( *changes.LocationID ) , location ); get_err != nil || location.HouseholdID != label.HouseholdID || location.InTrash() {
				err = errors.New( "no such location" )
				return
			}
		}
		label.LocationID = *changes.LocationID
	}
	if changes.ParentID != nil {
		seen := map[uint64]bool{}
		for next := *changes.ParentID ; next != 0 ; {
			if next == label.ID || seen[ next ] {
				err = ErrLabelLoop
				return
			}
			seen[ next ] = true
			parent := &Label{}
			if get_err := tx.Get( db.BucketLabels , idKey( next ) , parent ); get_err != nil || parent.HouseholdID != label.HouseholdID || parent.InTrash() {
				err = errors.New( "no such container to put it in" )
				return
			}
			next = parent.ParentID
		}
		label.ParentID = *changes.ParentID
	}
	if changes.Photos != nil {
		if label.Photos , err = cleanList( *changes.Photos , maxPhotos , 64 ); err != nil { return }
	}
	if changes.Notes != nil {
		notes := strings.TrimSpace( *changes.Notes )
		if len( notes ) > notesMax {
			err = ErrTooLong
			return
		}
		label.Notes = notes
	}
	if changes.MoveStatus != nil {
		switch *changes.MoveStatus {
		case MoveNone , MovePacked , MoveInTransit , MoveUnpacked:
			label.MoveStatus = *changes.MoveStatus
		default:
			err = ErrBadMove
			return
		}
	}
	if changes.Flags != nil {
		if label.Flags , err = cleanList( *changes.Flags , maxFlags , 40 ); err != nil { return }
	}
	if changes.Destination != nil {
		destination := strings.TrimSpace( *changes.Destination )
		if len( []rune( destination ) ) > 80 {
			err = ErrTooLong
			return
		}
		label.Destination = destination
	}
	return
}

// CreateLabel makes a container. code is a printed code waiting on one of the
// household's sheets, or a value to claim (a bought sticker); empty mints a
// fresh code.
func CreateLabel( store *db.Store , household_id uint64 , user_id uint64 , code string , changes LabelChanges ) ( label *Label , err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		next , tx_err := tx.NextSequence( db.BucketLabels )
		if tx_err != nil { return }
		now := time.Now().UTC()
		label = &Label{ ID: next , HouseholdID: household_id , CreatedAt: now , UpdatedAt: now , CreatedBy: user_id }
		if changes.Name == nil {
			untitled := "Container"
			changes.Name = &untitled
		}
		if tx_err = applyLabelChanges( tx , label , changes ); tx_err != nil { return }
		if strings.TrimSpace( code ) == "" {
			minted , mint_err := mintCodesTx( tx , 1 , CodeEntry{ HouseholdID: household_id , LabelID: next } )
			if mint_err != nil {
				tx_err = mint_err
				return
			}
			label.Code = minted[ 0 ]
		} else {
			stored , minted , claim_err := claimCodeTx( tx , household_id , next , code )
			if claim_err != nil {
				tx_err = claim_err
				return
			}
			if minted {
				label.Code = stored
			} else {
				// A bought sticker: the container still gets a code of its
				// own, for printing a replacement.
				fresh , mint_err := mintCodesTx( tx , 1 , CodeEntry{ HouseholdID: household_id , LabelID: next } )
				if mint_err != nil {
					tx_err = mint_err
					return
				}
				label.Code = fresh[ 0 ]
				label.Claimed = []string{ stored }
			}
		}
		tx_err = tx.Put( db.BucketLabels , idKey( next ) , label )
		return
	} )
	if err != nil { label = nil }
	return
}

// UpdateLabel edits a container that is not in Trash.
func UpdateLabel( store *db.Store , household_id uint64 , label_id uint64 , changes LabelChanges ) ( label *Label , err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		label = &Label{}
		if tx_err = tx.Get( db.BucketLabels , idKey( label_id ) , label ); tx_err != nil { return }
		if label.HouseholdID != household_id || label.InTrash() {
			tx_err = db.ErrNotFound
			return
		}
		if tx_err = applyLabelChanges( tx , label , changes ); tx_err != nil { return }
		label.UpdatedAt = time.Now().UTC()
		tx_err = tx.Put( db.BucketLabels , idKey( label_id ) , label )
		return
	} )
	if err != nil { label = nil }
	return
}

// ClaimCode adds a scanned value to a container, or (release) takes it off.
func ClaimCode( store *db.Store , household_id uint64 , label_id uint64 , raw string , release bool ) ( label *Label , err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		label = &Label{}
		if tx_err = tx.Get( db.BucketLabels , idKey( label_id ) , label ); tx_err != nil { return }
		if label.HouseholdID != household_id || label.InTrash() {
			tx_err = db.ErrNotFound
			return
		}
		if release {
			kept := []string{}
			for _ , claimed := range label.Claimed {
				if claimed == strings.TrimSpace( raw ) {
					if tx_err = releaseCodeTx( tx , label_id , claimed ); tx_err != nil { return }
					continue
				}
				kept = append( kept , claimed )
			}
			label.Claimed = kept
		} else {
			if len( label.Claimed ) >= maxClaimed {
				tx_err = ErrTooMany
				return
			}
			stored , _ , claim_err := claimCodeTx( tx , household_id , label_id , raw )
			if claim_err != nil {
				tx_err = claim_err
				return
			}
			if stored != label.Code {
				already := false
				for _ , claimed := range label.Claimed {
					if claimed == stored { already = true }
				}
				if already == false { label.Claimed = append( label.Claimed , stored ) }
			}
		}
		label.UpdatedAt = time.Now().UTC()
		tx_err = tx.Put( db.BucketLabels , idKey( label_id ) , label )
		return
	} )
	if err != nil { label = nil }
	return
}

// TrashLabel moves a container in or out of Trash. Its items go with it, so a
// restore brings back exactly what was deleted; its codes keep pointing at it,
// so a scan of a deleted container says where it went.
func TrashLabel( store *db.Store , household_id uint64 , label_id uint64 , actor_id uint64 , trashed bool ) ( label *Label , err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		label = &Label{}
		if tx_err = tx.Get( db.BucketLabels , idKey( label_id ) , label ); tx_err != nil { return }
		if label.HouseholdID != household_id {
			tx_err = db.ErrNotFound
			return
		}
		was := label.DeletedAt
		if trashed {
			label.throwAway( actor_id )
		} else {
			label.bringBack()
		}
		if tx_err = tx.Put( db.BucketLabels , idKey( label_id ) , label ); tx_err != nil { return }
		// The items deleted along with it (at the same moment) come back with
		// it; ones deleted on their own before stay deleted.
		items := []*Item{}
		tx_err = tx.ForEach( db.BucketItems , func() any { return &Item{} } , func( _ []byte , item any ) bool {
			found := item.( *Item )
			if found.LabelID != label_id { return true }
			if trashed && found.InTrash() == false {
				found.DeletedAt = label.DeletedAt
				found.DeletedBy = actor_id
				items = append( items , found )
			}
			if trashed == false && was != nil && found.DeletedAt != nil && found.DeletedAt.Equal( *was ) {
				found.bringBack()
				items = append( items , found )
			}
			return true
		} )
		if tx_err != nil { return }
		for _ , item := range items {
			if tx_err = tx.Put( db.BucketItems , idKey( item.ID ) , item ); tx_err != nil { return }
		}
		return
	} )
	if err != nil { label = nil }
	return
}

// purgeLabelTx removes a container for good, with its codes.
func purgeLabelTx( tx *db.Tx , label *Label ) ( err error ) {
	for _ , raw := range append( []string{ label.Code } , label.Claimed... ) {
		if err = releaseCodeTx( tx , label.ID , raw ); err != nil { return }
	}
	err = tx.Delete( db.BucketLabels , idKey( label.ID ) )
	return
}
