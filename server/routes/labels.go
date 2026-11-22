package routes

import (
	fiber "github.com/gofiber/fiber/v3"

	logs "qrcodeorganizer/server/logs"
	models "qrcodeorganizer/server/models"
	security "qrcodeorganizer/server/security"
)

// Containers: what a SmartLabel is stuck on. The routes call them labels,
// after the sticker; the screens call them containers.

type labelPath struct {
	householdPath
	LabelID uint64 `json:"label_id" path:"label_id" jsonschema:"minimum=1" jsonschema_description:"The container's id, from list_labels, search_inventory or scan_code." error:"invalid container id"`
}

// labelFields is what a person can set on a container. Only what is sent
// changes.
type labelFields struct {
	Name          *string   `json:"name,omitempty" jsonschema:"minLength=1,maxLength=80" error:"a name must be 1-80 characters"`
	Color         *string   `json:"color,omitempty" jsonschema_description:"Its label color, #rrggbb, or empty."`
	ContainerType *string   `json:"container_type,omitempty" jsonschema:"maxLength=40" jsonschema_description:"bin, box, drawer, shelf, bag, or anything else."`
	LocationID    *uint64   `json:"location_id,omitempty" jsonschema_description:"Where it sits; 0 for nowhere."`
	ParentID      *uint64   `json:"parent_id,omitempty" jsonschema_description:"The container it sits inside; 0 for none."`
	Photos        *[]string `json:"photos,omitempty" jsonschema_description:"File ids of its photos, in order (see attach_photo)."`
	Notes         *string   `json:"notes,omitempty" jsonschema:"maxLength=4000"`
	MoveStatus    *string   `json:"move_status,omitempty" jsonschema:"enum=,enum=packed,enum=in_transit,enum=unpacked" jsonschema_description:"For a move: packed, in_transit, unpacked, or empty."`
	Flags         *[]string `json:"flags,omitempty" jsonschema_description:"Handling notes printed on its label, such as fragile, this side up, heavy."`
	Destination   *string   `json:"destination,omitempty" jsonschema:"maxLength=80" jsonschema_description:"For a move: the room it goes to."`
}

func ( fields labelFields ) changes() ( changes models.LabelChanges ) {
	changes = models.LabelChanges{
		Name: fields.Name , Color: fields.Color , ContainerType: fields.ContainerType , LocationID: fields.LocationID ,
		ParentID: fields.ParentID , Photos: fields.Photos , Notes: fields.Notes , MoveStatus: fields.MoveStatus ,
		Flags: fields.Flags , Destination: fields.Destination ,
	}
	return
}

type listLabelsRequest struct {
	householdPath
	Page
	Q          string `json:"q,omitempty" query:"q" jsonschema_description:"Keep containers whose name, code, notes or place has these words."`
	LocationID uint64 `json:"location_id,omitempty" query:"location_id" jsonschema_description:"Only containers at this location or inside it."`
	ParentID   uint64 `json:"parent_id,omitempty" query:"parent_id" jsonschema_description:"Only containers sitting directly inside this one."`
	Color      string `json:"color,omitempty" query:"color" jsonschema_description:"Only this label color, #rrggbb."`
	MoveStatus string `json:"move_status,omitempty" query:"move_status" jsonschema:"enum=packed,enum=in_transit,enum=unpacked" jsonschema_description:"Only containers in this moving state."`
	Trash      bool   `json:"trash,omitempty" query:"trash" jsonschema_description:"List the ones in Trash instead."`
	Sort       string `json:"sort,omitempty" query:"sort" jsonschema:"enum=name,enum=path,enum=item_count,enum=value,enum=updated_at,enum=created_at" jsonschema_description:"Field to sort by."`
	Direction
}

type listLabelsReply struct {
	Labels []labelRow `json:"labels"`
	Paged
}

// @route("GET /api/households/{household_id}/labels", summary="List containers",
//     tag="Containers", errors=(404,),
//     description="Containers in a household, each with its code, where it is, how many items it holds "
//         "and what they are worth. Narrow by words, location, outer container, color or moving state.",
//     mcp="list_labels", read_only=True,
// )
func ( handlers *Handlers ) listLabels( c fiber.Ctx , in *listLabelsRequest ) ( reply *listLabelsReply , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needRead )
	if err != nil { return }
	found , err := handlers.loadInventory( household )
	if err != nil { return }
	rows := []labelRow{}
	for _ , label := range sortedLabels( found ) {
		if label.InTrash() != in.Trash { continue }
		if in.LocationID != 0 && found.underLocation( label.ID , in.LocationID ) == false { continue }
		if in.ParentID != 0 && label.ParentID != in.ParentID { continue }
		if in.Color != "" && label.Color != in.Color { continue }
		if in.MoveStatus != "" && label.MoveStatus != in.MoveStatus { continue }
		haystack := append( []string{ label.Name , label.Notes , label.Code , models.FormatCode( label.Code ) , label.ContainerType , label.Destination , household.Colors[ label.Color ] } , found.labelPath( label.ID )... )
		if matches( in.Q , haystack ) == false { continue }
		rows = append( rows , found.labelRow( label ) )
	}
	reply = &listLabelsReply{}
	reply.Labels , reply.Paged = sortedPage( rows , in.Sort , in.Direction , in.Page )
	return
}

type createLabelRequest struct {
	householdPath
	Code string `json:"code,omitempty" jsonschema:"maxLength=2000" jsonschema_description:"A printed code that is not on anything yet (from a print run), or any other QR or barcode value to claim, such as a bought sticker. Leave out to make a fresh code."`
	labelFields
}

// @route("POST /api/households/{household_id}/labels", summary="Set up a container",
//     tag="Containers", status=201, errors=(400, 403, 404, 409),
//     description="Make a container: give it a name, and optionally the code on its sticker, a color, "
//         "where it is, and the container it sits in.",
//     mcp="create_label",
// )
func ( handlers *Handlers ) createLabel( c fiber.Ctx , in *createLabelRequest ) ( reply *labelRow , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	label , create_err := models.CreateLabel( handlers.Store , household.ID , security.UserFrom( c ).ID , in.Code , in.labelFields.changes() )
	if err = inventoryRefusal( create_err , "that code is not one of this household's printed codes" ); err != nil { return }
	handlers.Webhooks.Emit( models.EventLabelCreated , map[string]any{ "household_id": household.ID , "label_id": label.ID , "name": label.Name } )
	reply , err = handlers.labelReply( household , label )
	return
}

func ( handlers *Handlers ) labelReply( household *models.Household , label *models.Label ) ( reply *labelRow , err error ) {
	found , err := handlers.loadInventory( household )
	if err != nil { return }
	row := found.labelRow( label )
	reply = &row
	return
}

type labelDetailReply struct {
	labelRow
	Items    []itemRow  `json:"items" jsonschema_description:"What is in it, not counting what is in Trash."`
	Children []labelRow `json:"children" jsonschema_description:"Containers sitting inside it."`
}

// @route("GET /api/households/{household_id}/labels/{label_id}", summary="One container and what is in it",
//     tag="Containers", errors=(404,),
//     description="A container with its items and the containers inside it.",
//     mcp="get_label", read_only=True,
// )
func ( handlers *Handlers ) getLabel( c fiber.Ctx , in *labelPath ) ( reply *labelDetailReply , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needRead )
	if err != nil { return }
	found , err := handlers.loadInventory( household )
	if err != nil { return }
	label , known := found.labels[ in.LabelID ]
	if known == false {
		err = refuse( fiber.StatusNotFound , "no such container" )
		return
	}
	reply = &labelDetailReply{ labelRow: found.labelRow( label ) , Items: []itemRow{} , Children: []labelRow{} }
	for _ , item := range found.items {
		if item.LabelID == label.ID && item.InTrash() == false { reply.Items = append( reply.Items , found.itemRow( item ) ) }
	}
	for _ , child := range sortedLabels( found ) {
		if child.ParentID == label.ID && child.InTrash() == false { reply.Children = append( reply.Children , found.labelRow( child ) ) }
	}
	return
}

type updateLabelRequest struct {
	labelPath
	labelFields
}

// @route("POST /api/households/{household_id}/labels/{label_id}", summary="Change a container",
//     tag="Containers", errors=(400, 403, 404),
//     description="Change a container's name, color, type, location, outer container, photos, notes, "
//         "moving state, handling flags or destination. Only what you send changes.",
//     mcp="update_label",
// )
func ( handlers *Handlers ) updateLabel( c fiber.Ctx , in *updateLabelRequest ) ( reply *labelRow , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	if err = handlers.checkPhotos( c , household , in.Photos ); err != nil { return }
	label , update_err := models.UpdateLabel( handlers.Store , household.ID , in.LabelID , in.labelFields.changes() )
	if err = inventoryRefusal( update_err , "no such container" ); err != nil { return }
	reply , err = handlers.labelReply( household , label )
	return
}

type claimCodeRequest struct {
	labelPath
	Code string `json:"code" jsonschema:"minLength=1,maxLength=2000" jsonschema_description:"The scanned value: a QR code's text or a barcode's digits." error:"a code to claim must be 1-2000 characters"`
}

// @route("POST /api/households/{household_id}/labels/{label_id}/codes", summary="Add a code to a container",
//     tag="Containers", errors=(400, 403, 404, 409),
//     description="Make another code open this container too: a store-bought QR sticker, a product's "
//         "barcode, or one of this household's printed codes.",
//     mcp="claim_code",
// )
func ( handlers *Handlers ) claimCode( c fiber.Ctx , in *claimCodeRequest ) ( reply *labelRow , err error ) {
	reply , err = handlers.changeCode( c , in , false )
	return
}

// @route("POST /api/households/{household_id}/labels/{label_id}/codes/release", summary="Take a code off a container",
//     tag="Containers", errors=(400, 403, 404),
//     description="Stop an added code opening this container. Its own code cannot be taken off.",
//     mcp="release_code",
// )
func ( handlers *Handlers ) releaseCode( c fiber.Ctx , in *claimCodeRequest ) ( reply *labelRow , err error ) {
	reply , err = handlers.changeCode( c , in , true )
	return
}

func ( handlers *Handlers ) changeCode( c fiber.Ctx , in *claimCodeRequest , release bool ) ( reply *labelRow , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	label , claim_err := models.ClaimCode( handlers.Store , household.ID , in.LabelID , in.Code , release )
	if err = inventoryRefusal( claim_err , "no such container" ); err != nil { return }
	reply , err = handlers.labelReply( household , label )
	return
}

type moveContentsRequest struct {
	labelPath
	ToLabelID uint64 `json:"to_label_id" jsonschema:"minimum=1" jsonschema_description:"The container everything goes into." error:"invalid container id"`
}

type movedReply struct {
	Moved   int      `json:"moved"`
	ItemIDs []uint64 `json:"item_ids"`
}

// @route("POST /api/households/{household_id}/labels/{label_id}/move-contents", summary="Move everything to another container",
//     tag="Containers", errors=(403, 404),
//     description="Move every item in this container into another one in the same household, in one go.",
//     mcp="move_contents",
// )
func ( handlers *Handlers ) moveContents( c fiber.Ctx , in *moveContentsRequest ) ( reply *movedReply , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	user_id := security.UserFrom( c ).ID
	moved , move_err := models.MoveContents( handlers.Store , household.ID , in.LabelID , in.ToLabelID , user_id )
	if err = inventoryRefusal( move_err , "no such container" ); err != nil { return }
	reply = &movedReply{ Moved: len( moved ) , ItemIDs: []uint64{} }
	for _ , item := range moved {
		reply.ItemIDs = append( reply.ItemIDs , item.ID )
		handlers.Webhooks.Emit( models.EventItemMoved , map[string]any{ "household_id": household.ID , "item_id": item.ID , "name": item.Name , "label_id": in.ToLabelID } )
	}
	return
}

// @route("POST /api/households/{household_id}/labels/{label_id}/delete", summary="Delete a container",
//     bulk=True,
//     tag="Containers", errors=(403, 404),
//     description="Put a container and the items in it in Trash. restore_label brings both back. Its codes "
//         "keep pointing at it until it is purged.",
//     mcp="delete_label", destructive=True,
// )
func ( handlers *Handlers ) deleteLabel( c fiber.Ctx , in *labelPath ) ( reply *labelRow , err error ) {
	reply , err = handlers.trashLabel( c , in , true )
	return
}

// @route("POST /api/households/{household_id}/labels/{label_id}/restore", summary="Restore a container",
//     bulk=True,
//     tag="Containers", errors=(403, 404),
//     description="Take a container back out of Trash, with the items that were deleted along with it.",
//     mcp="restore_label",
// )
func ( handlers *Handlers ) restoreLabel( c fiber.Ctx , in *labelPath ) ( reply *labelRow , err error ) {
	reply , err = handlers.trashLabel( c , in , false )
	return
}

func ( handlers *Handlers ) trashLabel( c fiber.Ctx , in *labelPath , trashed bool ) ( reply *labelRow , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	label , trash_err := models.TrashLabel( handlers.Store , household.ID , in.LabelID , security.UserFrom( c ).ID , trashed )
	if err = inventoryRefusal( trash_err , "no such container" ); err != nil { return }
	if trashed {
		handlers.audit( c , logs.KindAdmin , "container deleted" , fiber.Map{ "household_id": household.ID , "label_id": label.ID } )
		handlers.Webhooks.Emit( models.EventLabelDeleted , map[string]any{ "household_id": household.ID , "label_id": label.ID , "name": label.Name } )
	}
	reply , err = handlers.labelReply( household , label )
	return
}
