package routes

import (
	strings "strings"

	fiber "github.com/gofiber/fiber/v3"

	models "qrcodeorganizer/server/models"
	security "qrcodeorganizer/server/security"
)

type itemPath struct {
	householdPath
	ItemID uint64 `json:"item_id" path:"item_id" jsonschema:"minimum=1" jsonschema_description:"The item's id, from list_items, search_inventory or get_label." error:"invalid item id"`
}

// itemFields is what a person can set on an item. Only what is sent changes;
// the clear_ switches empty a number that has no "nothing" of its own.
type itemFields struct {
	Name           *string            `json:"name,omitempty" jsonschema:"minLength=1,maxLength=80" error:"a name must be 1-80 characters"`
	Description    *string            `json:"description,omitempty" jsonschema:"maxLength=4000"`
	Photos         *[]string          `json:"photos,omitempty" jsonschema_description:"File ids of its photos, in order."`
	Tags           *[]string          `json:"tags,omitempty" jsonschema_description:"Words to find it by. Stored in lower case."`
	Quantity       *float64           `json:"quantity,omitempty" jsonschema:"minimum=0" jsonschema_description:"How many there are. A new item has 1."`
	Unit           *string            `json:"unit,omitempty" jsonschema:"maxLength=20" jsonschema_description:"pcs, rolls, kg, ..."`
	LowStock       *float64           `json:"low_stock,omitempty" jsonschema:"minimum=0" jsonschema_description:"Remind when the quantity drops to this or below."`
	ClearLowStock  bool               `json:"clear_low_stock,omitempty" jsonschema_description:"Stop low-stock reminders."`
	Expires        *string            `json:"expires,omitempty" jsonschema_description:"Expiry date, YYYY-MM-DD, or empty."`
	RemindDays     *int               `json:"remind_days,omitempty" jsonschema:"minimum=0,maximum=365" jsonschema_description:"Days before the expiry to remind, instead of the household's lead."`
	ClearRemind    bool               `json:"clear_remind_days,omitempty" jsonschema_description:"Use the household's reminder lead again."`
	Fields         *map[string]string `json:"fields,omitempty" jsonschema_description:"Custom field values by field id (see the household's fields). Replaces them all."`
	Barcode        *string            `json:"barcode,omitempty" jsonschema:"maxLength=200"`
	Brand          *string            `json:"brand,omitempty" jsonschema:"maxLength=80"`
	Serial         *string            `json:"serial,omitempty" jsonschema:"maxLength=120"`
	PurchaseDate   *string            `json:"purchase_date,omitempty" jsonschema_description:"YYYY-MM-DD, or empty."`
	Value          *float64           `json:"value,omitempty" jsonschema:"minimum=0" jsonschema_description:"What one is worth, for the insurance list."`
	Receipts       *[]string          `json:"receipts,omitempty" jsonschema_description:"File ids of receipts or warranty papers."`
	LentTo         *string            `json:"lent_to,omitempty" jsonschema:"maxLength=80" jsonschema_description:"Who has it on loan; empty when it is back."`
	LentDue        *string            `json:"lent_due,omitempty" jsonschema_description:"When it should come back, YYYY-MM-DD, or empty."`
}

func ( fields itemFields ) changes() ( changes models.ItemChanges ) {
	changes = models.ItemChanges{
		Name: fields.Name , Description: fields.Description , Photos: fields.Photos , Tags: fields.Tags ,
		Quantity: fields.Quantity , Unit: fields.Unit , Expires: fields.Expires , Fields: fields.Fields ,
		Barcode: fields.Barcode , Brand: fields.Brand , Serial: fields.Serial , PurchaseDate: fields.PurchaseDate ,
		Value: fields.Value , Receipts: fields.Receipts , LentTo: fields.LentTo , LentDue: fields.LentDue ,
	}
	if fields.LowStock != nil || fields.ClearLowStock {
		low := fields.LowStock
		if fields.ClearLowStock { low = nil }
		changes.LowStock = &low
	}
	if fields.RemindDays != nil || fields.ClearRemind {
		days := fields.RemindDays
		if fields.ClearRemind { days = nil }
		changes.RemindDays = &days
	}
	return
}

func ( handlers *Handlers ) checkItemFiles( c fiber.Ctx , household *models.Household , fields itemFields ) ( err error ) {
	if err = handlers.checkPhotos( c , household , fields.Photos ); err != nil { return }
	err = handlers.checkPhotos( c , household , fields.Receipts )
	return
}

type listItemsRequest struct {
	householdPath
	Page
	Q              string `json:"q,omitempty" query:"q" jsonschema_description:"Words to find: in names, descriptions, tags, brands, barcodes, serials, custom fields, and where it is. Forgives a typo."`
	LabelID        uint64 `json:"label_id,omitempty" query:"label_id" jsonschema_description:"Only items in this container or the containers inside it."`
	LocationID     uint64 `json:"location_id,omitempty" query:"location_id" jsonschema_description:"Only items at this location or anywhere inside it."`
	Tag            string `json:"tag,omitempty" query:"tag" jsonschema_description:"Only items with this tag."`
	Color          string `json:"color,omitempty" query:"color" jsonschema_description:"Only items in containers of this label color."`
	ExpiringWithin *int   `json:"expiring_within,omitempty" query:"expiring_within" jsonschema:"minimum=0,maximum=3650" jsonschema_description:"Only items expiring within this many days (or already expired)."`
	LowStock       bool   `json:"low_stock,omitempty" query:"low_stock" jsonschema_description:"Only items at or below their low-stock level."`
	Lent           bool   `json:"lent,omitempty" query:"lent" jsonschema_description:"Only items lent out."`
	Trash          bool   `json:"trash,omitempty" query:"trash" jsonschema_description:"List the ones in Trash instead."`
	Sort           string `json:"sort,omitempty" query:"sort" jsonschema:"enum=name,enum=path,enum=quantity,enum=expires,enum=value,enum=updated_at,enum=created_at" jsonschema_description:"Field to sort by."`
	Direction
}

type listItemsReply struct {
	Items []itemRow `json:"items"`
	Paged
	TotalValue float64 `json:"total_value" jsonschema_description:"What every item matched is worth, value times quantity."`
}

// @route("GET /api/households/{household_id}/items", summary="List and search items",
//     tag="Items", errors=(404,),
//     description="Items in a household, each with the container it is in and where that is. Search by "
//         "words (a typo is forgiven) and narrow by container, location, tag, color, expiry, low stock or "
//         "loans.",
//     mcp="list_items", read_only=True,
// )
func ( handlers *Handlers ) listItems( c fiber.Ctx , in *listItemsRequest ) ( reply *listItemsReply , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needRead )
	if err != nil { return }
	found , err := handlers.loadInventory( household )
	if err != nil { return }
	rows := []itemRow{}
	reply = &listItemsReply{}
	tag := strings.ToLower( strings.TrimSpace( in.Tag ) )
	for _ , item := range found.items {
		if item.InTrash() != in.Trash { continue }
		if in.LabelID != 0 && found.insideLabel( item.LabelID , in.LabelID ) == false { continue }
		if in.LocationID != 0 && found.underLocation( item.LabelID , in.LocationID ) == false { continue }
		if tag != "" && containsString( item.Tags , tag ) == false { continue }
		if in.Color != "" && ( found.labels[ item.LabelID ] == nil || found.labels[ item.LabelID ].Color != in.Color ) { continue }
		row := found.itemRow( item )
		if in.ExpiringWithin != nil && ( row.ExpiresIn == nil || *row.ExpiresIn > *in.ExpiringWithin ) { continue }
		if in.LowStock && row.LowStockNow == false { continue }
		if in.Lent && item.LentTo == "" { continue }
		if matches( in.Q , found.itemHaystack( item ) ) == false { continue }
		rows = append( rows , row )
		reply.TotalValue += item.Value * max( item.Quantity , 1 )
	}
	sort_by := in.Sort
	if sort_by == "" { sort_by = "name" }
	reply.Items , reply.Paged = sortedPage( rows , sort_by , in.Direction , in.Page )
	return
}

func containsString( values []string , wanted string ) ( result bool ) {
	for _ , value := range values {
		if value == wanted {
			result = true
			return
		}
	}
	return
}

type createItemsRequest struct {
	labelPath
	Items []itemFields `json:"items" jsonschema:"minItems=1,maxItems=100" jsonschema_description:"The items to put in, each with at least a name." error:"send 1-100 items"`
}

type createItemsReply struct {
	Items []itemRow `json:"items"`
}

// @route("POST /api/households/{household_id}/labels/{label_id}/items", summary="Put items in a container",
//     tag="Items", status=201, errors=(400, 403, 404),
//     description="Add up to 100 items to a container at once, each with at least a name. A new item's "
//         "quantity is 1 unless you say.",
//     mcp="create_items",
// )
func ( handlers *Handlers ) createItems( c fiber.Ctx , in *createItemsRequest ) ( reply *createItemsReply , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	all := []models.ItemChanges{}
	for _ , fields := range in.Items {
		if fields.Name == nil {
			err = refuse( fiber.StatusBadRequest , "every item needs a name" )
			return
		}
		if err = handlers.checkItemFiles( c , household , fields ); err != nil { return }
		all = append( all , fields.changes() )
	}
	items , create_err := models.CreateItems( handlers.Store , household , in.LabelID , security.UserFrom( c ).ID , all )
	if err = inventoryRefusal( create_err , "no such container" ); err != nil { return }
	found , err := handlers.loadInventory( household )
	if err != nil { return }
	reply = &createItemsReply{ Items: []itemRow{} }
	for _ , item := range items {
		reply.Items = append( reply.Items , found.itemRow( item ) )
		handlers.Webhooks.Emit( models.EventItemCreated , map[string]any{ "household_id": household.ID , "item_id": item.ID , "name": item.Name , "label_id": item.LabelID } )
	}
	return
}

func ( handlers *Handlers ) itemReply( household *models.Household , item *models.Item ) ( reply *itemRow , err error ) {
	found , err := handlers.loadInventory( household )
	if err != nil { return }
	row := found.itemRow( item )
	reply = &row
	return
}

// @route("GET /api/households/{household_id}/items/{item_id}", summary="One item",
//     tag="Items", errors=(404,),
//     description="An item with everything known about it, where it is, and its history.",
//     mcp="get_item", read_only=True,
// )
func ( handlers *Handlers ) getItem( c fiber.Ctx , in *itemPath ) ( reply *itemRow , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needRead )
	if err != nil { return }
	item , get_err := models.GetItem( handlers.Store , in.ItemID )
	if get_err != nil || item.HouseholdID != household.ID {
		err = refuse( fiber.StatusNotFound , "no such item" )
		return
	}
	reply , err = handlers.itemReply( household , item )
	return
}

type updateItemRequest struct {
	itemPath
	itemFields
}

// @route("POST /api/households/{household_id}/items/{item_id}", summary="Change an item",
//     tag="Items", errors=(400, 403, 404),
//     description="Change anything about an item: name, photos, tags, quantity, low-stock level, expiry, "
//         "custom fields, value, receipts, who it is lent to. Only what you send changes.",
//     mcp="update_item",
// )
func ( handlers *Handlers ) updateItem( c fiber.Ctx , in *updateItemRequest ) ( reply *itemRow , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	if err = handlers.checkItemFiles( c , household , in.itemFields ); err != nil { return }
	item , _ , update_err := models.UpdateItem( handlers.Store , household , in.ItemID , security.UserFrom( c ).ID , in.itemFields.changes() )
	if err = inventoryRefusal( update_err , "no such item" ); err != nil { return }
	handlers.Webhooks.Emit( models.EventItemUpdated , map[string]any{ "household_id": household.ID , "item_id": item.ID , "name": item.Name } )
	reply , err = handlers.itemReply( household , item )
	return
}

type adjustRequest struct {
	itemPath
	Delta float64 `json:"delta" jsonschema_description:"How many to add; negative to take some out. Never goes below 0." error:"delta must be a number"`
}

// @route("POST /api/households/{household_id}/items/{item_id}/adjust", summary="Add to or take from an item's count",
//     tag="Items", errors=(400, 403, 404),
//     description="Change an item's quantity by an amount, such as -1 when one is used up. Two people "
//         "doing this at once both count.",
//     mcp="adjust_quantity",
// )
func ( handlers *Handlers ) adjustQuantity( c fiber.Ctx , in *adjustRequest ) ( reply *itemRow , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	item , adjust_err := models.AdjustQuantity( handlers.Store , household.ID , in.ItemID , security.UserFrom( c ).ID , in.Delta )
	if err = inventoryRefusal( adjust_err , "no such item" ); err != nil { return }
	reply , err = handlers.itemReply( household , item )
	return
}

type moveItemRequest struct {
	itemPath
	LabelID uint64 `json:"label_id" jsonschema:"minimum=1" jsonschema_description:"The container it goes into." error:"invalid container id"`
}

// @route("POST /api/households/{household_id}/items/{item_id}/move", summary="Move an item to another container",
//     bulk=True,
//     tag="Items", errors=(403, 404),
//     description="Put an item in another container of the same household.",
//     mcp="move_item",
// )
func ( handlers *Handlers ) moveItem( c fiber.Ctx , in *moveItemRequest ) ( reply *itemRow , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	_ , move_err := models.MoveItems( handlers.Store , household.ID , []uint64{ in.ItemID } , in.LabelID , security.UserFrom( c ).ID )
	if err = inventoryRefusal( move_err , "no such item or container" ); err != nil { return }
	item , err := models.GetItem( handlers.Store , in.ItemID )
	if err != nil { return }
	handlers.Webhooks.Emit( models.EventItemMoved , map[string]any{ "household_id": household.ID , "item_id": item.ID , "name": item.Name , "label_id": in.LabelID } )
	reply , err = handlers.itemReply( household , item )
	return
}

// @route("POST /api/households/{household_id}/items/{item_id}/delete", summary="Delete an item",
//     bulk=True,
//     tag="Items", errors=(403, 404),
//     description="Put an item in Trash; restore_item brings it back until it is purged.",
//     mcp="delete_item", destructive=True,
// )
func ( handlers *Handlers ) deleteItem( c fiber.Ctx , in *itemPath ) ( reply *itemRow , err error ) {
	reply , err = handlers.trashItem( c , in , true )
	return
}

// @route("POST /api/households/{household_id}/items/{item_id}/restore", summary="Restore an item",
//     bulk=True,
//     tag="Items", errors=(400, 403, 404),
//     description="Take an item back out of Trash. Its container has to be out of Trash first.",
//     mcp="restore_item",
// )
func ( handlers *Handlers ) restoreItem( c fiber.Ctx , in *itemPath ) ( reply *itemRow , err error ) {
	reply , err = handlers.trashItem( c , in , false )
	return
}

func ( handlers *Handlers ) trashItem( c fiber.Ctx , in *itemPath , trashed bool ) ( reply *itemRow , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	item , trash_err := models.TrashItem( handlers.Store , household.ID , in.ItemID , security.UserFrom( c ).ID , trashed )
	if err = inventoryRefusal( trash_err , "no such item" ); err != nil { return }
	if trashed {
		handlers.Webhooks.Emit( models.EventItemDeleted , map[string]any{ "household_id": household.ID , "item_id": item.ID , "name": item.Name } )
	}
	reply , err = handlers.itemReply( household , item )
	return
}
