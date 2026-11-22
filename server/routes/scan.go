package routes

import (
	fiber "github.com/gofiber/fiber/v3"

	models "qrcodeorganizer/server/models"
	security "qrcodeorganizer/server/security"
)

// Scanning. A code only means something to a member of the household it
// belongs to: to anyone else it is unknown, exactly like a code that exists
// nowhere, so a sticker on a box in a storage unit says nothing to whoever
// photographs it.

type scanRequest struct {
	Code string `json:"code" query:"code" jsonschema:"minLength=1,maxLength=2000" jsonschema_description:"What the camera read: a printed code (XXXXX-XXXXX, with or without the dash), a link ending /l/<code>, any other QR text, or a barcode's digits." error:"code must be 1-2000 characters"`
}

type scanReply struct {
	// Kind is label (a container you can see), unassigned (one of your
	// household's printed codes, not on anything yet), or unknown.
	Kind        string    `json:"kind" jsonschema:"enum=label,enum=unassigned,enum=unknown"`
	HouseholdID uint64    `json:"household_id" jsonschema_description:"0 when unknown."`
	Household   string    `json:"household" jsonschema_description:"Its name; empty when unknown."`
	Code        string    `json:"code" jsonschema_description:"The code as printed, for one of ours; what was scanned, otherwise."`
	Label       *labelRow `json:"label" jsonschema_description:"The container, when kind is label."`
	InTrash     bool      `json:"in_trash" jsonschema_description:"The container is in Trash."`
	CanEdit     bool      `json:"can_edit" jsonschema_description:"Whether you may set it up or change it."`
}

// @route("GET /api/scan", summary="Look up a scanned code",
//     tag="Scan",
//     description="What a scanned or typed code is: one of your containers (with where it is), one of your "
//         "household's printed codes still waiting to be stuck on something (set it up with create_label "
//         "and this code), or unknown -- which may be a store-bought sticker or barcode you can claim for a "
//         "container with claim_code.",
//     mcp="scan_code", read_only=True,
// )
func ( handlers *Handlers ) scanCode( c fiber.Ctx , in *scanRequest ) ( reply *scanReply , err error ) {
	reply = &scanReply{ Kind: "unknown" , Code: in.Code }
	if code , ok := models.NormalizeCode( in.Code ); ok { reply.Code = models.FormatCode( code ) }
	entry , lookup_err := models.LookupCode( handlers.Store , in.Code )
	if lookup_err != nil { return }
	household , get_err := models.GetHousehold( handlers.Store , entry.HouseholdID )
	if get_err != nil { return }
	role := household.RoleOf( security.UserFrom( c ).ID )
	if role == "" { return }
	reply.HouseholdID = household.ID
	reply.Household = household.Name
	reply.CanEdit = models.CanEdit( role )
	if entry.LabelID == 0 {
		reply.Kind = "unassigned"
		return
	}
	found , err := handlers.loadInventory( household )
	if err != nil { return }
	label , known := found.labels[ entry.LabelID ]
	if known == false {
		reply.Kind = "unknown"
		reply.HouseholdID , reply.Household = 0 , ""
		return
	}
	row := found.labelRow( label )
	reply.Kind = "label"
	reply.Label = &row
	reply.InTrash = label.InTrash()
	return
}

type searchRequest struct {
	Page
	Q string `json:"q" query:"q" jsonschema:"minLength=1,maxLength=200" jsonschema_description:"What you are looking for, such as \"drill\" or \"christmas lights\". A typo is forgiven." error:"q must be 1-200 characters"`
}

type searchHit struct {
	Kind        string    `json:"kind" jsonschema:"enum=item,enum=label"`
	HouseholdID uint64    `json:"household_id"`
	Household   string    `json:"household"`
	Item        *itemRow  `json:"item" jsonschema_description:"When kind is item."`
	Label       *labelRow `json:"label" jsonschema_description:"When kind is label."`
	Where       string    `json:"where" jsonschema_description:"The whole answer in words: household › location › containers."`
}

type searchReply struct {
	Hits []searchHit `json:"hits"`
	Paged
}

// @route("GET /api/search", summary="Find where something is",
//     tag="Scan",
//     description="Search every household you are in for items and containers, and say where each one "
//         "is. The question \"where is my drill?\" is search_inventory with q=drill.",
//     mcp="search_inventory", read_only=True,
// )
func ( handlers *Handlers ) searchInventory( c fiber.Ctx , in *searchRequest ) ( reply *searchReply , err error ) {
	households , err := models.ListHouseholds( handlers.Store , security.UserFrom( c ).ID )
	if err != nil { return }
	hits := []searchHit{}
	for _ , household := range households {
		found , load_err := handlers.loadInventory( household )
		if load_err != nil {
			err = load_err
			return
		}
		for _ , item := range found.items {
			if item.InTrash() || matches( in.Q , found.itemHaystack( item ) ) == false { continue }
			row := found.itemRow( item )
			hits = append( hits , searchHit{ Kind: "item" , HouseholdID: household.ID , Household: household.Name , Item: &row ,
				Where: joinPath( append( []string{ household.Name } , found.labelPath( item.LabelID )... ) ) } )
		}
		for _ , label := range sortedLabels( found ) {
			if label.InTrash() { continue }
			haystack := append( []string{ label.Name , label.Notes , label.Code , models.FormatCode( label.Code ) , household.Colors[ label.Color ] } , found.labelPath( label.ID )... )
			if matches( in.Q , haystack ) == false { continue }
			row := found.labelRow( label )
			hits = append( hits , searchHit{ Kind: "label" , HouseholdID: household.ID , Household: household.Name , Label: &row ,
				Where: joinPath( append( []string{ household.Name } , found.labelPath( label.ID )... ) ) } )
		}
	}
	reply = &searchReply{}
	reply.Hits , reply.Paged = cutPage( hits , in.Page )
	return
}
