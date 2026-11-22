package routes

import (
	sort "sort"
	time "time"

	fiber "github.com/gofiber/fiber/v3"

	logs "qrcodeorganizer/server/logs"
	models "qrcodeorganizer/server/models"
	security "qrcodeorganizer/server/security"
)

// Households: shared inventories, their members, and what their label colors
// and custom fields mean.

type memberRow struct {
	UserID      uint64 `json:"user_id"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role" jsonschema:"enum=owner,enum=editor,enum=viewer"`
}

type householdRow struct {
	ID           uint64            `json:"id"`
	Name         string            `json:"name"`
	MyRole       string            `json:"my_role" jsonschema:"enum=owner,enum=editor,enum=viewer"`
	Members      []memberRow       `json:"members"`
	Colors       map[string]string `json:"colors" jsonschema_description:"Label color (#rrggbb) to what it means here."`
	Fields       []models.FieldDef `json:"fields" jsonschema_description:"Custom fields every item here can fill in."`
	ReminderDays *int              `json:"reminder_days" jsonschema_description:"Days before an expiry its reminder goes out; null uses the app's default."`
	CreatedAt    time.Time         `json:"created_at"`
}

func ( householdRow ) APIName() ( name string ) { name = "Household"; return }

func ( handlers *Handlers ) householdRowFor( household *models.Household , user_id uint64 ) ( row householdRow ) {
	members := []memberRow{}
	for member_id , role := range household.Members {
		name := ""
		if user , err := models.GetUser( handlers.Store , member_id ); err == nil { name = user.DisplayName }
		members = append( members , memberRow{ UserID: member_id , DisplayName: name , Role: role } )
	}
	sort.Slice( members , func( i int , j int ) bool { return members[ i ].UserID < members[ j ].UserID } )
	colors := household.Colors
	if colors == nil { colors = map[string]string{} }
	fields := household.Fields
	if fields == nil { fields = []models.FieldDef{} }
	row = householdRow{
		ID: household.ID , Name: household.Name , MyRole: household.RoleOf( user_id ) , Members: members ,
		Colors: colors , Fields: fields , ReminderDays: household.ReminderDays , CreatedAt: household.CreatedAt ,
	}
	return
}

type listHouseholdsReply struct {
	Households []householdRow `json:"households"`
	Paged
}

// @route("GET /api/households", summary="List your households",
//     tag="Households",
//     description="The households you are a member of, each with your role, its members, color meanings and "
//         "custom fields. Every other inventory tool takes one of these ids.",
//     mcp="list_households", read_only=True,
// )
func ( handlers *Handlers ) listHouseholds( c fiber.Ctx , in *Page ) ( reply *listHouseholdsReply , err error ) {
	user_id := security.UserFrom( c ).ID
	households , err := models.ListHouseholds( handlers.Store , user_id )
	if err != nil { return }
	rows := []householdRow{}
	for _ , household := range households { rows = append( rows , handlers.householdRowFor( household , user_id ) ) }
	reply = &listHouseholdsReply{}
	reply.Households , reply.Paged = cutPage( rows , *in )
	return
}

type createHouseholdRequest struct {
	Name string `json:"name" jsonschema:"minLength=1,maxLength=80" error:"a name must be 1-80 characters"`
}

// @route("POST /api/households", summary="Create a household",
//     tag="Households", status=201,
//     description="Make a new shared inventory. You are its first owner.",
//     mcp="create_household",
// )
func ( handlers *Handlers ) createHousehold( c fiber.Ctx , in *createHouseholdRequest ) ( reply *householdRow , err error ) {
	user_id := security.UserFrom( c ).ID
	household , create_err := models.CreateHousehold( handlers.Store , in.Name , user_id )
	if err = inventoryRefusal( create_err , "" ); err != nil { return }
	handlers.audit( c , logs.KindAdmin , "household created" , fiber.Map{ "household_id": household.ID } )
	row := handlers.householdRowFor( household , user_id )
	reply = &row
	return
}

// @route("GET /api/households/{household_id}", summary="One household",
//     tag="Households", errors=(404,),
//     description="A household you are in: your role, its members, color meanings and custom fields.",
//     mcp="get_household", read_only=True,
// )
func ( handlers *Handlers ) getHousehold( c fiber.Ctx , in *householdPath ) ( reply *householdRow , err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needRead )
	if err != nil { return }
	row := handlers.householdRowFor( household , security.UserFrom( c ).ID )
	reply = &row
	return
}

type renameHouseholdRequest struct {
	householdPath
	Name string `json:"name" jsonschema:"minLength=1,maxLength=80" error:"a name must be 1-80 characters"`
}

// @route("POST /api/households/{household_id}/rename", summary="Rename a household",
//     tag="Households", errors=(403, 404),
//     description="Give a household a new name. Owners and editors.",
//     mcp="rename_household",
// )
func ( handlers *Handlers ) renameHousehold( c fiber.Ctx , in *renameHouseholdRequest ) ( reply *householdRow , err error ) {
	if _ , _ , err = handlers.household( c , in.HouseholdID , needEdit ); err != nil { return }
	household , rename_err := models.RenameHousehold( handlers.Store , in.HouseholdID , in.Name )
	if err = inventoryRefusal( rename_err , "no such household" ); err != nil { return }
	row := handlers.householdRowFor( household , security.UserFrom( c ).ID )
	reply = &row
	return
}

// @route("POST /api/households/{household_id}/delete", summary="Delete a household",
//     tag="Households", errors=(403, 404),
//     description="Delete a household and everything in it, for good: its locations, containers, codes, "
//         "print runs and items. Owners only. Photos stay in their uploaders' Files.",
//     mcp="delete_household", destructive=True,
// )
func ( handlers *Handlers ) deleteHousehold( c fiber.Ctx , in *householdPath ) ( reply *okReply , err error ) {
	if _ , _ , err = handlers.household( c , in.HouseholdID , needOwner ); err != nil { return }
	if err = models.DeleteHousehold( handlers.Store , in.HouseholdID ); err != nil { return }
	handlers.audit( c , logs.KindAdmin , "household deleted" , fiber.Map{ "household_id": in.HouseholdID } )
	reply = &okReply{ Ok: true }
	return
}

type setMemberRequest struct {
	householdPath
	UserID uint64 `json:"user_id" jsonschema:"minimum=1" jsonschema_description:"Their account id, from GET /api/team (list_team over MCP)." error:"invalid user id"`
	Role   string `json:"role" jsonschema:"enum=owner,enum=editor,enum=viewer,enum=" jsonschema_description:"owner, editor or viewer; empty takes them out." error:"role must be owner, editor, viewer or empty"`
}

// @route("POST /api/households/{household_id}/members", summary="Add, change or remove a member",
//     tag="Households", errors=(403, 404, 409),
//     description="Give someone with an account a role here -- owner (everything), editor (change the "
//         "inventory) or viewer (look only) -- or an empty role to take them out. Owners only. A household "
//         "always keeps at least one owner.",
//     mcp="set_household_member",
// )
func ( handlers *Handlers ) setHouseholdMember( c fiber.Ctx , in *setMemberRequest ) ( reply *householdRow , err error ) {
	if _ , _ , err = handlers.household( c , in.HouseholdID , needOwner ); err != nil { return }
	household , set_err := models.SetMember( handlers.Store , in.HouseholdID , in.UserID , in.Role )
	if err = inventoryRefusal( set_err , "no such household" ); err != nil { return }
	handlers.audit( c , logs.KindAdmin , "household member set" , fiber.Map{ "household_id": in.HouseholdID , "user_id": in.UserID , "role": in.Role } )
	row := handlers.householdRowFor( household , security.UserFrom( c ).ID )
	reply = &row
	return
}

// @route("POST /api/households/{household_id}/leave", summary="Leave a household",
//     tag="Households", errors=(404, 409),
//     description="Stop being a member. The last owner cannot leave: make someone else an owner, or delete "
//         "the household.",
//     mcp="leave_household",
// )
func ( handlers *Handlers ) leaveHousehold( c fiber.Ctx , in *householdPath ) ( reply *okReply , err error ) {
	if _ , _ , err = handlers.household( c , in.HouseholdID , needRead ); err != nil { return }
	_ , leave_err := models.SetMember( handlers.Store , in.HouseholdID , security.UserFrom( c ).ID , "" )
	if err = inventoryRefusal( leave_err , "no such household" ); err != nil { return }
	reply = &okReply{ Ok: true }
	return
}

type setColorsRequest struct {
	householdPath
	Colors map[string]string `json:"colors" jsonschema_description:"Label color (#rrggbb) to what it means here, such as \"#d32f2f\": \"Kitchen\". Replaces the whole list."`
}

// @route("POST /api/households/{household_id}/colors", summary="Say what label colors mean",
//     tag="Households", errors=(403, 404),
//     description="Replace the household's color legend: which room or kind of thing each label color "
//         "stands for. Owners and editors.",
//     mcp="set_household_colors",
// )
func ( handlers *Handlers ) setHouseholdColors( c fiber.Ctx , in *setColorsRequest ) ( reply *householdRow , err error ) {
	if _ , _ , err = handlers.household( c , in.HouseholdID , needEdit ); err != nil { return }
	household , set_err := models.SetColors( handlers.Store , in.HouseholdID , in.Colors )
	if err = inventoryRefusal( set_err , "no such household" ); err != nil { return }
	row := handlers.householdRowFor( household , security.UserFrom( c ).ID )
	reply = &row
	return
}

type setFieldsRequest struct {
	householdPath
	Fields       []models.FieldDef `json:"fields" jsonschema_description:"The whole list. Each has a name, a type (text, number, date, select or bool), options for a select, and the id it already has, if any, so values survive a rename."`
	ReminderDays *int              `json:"reminder_days" jsonschema_description:"Days before an expiry its reminder goes out, 0-365; null uses the app's default."`
}

// @route("POST /api/households/{household_id}/fields", summary="Set custom fields and reminder lead",
//     tag="Households", errors=(403, 404),
//     description="Replace the custom fields items here can fill in (serial numbers, warranty dates, "
//         "anything), and how many days before an expiry reminders go out. Owners and editors.",
//     mcp="set_household_fields",
// )
func ( handlers *Handlers ) setHouseholdFields( c fiber.Ctx , in *setFieldsRequest ) ( reply *householdRow , err error ) {
	if _ , _ , err = handlers.household( c , in.HouseholdID , needEdit ); err != nil { return }
	if in.ReminderDays != nil && ( *in.ReminderDays < 0 || *in.ReminderDays > 365 ) {
		err = refuse( fiber.StatusBadRequest , "reminder_days must be 0-365" )
		return
	}
	household , set_err := models.SetFields( handlers.Store , in.HouseholdID , in.Fields , in.ReminderDays )
	if err = inventoryRefusal( set_err , "no such household" ); err != nil { return }
	row := handlers.householdRowFor( household , security.UserFrom( c ).ID )
	reply = &row
	return
}
