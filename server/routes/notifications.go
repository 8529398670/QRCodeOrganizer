package routes

import (
	fmt "fmt"
	strings "strings"
	time "time"

	fiber "github.com/gofiber/fiber/v3"

	apidoc "qrcodeorganizer/server/apidoc"

	models "qrcodeorganizer/server/models"
	security "qrcodeorganizer/server/security"
)

// Reminders without email: the bell in the navbar (these routes), webhooks,
// and a calendar feed a phone subscribes to.

type notificationRow struct {
	ID          uint64    `json:"id"`
	HouseholdID uint64    `json:"household_id"`
	ItemID      uint64    `json:"item_id"`
	LabelID     uint64    `json:"label_id"`
	Kind        string    `json:"kind" jsonschema:"enum=expiring,enum=expired,enum=low_stock,enum=overdue"`
	ItemName    string    `json:"item_name"`
	Due         string    `json:"due"`
	Read        bool      `json:"read"`
	CreatedAt   time.Time `json:"created_at"`
}

func ( notificationRow ) APIName() ( name string ) { name = "Notification"; return }

type listNotificationsRequest struct {
	Page
	Unread bool `json:"unread,omitempty" query:"unread" jsonschema_description:"Only the ones not read yet."`
}

type listNotificationsReply struct {
	Notifications []notificationRow `json:"notifications"`
	Paged
	Unread int `json:"unread" jsonschema_description:"How many are not read yet, in all."`
}

// @route("GET /api/notifications", summary="Your reminders",
//     tag="Reminders",
//     description="Reminders about items in your households, newest first: expiring or expired, run low, "
//         "or overdue back from a loan.",
//     mcp="list_notifications", read_only=True,
// )
func ( handlers *Handlers ) listNotifications( c fiber.Ctx , in *listNotificationsRequest ) ( reply *listNotificationsReply , err error ) {
	user_id := security.UserFrom( c ).ID
	notices , err := models.ListNotifications( handlers.Store , user_id )
	if err != nil { return }
	rows := []notificationRow{}
	reply = &listNotificationsReply{}
	for _ , notice := range notices {
		// Someone who left the household stops hearing about it.
		household , get_err := models.GetHousehold( handlers.Store , notice.HouseholdID )
		if get_err != nil || household.RoleOf( user_id ) == "" { continue }
		if notice.Read == false { reply.Unread += 1 }
		if in.Unread && notice.Read { continue }
		rows = append( rows , notificationRow{ ID: notice.ID , HouseholdID: notice.HouseholdID , ItemID: notice.ItemID , LabelID: notice.LabelID ,
			Kind: notice.Kind , ItemName: notice.ItemName , Due: notice.Due , Read: notice.Read , CreatedAt: notice.CreatedAt } )
	}
	reply.Notifications , reply.Paged = cutPage( rows , in.Page )
	return
}

type markReadRequest struct {
	NotificationID uint64 `json:"notification_id,omitempty" jsonschema_description:"The one to mark read; leave out to mark them all."`
}

type markReadReply struct {
	Marked int `json:"marked"`
}

// @route("POST /api/notifications/read", summary="Mark reminders read",
//     tag="Reminders",
//     description="Mark one reminder read, or all of them.",
//     mcp="mark_notifications_read",
// )
func ( handlers *Handlers ) markNotificationsRead( c fiber.Ctx , in *markReadRequest ) ( reply *markReadReply , err error ) {
	marked , err := models.MarkRead( handlers.Store , security.UserFrom( c ).ID , in.NotificationID )
	if err != nil { return }
	reply = &markReadReply{ Marked: marked }
	return
}

type feedReply struct {
	On   bool   `json:"on"`
	Path string `json:"path" jsonschema_description:"The feed's path, only right after it is made: it is not stored and cannot be shown again. A credential: treat it like a password."`
}

// @route("GET /api/account/calendar-feed", summary="Whether your calendar feed is on",
//     tag="Reminders",
//     description="Whether you have a calendar feed address switched on.",
//     mcp="calendar_feed_status", read_only=True,
// )
func ( handlers *Handlers ) calendarFeedStatus( c fiber.Ctx , _ *apidoc.None ) ( reply *feedReply , err error ) {
	on , err := models.HasFeedToken( handlers.Store , security.UserFrom( c ).ID )
	reply = &feedReply{ On: on }
	return
}

// @route("POST /api/account/calendar-feed", summary="Make a calendar feed address",
//     tag="Reminders", access=SESSION_ONLY,
//     description="Make a new address for a calendar app to subscribe to, listing expiry dates and loan "
//         "due dates. Any address you had before stops working.",
// )
func ( handlers *Handlers ) createCalendarFeed( c fiber.Ctx , _ *apidoc.None ) ( reply *feedReply , err error ) {
	credential , err := models.IssueFeedToken( handlers.Store , security.UserFrom( c ).ID )
	if err != nil { return }
	reply = &feedReply{ On: true , Path: "/api/feed/" + credential + ".ics" }
	return
}

// @route("POST /api/account/calendar-feed/off", summary="Switch your calendar feed off",
//     tag="Reminders", access=SESSION_ONLY,
//     description="Stop your calendar feed address working.",
// )
func ( handlers *Handlers ) revokeCalendarFeed( c fiber.Ctx , _ *apidoc.None ) ( reply *feedReply , err error ) {
	_ , err = models.RevokeFeedToken( handlers.Store , security.UserFrom( c ).ID )
	reply = &feedReply{ On: false }
	return
}

type feedPath struct {
	Token string `json:"token" path:"token" jsonschema_description:"The feed's credential, ending .ics."`
}

// icsText escapes text for a calendar line (RFC 5545 section 3.3.11).
func icsText( value string ) ( escaped string ) {
	escaped = strings.NewReplacer( "\\" , "\\\\" , ";" , "\\;" , "," , "\\," , "\n" , "\\n" , "\r" , "" ).Replace( value )
	return
}

// @route("GET /api/feed/{token}", summary="A calendar feed",
//     tag="Reminders", access=PUBLIC, errors=(404,), reply_type="text/calendar",
//     reply_description="An iCalendar file: one all-day event per expiry date and loan due date",
// )
func ( handlers *Handlers ) calendarFeed( c fiber.Ctx , in *feedPath ) ( err error ) {
	user_id := models.ResolveFeedToken( handlers.Store , strings.TrimSuffix( in.Token , ".ics" ) )
	if user_id == 0 {
		err = notFound( c , "no such feed" )
		return
	}
	households , err := models.ListHouseholds( handlers.Store , user_id )
	if err != nil { return }
	builder := strings.Builder{}
	line := func( text string ) { builder.WriteString( text + "\r\n" ) }
	line( "BEGIN:VCALENDAR" )
	line( "VERSION:2.0" )
	line( "PRODID:-//" + icsText( handlers.cfg().ProjectName ) + "//Reminders//EN" )
	line( "X-WR-CALNAME:" + icsText( handlers.cfg().ProjectName + " reminders" ) )
	stamp := time.Now().UTC().Format( "20060102T150405Z" )
	event := func( uid string , date string , summary string , description string ) {
		when , parse_err := time.Parse( models.DateLayout , date )
		if parse_err != nil { return }
		line( "BEGIN:VEVENT" )
		line( "UID:" + uid )
		line( "DTSTAMP:" + stamp )
		line( "DTSTART;VALUE=DATE:" + when.Format( "20060102" ) )
		line( "SUMMARY:" + icsText( summary ) )
		line( "DESCRIPTION:" + icsText( description ) )
		line( "END:VEVENT" )
	}
	for _ , household := range households {
		found , load_err := handlers.loadInventory( household )
		if load_err != nil {
			err = load_err
			return
		}
		for _ , item := range found.items {
			if item.InTrash() { continue }
			where := joinPath( append( []string{ household.Name } , found.labelPath( item.LabelID )... ) )
			if item.Expires != "" {
				event( fmt.Sprintf( "expires-%d-%s@qrcodeorganizer" , item.ID , item.Expires ) , item.Expires , item.Name + " expires" , where )
			}
			if item.LentTo != "" && item.LentDue != "" {
				event( fmt.Sprintf( "loan-%d-%s@qrcodeorganizer" , item.ID , item.LentDue ) , item.LentDue , item.Name + " due back from " + item.LentTo , where )
			}
		}
	}
	line( "END:VCALENDAR" )
	c.Set( fiber.HeaderContentType , "text/calendar; charset=utf-8" )
	c.Set( fiber.HeaderCacheControl , "private, no-store" )
	err = c.SendString( builder.String() )
	return
}
