package models

// The organizer's records. A household is one shared inventory: its members,
// the meaning of its label colors, and the extra fields its items carry. Every
// location, container (label) and item belongs to exactly one household, and
// who may read or change them is decided here, by the caller's role in it --
// never by the account's admin role, which is about who gets into the app.

import (
	errors "errors"
	strings "strings"
	time "time"

	db "qrcodeorganizer/server/db"
	encryption "qrcodeorganizer/server/encryption"
)

const (
	HouseholdOwner  = "owner"
	HouseholdEditor = "editor"
	HouseholdViewer = "viewer"

	householdNameMax = 80
	fieldNameMax     = 40
	maxCustomFields  = 30
	maxColorMeanings = 40
)

var (
	ErrBadName        = errors.New( "a name must be 1-80 characters" )
	ErrNotMember      = errors.New( "not a member of that household" )
	ErrLastOwner      = errors.New( "a household needs at least one owner" )
	ErrBadRole        = errors.New( "role must be owner, editor or viewer" )
	ErrBadField       = errors.New( "a custom field needs a name of 1-40 characters and a type of text, number, date, select or bool" )
	ErrTooManyFields  = errors.New( "a household can have at most 30 custom fields" )
	ErrTooManyColors  = errors.New( "a household can name at most 40 colors" )
	ErrNoSuchUser     = errors.New( "no such user" )
)

// FieldDef is one custom field every item in the household can fill in.
type FieldDef struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Options []string `json:"options,omitempty"`
}

type Household struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	// Members is user id -> owner, editor or viewer.
	Members map[uint64]string `json:"members"`
	// Colors is what each label color means here: "#d32f2f" -> "Kitchen".
	Colors map[string]string `json:"colors,omitempty"`
	Fields []FieldDef        `json:"fields,omitempty"`
	// ReminderDays overrides inventory.reminder_days for this household.
	ReminderDays *int      `json:"reminder_days,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    uint64    `json:"created_by"`
}

func ValidHouseholdRole( role string ) ( result bool ) {
	result = role == HouseholdOwner || role == HouseholdEditor || role == HouseholdViewer
	return
}

// RoleOf is the user's role here, or "" for someone who is not a member.
func ( household *Household ) RoleOf( user_id uint64 ) ( role string ) {
	role = household.Members[ user_id ]
	return
}

func CanEdit( role string ) ( result bool ) {
	result = role == HouseholdOwner || role == HouseholdEditor
	return
}

func cleanName( name string ) ( cleaned string , err error ) {
	cleaned = strings.TrimSpace( name )
	if len( cleaned ) < 1 || len( []rune( cleaned ) ) > householdNameMax {
		err = ErrBadName
	}
	return
}

func idKey( id uint64 ) ( key []byte ) {
	key = encryption.Uint64ToBytes( id )
	return
}

func CreateHousehold( store *db.Store , name string , user_id uint64 ) ( household *Household , err error ) {
	cleaned , err := cleanName( name )
	if err != nil { return }
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		next , tx_err := tx.NextSequence( db.BucketHouseholds )
		if tx_err != nil { return }
		household = &Household{
			ID:        next,
			Name:      cleaned,
			Members:   map[uint64]string{ user_id: HouseholdOwner },
			Colors:    map[string]string{},
			CreatedAt: time.Now().UTC(),
			CreatedBy: user_id,
		}
		tx_err = tx.Put( db.BucketHouseholds , idKey( next ) , household )
		return
	} )
	if err != nil { household = nil }
	return
}

func GetHousehold( store *db.Store , household_id uint64 ) ( household *Household , err error ) {
	household = &Household{}
	if err = store.Get( db.BucketHouseholds , idKey( household_id ) , household ); err != nil {
		household = nil
	}
	return
}

// ListHouseholds is every household the user is a member of, oldest first.
func ListHouseholds( store *db.Store , user_id uint64 ) ( households []*Household , err error ) {
	households = []*Household{}
	err = store.ForEach( db.BucketHouseholds , func() any { return &Household{} } , func( _ []byte , item any ) bool {
		household := item.( *Household )
		if household.RoleOf( user_id ) != "" { households = append( households , household ) }
		return true
	} )
	return
}

// AllHouseholds is every household, for the background jobs.
func AllHouseholds( store *db.Store ) ( households []*Household , err error ) {
	households = []*Household{}
	err = store.ForEach( db.BucketHouseholds , func() any { return &Household{} } , func( _ []byte , item any ) bool {
		households = append( households , item.( *Household ) )
		return true
	} )
	return
}

// UpdateHousehold changes one household in a single transaction.
func UpdateHousehold( store *db.Store , household_id uint64 , mutate func( household *Household ) error ) ( updated *Household , err error ) {
	err = store.UpdateValue( db.BucketHouseholds , idKey( household_id ) , func() any { return &Household{} } , func( item any ) ( mutate_err error ) {
		updated = item.( *Household )
		if updated.Members == nil { updated.Members = map[uint64]string{} }
		if updated.Colors == nil { updated.Colors = map[string]string{} }
		mutate_err = mutate( updated )
		return
	} )
	if err != nil { updated = nil }
	return
}

func RenameHousehold( store *db.Store , household_id uint64 , name string ) ( updated *Household , err error ) {
	cleaned , err := cleanName( name )
	if err != nil { return }
	updated , err = UpdateHousehold( store , household_id , func( household *Household ) ( mutate_err error ) {
		household.Name = cleaned
		return
	} )
	return
}

// SetMember adds someone, changes their role, or (role "") takes them out. A
// household always keeps at least one owner.
func SetMember( store *db.Store , household_id uint64 , user_id uint64 , role string ) ( updated *Household , err error ) {
	if role != "" && ValidHouseholdRole( role ) == false {
		err = ErrBadRole
		return
	}
	if role != "" {
		if _ , get_err := GetUser( store , user_id ); get_err != nil {
			err = ErrNoSuchUser
			return
		}
	}
	updated , err = UpdateHousehold( store , household_id , func( household *Household ) ( mutate_err error ) {
		if role == "" {
			if household.RoleOf( user_id ) == "" {
				mutate_err = ErrNotMember
				return
			}
			delete( household.Members , user_id )
		} else {
			household.Members[ user_id ] = role
		}
		owners := 0
		for _ , each := range household.Members {
			if each == HouseholdOwner { owners += 1 }
		}
		if owners == 0 { mutate_err = ErrLastOwner }
		return
	} )
	return
}

// AddMemberTx is SetMember for someone joining through an invite link, inside
// the transaction that creates their account. A household that is gone is
// skipped rather than failing the join.
func AddMemberTx( tx *db.Tx , household_id uint64 , user_id uint64 , role string ) ( err error ) {
	if ValidHouseholdRole( role ) == false { role = HouseholdEditor }
	household := &Household{}
	if get_err := tx.Get( db.BucketHouseholds , idKey( household_id ) , household ); get_err != nil { return }
	if household.Members == nil { household.Members = map[uint64]string{} }
	if household.RoleOf( user_id ) != "" { return }
	household.Members[ user_id ] = role
	err = tx.Put( db.BucketHouseholds , idKey( household_id ) , household )
	return
}

// SetColors replaces what the household's label colors mean.
func SetColors( store *db.Store , household_id uint64 , colors map[string]string ) ( updated *Household , err error ) {
	if len( colors ) > maxColorMeanings {
		err = ErrTooManyColors
		return
	}
	cleaned := map[string]string{}
	for color , meaning := range colors {
		color = strings.ToLower( strings.TrimSpace( color ) )
		meaning = strings.TrimSpace( meaning )
		if ValidColor( color ) == false || meaning == "" { continue }
		if len( []rune( meaning ) ) > householdNameMax { meaning = string( []rune( meaning )[ :householdNameMax ] ) }
		cleaned[ color ] = meaning
	}
	updated , err = UpdateHousehold( store , household_id , func( household *Household ) ( mutate_err error ) {
		household.Colors = cleaned
		return
	} )
	return
}

// SetFields replaces the household's custom field definitions. A field keeps
// its id across renames, so the values items hold for it survive.
func SetFields( store *db.Store , household_id uint64 , fields []FieldDef , reminder_days *int ) ( updated *Household , err error ) {
	if len( fields ) > maxCustomFields {
		err = ErrTooManyFields
		return
	}
	cleaned := []FieldDef{}
	for _ , field := range fields {
		field.Name = strings.TrimSpace( field.Name )
		switch field.Type {
		case "text" , "number" , "date" , "select" , "bool":
		default:
			err = ErrBadField
			return
		}
		if field.Name == "" || len( []rune( field.Name ) ) > fieldNameMax {
			err = ErrBadField
			return
		}
		if field.ID == "" { field.ID = strings.ToLower( encryption.GenerateRandomString( 8 ) ) }
		options := []string{}
		for _ , option := range field.Options {
			option = strings.TrimSpace( option )
			if option != "" && len( options ) < 50 { options = append( options , option ) }
		}
		field.Options = options
		cleaned = append( cleaned , field )
	}
	updated , err = UpdateHousehold( store , household_id , func( household *Household ) ( mutate_err error ) {
		household.Fields = cleaned
		household.ReminderDays = reminder_days
		return
	} )
	return
}

// DeleteHousehold removes a household and everything in it at once: its
// locations, containers, codes, print sheets, items, drafts and notifications.
// Photos stay in their uploaders' Files, since the files are theirs.
func DeleteHousehold( store *db.Store , household_id uint64 ) ( err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		if tx_err = tx.Delete( db.BucketHouseholds , idKey( household_id ) ); tx_err != nil { return }
		in := func( id uint64 ) bool { return id == household_id }
		if _ , tx_err = tx.DeleteWhere( db.BucketLocations , func() any { return &Location{} } , func( _ []byte , item any ) bool { return in( item.( *Location ).HouseholdID ) } ); tx_err != nil { return }
		if _ , tx_err = tx.DeleteWhere( db.BucketLabels , func() any { return &Label{} } , func( _ []byte , item any ) bool { return in( item.( *Label ).HouseholdID ) } ); tx_err != nil { return }
		if _ , tx_err = tx.DeleteWhere( db.BucketCodes , func() any { return &CodeEntry{} } , func( _ []byte , item any ) bool { return in( item.( *CodeEntry ).HouseholdID ) } ); tx_err != nil { return }
		if _ , tx_err = tx.DeleteWhere( db.BucketSheets , func() any { return &LabelSheet{} } , func( _ []byte , item any ) bool { return in( item.( *LabelSheet ).HouseholdID ) } ); tx_err != nil { return }
		if _ , tx_err = tx.DeleteWhere( db.BucketItems , func() any { return &Item{} } , func( _ []byte , item any ) bool { return in( item.( *Item ).HouseholdID ) } ); tx_err != nil { return }
		if _ , tx_err = tx.DeleteWhere( db.BucketExtractions , func() any { return &Extraction{} } , func( _ []byte , item any ) bool { return in( item.( *Extraction ).HouseholdID ) } ); tx_err != nil { return }
		_ , tx_err = tx.DeleteWhere( db.BucketNotifications , func() any { return &Notification{} } , func( _ []byte , item any ) bool { return in( item.( *Notification ).HouseholdID ) } )
		return
	} )
	return
}

// Trash marks a location, container or item deleted but still restorable,
// until the purge job removes it after inventory.trash_days.
type Trash struct {
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	DeletedBy uint64     `json:"deleted_by,omitempty"`
}

func ( trash *Trash ) InTrash() ( result bool ) {
	result = trash.DeletedAt != nil
	return
}

func ( trash *Trash ) throwAway( actor_id uint64 ) {
	if trash.DeletedAt != nil { return }
	now := time.Now().UTC()
	trash.DeletedAt = &now
	trash.DeletedBy = actor_id
}

func ( trash *Trash ) bringBack() {
	trash.DeletedAt = nil
	trash.DeletedBy = 0
}

func ( trash *Trash ) expired( keep time.Duration ) ( result bool ) {
	result = trash.DeletedAt != nil && time.Since( *trash.DeletedAt ) > keep
	return
}

// ValidColor is a #rrggbb color, lower case.
func ValidColor( color string ) ( result bool ) {
	if len( color ) != 7 || color[ 0 ] != '#' { return }
	for _ , character := range color[ 1: ] {
		if ( character < '0' || character > '9' ) && ( character < 'a' || character > 'f' ) { return }
	}
	result = true
	return
}
