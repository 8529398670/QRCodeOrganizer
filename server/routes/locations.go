package routes

import (
	fiber "github.com/gofiber/fiber/v3"

	models "qrcodeorganizer/server/models"
	security "qrcodeorganizer/server/security"
)

type locationPath struct {
	householdPath
	LocationID uint64 `json:"location_id" path:"location_id" jsonschema:"minimum=1" jsonschema_description:"The location's id, from list_locations." error:"invalid location id"`
}

type listLocationsRequest struct {
	householdPath
	Page
	Trash bool `json:"trash,omitempty" query:"trash" jsonschema_description:"List the ones in Trash instead."`
}

type listLocationsReply struct {
	Locations []locationRow `json:"locations"`
	Paged
}

// @route("GET /api/households/{household_id}/locations", summary="List locations",
//     tag="Locations", errors=(404,),
//     description="The places containers sit, such as House › Garage › Shelf 2, each with its full path, in "
//         "path order.",
//     mcp="list_locations", read_only=True,
// )
func ( handlers *Handlers ) listLocations( c fiber.Ctx , in *listLocationsRequest ) ( reply *listLocationsReply , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needRead )
	if err != nil { return }
	found , err := handlers.loadInventory( household )
	if err != nil { return }
	rows := []locationRow{}
	for _ , location := range found.locations {
		if location.InTrash() != in.Trash { continue }
		rows = append( rows , found.locationRow( location ) )
	}
	reply = &listLocationsReply{}
	reply.Locations , reply.Paged = sortedPage( rows , "path" , Direction{} , in.Page )
	return
}

type createLocationRequest struct {
	householdPath
	Name     string `json:"name" jsonschema:"minLength=1,maxLength=80" error:"a name must be 1-80 characters"`
	ParentID uint64 `json:"parent_id,omitempty" jsonschema_description:"The location it is inside; leave out for a top-level one."`
}

// @route("POST /api/households/{household_id}/locations", summary="Add a location",
//     tag="Locations", status=201, errors=(400, 403, 404),
//     description="Add a place containers can sit, optionally inside another (Garage inside House).",
//     mcp="create_location",
// )
func ( handlers *Handlers ) createLocation( c fiber.Ctx , in *createLocationRequest ) ( reply *locationRow , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	location , create_err := models.CreateLocation( handlers.Store , household.ID , in.Name , in.ParentID )
	if err = inventoryRefusal( create_err , "no such parent location" ); err != nil { return }
	reply , err = handlers.locationReply( household , location )
	return
}

func ( handlers *Handlers ) locationReply( household *models.Household , location *models.Location ) ( reply *locationRow , err error ) {
	found , err := handlers.loadInventory( household )
	if err != nil { return }
	row := found.locationRow( location )
	reply = &row
	return
}

type updateLocationRequest struct {
	locationPath
	Name     *string `json:"name,omitempty" jsonschema:"minLength=1,maxLength=80" error:"a name must be 1-80 characters"`
	ParentID *uint64 `json:"parent_id,omitempty" jsonschema_description:"Move it inside another location; 0 makes it top-level."`
}

// @route("POST /api/households/{household_id}/locations/{location_id}", summary="Rename or move a location",
//     tag="Locations", errors=(400, 403, 404),
//     description="Change a location's name, or move it inside another (0 for top-level). Only what you "
//         "send changes.",
//     mcp="update_location",
// )
func ( handlers *Handlers ) updateLocation( c fiber.Ctx , in *updateLocationRequest ) ( reply *locationRow , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	location , update_err := models.UpdateLocation( handlers.Store , household.ID , in.LocationID , in.Name , in.ParentID )
	if err = inventoryRefusal( update_err , "no such location" ); err != nil { return }
	reply , err = handlers.locationReply( household , location )
	return
}

// @route("POST /api/households/{household_id}/locations/{location_id}/delete", summary="Delete a location",
//     bulk=True,
//     tag="Locations", errors=(403, 404),
//     description="Put a location in Trash. Containers there keep it until it is purged, and show no "
//         "location meanwhile; restore_location puts it back.",
//     mcp="delete_location", destructive=True,
// )
func ( handlers *Handlers ) deleteLocation( c fiber.Ctx , in *locationPath ) ( reply *locationRow , err error ) {
	reply , err = handlers.trashLocation( c , in , true )
	return
}

// @route("POST /api/households/{household_id}/locations/{location_id}/restore", summary="Restore a location",
//     bulk=True,
//     tag="Locations", errors=(403, 404),
//     description="Take a location back out of Trash.",
//     mcp="restore_location",
// )
func ( handlers *Handlers ) restoreLocation( c fiber.Ctx , in *locationPath ) ( reply *locationRow , err error ) {
	reply , err = handlers.trashLocation( c , in , false )
	return
}

func ( handlers *Handlers ) trashLocation( c fiber.Ctx , in *locationPath , trashed bool ) ( reply *locationRow , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needEdit )
	if err != nil { return }
	location , trash_err := models.TrashLocation( handlers.Store , household.ID , in.LocationID , security.UserFrom( c ).ID , trashed )
	if err = inventoryRefusal( trash_err , "no such location" ); err != nil { return }
	reply , err = handlers.locationReply( household , location )
	return
}
