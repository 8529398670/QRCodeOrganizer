package routes

import (
	errors "errors"

	fiber "github.com/gofiber/fiber/v3"

	db "qrcodeorganizer/server/db"
	models "qrcodeorganizer/server/models"
	security "qrcodeorganizer/server/security"
)

// Who may touch a household's inventory is decided here, for every route
// under /api/households/{household_id}: by the caller's role in that
// household. Someone who is not a member gets the 404 an unknown household
// gets, so a guessed id says nothing. Being an admin of the app gives no way
// into anyone's inventory; an API key acts as the account it belongs to.

const (
	needRead  = "read"
	needEdit  = "edit"
	needOwner = "owner"
)

// householdPath is a route under /api/households/{household_id}.
type householdPath struct {
	HouseholdID uint64 `json:"household_id" path:"household_id" jsonschema:"minimum=1" jsonschema_description:"The household's id, from GET /api/households (list_households over MCP)." error:"invalid household id"`
}

// household is the household the request names, if the caller may do what
// need says there. role is the caller's role in it.
func ( handlers *Handlers ) household( c fiber.Ctx , household_id uint64 , need string ) ( household *models.Household , role string , err error ) {
	found , get_err := models.GetHousehold( handlers.Store , household_id )
	if get_err != nil {
		err = refuse( fiber.StatusNotFound , "no such household" )
		return
	}
	role = found.RoleOf( security.UserFrom( c ).ID )
	switch {
	case role == "":
		err = refuse( fiber.StatusNotFound , "no such household" )
	case need == needEdit && models.CanEdit( role ) == false:
		err = refuse( fiber.StatusForbidden , "viewers can look but not change anything" )
	case need == needOwner && role != models.HouseholdOwner:
		err = refuse( fiber.StatusForbidden , "only an owner of this household can do that" )
	default:
		household = found
	}
	return
}

// inventoryRefusal turns a model's error into the answer every inventory
// route gives. Anything it does not know is a 500, whose cause is logged.
func inventoryRefusal( cause error , missing string ) ( err error ) {
	switch {
	case cause == nil:
	case errors.Is( cause , db.ErrNotFound ) , errors.Is( cause , models.ErrNoSuchLabel ) , errors.Is( cause , models.ErrCodeUnknown ):
		err = refuse( fiber.StatusNotFound , missing )
	case errors.Is( cause , models.ErrCodeTaken ):
		err = refuse( fiber.StatusConflict , cause.Error() )
	case errors.Is( cause , models.ErrLastOwner ) , errors.Is( cause , models.ErrNotMember ):
		err = refuse( fiber.StatusConflict , cause.Error() )
	case isValidation( cause ):
		err = refuse( fiber.StatusBadRequest , cause.Error() )
	default:
		err = cause
	}
	return
}

// isValidation is an error that says what was wrong with what was sent.
func isValidation( cause error ) ( result bool ) {
	for _ , known := range []error{
		models.ErrBadName , models.ErrBadRole , models.ErrBadField , models.ErrTooManyFields , models.ErrTooManyColors ,
		models.ErrNoSuchUser , models.ErrLocationLoop , models.ErrLabelLoop , models.ErrBadMove , models.ErrTooMany ,
		models.ErrTooLong , models.ErrBadDate , models.ErrBadQuantity , models.ErrBadClaim , models.ErrSheetSize ,
		models.ErrBadAPIKey , models.ErrBadModel , models.ErrBadEffort , models.ErrBadAccess ,
	} {
		if errors.Is( cause , known ) {
			result = true
			return
		}
	}
	// The models' one-off messages are written for the person, so they can
	// be shown as they are.
	result = modelMessages[ cause.Error() ]
	return
}

var modelMessages = map[string]bool{
	"no such location": true,
	"no such container to put it in": true,
	"color must look like #1e88e5": true,
	"remind_days must be 0-365": true,
	"value must be 0 or more": true,
	"its container is in Trash: restore the container first": true,
}
