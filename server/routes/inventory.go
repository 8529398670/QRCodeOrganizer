package routes

import (
	sort "sort"
	strings "strings"
	time "time"

	models "qrcodeorganizer/server/models"
)

// inventory is one household's whole inventory, read once per request. bolt
// has no query language, so lists, search and paths are worked out here, in
// memory: a household's worth of records is small.
type inventory struct {
	household *models.Household
	locations map[uint64]*models.Location
	labels    map[uint64]*models.Label
	items     []*models.Item
}

func ( handlers *Handlers ) loadInventory( household *models.Household ) ( found *inventory , err error ) {
	found = &inventory{ household: household , locations: map[uint64]*models.Location{} , labels: map[uint64]*models.Label{} }
	locations , err := models.ListLocations( handlers.Store , household.ID )
	if err != nil { return }
	for _ , location := range locations { found.locations[ location.ID ] = location }
	labels , err := models.ListLabels( handlers.Store , household.ID )
	if err != nil { return }
	for _ , label := range labels { found.labels[ label.ID ] = label }
	found.items , err = models.ListItems( handlers.Store , household.ID )
	return
}

const pathSeparator = " › "

// locationPath is House › Garage › Shelf 2.
func ( found *inventory ) locationPath( location_id uint64 ) ( names []string ) {
	seen := map[uint64]bool{}
	for next := location_id ; next != 0 && seen[ next ] == false ; {
		seen[ next ] = true
		location , known := found.locations[ next ]
		if known == false || location.InTrash() { break }
		names = append( []string{ location.Name } , names... )
		next = location.ParentID
	}
	return
}

// labelPath is where a container is: the location of the outermost container
// it sits in, then each container around it, then itself.
func ( found *inventory ) labelPath( label_id uint64 ) ( names []string ) {
	chain := []*models.Label{}
	seen := map[uint64]bool{}
	for next := label_id ; next != 0 && seen[ next ] == false ; {
		seen[ next ] = true
		label , known := found.labels[ next ]
		if known == false { break }
		chain = append( []*models.Label{ label } , chain... )
		next = label.ParentID
	}
	if len( chain ) == 0 { return }
	names = found.locationPath( chain[ 0 ].LocationID )
	for _ , label := range chain { names = append( names , label.Name ) }
	return
}

func joinPath( names []string ) ( path string ) {
	path = strings.Join( names , pathSeparator )
	return
}

// locationRow is a location with where it is.
type locationRow struct {
	models.Location
	Path        string `json:"path"`
	LabelCount  int    `json:"label_count" jsonschema_description:"Containers directly here."`
}

func ( locationRow ) APIName() ( name string ) { name = "Location"; return }

func ( found *inventory ) locationRow( location *models.Location ) ( row locationRow ) {
	row = locationRow{ Location: *location , Path: joinPath( found.locationPath( location.ID ) ) }
	for _ , label := range found.labels {
		if label.LocationID == location.ID && label.InTrash() == false { row.LabelCount += 1 }
	}
	return
}

// labelRow is a container with where it is and how full.
type labelRow struct {
	models.Label
	PrintedCode string  `json:"printed_code" jsonschema_description:"The code as printed under its QR, XXXXX-XXXXX."`
	Path        string  `json:"path" jsonschema_description:"Where it is: location › outer containers › this one."`
	ItemCount   int     `json:"item_count"`
	ChildCount  int     `json:"child_count" jsonschema_description:"Containers sitting inside this one."`
	Value       float64 `json:"value" jsonschema_description:"What its items are worth together."`
	ColorMeans  string  `json:"color_means" jsonschema_description:"What its color stands for in this household."`
}

func ( labelRow ) APIName() ( name string ) { name = "Container"; return }

func ( found *inventory ) labelRow( label *models.Label ) ( row labelRow ) {
	row = labelRow{ Label: *label , PrintedCode: models.FormatCode( label.Code ) , Path: joinPath( found.labelPath( label.ID ) ) }
	if row.Claimed == nil { row.Claimed = []string{} }
	if row.Photos == nil { row.Photos = []string{} }
	if row.Flags == nil { row.Flags = []string{} }
	row.ColorMeans = found.household.Colors[ label.Color ]
	for _ , item := range found.items {
		if item.LabelID == label.ID && item.InTrash() == false {
			row.ItemCount += 1
			row.Value += item.Value * max( item.Quantity , 1 )
		}
	}
	for _ , other := range found.labels {
		if other.ParentID == label.ID && other.InTrash() == false { row.ChildCount += 1 }
	}
	return
}

// itemRow is an item with the container it is in and where that is.
type itemRow struct {
	models.Item
	LabelName   string `json:"label_name"`
	LabelCode   string `json:"label_code" jsonschema_description:"Its container's printed code."`
	LabelColor  string `json:"label_color"`
	Path        string `json:"path" jsonschema_description:"Where it is: location › containers."`
	ExpiresIn   *int   `json:"expires_in" jsonschema_description:"Days until it expires (negative once it has); null without a date."`
	LowStockNow bool   `json:"low_stock_now" jsonschema_description:"At or below its low-stock level."`
	Overdue     bool   `json:"overdue" jsonschema_description:"Lent out and past its due date."`
}

func ( itemRow ) APIName() ( name string ) { name = "Item"; return }

func ( found *inventory ) itemRow( item *models.Item ) ( row itemRow ) {
	row = itemRow{ Item: *item , Path: joinPath( found.labelPath( item.LabelID ) ) }
	if row.Photos == nil { row.Photos = []string{} }
	if row.Tags == nil { row.Tags = []string{} }
	if row.Receipts == nil { row.Receipts = []string{} }
	if row.Fields == nil { row.Fields = map[string]string{} }
	if row.History == nil { row.History = []models.ItemEvent{} }
	if label , known := found.labels[ item.LabelID ]; known {
		row.LabelName = label.Name
		row.LabelCode = models.FormatCode( label.Code )
		row.LabelColor = label.Color
	}
	row.ExpiresIn = daysUntil( item.Expires )
	row.LowStockNow = item.LowStock != nil && item.Quantity <= *item.LowStock
	due := daysUntil( item.LentDue )
	row.Overdue = item.LentTo != "" && due != nil && *due < 0
	return
}

// daysUntil is whole days from today (UTC) to a YYYY-MM-DD date.
func daysUntil( date string ) ( days *int ) {
	if date == "" { return }
	when , err := time.Parse( models.DateLayout , date )
	if err != nil { return }
	today , _ := time.Parse( models.DateLayout , time.Now().UTC().Format( models.DateLayout ) )
	count := int( when.Sub( today ).Hours() / 24 )
	days = &count
	return
}

// matches is a forgiving search: every word typed must appear somewhere in
// the item -- name, description, tags, brand, barcode, serial, custom field
// values, its container and where that is -- or be one letter off a word
// there, so "screwdriver" finds "screw driver" and "hamer" finds "hammer".
func matches( query string , haystack []string ) ( result bool ) {
	words := strings.Fields( strings.ToLower( query ) )
	if len( words ) == 0 {
		result = true
		return
	}
	text := strings.ToLower( strings.Join( haystack , " " ) )
	squashed := strings.ReplaceAll( text , " " , "" )
	tokens := strings.FieldsFunc( text , func( character rune ) bool {
		return ( character < 'a' || character > 'z' ) && ( character < '0' || character > '9' ) && character < 128
	} )
	for _ , word := range words {
		if strings.Contains( text , word ) || strings.Contains( squashed , word ) { continue }
		close := false
		if len( word ) >= 4 {
			for _ , token := range tokens {
				if oneEditApart( word , token ) || ( len( token ) > len( word ) && oneEditApart( word , token[ :len( word ) ] ) ) {
					close = true
					break
				}
			}
		}
		if close == false { return }
	}
	result = true
	return
}

// oneEditApart is two words a single typo apart: one letter added, dropped or
// changed.
func oneEditApart( a string , b string ) ( result bool ) {
	if a == b {
		result = true
		return
	}
	if len( a ) > len( b ) { a , b = b , a }
	if len( b ) - len( a ) > 1 { return }
	edits := 0
	for i , j := 0 , 0 ; i < len( a ) || j < len( b ) ; {
		switch {
		case i < len( a ) && j < len( b ) && a[ i ] == b[ j ]:
			i , j = i + 1 , j + 1
			continue
		case len( a ) == len( b ):
			i , j = i + 1 , j + 1
		default:
			j += 1
		}
		edits += 1
		if edits > 1 { return }
	}
	result = true
	return
}

func ( found *inventory ) itemHaystack( item *models.Item ) ( haystack []string ) {
	haystack = []string{ item.Name , item.Description , item.Brand , item.Barcode , item.Serial , item.LentTo , item.Unit }
	haystack = append( haystack , item.Tags... )
	for _ , value := range item.Fields { haystack = append( haystack , value ) }
	if label , known := found.labels[ item.LabelID ]; known {
		haystack = append( haystack , label.Name , label.Notes , models.FormatCode( label.Code ) , label.Code , found.household.Colors[ label.Color ] )
	}
	haystack = append( haystack , found.labelPath( item.LabelID )... )
	return
}

// underLocation is whether a container is at a location or anywhere inside it.
func ( found *inventory ) underLocation( label_id uint64 , location_id uint64 ) ( result bool ) {
	outer := found.labels[ label_id ]
	climbed := map[uint64]bool{}
	for outer != nil && outer.ParentID != 0 && climbed[ outer.ID ] == false {
		climbed[ outer.ID ] = true
		outer = found.labels[ outer.ParentID ]
	}
	if outer == nil { return }
	visited := map[uint64]bool{}
	for next := outer.LocationID ; next != 0 && visited[ next ] == false ; {
		if next == location_id {
			result = true
			return
		}
		visited[ next ] = true
		location := found.locations[ next ]
		if location == nil { return }
		next = location.ParentID
	}
	return
}

// insideLabel is whether a container is this one or sits anywhere inside it.
func ( found *inventory ) insideLabel( label_id uint64 , outer_id uint64 ) ( result bool ) {
	seen := map[uint64]bool{}
	for next := label_id ; next != 0 && seen[ next ] == false ; {
		if next == outer_id {
			result = true
			return
		}
		seen[ next ] = true
		label := found.labels[ next ]
		if label == nil { return }
		next = label.ParentID
	}
	return
}

func sortedLabels( found *inventory ) ( labels []*models.Label ) {
	for _ , label := range found.labels { labels = append( labels , label ) }
	sort.Slice( labels , func( i int , j int ) bool { return labels[ i ].ID < labels[ j ].ID } )
	return
}
