package jobs

import (
	errors "errors"
	fmt "fmt"
	time "time"

	config "qrcodeorganizer/server/config"
	db "qrcodeorganizer/server/db"
	models "qrcodeorganizer/server/models"
)

// Emit is how a job says something happened (webhooks.Dispatcher.Emit), handed
// in so this package does not need the webhooks one.
type Emit func( event string , data map[string]any )

// Inventory registers the organizer's own jobs: the reminders that stand in
// for email, and clearing out Trash and old drafts.
func Inventory( registry *Registry , store *db.Store , settings *config.Holder , emit Emit ) {
	registry.Register( "reminders" , time.Hour , reminders( store , settings , emit ) )
	registry.Register( "inventory-trash" , time.Hour , inventoryTrash( store , settings ) )
}

// reminders looks at every item once an hour and tells each member of its
// household -- by a notification, and to webhooks -- when it is about to
// expire or has, has run low, or is overdue back from a loan. Each reminder
// goes out once: its key names the item, the kind and the date it is about.
func reminders( store *db.Store , settings *config.Holder , emit Emit ) ( run Func ) {
	run = func() ( note string , err error ) {
		households , err := models.AllHouseholds( store )
		if err != nil { return }
		today , _ := time.Parse( models.DateLayout , time.Now().UTC().Format( models.DateLayout ) )
		sent := 0
		for _ , household := range households {
			members := []uint64{}
			for user_id := range household.Members { members = append( members , user_id ) }
			lead := settings.Get().Inventory.ReminderDays
			if household.ReminderDays != nil { lead = *household.ReminderDays }
			items , list_err := models.ListItems( store , household.ID )
			if list_err != nil {
				err = errors.Join( err , list_err )
				continue
			}
			for _ , item := range items {
				if item.InTrash() { continue }
				label , _ := models.GetLabel( store , item.LabelID )
				if label == nil || label.InTrash() { continue }
				notice := models.Notification{ HouseholdID: household.ID , ItemID: item.ID , LabelID: item.LabelID , ItemName: item.Name }
				data := map[string]any{ "household_id": household.ID , "item_id": item.ID , "name": item.Name , "label_id": item.LabelID }
				send := func( kind string , due string , event string ) {
					notice.Kind , notice.Due = kind , due
					notice.Key = fmt.Sprintf( "%d:%s:%s" , item.ID , kind , due )
					added , notify_err := models.Notify( store , members , notice )
					if notify_err != nil {
						err = errors.Join( err , notify_err )
						return
					}
					if added > 0 {
						sent += 1
						if event != "" {
							data[ "due" ] = due
							emit( event , data )
						}
					}
				}
				if expires , parse_err := time.Parse( models.DateLayout , item.Expires ); item.Expires != "" && parse_err == nil {
					item_lead := lead
					if item.RemindDays != nil { item_lead = *item.RemindDays }
					days := int( expires.Sub( today ).Hours() / 24 )
					switch {
					case days < 0:
						send( models.NoticeExpired , item.Expires , models.EventItemExpiring )
					case days <= item_lead:
						send( models.NoticeExpiring , item.Expires , models.EventItemExpiring )
					}
				}
				if item.LowStock != nil && item.Quantity <= *item.LowStock {
					// Once per time it runs low: the key carries the day it did,
					// approximated by the last change to the item.
					send( models.NoticeLowStock , item.UpdatedAt.Format( models.DateLayout ) , models.EventItemLowStock )
				}
				if due , parse_err := time.Parse( models.DateLayout , item.LentDue ); item.LentTo != "" && item.LentDue != "" && parse_err == nil && due.Before( today ) {
					send( models.NoticeOverdue , item.LentDue , models.EventLoanOverdue )
				}
			}
		}
		pruned , prune_err := models.PruneNotifications( store )
		err = errors.Join( err , prune_err )
		note = fmt.Sprintf( "%d new reminders, %d old notifications removed" , sent , pruned )
		return
	}
	return
}

// inventoryTrash empties Trash of what has been there past
// inventory.trash_days, and drops week-old photo drafts.
func inventoryTrash( store *db.Store , settings *config.Holder ) ( run Func ) {
	run = func() ( note string , err error ) {
		keep := time.Duration( settings.Get().Inventory.TrashDays ) * 24 * time.Hour
		removed , trash_err := models.PurgeTrash( store , keep )
		drafts , drafts_err := models.PruneExtractions( store )
		err = errors.Join( trash_err , drafts_err )
		note = fmt.Sprintf( "removed %d from Trash, %d photo drafts" , removed , drafts )
		return
	}
	return
}
