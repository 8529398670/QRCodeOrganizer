package models

// Codes: what a scan finds. Every container has one code of its own, printed
// under its QR as XXXXX-XXXXX and inside it as <public url>/l/<code>, and may
// also claim codes it did not come with -- a store-bought QR sticker or a
// product barcode. Each code points at one container, in one household, and a
// code minted for a print sheet points at the sheet until someone scans it and
// sets up the container it went on.
//
// Codes are bolt keys, so a lookup is one read. A minted code is printed on
// paper anyway; a claimed one is stored as a hash, since it might be anything.

import (
	errors "errors"
	strings "strings"

	db "qrcodeorganizer/server/db"
	encryption "qrcodeorganizer/server/encryption"
)

// codeAlphabet leaves out 0, 1, I, L, O and U, which are read wrong or spell
// things, so a code read out over the phone survives.
const (
	codeAlphabet  = "23456789ABCDEFGHJKMNPQRSTVWXYZ"
	codeLength    = 10
	claimedMaxLen = 2000
)

var (
	ErrCodeTaken   = errors.New( "that code is already on another container" )
	ErrCodeUnknown = errors.New( "no such code" )
	ErrBadClaim    = errors.New( "a code to claim must be 1-2000 characters" )
)

// CodeEntry is what one code points at.
type CodeEntry struct {
	HouseholdID uint64 `json:"household_id"`
	// LabelID is 0 while a printed code waits to be stuck on something.
	LabelID uint64 `json:"label_id,omitempty"`
	SheetID uint64 `json:"sheet_id,omitempty"`
	// Claimed is a code the container did not come with.
	Claimed bool `json:"claimed,omitempty"`
}

// NewCode is a fresh random code, not yet checked against the index.
func NewCode() ( code string ) {
	raw := encryption.GenerateRandomBytes( codeLength )
	builder := strings.Builder{}
	for _ , value := range raw {
		builder.WriteByte( codeAlphabet[ int( value ) % len( codeAlphabet ) ] )
	}
	code = builder.String()
	return
}

// FormatCode is a code as printed: XXXXX-XXXXX.
func FormatCode( code string ) ( printed string ) {
	printed = code
	if len( code ) == codeLength { printed = code[ :5 ] + "-" + code[ 5: ] }
	return
}

// NormalizeCode turns whatever was scanned or typed into one of our codes,
// when it is one: the bare code in any case, with or without its dash, or a
// link ending /l/<code>. ok is false for anything else.
func NormalizeCode( raw string ) ( code string , ok bool ) {
	text := strings.TrimSpace( raw )
	if index := strings.LastIndex( text , "/l/" ); index >= 0 { text = text[ index + 3: ] }
	text = strings.TrimRight( text , "/" )
	builder := strings.Builder{}
	for _ , character := range strings.ToUpper( text ) {
		if character == '-' || character == ' ' { continue }
		if strings.ContainsRune( codeAlphabet , character ) == false { return }
		builder.WriteRune( character )
	}
	code = builder.String()
	ok = len( code ) == codeLength
	return
}

func mintedKey( code string ) ( key []byte ) {
	key = []byte( "c:" + code )
	return
}

func claimedKey( raw string ) ( key []byte ) {
	key = []byte( "x:" + encryption.Sha256Hex( strings.TrimSpace( raw ) ) )
	return
}

// codeKey is where a scanned value lives: one of ours by its code, anything
// else by its hash.
func codeKey( raw string ) ( key []byte ) {
	if code , ok := NormalizeCode( raw ); ok {
		key = mintedKey( code )
		return
	}
	key = claimedKey( raw )
	return
}

// LookupCode finds what a scanned value points at.
func LookupCode( store *db.Store , raw string ) ( entry *CodeEntry , err error ) {
	if strings.TrimSpace( raw ) == "" {
		err = ErrCodeUnknown
		return
	}
	entry = &CodeEntry{}
	if err = store.Get( db.BucketCodes , codeKey( raw ) , entry ); err != nil {
		entry = nil
		if errors.Is( err , db.ErrNotFound ) { err = ErrCodeUnknown }
	}
	return
}

// mintCodesTx makes count codes no one has, pointing at a sheet or a label.
func mintCodesTx( tx *db.Tx , count int , entry CodeEntry ) ( codes []string , err error ) {
	codes = []string{}
	for len( codes ) < count {
		code := NewCode()
		probe := &CodeEntry{}
		if get_err := tx.Get( db.BucketCodes , mintedKey( code ) , probe ); get_err == nil { continue }
		if err = tx.Put( db.BucketCodes , mintedKey( code ) , entry ); err != nil { return }
		codes = append( codes , code )
	}
	return
}

// ClaimCodeTx points a scanned value at a container. A value already on
// another container is refused; on this one, it is a no-op. A printed code
// waiting on a sheet in the same household is taken off the sheet.
func claimCodeTx( tx *db.Tx , household_id uint64 , label_id uint64 , raw string ) ( stored string , minted bool , err error ) {
	raw = strings.TrimSpace( raw )
	if raw == "" || len( raw ) > claimedMaxLen {
		err = ErrBadClaim
		return
	}
	key := codeKey( raw )
	code , minted := NormalizeCode( raw )
	existing := &CodeEntry{}
	get_err := tx.Get( db.BucketCodes , key , existing )
	switch {
	case get_err == nil && existing.LabelID == label_id:
	case get_err == nil && existing.LabelID == 0 && existing.HouseholdID == household_id:
		existing.LabelID = label_id
		err = tx.Put( db.BucketCodes , key , existing )
	case get_err == nil:
		err = ErrCodeTaken
		return
	case minted:
		// One of ours that was never minted here: refuse it, rather than
		// let a guessed code become a real one.
		err = ErrCodeUnknown
		return
	default:
		err = tx.Put( db.BucketCodes , key , CodeEntry{ HouseholdID: household_id , LabelID: label_id , Claimed: true } )
	}
	stored = raw
	if minted { stored = code }
	return
}

// releaseCodeTx frees a value a container claimed.
func releaseCodeTx( tx *db.Tx , label_id uint64 , raw string ) ( err error ) {
	key := codeKey( raw )
	existing := &CodeEntry{}
	if get_err := tx.Get( db.BucketCodes , key , existing ); get_err != nil || existing.LabelID != label_id { return }
	if existing.SheetID != 0 && existing.Claimed == false {
		// A printed code goes back to its sheet, ready to be stuck elsewhere.
		existing.LabelID = 0
		err = tx.Put( db.BucketCodes , key , existing )
		return
	}
	err = tx.Delete( db.BucketCodes , key )
	return
}
