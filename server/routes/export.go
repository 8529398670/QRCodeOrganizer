package routes

import (
	bytes "bytes"
	csv "encoding/csv"
	fmt "fmt"
	io "io"
	sort "sort"
	strconv "strconv"
	strings "strings"
	time "time"

	fiber "github.com/gofiber/fiber/v3"

	models "qrcodeorganizer/server/models"
	security "qrcodeorganizer/server/security"
)

// A household's inventory as a spreadsheet, out and back in. The same columns
// go both ways, so an export can be edited and imported into another
// household; the import also understands the usual names other apps use
// (SmartLabels' export among them), and makes containers it does not know.

var exportColumns = []string{
	"Item" , "Description" , "Quantity" , "Unit" , "Container" , "Container code" , "Color" , "Location" ,
	"Tags" , "Brand" , "Barcode" , "Serial" , "Expires" , "Low stock" , "Purchase date" , "Value" ,
	"Lent to" , "Lent due" ,
}

func amount( value float64 ) ( text string ) {
	text = strconv.FormatFloat( value , 'f' , -1 , 64 )
	return
}

// csvSafe stops a cell being read as a formula when the file is opened in a
// spreadsheet: a value starting with = + - @ gets a leading quote mark.
func csvSafe( value string ) ( safe string ) {
	safe = value
	if value != "" && strings.ContainsRune( "=+-@\t\r" , rune( value[ 0 ] ) ) { safe = "'" + value }
	return
}

// @route("GET /api/households/{household_id}/export.csv", summary="Download the inventory as a spreadsheet",
//     tag="Reports", errors=(404,), reply_type="text/csv",
//     reply_description="Every item not in Trash, one row each, as CSV",
//     mcp_http="A CSV file is not JSON: GET this address with the same Authorization header.",
// )
func ( handlers *Handlers ) exportCSV( c fiber.Ctx , in *householdPath ) ( err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needRead )
	if err != nil { return }
	found , err := handlers.loadInventory( household )
	if err != nil { return }
	header := append( []string{} , exportColumns... )
	for _ , field := range household.Fields { header = append( header , field.Name ) }
	buffer := &bytes.Buffer{}
	writer := csv.NewWriter( buffer )
	writer.Write( header )
	items := []*models.Item{}
	for _ , item := range found.items {
		if item.InTrash() == false { items = append( items , item ) }
	}
	sort.Slice( items , func( i int , j int ) bool {
		a , b := joinPath( found.labelPath( items[ i ].LabelID ) ) , joinPath( found.labelPath( items[ j ].LabelID ) )
		if a != b { return a < b }
		return strings.ToLower( items[ i ].Name ) < strings.ToLower( items[ j ].Name )
	} )
	for _ , item := range items {
		label := found.labels[ item.LabelID ]
		container , code , color , location := "" , "" , "" , ""
		if label != nil {
			container , code , color = label.Name , models.FormatCode( label.Code ) , label.Color
			location = joinPath( found.locationPath( label.LocationID ) )
		}
		low := ""
		if item.LowStock != nil { low = amount( *item.LowStock ) }
		value := ""
		if item.Value > 0 { value = amount( item.Value ) }
		row := []string{
			item.Name , item.Description , amount( item.Quantity ) , item.Unit , container , code , color , location ,
			strings.Join( item.Tags , ", " ) , item.Brand , item.Barcode , item.Serial , item.Expires , low ,
			item.PurchaseDate , value , item.LentTo , item.LentDue ,
		}
		for _ , field := range household.Fields { row = append( row , item.Fields[ field.ID ] ) }
		for index := range row { row[ index ] = csvSafe( row[ index ] ) }
		writer.Write( row )
	}
	writer.Flush()
	name := fmt.Sprintf( "%s-inventory-%s.csv" , strings.ReplaceAll( strings.ToLower( household.Name ) , " " , "-" ) , time.Now().UTC().Format( models.DateLayout ) )
	c.Set( fiber.HeaderContentType , "text/csv; charset=utf-8" )
	c.Set( fiber.HeaderContentDisposition , contentDisposition( name ) )
	c.Set( fiber.HeaderCacheControl , "private, no-store" )
	err = c.Send( buffer.Bytes() )
	return
}

// importAliases is every header name a column is known by, in lower case.
var importAliases = map[string][]string{
	"name":          { "item" , "name" , "item name" , "title" , "product" },
	"description":   { "description" , "details" , "notes" , "item description" },
	"quantity":      { "quantity" , "qty" , "count" , "amount" , "inventory count" },
	"unit":          { "unit" , "units" },
	"container":     { "container" , "label" , "label name" , "box" , "bin" , "smartlabel" , "container name" },
	"code":          { "container code" , "code" , "label code" , "qr code" },
	"color":         { "color" , "colour" , "label color" },
	"location":      { "location" , "place" , "room" , "label location" },
	"tags":          { "tags" , "tag" , "category" , "categories" },
	"brand":         { "brand" , "make" , "manufacturer" },
	"barcode":       { "barcode" , "upc" , "ean" },
	"serial":        { "serial" , "serial number" , "s/n" },
	"expires":       { "expires" , "expiration" , "expiration date" , "expiry" , "expiry date" , "best before" },
	"low_stock":     { "low stock" , "minimum" , "reorder at" },
	"purchase_date": { "purchase date" , "purchased" , "bought" },
	"value":         { "value" , "price" , "cost" , "worth" },
	"lent_to":       { "lent to" , "loaned to" },
	"lent_due":      { "lent due" , "due back" },
}

type importRequest struct {
	householdPath
	FileID string `json:"file_id" jsonschema:"minLength=1" jsonschema_description:"A CSV file you uploaded (files_start_upload), with a header row." error:"file_id is required"`
}

type importReply struct {
	Items      int      `json:"items" jsonschema_description:"Items added."`
	Containers int      `json:"containers" jsonschema_description:"Containers made for them."`
	Locations  int      `json:"locations" jsonschema_description:"Locations made for them."`
	Skipped    []string `json:"skipped" jsonschema_description:"Rows left out, and why."`
}

// @route("POST /api/households/{household_id}/import", summary="Add items from a spreadsheet",
//     tag="Reports", errors=(400, 403, 404),
//     description="Add every row of an uploaded CSV as an item. A header row names the columns (Item, "
//         "Quantity, Container, Location, Tags, Expires, Value ... as in an export, and the usual names "
//         "other apps use). Containers and locations it does not know are made; nothing is changed or "
//         "removed.",
//     mcp="import_csv",
// )
func ( handlers *Handlers ) importCSV( c fiber.Ctx , in *importRequest ) ( reply *importReply , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	user_id := security.UserFrom( c ).ID
	file , get_err := models.GetFile( handlers.Store , in.FileID )
	if get_err != nil || file.Deleted() || file.State != models.FileReady || file.UserID != user_id {
		err = refuse( fiber.StatusBadRequest , "no such file: upload the CSV first, then name its file id" )
		return
	}
	if file.Size > 20 << 20 {
		err = refuse( fiber.StatusBadRequest , "that spreadsheet is over 20 MB" )
		return
	}
	reader , err := models.OpenFile( handlers.cfg() , file )
	if err != nil { return }
	defer reader.Close()
	raw , err := io.ReadAll( reader )
	if err != nil { return }
	raw = bytes.TrimPrefix( raw , []byte( "\xEF\xBB\xBF" ) )
	parser := csv.NewReader( bytes.NewReader( raw ) )
	parser.FieldsPerRecord = -1
	parser.LazyQuotes = true
	records , parse_err := parser.ReadAll()
	if parse_err != nil || len( records ) < 2 {
		err = refuse( fiber.StatusBadRequest , "that is not a CSV with a header row and at least one item" )
		return
	}
	columns := map[string]int{}
	custom := map[int]string{}
	for index , heading := range records[ 0 ] {
		heading = strings.ToLower( strings.TrimSpace( heading ) )
		for column , aliases := range importAliases {
			if _ , taken := columns[ column ]; taken == false && containsString( aliases , heading ) { columns[ column ] = index }
		}
		for _ , field := range household.Fields {
			if strings.ToLower( field.Name ) == heading { custom[ index ] = field.ID }
		}
	}
	if _ , has := columns[ "name" ]; has == false {
		err = refuse( fiber.StatusBadRequest , "the header row needs an Item (or Name) column" )
		return
	}
	reply = &importReply{ Skipped: []string{} }
	found , err := handlers.loadInventory( household )
	if err != nil { return }
	importer := &importer{ handlers: handlers , household: household , found: found , user_id: user_id , reply: reply }
	for number , record := range records[ 1: ] {
		cell := func( column string ) ( value string ) {
			if index , has := columns[ column ]; has && index < len( record ) { value = strings.TrimPrefix( strings.TrimSpace( record[ index ] ) , "'" ) }
			return
		}
		if cell( "name" ) == "" { continue }
		if row_err := importer.add( cell , record , custom ); row_err != nil {
			reply.Skipped = append( reply.Skipped , fmt.Sprintf( "row %d: %s" , number + 2 , row_err.Error() ) )
		}
	}
	return
}

type importer struct {
	handlers  *Handlers
	household *models.Household
	found     *inventory
	user_id   uint64
	reply     *importReply
	locations map[string]uint64
	labels    map[string]uint64
}

func ( importer *importer ) location( path string ) ( location_id uint64 , err error ) {
	if path == "" { return }
	if importer.locations == nil {
		importer.locations = map[string]uint64{}
		for _ , location := range importer.found.locations {
			if location.InTrash() == false { importer.locations[ strings.ToLower( joinPath( importer.found.locationPath( location.ID ) ) ) ] = location.ID }
		}
	}
	parts := strings.FieldsFunc( path , func( character rune ) bool { return character == '›' || character == '>' || character == '/' } )
	so_far := []string{}
	for _ , part := range parts {
		part = strings.TrimSpace( part )
		if part == "" { continue }
		so_far = append( so_far , part )
		key := strings.ToLower( joinPath( so_far ) )
		if known , has := importer.locations[ key ]; has {
			location_id = known
			continue
		}
		made , make_err := models.CreateLocation( importer.handlers.Store , importer.household.ID , part , location_id )
		if make_err != nil {
			err = make_err
			return
		}
		importer.locations[ key ] = made.ID
		importer.reply.Locations += 1
		location_id = made.ID
	}
	return
}

func ( importer *importer ) label( name string , code string , color string , location string ) ( label_id uint64 , err error ) {
	if importer.labels == nil {
		importer.labels = map[string]uint64{}
		for _ , label := range importer.found.labels {
			if label.InTrash() { continue }
			importer.labels[ "code:" + label.Code ] = label.ID
			importer.labels[ "name:" + strings.ToLower( label.Name ) ] = label.ID
		}
	}
	if normal , ok := models.NormalizeCode( code ); ok {
		if known , has := importer.labels[ "code:" + normal ]; has {
			label_id = known
			return
		}
	}
	if name == "" { name = "Imported" }
	if known , has := importer.labels[ "name:" + strings.ToLower( name ) ]; has {
		label_id = known
		return
	}
	location_id , err := importer.location( location )
	if err != nil { return }
	changes := models.LabelChanges{ Name: &name , LocationID: &location_id }
	color = strings.ToLower( color )
	if models.ValidColor( color ) { changes.Color = &color }
	made , err := models.CreateLabel( importer.handlers.Store , importer.household.ID , importer.user_id , "" , changes )
	if err != nil { return }
	importer.labels[ "name:" + strings.ToLower( name ) ] = made.ID
	importer.reply.Containers += 1
	label_id = made.ID
	return
}

func ( importer *importer ) add( cell func( string ) string , record []string , custom map[int]string ) ( err error ) {
	label_id , err := importer.label( cell( "container" ) , cell( "code" ) , cell( "color" ) , cell( "location" ) )
	if err != nil { return }
	text := func( column string ) ( value *string ) {
		if found := cell( column ); found != "" { value = &found }
		return
	}
	number := func( column string ) ( value *float64 , bad bool ) {
		found := strings.TrimSpace( strings.TrimLeft( cell( column ) , "$€£" ) )
		if found == "" { return }
		parsed , parse_err := strconv.ParseFloat( strings.ReplaceAll( found , "," , "" ) , 64 )
		if parse_err != nil {
			bad = true
			return
		}
		value = &parsed
		return
	}
	changes := models.ItemChanges{
		Name: text( "name" ) , Description: text( "description" ) , Unit: text( "unit" ) , Brand: text( "brand" ) ,
		Barcode: text( "barcode" ) , Serial: text( "serial" ) , Expires: text( "expires" ) , PurchaseDate: text( "purchase_date" ) ,
		LentTo: text( "lent_to" ) , LentDue: text( "lent_due" ) ,
	}
	if tags := cell( "tags" ); tags != "" {
		list := strings.FieldsFunc( tags , func( character rune ) bool { return character == ',' || character == ';' } )
		changes.Tags = &list
	}
	quantity , bad := number( "quantity" )
	if bad {
		err = fmt.Errorf( "quantity is not a number" )
		return
	}
	changes.Quantity = quantity
	if value , bad := number( "value" ); bad == false {
		changes.Value = value
	}
	if low , bad := number( "low_stock" ); bad == false && low != nil {
		changes.LowStock = &low
	}
	if len( custom ) > 0 {
		values := map[string]string{}
		for index , field_id := range custom {
			if index < len( record ) && strings.TrimSpace( record[ index ] ) != "" { values[ field_id ] = strings.TrimSpace( record[ index ] ) }
		}
		changes.Fields = &values
	}
	_ , err = models.CreateItems( importer.handlers.Store , importer.household , label_id , importer.user_id , []models.ItemChanges{ changes } )
	if err == nil { importer.reply.Items += 1 }
	return
}
