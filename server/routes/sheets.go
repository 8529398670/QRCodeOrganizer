package routes

import (
	time "time"

	fiber "github.com/gofiber/fiber/v3"

	models "qrcodeorganizer/server/models"
	security "qrcodeorganizer/server/security"
)

// Print runs: sheets of brand-new codes to print and stick on containers. The
// page lays them out and prints them; the server only mints the codes, so
// none is ever printed twice.

type sheetPath struct {
	householdPath
	SheetID uint64 `json:"sheet_id" path:"sheet_id" jsonschema:"minimum=1" jsonschema_description:"The print run's id, from list_sheets." error:"invalid print run id"`
}

type sheetRow struct {
	ID        uint64    `json:"id"`
	Layout    string    `json:"layout"`
	Palette   string    `json:"palette"`
	Codes     []string  `json:"codes" jsonschema_description:"Each code as printed, XXXXX-XXXXX."`
	Names     []string  `json:"names"`
	Used      int       `json:"used" jsonschema_description:"How many are on containers by now."`
	CreatedAt time.Time `json:"created_at"`
}

func ( sheetRow ) APIName() ( name string ) { name = "PrintRun"; return }

func ( handlers *Handlers ) sheetRow( sheet *models.LabelSheet ) ( row sheetRow , err error ) {
	row = sheetRow{ ID: sheet.ID , Layout: sheet.Layout , Palette: sheet.Palette , Codes: []string{} , Names: sheet.Names , CreatedAt: sheet.CreatedAt }
	if row.Names == nil { row.Names = []string{} }
	for _ , code := range sheet.Codes { row.Codes = append( row.Codes , models.FormatCode( code ) ) }
	row.Used , err = models.SheetUse( handlers.Store , sheet )
	return
}

type listSheetsReply struct {
	Sheets []sheetRow `json:"sheets"`
	Paged
}

type listSheetsRequest struct {
	householdPath
	Page
}

// @route("GET /api/households/{household_id}/sheets", summary="List print runs",
//     tag="Labels", errors=(404,),
//     description="Sheets of codes made for printing, newest first, with how many are on containers.",
//     mcp="list_sheets", read_only=True,
// )
func ( handlers *Handlers ) listSheets( c fiber.Ctx , in *listSheetsRequest ) ( reply *listSheetsReply , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needRead )
	if err != nil { return }
	sheets , err := models.ListSheets( handlers.Store , household.ID )
	if err != nil { return }
	rows := []sheetRow{}
	for index := len( sheets ) - 1 ; index >= 0 ; index-- {
		row , row_err := handlers.sheetRow( sheets[ index ] )
		if row_err != nil {
			err = row_err
			return
		}
		rows = append( rows , row )
	}
	reply = &listSheetsReply{}
	reply.Sheets , reply.Paged = cutPage( rows , in.Page )
	return
}

type createSheetRequest struct {
	householdPath
	Count   int      `json:"count" jsonschema:"minimum=1,maximum=300" jsonschema_description:"How many labels." error:"a print run is 1-300 labels"`
	Layout  string   `json:"layout,omitempty" jsonschema:"maxLength=40" jsonschema_description:"Which sheet it is printed on, as the Print screen names it."`
	Palette string   `json:"palette,omitempty" jsonschema:"maxLength=40" jsonschema_description:"original, neon, modern, or a single #rrggbb."`
	Names   []string `json:"names,omitempty" jsonschema_description:"Names to print under the codes, in order (optional)."`
}

// @route("POST /api/households/{household_id}/sheets", summary="Make codes to print",
//     tag="Labels", status=201, errors=(400, 403, 404),
//     description="Mint 1-300 brand-new codes for printing. Each one, once stuck on something and scanned, "
//         "offers to set up that container.",
//     mcp="create_sheet",
// )
func ( handlers *Handlers ) createSheet( c fiber.Ctx , in *createSheetRequest ) ( reply *sheetRow , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	sheet , create_err := models.CreateSheet( handlers.Store , household.ID , security.UserFrom( c ).ID , in.Layout , in.Palette , in.Count , in.Names )
	if err = inventoryRefusal( create_err , "" ); err != nil { return }
	row , err := handlers.sheetRow( sheet )
	if err != nil { return }
	reply = &row
	return
}

// @route("POST /api/households/{household_id}/sheets/{sheet_id}/delete", summary="Forget a print run",
//     tag="Labels", errors=(403, 404),
//     description="Forget a print run. Its codes already on containers keep working; the rest stop, so "
//         "stickers thrown away cannot be scanned later.",
//     mcp="delete_sheet", destructive=True,
// )
func ( handlers *Handlers ) deleteSheet( c fiber.Ctx , in *sheetPath ) ( reply *okReply , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	if err = inventoryRefusal( models.DeleteSheet( handlers.Store , household.ID , in.SheetID ) , "no such print run" ); err != nil { return }
	reply = &okReply{ Ok: true }
	return
}

type reprintRequest struct {
	householdPath
	LabelIDs []uint64 `json:"label_ids" jsonschema:"minItems=1,maxItems=300" jsonschema_description:"Containers whose own labels to print again." error:"send 1-300 container ids"`
}

type reprintRow struct {
	LabelID uint64   `json:"label_id"`
	Code    string   `json:"code" jsonschema_description:"As printed, XXXXX-XXXXX."`
	Name    string   `json:"name"`
	Color   string   `json:"color"`
	Flags   []string `json:"flags"`
}

type reprintReply struct {
	Labels []reprintRow `json:"labels"`
}

// @route("POST /api/households/{household_id}/labels/reprint", summary="Codes to print again",
//     tag="Labels", errors=(403, 404),
//     description="The codes, names, colors and handling flags of containers you already have, to print "
//         "replacement labels.",
//     mcp="reprint_labels", read_only=True,
// )
func ( handlers *Handlers ) reprintLabels( c fiber.Ctx , in *reprintRequest ) ( reply *reprintReply , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needRead )
	if err != nil { return }
	reply = &reprintReply{ Labels: []reprintRow{} }
	for _ , label_id := range in.LabelIDs {
		label , get_err := models.GetLabel( handlers.Store , label_id )
		if get_err != nil || label.HouseholdID != household.ID || label.InTrash() { continue }
		flags := label.Flags
		if flags == nil { flags = []string{} }
		reply.Labels = append( reply.Labels , reprintRow{ LabelID: label.ID , Code: models.FormatCode( label.Code ) , Name: label.Name , Color: label.Color , Flags: flags } )
	}
	return
}
