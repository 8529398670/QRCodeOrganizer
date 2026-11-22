package models

import (
	time "time"

	db "qrcodeorganizer/server/db"
)

// An Extraction is Smart Photo Extraction's draft: what Claude saw in one
// photo, waiting for a person to tick what is right and press Done. Nothing
// reaches the inventory until they do.
const (
	ExtractionReady    = "ready"
	ExtractionAccepted = "accepted"
	ExtractionFailed   = "failed"

	extractionKeep = 7 * 24 * time.Hour
)

// ProposedItem is one thing found in the photo. Box is where, as fractions of
// the photo's width and height (x, y, width, height), so the page can crop a
// picture of each item from the photo.
type ProposedItem struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Quantity    float64   `json:"quantity"`
	Brand       string    `json:"brand,omitempty"`
	Category    string    `json:"category,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Box         []float64 `json:"box,omitempty"`
}

type Extraction struct {
	ID          uint64         `json:"id"`
	HouseholdID uint64         `json:"household_id"`
	LabelID     uint64         `json:"label_id"`
	FileID      string         `json:"file_id"`
	Status      string         `json:"status"`
	Items       []ProposedItem `json:"items"`
	Summary     string         `json:"summary,omitempty"`
	Error       string         `json:"error,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	CreatedBy   uint64         `json:"created_by"`
	// SharedKey is a photo read with the server's key, which is what
	// ai.daily_limit_per_household limits; one read with someone's own key
	// is on their own bill.
	SharedKey bool `json:"shared_key,omitempty"`
}

func SaveExtraction( store *db.Store , extraction *Extraction ) ( err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		if extraction.ID == 0 {
			if extraction.ID , tx_err = tx.NextSequence( db.BucketExtractions ); tx_err != nil { return }
		}
		tx_err = tx.Put( db.BucketExtractions , idKey( extraction.ID ) , extraction )
		return
	} )
	return
}

func GetExtraction( store *db.Store , extraction_id uint64 ) ( extraction *Extraction , err error ) {
	extraction = &Extraction{}
	if err = store.Get( db.BucketExtractions , idKey( extraction_id ) , extraction ); err != nil { extraction = nil }
	return
}

// ExtractionsToday is how many photos a household sent today (UTC), for
// ai.daily_limit_per_household.
func ExtractionsToday( store *db.Store , household_id uint64 ) ( count int , err error ) {
	today := time.Now().UTC().Format( DateLayout )
	err = store.ForEach( db.BucketExtractions , func() any { return &Extraction{} } , func( _ []byte , item any ) bool {
		extraction := item.( *Extraction )
		if extraction.HouseholdID == household_id && extraction.SharedKey && extraction.CreatedAt.Format( DateLayout ) == today { count += 1 }
		return true
	} )
	return
}

// PruneExtractions drops drafts older than a week.
func PruneExtractions( store *db.Store ) ( removed int , err error ) {
	removed , err = store.DeleteWhere( db.BucketExtractions , func() any { return &Extraction{} } , func( _ []byte , item any ) bool {
		return time.Since( item.( *Extraction ).CreatedAt ) > extractionKeep
	} )
	return
}
