# QRCodeOrganizer MCP tools

This file is generated: do not edit it. Change the tools next to their handlers and run `scripts/api_docs.sh`. Every route is in [api.md](api.md).

Every tool POST /mcp offers. Written by scripts/api_docs.sh from the registered tools; do not edit by hand. A key sees the tools its role allows (adminOnly ones need an admin key), and a tool whose feature is switched off answers like one that does not exist. route is the HTTP route the tool stands for, with the same rules; http_only are the routes a key may call that have no tool, and why.

|  |  |
|---|---|
| Endpoint | `POST /mcp` |
| Transport | Streamable HTTP, stateless: one JSON-RPC message per POST, a JSON reply |
| Auth | Authorization: Bearer <API key> |

| Tool | Route | Who | What |
|---|---|---|---|
| [`accept_extraction`](#accept_extraction) | [`POST /api/households/{household_id}/extractions/{extraction_id}/accept`](api.md#post-apihouseholdshousehold_idextractionsextraction_idaccept) | any key | Save a photo draft's items |
| [`adjust_quantity`](#adjust_quantity) | [`POST /api/households/{household_id}/items/{item_id}/adjust`](api.md#post-apihouseholdshousehold_iditemsitem_idadjust) | any key | Add to or take from an item's count |
| [`admin_cancel_login_link`](#admin_cancel_login_link) | [`POST /api/admin/users/{user_id}/cancel-login-link`](api.md#post-apiadminusersuser_idcancel-login-link) | admin keys | Stop an account's login link working (admin) |
| [`admin_cancel_login_link_bulk`](#admin_cancel_login_link_bulk) | [`POST /api/bulk/admin/users/cancel-login-link`](api.md#post-apibulkadminuserscancel-login-link) | admin keys | Stop an account's login link working, for many at once (admin) |
| [`admin_create_invite_link`](#admin_create_invite_link) | [`POST /api/admin/invite-links`](api.md#post-apiadmininvite-links) | admin keys | Create an invite link (admin) |
| [`admin_create_user`](#admin_create_user) | [`POST /api/admin/users`](api.md#post-apiadminusers) | admin keys | Create an account (admin) |
| [`admin_delete_invite_link`](#admin_delete_invite_link) | [`POST /api/admin/invite-links/{link_id}/delete`](api.md#post-apiadmininvite-linkslink_iddelete) | admin keys | Delete an invite link (admin) |
| [`admin_delete_invite_link_bulk`](#admin_delete_invite_link_bulk) | [`POST /api/bulk/admin/invite-links/delete`](api.md#post-apibulkadmininvite-linksdelete) | admin keys | Delete an invite link, for many at once (admin) |
| [`admin_delete_user`](#admin_delete_user) | [`POST /api/admin/users/{user_id}/delete`](api.md#post-apiadminusersuser_iddelete) | admin keys | Delete an account (admin) |
| [`admin_delete_user_bulk`](#admin_delete_user_bulk) | [`POST /api/bulk/admin/users/delete`](api.md#post-apibulkadminusersdelete) | admin keys | Delete an account, for many at once (admin) |
| [`admin_get_theme_defaults`](#admin_get_theme_defaults) | [`GET /api/admin/theme-defaults`](api.md#get-apiadmintheme-defaults) | admin keys | Default themes (admin) |
| [`admin_list_ai_access`](#admin_list_ai_access) | [`GET /api/admin/ai-access`](api.md#get-apiadminai-access) | admin keys | Who may use the server's AI key (admin) |
| [`admin_list_files`](#admin_list_files) | [`GET /api/admin/files`](api.md#get-apiadminfiles) | admin keys | List everyone's files (admin) |
| [`admin_list_invite_links`](#admin_list_invite_links) | [`GET /api/admin/invite-links`](api.md#get-apiadmininvite-links) | admin keys | List invite links (admin) |
| [`admin_list_user_sessions`](#admin_list_user_sessions) | [`GET /api/admin/users/{user_id}/sessions`](api.md#get-apiadminusersuser_idsessions) | admin keys | An account's signed-in devices (admin) |
| [`admin_list_users`](#admin_list_users) | [`GET /api/admin/users`](api.md#get-apiadminusers) | admin keys | List users (admin) |
| [`admin_reissue_login`](#admin_reissue_login) | [`POST /api/admin/users/{user_id}/reissue-login`](api.md#post-apiadminusersuser_idreissue-login) | admin keys | A new login link for an account (admin) |
| [`admin_rename_invite_link`](#admin_rename_invite_link) | [`POST /api/admin/invite-links/{link_id}/rename`](api.md#post-apiadmininvite-linkslink_idrename) | admin keys | Rename an invite link (admin) |
| [`admin_rename_user`](#admin_rename_user) | [`POST /api/admin/users/{user_id}/rename`](api.md#post-apiadminusersuser_idrename) | admin keys | Rename someone (admin) |
| [`admin_reset_two_factor`](#admin_reset_two_factor) | [`POST /api/admin/users/{user_id}/reset-two-factor`](api.md#post-apiadminusersuser_idreset-two-factor) | admin keys | Turn off an account's two-factor (admin) |
| [`admin_reset_two_factor_bulk`](#admin_reset_two_factor_bulk) | [`POST /api/bulk/admin/users/reset-two-factor`](api.md#post-apibulkadminusersreset-two-factor) | admin keys | Turn off an account's two-factor, for many at once (admin) |
| [`admin_restore_invite_link`](#admin_restore_invite_link) | [`POST /api/admin/invite-links/{link_id}/restore`](api.md#post-apiadmininvite-linkslink_idrestore) | admin keys | Undo deleting an invite link (admin) |
| [`admin_restore_invite_link_bulk`](#admin_restore_invite_link_bulk) | [`POST /api/bulk/admin/invite-links/restore`](api.md#post-apibulkadmininvite-linksrestore) | admin keys | Undo deleting an invite link, for many at once (admin) |
| [`admin_revoke_access`](#admin_revoke_access) | [`POST /api/admin/users/{user_id}/revoke-access`](api.md#post-apiadminusersuser_idrevoke-access) | admin keys | Revoke a user's access (admin) |
| [`admin_revoke_access_bulk`](#admin_revoke_access_bulk) | [`POST /api/bulk/admin/users/revoke-access`](api.md#post-apibulkadminusersrevoke-access) | admin keys | Revoke a user's access, for many at once (admin) |
| [`admin_revoke_invite_link`](#admin_revoke_invite_link) | [`POST /api/admin/invite-links/{link_id}/revoke`](api.md#post-apiadmininvite-linkslink_idrevoke) | admin keys | Switch off an invite link (admin) |
| [`admin_revoke_invite_link_bulk`](#admin_revoke_invite_link_bulk) | [`POST /api/bulk/admin/invite-links/revoke`](api.md#post-apibulkadmininvite-linksrevoke) | admin keys | Switch off an invite link, for many at once (admin) |
| [`admin_revoke_user_session`](#admin_revoke_user_session) | [`POST /api/admin/users/{user_id}/sessions/{session_id}/revoke`](api.md#post-apiadminusersuser_idsessionssession_idrevoke) | admin keys | Sign out one of an account's devices (admin) |
| [`admin_revoke_user_session_bulk`](#admin_revoke_user_session_bulk) | [`POST /api/bulk/admin/users/sessions/revoke`](api.md#post-apibulkadminuserssessionsrevoke) | admin keys | Sign out one of an account's devices, for many at once (admin) |
| [`admin_set_ai_access`](#admin_set_ai_access) | [`POST /api/admin/ai-access/{user_id}`](api.md#post-apiadminai-accessuser_id) | admin keys | Set one person's use of the server's AI key (admin) |
| [`admin_set_ai_access_bulk`](#admin_set_ai_access_bulk) | [`POST /api/bulk/admin/ai-access`](api.md#post-apibulkadminai-access) | admin keys | Set one person's use of the server's AI key, for many at once (admin) |
| [`admin_set_ai_default`](#admin_set_ai_default) | [`POST /api/admin/ai-access`](api.md#post-apiadminai-access) | admin keys | Set the default for the server's AI key (admin) |
| [`admin_set_invite_link_limits`](#admin_set_invite_link_limits) | [`POST /api/admin/invite-links/{link_id}/limits`](api.md#post-apiadmininvite-linkslink_idlimits) | admin keys | Change an invite link's places and expiry (admin) |
| [`admin_set_theme_defaults`](#admin_set_theme_defaults) | [`POST /api/admin/theme-defaults`](api.md#post-apiadmintheme-defaults) | admin keys | Replace the default themes (admin) |
| [`admin_set_user_disabled`](#admin_set_user_disabled) | [`POST /api/admin/users/{user_id}/disabled`](api.md#post-apiadminusersuser_iddisabled) | admin keys | Disable or enable a user (admin) |
| [`admin_set_user_disabled_bulk`](#admin_set_user_disabled_bulk) | [`POST /api/bulk/admin/users/disabled`](api.md#post-apibulkadminusersdisabled) | admin keys | Disable or enable a user, for many at once (admin) |
| [`admin_set_user_role`](#admin_set_user_role) | [`POST /api/admin/users/{user_id}/role`](api.md#post-apiadminusersuser_idrole) | admin keys | Change an account's role (admin) |
| [`admin_set_user_role_bulk`](#admin_set_user_role_bulk) | [`POST /api/bulk/admin/users/role`](api.md#post-apibulkadminusersrole) | admin keys | Change an account's role, for many at once (admin) |
| [`admin_storage`](#admin_storage) | [`GET /api/admin/storage`](api.md#get-apiadminstorage) | admin keys | Storage for everyone (admin) |
| [`admin_switch_on_invite_link`](#admin_switch_on_invite_link) | [`POST /api/admin/invite-links/{link_id}/switch-on`](api.md#post-apiadmininvite-linkslink_idswitch-on) | admin keys | Switch an invite link back on (admin) |
| [`admin_switch_on_invite_link_bulk`](#admin_switch_on_invite_link_bulk) | [`POST /api/bulk/admin/invite-links/switch-on`](api.md#post-apibulkadmininvite-linksswitch-on) | admin keys | Switch an invite link back on, for many at once (admin) |
| [`ai_status`](#ai_status) | [`GET /api/ai`](api.md#get-apiai) | any key | Whether Smart Photo Extraction is ready for you |
| [`calendar_feed_status`](#calendar_feed_status) | [`GET /api/account/calendar-feed`](api.md#get-apiaccountcalendar-feed) | any key | Whether your calendar feed is on |
| [`cancel_my_login_link`](#cancel_my_login_link) | [`POST /api/account/login-link/cancel`](api.md#post-apiaccountlogin-linkcancel) | any key | Stop your own login link working |
| [`claim_code`](#claim_code) | [`POST /api/households/{household_id}/labels/{label_id}/codes`](api.md#post-apihouseholdshousehold_idlabelslabel_idcodes) | any key | Add a code to a container |
| [`create_household`](#create_household) | [`POST /api/households`](api.md#post-apihouseholds) | any key | Create a household |
| [`create_items`](#create_items) | [`POST /api/households/{household_id}/labels/{label_id}/items`](api.md#post-apihouseholdshousehold_idlabelslabel_iditems) | any key | Put items in a container |
| [`create_label`](#create_label) | [`POST /api/households/{household_id}/labels`](api.md#post-apihouseholdshousehold_idlabels) | any key | Set up a container |
| [`create_location`](#create_location) | [`POST /api/households/{household_id}/locations`](api.md#post-apihouseholdshousehold_idlocations) | any key | Add a location |
| [`create_my_login_link`](#create_my_login_link) | [`POST /api/account/login-link`](api.md#post-apiaccountlogin-link) | any key | A login link for yourself |
| [`create_sheet`](#create_sheet) | [`POST /api/households/{household_id}/sheets`](api.md#post-apihouseholdshousehold_idsheets) | any key | Make codes to print |
| [`delete_household`](#delete_household) | [`POST /api/households/{household_id}/delete`](api.md#post-apihouseholdshousehold_iddelete) | any key | Delete a household |
| [`delete_item`](#delete_item) | [`POST /api/households/{household_id}/items/{item_id}/delete`](api.md#post-apihouseholdshousehold_iditemsitem_iddelete) | any key | Delete an item |
| [`delete_item_bulk`](#delete_item_bulk) | [`POST /api/bulk/households/items/delete`](api.md#post-apibulkhouseholdsitemsdelete) | any key | Delete an item, for many at once |
| [`delete_label`](#delete_label) | [`POST /api/households/{household_id}/labels/{label_id}/delete`](api.md#post-apihouseholdshousehold_idlabelslabel_iddelete) | any key | Delete a container |
| [`delete_label_bulk`](#delete_label_bulk) | [`POST /api/bulk/households/labels/delete`](api.md#post-apibulkhouseholdslabelsdelete) | any key | Delete a container, for many at once |
| [`delete_location`](#delete_location) | [`POST /api/households/{household_id}/locations/{location_id}/delete`](api.md#post-apihouseholdshousehold_idlocationslocation_iddelete) | any key | Delete a location |
| [`delete_location_bulk`](#delete_location_bulk) | [`POST /api/bulk/households/locations/delete`](api.md#post-apibulkhouseholdslocationsdelete) | any key | Delete a location, for many at once |
| [`delete_sheet`](#delete_sheet) | [`POST /api/households/{household_id}/sheets/{sheet_id}/delete`](api.md#post-apihouseholdshousehold_idsheetssheet_iddelete) | any key | Forget a print run |
| [`extract_photo`](#extract_photo) | [`POST /api/households/{household_id}/labels/{label_id}/extract`](api.md#post-apihouseholdshousehold_idlabelslabel_idextract) | any key | Read a photo of a container's contents |
| [`files_delete`](#files_delete) | [`POST /api/files/{file_id}/delete`](api.md#post-apifilesfile_iddelete) | any key | Delete a file |
| [`files_delete_bulk`](#files_delete_bulk) | [`POST /api/bulk/files/delete`](api.md#post-apibulkfilesdelete) | any key | Delete a file, for many at once |
| [`files_download_url`](#files_download_url) | [`GET /api/files/{file_id}/download`](api.md#get-apifilesfile_iddownload) | any key | Where to download a file |
| [`files_finish_upload`](#files_finish_upload) | [`POST /api/files/{file_id}/finish`](api.md#post-apifilesfile_idfinish) | any key | Finish an upload |
| [`files_get`](#files_get) | [`GET /api/files/{file_id}`](api.md#get-apifilesfile_id) | any key | One of your files |
| [`files_list`](#files_list) | [`GET /api/files`](api.md#get-apifiles) | any key | List your files |
| [`files_rename`](#files_rename) | [`POST /api/files/{file_id}/rename`](api.md#post-apifilesfile_idrename) | any key | Rename a file |
| [`files_restore`](#files_restore) | [`POST /api/files/{file_id}/restore`](api.md#post-apifilesfile_idrestore) | any key | Restore a deleted file |
| [`files_restore_bulk`](#files_restore_bulk) | [`POST /api/bulk/files/restore`](api.md#post-apibulkfilesrestore) | any key | Restore a deleted file, for many at once |
| [`files_start_upload`](#files_start_upload) | [`POST /api/files`](api.md#post-apifiles) | any key | Start uploading a file |
| [`get_extraction`](#get_extraction) | [`GET /api/households/{household_id}/extractions/{extraction_id}`](api.md#get-apihouseholdshousehold_idextractionsextraction_id) | any key | A photo draft |
| [`get_household`](#get_household) | [`GET /api/households/{household_id}`](api.md#get-apihouseholdshousehold_id) | any key | One household |
| [`get_item`](#get_item) | [`GET /api/households/{household_id}/items/{item_id}`](api.md#get-apihouseholdshousehold_iditemsitem_id) | any key | One item |
| [`get_label`](#get_label) | [`GET /api/households/{household_id}/labels/{label_id}`](api.md#get-apihouseholdshousehold_idlabelslabel_id) | any key | One container and what is in it |
| [`import_csv`](#import_csv) | [`POST /api/households/{household_id}/import`](api.md#post-apihouseholdshousehold_idimport) | any key | Add items from a spreadsheet |
| [`label_my_session`](#label_my_session) | [`POST /api/account/sessions/{session_id}/label`](api.md#post-apiaccountsessionssession_idlabel) | any key | Name one of your devices |
| [`leave_household`](#leave_household) | [`POST /api/households/{household_id}/leave`](api.md#post-apihouseholdshousehold_idleave) | any key | Leave a household |
| [`list_households`](#list_households) | [`GET /api/households`](api.md#get-apihouseholds) | any key | List your households |
| [`list_items`](#list_items) | [`GET /api/households/{household_id}/items`](api.md#get-apihouseholdshousehold_iditems) | any key | List and search items |
| [`list_labels`](#list_labels) | [`GET /api/households/{household_id}/labels`](api.md#get-apihouseholdshousehold_idlabels) | any key | List containers |
| [`list_locations`](#list_locations) | [`GET /api/households/{household_id}/locations`](api.md#get-apihouseholdshousehold_idlocations) | any key | List locations |
| [`list_my_sessions`](#list_my_sessions) | [`GET /api/account/sessions`](api.md#get-apiaccountsessions) | any key | Your signed-in devices |
| [`list_notifications`](#list_notifications) | [`GET /api/notifications`](api.md#get-apinotifications) | any key | Your reminders |
| [`list_sheets`](#list_sheets) | [`GET /api/households/{household_id}/sheets`](api.md#get-apihouseholdshousehold_idsheets) | any key | List print runs |
| [`list_team`](#list_team) | [`GET /api/team`](api.md#get-apiteam) | any key | List team |
| [`mark_notifications_read`](#mark_notifications_read) | [`POST /api/notifications/read`](api.md#post-apinotificationsread) | any key | Mark reminders read |
| [`move_contents`](#move_contents) | [`POST /api/households/{household_id}/labels/{label_id}/move-contents`](api.md#post-apihouseholdshousehold_idlabelslabel_idmove-contents) | any key | Move everything to another container |
| [`move_item`](#move_item) | [`POST /api/households/{household_id}/items/{item_id}/move`](api.md#post-apihouseholdshousehold_iditemsitem_idmove) | any key | Move an item to another container |
| [`move_item_bulk`](#move_item_bulk) | [`POST /api/bulk/households/items/move`](api.md#post-apibulkhouseholdsitemsmove) | any key | Move an item to another container, for many at once |
| [`my_ai_settings`](#my_ai_settings) | [`GET /api/account/ai`](api.md#get-apiaccountai) | any key | Your photo-reading settings |
| [`my_storage`](#my_storage) | [`GET /api/account/storage`](api.md#get-apiaccountstorage) | any key | Your storage |
| [`release_code`](#release_code) | [`POST /api/households/{household_id}/labels/{label_id}/codes/release`](api.md#post-apihouseholdshousehold_idlabelslabel_idcodesrelease) | any key | Take a code off a container |
| [`rename_account`](#rename_account) | [`POST /api/account/rename`](api.md#post-apiaccountrename) | any key | Rename your account |
| [`rename_household`](#rename_household) | [`POST /api/households/{household_id}/rename`](api.md#post-apihouseholdshousehold_idrename) | any key | Rename a household |
| [`reprint_labels`](#reprint_labels) | [`POST /api/households/{household_id}/labels/reprint`](api.md#post-apihouseholdshousehold_idlabelsreprint) | any key | Codes to print again |
| [`restore_item`](#restore_item) | [`POST /api/households/{household_id}/items/{item_id}/restore`](api.md#post-apihouseholdshousehold_iditemsitem_idrestore) | any key | Restore an item |
| [`restore_item_bulk`](#restore_item_bulk) | [`POST /api/bulk/households/items/restore`](api.md#post-apibulkhouseholdsitemsrestore) | any key | Restore an item, for many at once |
| [`restore_label`](#restore_label) | [`POST /api/households/{household_id}/labels/{label_id}/restore`](api.md#post-apihouseholdshousehold_idlabelslabel_idrestore) | any key | Restore a container |
| [`restore_label_bulk`](#restore_label_bulk) | [`POST /api/bulk/households/labels/restore`](api.md#post-apibulkhouseholdslabelsrestore) | any key | Restore a container, for many at once |
| [`restore_location`](#restore_location) | [`POST /api/households/{household_id}/locations/{location_id}/restore`](api.md#post-apihouseholdshousehold_idlocationslocation_idrestore) | any key | Restore a location |
| [`restore_location_bulk`](#restore_location_bulk) | [`POST /api/bulk/households/locations/restore`](api.md#post-apibulkhouseholdslocationsrestore) | any key | Restore a location, for many at once |
| [`revoke_my_other_sessions`](#revoke_my_other_sessions) | [`POST /api/account/sessions/revoke-others`](api.md#post-apiaccountsessionsrevoke-others) | any key | Sign out all your devices |
| [`revoke_my_session`](#revoke_my_session) | [`POST /api/account/sessions/{session_id}/revoke`](api.md#post-apiaccountsessionssession_idrevoke) | any key | Sign out one of your devices |
| [`revoke_my_session_bulk`](#revoke_my_session_bulk) | [`POST /api/bulk/account/sessions/revoke`](api.md#post-apibulkaccountsessionsrevoke) | any key | Sign out one of your devices, for many at once |
| [`scan_code`](#scan_code) | [`GET /api/scan`](api.md#get-apiscan) | any key | Look up a scanned code |
| [`search_inventory`](#search_inventory) | [`GET /api/search`](api.md#get-apisearch) | any key | Find where something is |
| [`set_household_colors`](#set_household_colors) | [`POST /api/households/{household_id}/colors`](api.md#post-apihouseholdshousehold_idcolors) | any key | Say what label colors mean |
| [`set_household_fields`](#set_household_fields) | [`POST /api/households/{household_id}/fields`](api.md#post-apihouseholdshousehold_idfields) | any key | Set custom fields and reminder lead |
| [`set_household_member`](#set_household_member) | [`POST /api/households/{household_id}/members`](api.md#post-apihouseholdshousehold_idmembers) | any key | Add, change or remove a member |
| [`set_shortcuts`](#set_shortcuts) | [`POST /api/account/shortcuts`](api.md#post-apiaccountshortcuts) | any key | Change your keyboard shortcuts |
| [`set_theme`](#set_theme) | [`POST /api/account/theme`](api.md#post-apiaccounttheme) | any key | Pick your theme |
| [`set_themes`](#set_themes) | [`POST /api/account/themes`](api.md#post-apiaccountthemes) | any key | Replace your own themes |
| [`update_item`](#update_item) | [`POST /api/households/{household_id}/items/{item_id}`](api.md#post-apihouseholdshousehold_iditemsitem_id) | any key | Change an item |
| [`update_label`](#update_label) | [`POST /api/households/{household_id}/labels/{label_id}`](api.md#post-apihouseholdshousehold_idlabelslabel_id) | any key | Change a container |
| [`update_location`](#update_location) | [`POST /api/households/{household_id}/locations/{location_id}`](api.md#post-apihouseholdshousehold_idlocationslocation_id) | any key | Rename or move a location |
| [`whoami`](#whoami) | [`GET /api/me`](api.md#get-apime) | any key | Who am I |

## accept_extraction

Put the items you kept from a draft into its container, and mark the draft done.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/extractions/{extraction_id}/accept`](api.md#post-apihouseholdshousehold_idextractionsextraction_idaccept) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `extraction_id` | integer | yes | at least 1 | The draft's id, from extract_photo. |
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `items` | array of object | yes |  | The items to keep, as they should be saved -- the draft's, edited as needed, with photos cropped from it if you like. |
| `items[].barcode` | string or null | no | at most 200 characters |  |
| `items[].brand` | string or null | no | at most 80 characters |  |
| `items[].clear_low_stock` | boolean | no |  | Stop low-stock reminders. |
| `items[].clear_remind_days` | boolean | no |  | Use the household's reminder lead again. |
| `items[].description` | string or null | no | at most 4000 characters |  |
| `items[].expires` | string or null | no |  | Expiry date, YYYY-MM-DD, or empty. |
| `items[].fields` | map of string or null | no |  | Custom field values by field id (see the household's fields). Replaces them all. |
| `items[].lent_due` | string or null | no |  | When it should come back, YYYY-MM-DD, or empty. |
| `items[].lent_to` | string or null | no | at most 80 characters | Who has it on loan; empty when it is back. |
| `items[].low_stock` | number or null | no | at least 0 | Remind when the quantity drops to this or below. |
| `items[].name` | string or null | no | 1 to 80 characters |  |
| `items[].photos` | array of string or null | no |  | File ids of its photos, in order. |
| `items[].purchase_date` | string or null | no |  | YYYY-MM-DD, or empty. |
| `items[].quantity` | number or null | no | at least 0 | How many there are. A new item has 1. |
| `items[].receipts` | array of string or null | no |  | File ids of receipts or warranty papers. |
| `items[].remind_days` | integer or null | no | 0 to 365 | Days before the expiry to remind, instead of the household's lead. |
| `items[].serial` | string or null | no | at most 120 characters |  |
| `items[].tags` | array of string or null | no |  | Words to find it by. Stored in lower case. |
| `items[].unit` | string or null | no | at most 20 characters | pcs, rolls, kg, ... |
| `items[].value` | number or null | no | at least 0 | What one is worth, for the insurance list. |

## adjust_quantity

Change an item's quantity by an amount, such as -1 when one is used up. Two people doing this at once both count.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/items/{item_id}/adjust`](api.md#post-apihouseholdshousehold_iditemsitem_idadjust) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `delta` | number | yes |  | How many to add; negative to take some out. Never goes below 0. |
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `item_id` | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |

## admin_cancel_login_link

Void an account's live login link without making a new one: a link that went somewhere it should not have. Devices it already signed in stay signed in; admin_revoke_access ends those too. Answers with removed: 0 or 1.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/cancel-login-link`](api.md#post-apiadminusersuser_idcancel-login-link) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_cancel_login_link_bulk

Runs admin_cancel_login_link once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/users/cancel-login-link`](api.md#post-apibulkadminuserscancel-login-link) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_create_invite_link

One link for a group: each person who opens it joins as a new user, up to max_uses people, for days days. Answers with join_url (shown only now) and the link's details. Revoking or deleting this API key switches the link off.

|  |  |
|---|---|
| Route | [`POST /api/admin/invite-links`](api.md#post-apiadmininvite-links) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `days` | integer | yes | 1 to 365 |  |
| `label` | string | yes | 1 to 80 characters |  |
| `max_uses` | integer | yes | 1 to 1000 |  |

## admin_create_user

Make a new account and its first login link, to hand to the person. role is admin or user. Answers with user_id and the link. Answers with login_url (open it to sign in; it works once unless max_uses says otherwise), expires_at and max_uses. Any older link the account could still use stops working; its signed-in devices stay signed in. Revoking or deleting this API key voids the link while it is unused, and signs out every device it signed in.

|  |  |
|---|---|
| Route | [`POST /api/admin/users`](api.md#post-apiadminusers) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `display_name` | string | yes | 1 to 80 characters |  |
| `expires_in_minutes` | integer or null | no | 0 to 525600 | 0 never expires. Default auth.login_token_ttl_seconds for a link that works once, auth.multi_use_login_token_ttl_seconds otherwise. |
| `max_uses` | integer or null | no | 0 to 1000 | Sign-ins the link allows: 1-1000, or 0 for unlimited. Default 1. |
| `role` | string | yes | one of `admin`, `user` |  |

## admin_delete_invite_link

Take an invite link off the list and stop it working. The accounts it made stay. admin_restore_invite_link undoes it for a few minutes.

|  |  |
|---|---|
| Route | [`POST /api/admin/invite-links/{link_id}/delete`](api.md#post-apiadmininvite-linkslink_iddelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

## admin_delete_invite_link_bulk

Runs admin_delete_invite_link once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/invite-links/delete`](api.md#post-apibulkadmininvite-linksdelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_delete_user

Remove an account for good, with its sessions, login link and API keys. Cannot be undone: admin_set_user_disabled is the reversible version. Only the owner deletes an admin; nobody deletes the owner or themselves.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/delete`](api.md#post-apiadminusersuser_iddelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_delete_user_bulk

Runs admin_delete_user once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/users/delete`](api.md#post-apibulkadminusersdelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_get_theme_defaults

The theme list everyone without their own list sees. themes is null while the built-in ones apply.

|  |  |
|---|---|
| Route | [`GET /api/admin/theme-defaults`](api.md#get-apiadmintheme-defaults) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|

## admin_list_ai_access

The default for everyone, and each person's choice: follow the default, may use the server's key, or must bring their own. Someone with their own key always uses it.

|  |  |
|---|---|
| Route | [`GET /api/admin/ai-access`](api.md#get-apiadminai-access) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

## admin_list_files

Every account's files, a page at a time, with whose each one is (user_id, user_name). next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Route | [`GET /api/admin/files`](api.md#get-apiadminfiles) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `dir` | string | no | one of `asc`, `desc` | Default asc. |
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | string | no |  | Keep files whose name or owner contains this, in any case. |
| `sort` | string | no | one of `name`, `user_name`, `size`, `created_at` | Field to sort the whole list by. Rows without it go last. |

## admin_list_invite_links

Group invite links, a page at a time: label, places used and left, expiry, and whether each still works. The links themselves are never shown again after creation. next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Route | [`GET /api/admin/invite-links`](api.md#get-apiadmininvite-links) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `dir` | string | no | one of `asc`, `desc` | Default asc. |
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | string | no |  | Keep links whose label or maker contains this, in any case. |
| `sort` | string | no | one of `label`, `used`, `expires_at` | Field to sort the whole list by. Rows without it go last. |

## admin_list_user_sessions

Where an account is signed in: every session, with when it started and was last seen. Addresses are shown to the owner only, and only the owner sees the owner's.

|  |  |
|---|---|
| Route | [`GET /api/admin/users/{user_id}/sessions`](api.md#get-apiadminusersuser_idsessions) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_list_users

Accounts, a page at a time, including disabled ones, with each one's role and which is the owner. next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Route | [`GET /api/admin/users`](api.md#get-apiadminusers) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `dir` | string | no | one of `asc`, `desc` | Default asc. |
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | string | no |  | Keep accounts whose name or username contains this, in any case. |
| `sort` | string | no | one of `display_name`, `role`, `disabled`, `created_at` | Field to sort the whole list by. Rows without it go last. |

## admin_reissue_login

For someone who lost their device or their link: a fresh login link for their account. Answers with login_url (open it to sign in; it works once unless max_uses says otherwise), expires_at and max_uses. Any older link the account could still use stops working; its signed-in devices stay signed in. Revoking or deleting this API key voids the link while it is unused, and signs out every device it signed in.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/reissue-login`](api.md#post-apiadminusersuser_idreissue-login) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expires_in_minutes` | integer or null | no | 0 to 525600 | 0 never expires. Default auth.login_token_ttl_seconds for a link that works once, auth.multi_use_login_token_ttl_seconds otherwise. |
| `max_uses` | integer or null | no | 0 to 1000 | Sign-ins the link allows: 1-1000, or 0 for unlimited. Default 1. |
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_rename_invite_link

Change the label people see on an invite link's join page (1-80 characters). expect, when given, is the label you think it has now.

|  |  |
|---|---|
| Route | [`POST /api/admin/invite-links/{link_id}/rename`](api.md#post-apiadmininvite-linkslink_idrename) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expect` | string or null | no |  | The label you think it has. Answers 409 if it has another. |
| `label` | string | yes | 1 to 80 characters | 1-80 characters. |
| `link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

## admin_rename_user

Change another account's display name (1-80 characters). Only the owner renames an admin. expect, when given, is the name you think it has now; if it changed meanwhile the rename is refused with the current name.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/rename`](api.md#post-apiadminusersuser_idrename) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `display_name` | string | yes | 1 to 80 characters | 1-80 characters. |
| `expect` | string or null | no |  | The name you think they have. Answers 409 if it is another. |
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_reset_two_factor

For someone who lost both their phone and their recovery codes: switches their two-factor sign-in off so a login link lets them in again. Only the owner resets an admin's; the owner's is reset from a shell.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/reset-two-factor`](api.md#post-apiadminusersuser_idreset-two-factor) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_reset_two_factor_bulk

Runs admin_reset_two_factor once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/users/reset-two-factor`](api.md#post-apibulkadminusersreset-two-factor) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_restore_invite_link

Bring back an invite link deleted in the last few minutes, working as before. The link then counts as this key's: revoking the key switches it off.

|  |  |
|---|---|
| Route | [`POST /api/admin/invite-links/{link_id}/restore`](api.md#post-apiadmininvite-linkslink_idrestore) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

## admin_restore_invite_link_bulk

Runs admin_restore_invite_link once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/invite-links/restore`](api.md#post-apibulkadmininvite-linksrestore) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_revoke_access

For an account that may be compromised: ends every session and voids its live login link and every API key it holds, at once. The account stays enabled; admin_reissue_login lets its owner back in. The owner and your own account cannot be revoked.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/revoke-access`](api.md#post-apiadminusersuser_idrevoke-access) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_revoke_access_bulk

Runs admin_revoke_access once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/users/revoke-access`](api.md#post-apibulkadminusersrevoke-access) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_revoke_invite_link

Stop an invite link from letting anyone else join or sign in through it. Accounts it already made keep working. Cannot be switched back on.

|  |  |
|---|---|
| Route | [`POST /api/admin/invite-links/{link_id}/revoke`](api.md#post-apiadmininvite-linkslink_idrevoke) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

## admin_revoke_invite_link_bulk

Runs admin_revoke_invite_link once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/invite-links/revoke`](api.md#post-apibulkadmininvite-linksrevoke) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_revoke_user_session

End one of an account's sessions, rather than all of them with admin_revoke_access.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/sessions/{session_id}/revoke`](api.md#post-apiadminusersuser_idsessionssession_idrevoke) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `session_id` | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_revoke_user_session_bulk

Runs admin_revoke_user_session once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/users/sessions/revoke`](api.md#post-apibulkadminuserssessionsrevoke) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].session_id` | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_set_ai_access

Let one person use the server's key, make them bring their own, or have them follow the default.

|  |  |
|---|---|
| Route | [`POST /api/admin/ai-access/{user_id}`](api.md#post-apiadminai-accessuser_id) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `access` | string | yes | one of `default`, `shared`, `own` | default follows the default; shared may use the server's key; own must bring their own. |
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_set_ai_access_bulk

Runs admin_set_ai_access once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/ai-access`](api.md#post-apibulkadminai-access) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].access` | string | no | one of `default`, `shared`, `own` | default follows the default; shared may use the server's key; own must bring their own. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |
| `shared.access` | string | no | one of `default`, `shared`, `own` | default follows the default; shared may use the server's key; own must bring their own. |

## admin_set_ai_default

Whether people without their own key may use the server's by default.

|  |  |
|---|---|
| Route | [`POST /api/admin/ai-access`](api.md#post-apiadminai-access) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `default_shared` | boolean | yes |  | true lets everyone without their own key use the server's; false makes everyone bring their own, unless their row says otherwise. |

## admin_set_invite_link_limits

Set how many people the link may take in (never fewer than have joined) and, when days is given, a new expiry counted from now. The link then counts as this key's: revoking the key switches it off.

|  |  |
|---|---|
| Route | [`POST /api/admin/invite-links/{link_id}/limits`](api.md#post-apiadmininvite-linkslink_idlimits) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `days` | integer or null | no | 1 to 365 | Counted from now. Leave out to keep the expiry. |
| `link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `max_uses` | integer | yes | 1 to 1000 | Not fewer than have joined. |

## admin_set_theme_defaults

Replace the theme list everyone without their own list sees, 1 to 40 themes in the Theme schema of the API reference. null puts the built-in themes back.

|  |  |
|---|---|
| Route | [`POST /api/admin/theme-defaults`](api.md#post-apiadmintheme-defaults) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `builtins` | integer | no |  | Which built-in themes this list has seen: the highest since in Theme.builtIn (static/js/theme.js) when it was saved. Built-in themes added after that join the list the browser shows. Left out: 0, the first ones. |
| `themes` | array of object or null | no |  | 1 to 40 themes. null goes back to the built-in ones. |
| `themes[].accent` | string | yes |  | The main color: what you can press. |
| `themes[].background` | string | no |  | The page. Left out: the mode's own. |
| `themes[].button` | object or null | no |  | How the theme's buttons look. A choice left out is the default (rounded, normal, raised, normal). |
| `themes[].button.case` | string | no | one of `normal`, `upper` |  |
| `themes[].button.depth` | string | no | one of `flat`, `raised`, `glossy` |  |
| `themes[].button.shape` | string | no | one of `square`, `soft`, `rounded`, `pill` |  |
| `themes[].button.weight` | string | no | one of `normal`, `medium`, `bold` |  |
| `themes[].dark` | object or null | no |  | A light theme's colors for dark mode, kept while it is light. Switching the theme to dark swaps them in. A light theme's light (or a dark theme's dark) is dropped. |
| `themes[].dark.accent` | string | yes |  | The main color: what you can press. |
| `themes[].dark.background` | string | no |  | The page. Left out: the mode's own. |
| `themes[].dark.overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `themes[].dark.second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `themes[].dark.surface` | string | no |  | Cards. Left out: the mode's own. |
| `themes[].id` | string | yes |  | Lower-case letters, digits, - and _, up to 32. |
| `themes[].light` | object or null | no |  | A dark theme's colors for light mode, kept while it is dark. Switching the theme to light swaps them in. |
| `themes[].light.accent` | string | yes |  | The main color: what you can press. |
| `themes[].light.background` | string | no |  | The page. Left out: the mode's own. |
| `themes[].light.overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `themes[].light.second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `themes[].light.surface` | string | no |  | Cards. Left out: the mode's own. |
| `themes[].mode` | string | yes | one of `light`, `dark` |  |
| `themes[].name` | string | yes |  | 1 to 60 characters. |
| `themes[].overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `themes[].second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `themes[].surface` | string | no |  | Cards. Left out: the mode's own. |

## admin_set_user_disabled

Switch an account off (disabled: true) or back on (false). Disabling signs it out everywhere and voids its live login link; its API keys stop working while it is off. Reversible. The owner and your own account cannot be disabled.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/disabled`](api.md#post-apiadminusersuser_iddisabled) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `disabled` | boolean | yes |  |  |
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_set_user_disabled_bulk

Runs admin_set_user_disabled once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/users/disabled`](api.md#post-apibulkadminusersdisabled) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].disabled` | boolean | no |  |  |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |
| `shared.disabled` | boolean | no |  |  |

## admin_set_user_role

Make an account an admin or an ordinary user. Only the owner takes an admin's rights away; the owner is always an admin, and nobody changes their own role. Becoming an admin ends the account's sessions and live login link.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/role`](api.md#post-apiadminusersuser_idrole) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `role` | string | yes | one of `admin`, `user` |  |
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_set_user_role_bulk

Runs admin_set_user_role once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/users/role`](api.md#post-apibulkadminusersrole) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].role` | string | no | one of `admin`, `user` |  |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |
| `shared.role` | string | no | one of `admin`, `user` |  |

## admin_storage

How many bytes everyone's files use together, and the disk's free and total bytes where the system says.

|  |  |
|---|---|
| Route | [`GET /api/admin/storage`](api.md#get-apiadminstorage) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|

## admin_switch_on_invite_link

Undo admin_revoke_invite_link: anyone still holding the link can use it again. The link then counts as this key's: revoking the key switches it off.

|  |  |
|---|---|
| Route | [`POST /api/admin/invite-links/{link_id}/switch-on`](api.md#post-apiadmininvite-linkslink_idswitch-on) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

## admin_switch_on_invite_link_bulk

Runs admin_switch_on_invite_link once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/invite-links/switch-on`](api.md#post-apibulkadmininvite-linksswitch-on) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## ai_status

Whether your photos can be read, whose key they would go on (your own or the server's), and which service and model would read them.

|  |  |
|---|---|
| Route | [`GET /api/ai`](api.md#get-apiai) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|

## calendar_feed_status

Whether you have a calendar feed address switched on.

|  |  |
|---|---|
| Route | [`GET /api/account/calendar-feed`](api.md#get-apiaccountcalendar-feed) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|

## cancel_my_login_link

Void your account's live login link without making a new one. Devices it already signed in stay signed in. Answers with removed: 0 or 1.

|  |  |
|---|---|
| Route | [`POST /api/account/login-link/cancel`](api.md#post-apiaccountlogin-linkcancel) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|

## claim_code

Make another code open this container too: a store-bought QR sticker, a product's barcode, or one of this household's printed codes.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/labels/{label_id}/codes`](api.md#post-apihouseholdshousehold_idlabelslabel_idcodes) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `code` | string | yes | 1 to 2000 characters | The scanned value: a QR code's text or a barcode's digits. |
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |

## create_household

Make a new shared inventory. You are its first owner.

|  |  |
|---|---|
| Route | [`POST /api/households`](api.md#post-apihouseholds) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `name` | string | yes | 1 to 80 characters |  |

## create_items

Add up to 100 items to a container at once, each with at least a name. A new item's quantity is 1 unless you say.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/labels/{label_id}/items`](api.md#post-apihouseholdshousehold_idlabelslabel_iditems) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `items` | array of object | yes |  | The items to put in, each with at least a name. |
| `items[].barcode` | string or null | no | at most 200 characters |  |
| `items[].brand` | string or null | no | at most 80 characters |  |
| `items[].clear_low_stock` | boolean | no |  | Stop low-stock reminders. |
| `items[].clear_remind_days` | boolean | no |  | Use the household's reminder lead again. |
| `items[].description` | string or null | no | at most 4000 characters |  |
| `items[].expires` | string or null | no |  | Expiry date, YYYY-MM-DD, or empty. |
| `items[].fields` | map of string or null | no |  | Custom field values by field id (see the household's fields). Replaces them all. |
| `items[].lent_due` | string or null | no |  | When it should come back, YYYY-MM-DD, or empty. |
| `items[].lent_to` | string or null | no | at most 80 characters | Who has it on loan; empty when it is back. |
| `items[].low_stock` | number or null | no | at least 0 | Remind when the quantity drops to this or below. |
| `items[].name` | string or null | no | 1 to 80 characters |  |
| `items[].photos` | array of string or null | no |  | File ids of its photos, in order. |
| `items[].purchase_date` | string or null | no |  | YYYY-MM-DD, or empty. |
| `items[].quantity` | number or null | no | at least 0 | How many there are. A new item has 1. |
| `items[].receipts` | array of string or null | no |  | File ids of receipts or warranty papers. |
| `items[].remind_days` | integer or null | no | 0 to 365 | Days before the expiry to remind, instead of the household's lead. |
| `items[].serial` | string or null | no | at most 120 characters |  |
| `items[].tags` | array of string or null | no |  | Words to find it by. Stored in lower case. |
| `items[].unit` | string or null | no | at most 20 characters | pcs, rolls, kg, ... |
| `items[].value` | number or null | no | at least 0 | What one is worth, for the insurance list. |
| `label_id` | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |

## create_label

Make a container: give it a name, and optionally the code on its sticker, a color, where it is, and the container it sits in.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/labels`](api.md#post-apihouseholdshousehold_idlabels) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `code` | string | no | at most 2000 characters | A printed code that is not on anything yet (from a print run), or any other QR or barcode value to claim, such as a bought sticker. Leave out to make a fresh code. |
| `color` | string or null | no |  | Its label color, #rrggbb, or empty. |
| `container_type` | string or null | no | at most 40 characters | bin, box, drawer, shelf, bag, or anything else. |
| `destination` | string or null | no | at most 80 characters | For a move: the room it goes to. |
| `flags` | array of string or null | no |  | Handling notes printed on its label, such as fragile, this side up, heavy. |
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `location_id` | integer or null | no |  | Where it sits; 0 for nowhere. |
| `move_status` | string or null | no | one of ``, `packed`, `in_transit`, `unpacked` | For a move: packed, in_transit, unpacked, or empty. |
| `name` | string or null | no | 1 to 80 characters |  |
| `notes` | string or null | no | at most 4000 characters |  |
| `parent_id` | integer or null | no |  | The container it sits inside; 0 for none. |
| `photos` | array of string or null | no |  | File ids of its photos, in order (see attach_photo). |

## create_location

Add a place containers can sit, optionally inside another (Garage inside House).

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/locations`](api.md#post-apihouseholdshousehold_idlocations) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `name` | string | yes | 1 to 80 characters |  |
| `parent_id` | integer | no |  | The location it is inside; leave out for a top-level one. |

## create_my_login_link

A link that signs your own account in on a browser or another device. An API key may ask only if your account let it, from a browser session (POST /api/keys/{key_id}/sign-in); its link works once, for at most 15 minutes. Only works if this key was allowed to make sign-in links, on the API keys card. Answers with login_url (open it to sign in; it works once, for at most 15 minutes), expires_at and max_uses. Any older link the account could still use stops working; its signed-in devices stay signed in. Revoking or deleting this API key voids the link while it is unused, and signs out every device it signed in.

|  |  |
|---|---|
| Route | [`POST /api/account/login-link`](api.md#post-apiaccountlogin-link) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expires_in_minutes` | integer or null | no | 0 to 525600 | 0 never expires. Default auth.login_token_ttl_seconds for a link that works once, auth.multi_use_login_token_ttl_seconds otherwise. |
| `max_uses` | integer or null | no | 0 to 1000 | Sign-ins the link allows: 1-1000, or 0 for unlimited. Default 1. |

## create_sheet

Mint 1-300 brand-new codes for printing. Each one, once stuck on something and scanned, offers to set up that container.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/sheets`](api.md#post-apihouseholdshousehold_idsheets) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `count` | integer | yes | 1 to 300 | How many labels. |
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `layout` | string | no | at most 40 characters | Which sheet it is printed on, as the Print screen names it. |
| `names` | array of string | no |  | Names to print under the codes, in order (optional). |
| `palette` | string | no | at most 40 characters | original, neon, modern, or a single #rrggbb. |

## delete_household

Delete a household and everything in it, for good: its locations, containers, codes, print runs and items. Owners only. Photos stay in their uploaders' Files.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/delete`](api.md#post-apihouseholdshousehold_iddelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

## delete_item

Put an item in Trash; restore_item brings it back until it is purged.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/items/{item_id}/delete`](api.md#post-apihouseholdshousehold_iditemsitem_iddelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `item_id` | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |

## delete_item_bulk

Runs delete_item once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/households/items/delete`](api.md#post-apibulkhouseholdsitemsdelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `items[].item_id` | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## delete_label

Put a container and the items in it in Trash. restore_label brings both back. Its codes keep pointing at it until it is purged.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/labels/{label_id}/delete`](api.md#post-apihouseholdshousehold_idlabelslabel_iddelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |

## delete_label_bulk

Runs delete_label once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/households/labels/delete`](api.md#post-apibulkhouseholdslabelsdelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `items[].label_id` | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## delete_location

Put a location in Trash. Containers there keep it until it is purged, and show no location meanwhile; restore_location puts it back.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/locations/{location_id}/delete`](api.md#post-apihouseholdshousehold_idlocationslocation_iddelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `location_id` | integer | yes | at least 1 | The location's id, from list_locations. |

## delete_location_bulk

Runs delete_location once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/households/locations/delete`](api.md#post-apibulkhouseholdslocationsdelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `items[].location_id` | integer | yes | at least 1 | The location's id, from list_locations. |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## delete_sheet

Forget a print run. Its codes already on containers keep working; the rest stop, so stickers thrown away cannot be scanned later.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/sheets/{sheet_id}/delete`](api.md#post-apihouseholdshousehold_idsheetssheet_iddelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `sheet_id` | integer | yes | at least 1 | The print run's id, from list_sheets. |

## extract_photo

Ask Claude to list the items in an uploaded top-down photo, with your own API key if you set one, else the server's if an admin lets you. Answers at once with a draft whose status is working; ask get_extraction until it is ready, then accept_extraction with the items that are right.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/labels/{label_id}/extract`](api.md#post-apihouseholdshousehold_idlabelslabel_idextract) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `file_id` | string | yes | at least 1 characters | The photo, already uploaded (files_start_upload): JPEG, PNG, WebP or GIF, taken from above. |
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |

## files_delete

Move one of your files to the bin. It can be restored for ten minutes; after that it is gone. files_restore brings it back.

|  |  |
|---|---|
| Route | [`POST /api/files/{file_id}/delete`](api.md#post-apifilesfile_iddelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `file_id` | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |

## files_delete_bulk

Runs files_delete once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/files/delete`](api.md#post-apibulkfilesdelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].file_id` | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## files_download_url

The address of one of your files. GET it over HTTP with the same Authorization: Bearer header this MCP connection uses; the bytes do not come through MCP.

|  |  |
|---|---|
| Route | [`GET /api/files/{file_id}/download`](api.md#get-apifilesfile_iddownload) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `file_id` | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |

## files_finish_upload

Seal an upload once every piece has arrived; the file is then ready to download. Asked again, it answers the same. If pieces are missing it says how many arrived (received).

|  |  |
|---|---|
| Route | [`POST /api/files/{file_id}/finish`](api.md#post-apifilesfile_idfinish) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `file_id` | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |

## files_get

One file: its details, and for an unfinished upload how many pieces have arrived (received), which is where to resume.

|  |  |
|---|---|
| Route | [`GET /api/files/{file_id}`](api.md#get-apifilesfile_id) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `file_id` | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |

## files_list

Your uploaded files, a page at a time: id, name, type, size, state (uploading or ready) and when. next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Route | [`GET /api/files`](api.md#get-apifiles) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `dir` | string | no | one of `asc`, `desc` | Default asc. |
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | string | no |  | Keep files whose name contains this, in any case. |
| `sort` | string | no | one of `name`, `size`, `created_at` | Field to sort the whole list by. Rows without it go last. |

## files_rename

Give one of your files a new name. expect, when given, is the name you think it has now; a rename made meanwhile is refused with the current name.

|  |  |
|---|---|
| Route | [`POST /api/files/{file_id}/rename`](api.md#post-apifilesfile_idrename) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expect` | string or null | no |  | The name you think it has. Answers 409 if it has another. |
| `file_id` | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |
| `name` | string | yes |  | Cleaned like an upload's name. 1-255 bytes. |

## files_restore

Bring back a file you deleted, while it is still in the bin and your storage quota has room.

|  |  |
|---|---|
| Route | [`POST /api/files/{file_id}/restore`](api.md#post-apifilesfile_idrestore) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `file_id` | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |

## files_restore_bulk

Runs files_restore once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/files/restore`](api.md#post-apibulkfilesrestore) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].file_id` | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## files_start_upload

Reserve room for a file and get where to send it. The bytes do not go through MCP: POST each piece of chunk_bytes bytes (the last may be shorter), in order from index 0, as the raw body (Content-Type: application/octet-stream) to chunk_url with {index} filled in and the same Authorization: Bearer header this MCP connection uses. curl is a shell recipe that does all of it. Then call files_finish_upload.

|  |  |
|---|---|
| Route | [`POST /api/files`](api.md#post-apifiles) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `name` | string | yes |  |  |
| `size` | integer | yes | at least 0 |  |
| `type` | string | no |  | Shown in the list. Downloads are always application/octet-stream. |

## get_extraction

What Claude found in a photo so far: status working, then ready (with items) or failed (with error).

|  |  |
|---|---|
| Route | [`GET /api/households/{household_id}/extractions/{extraction_id}`](api.md#get-apihouseholdshousehold_idextractionsextraction_id) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `extraction_id` | integer | yes | at least 1 | The draft's id, from extract_photo. |
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

## get_household

A household you are in: your role, its members, color meanings and custom fields.

|  |  |
|---|---|
| Route | [`GET /api/households/{household_id}`](api.md#get-apihouseholdshousehold_id) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

## get_item

An item with everything known about it, where it is, and its history.

|  |  |
|---|---|
| Route | [`GET /api/households/{household_id}/items/{item_id}`](api.md#get-apihouseholdshousehold_iditemsitem_id) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `item_id` | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |

## get_label

A container with its items and the containers inside it.

|  |  |
|---|---|
| Route | [`GET /api/households/{household_id}/labels/{label_id}`](api.md#get-apihouseholdshousehold_idlabelslabel_id) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |

## import_csv

Add every row of an uploaded CSV as an item. A header row names the columns (Item, Quantity, Container, Location, Tags, Expires, Value ... as in an export, and the usual names other apps use). Containers and locations it does not know are made; nothing is changed or removed.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/import`](api.md#post-apihouseholdshousehold_idimport) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `file_id` | string | yes | at least 1 characters | A CSV file you uploaded (files_start_upload), with a header row. |
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

## label_my_session

Give one of your devices a name (at most 80 characters; empty clears it). expect, when given, is the name you think it has now.

|  |  |
|---|---|
| Route | [`POST /api/account/sessions/{session_id}/label`](api.md#post-apiaccountsessionssession_idlabel) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expect` | string or null | no |  | The name you think it has. Answers 409 if it has another. |
| `label` | string | yes | at most 80 characters | At most 80 characters. |
| `session_id` | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |

## leave_household

Stop being a member. The last owner cannot leave: make someone else an owner, or delete the household.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/leave`](api.md#post-apihouseholdshousehold_idleave) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

## list_households

The households you are a member of, each with your role, its members, color meanings and custom fields. Every other inventory tool takes one of these ids.

|  |  |
|---|---|
| Route | [`GET /api/households`](api.md#get-apihouseholds) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

## list_items

Items in a household, each with the container it is in and where that is. Search by words (a typo is forgiven) and narrow by container, location, tag, color, expiry, low stock or loans.

|  |  |
|---|---|
| Route | [`GET /api/households/{household_id}/items`](api.md#get-apihouseholdshousehold_iditems) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `color` | string | no |  | Only items in containers of this label color. |
| `dir` | string | no | one of `asc`, `desc` | Default asc. |
| `expiring_within` | integer or null | no | 0 to 3650 | Only items expiring within this many days (or already expired). |
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | integer | no |  | Only items in this container or the containers inside it. |
| `lent` | boolean | no |  | Only items lent out. |
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `location_id` | integer | no |  | Only items at this location or anywhere inside it. |
| `low_stock` | boolean | no |  | Only items at or below their low-stock level. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | string | no |  | Words to find: in names, descriptions, tags, brands, barcodes, serials, custom fields, and where it is. Forgives a typo. |
| `sort` | string | no | one of `name`, `path`, `quantity`, `expires`, `value`, `updated_at`, `created_at` | Field to sort by. |
| `tag` | string | no |  | Only items with this tag. |
| `trash` | boolean | no |  | List the ones in Trash instead. |

## list_labels

Containers in a household, each with its code, where it is, how many items it holds and what they are worth. Narrow by words, location, outer container, color or moving state.

|  |  |
|---|---|
| Route | [`GET /api/households/{household_id}/labels`](api.md#get-apihouseholdshousehold_idlabels) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `color` | string | no |  | Only this label color, #rrggbb. |
| `dir` | string | no | one of `asc`, `desc` | Default asc. |
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `location_id` | integer | no |  | Only containers at this location or inside it. |
| `move_status` | string | no | one of `packed`, `in_transit`, `unpacked` | Only containers in this moving state. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `parent_id` | integer | no |  | Only containers sitting directly inside this one. |
| `q` | string | no |  | Keep containers whose name, code, notes or place has these words. |
| `sort` | string | no | one of `name`, `path`, `item_count`, `value`, `updated_at`, `created_at` | Field to sort by. |
| `trash` | boolean | no |  | List the ones in Trash instead. |

## list_locations

The places containers sit, such as House › Garage › Shelf 2, each with its full path, in path order.

|  |  |
|---|---|
| Route | [`GET /api/households/{household_id}/locations`](api.md#get-apihouseholdshousehold_idlocations) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `trash` | boolean | no |  | List the ones in Trash instead. |

## list_my_sessions

Where your account is signed in, a page at a time: id, when it started and was last seen, the browser, and the name you gave it. Through an API key no device is current. next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Route | [`GET /api/account/sessions`](api.md#get-apiaccountsessions) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

## list_notifications

Reminders about items in your households, newest first: expiring or expired, run low, or overdue back from a loan.

|  |  |
|---|---|
| Route | [`GET /api/notifications`](api.md#get-apinotifications) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `unread` | boolean | no |  | Only the ones not read yet. |

## list_sheets

Sheets of codes made for printing, newest first, with how many are on containers.

|  |  |
|---|---|
| Route | [`GET /api/households/{household_id}/sheets`](api.md#get-apihouseholdshousehold_idsheets) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

## list_team

People with an active account, a page at a time: id and display name. Use the id wherever another tool asks for a user. next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Route | [`GET /api/team`](api.md#get-apiteam) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

## mark_notifications_read

Mark one reminder read, or all of them.

|  |  |
|---|---|
| Route | [`POST /api/notifications/read`](api.md#post-apinotificationsread) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `notification_id` | integer | no |  | The one to mark read; leave out to mark them all. |

## move_contents

Move every item in this container into another one in the same household, in one go.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/labels/{label_id}/move-contents`](api.md#post-apihouseholdshousehold_idlabelslabel_idmove-contents) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |
| `to_label_id` | integer | yes | at least 1 | The container everything goes into. |

## move_item

Put an item in another container of the same household.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/items/{item_id}/move`](api.md#post-apihouseholdshousehold_iditemsitem_idmove) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `item_id` | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |
| `label_id` | integer | yes | at least 1 | The container it goes into. |

## move_item_bulk

Runs move_item once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/households/items/move`](api.md#post-apibulkhouseholdsitemsmove) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `items[].item_id` | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |
| `items[].label_id` | integer | no | at least 1 | The container it goes into. |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |
| `shared.label_id` | integer | no | at least 1 | The container it goes into. |

## my_ai_settings

Whether you have your own AI key (never the key itself) and for which service, the model and effort you picked, whether you may use the server's key, and whose key your photos would go on now.

|  |  |
|---|---|
| Route | [`GET /api/account/ai`](api.md#get-apiaccountai) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|

## my_storage

How many bytes your files use, and your quota (0 is no limit).

|  |  |
|---|---|
| Route | [`GET /api/account/storage`](api.md#get-apiaccountstorage) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|

## release_code

Stop an added code opening this container. Its own code cannot be taken off.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/labels/{label_id}/codes/release`](api.md#post-apihouseholdshousehold_idlabelslabel_idcodesrelease) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `code` | string | yes | 1 to 2000 characters | The scanned value: a QR code's text or a barcode's digits. |
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |

## rename_account

Change your own display name (1-80 characters). expect, when given, is the name you think it has now; if someone changed it meanwhile the rename is refused with the current name, so nothing is overwritten unseen.

|  |  |
|---|---|
| Route | [`POST /api/account/rename`](api.md#post-apiaccountrename) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `display_name` | string | yes | 1 to 80 characters |  |
| `expect` | string or null | no |  |  |

## rename_household

Give a household a new name. Owners and editors.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/rename`](api.md#post-apihouseholdshousehold_idrename) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `name` | string | yes | 1 to 80 characters |  |

## reprint_labels

The codes, names, colors and handling flags of containers you already have, to print replacement labels.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/labels/reprint`](api.md#post-apihouseholdshousehold_idlabelsreprint) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_ids` | array of integer | yes |  | Containers whose own labels to print again. |

## restore_item

Take an item back out of Trash. Its container has to be out of Trash first.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/items/{item_id}/restore`](api.md#post-apihouseholdshousehold_iditemsitem_idrestore) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `item_id` | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |

## restore_item_bulk

Runs restore_item once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/households/items/restore`](api.md#post-apibulkhouseholdsitemsrestore) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `items[].item_id` | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## restore_label

Take a container back out of Trash, with the items that were deleted along with it.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/labels/{label_id}/restore`](api.md#post-apihouseholdshousehold_idlabelslabel_idrestore) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |

## restore_label_bulk

Runs restore_label once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/households/labels/restore`](api.md#post-apibulkhouseholdslabelsrestore) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `items[].label_id` | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## restore_location

Take a location back out of Trash.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/locations/{location_id}/restore`](api.md#post-apihouseholdshousehold_idlocationslocation_idrestore) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `location_id` | integer | yes | at least 1 | The location's id, from list_locations. |

## restore_location_bulk

Runs restore_location once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/households/locations/restore`](api.md#post-apibulkhouseholdslocationsrestore) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `items[].location_id` | integer | yes | at least 1 | The location's id, from list_locations. |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## revoke_my_other_sessions

End every session your account has: through an API key there is no current device to keep, so every browser signs in again. API keys keep working. Answers with how many were removed.

|  |  |
|---|---|
| Route | [`POST /api/account/sessions/revoke-others`](api.md#post-apiaccountsessionsrevoke-others) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|

## revoke_my_session

End one of your signed-in sessions; that browser has to sign in again.

|  |  |
|---|---|
| Route | [`POST /api/account/sessions/{session_id}/revoke`](api.md#post-apiaccountsessionssession_idrevoke) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `session_id` | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |

## revoke_my_session_bulk

Runs revoke_my_session once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/account/sessions/revoke`](api.md#post-apibulkaccountsessionsrevoke) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].session_id` | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## scan_code

What a scanned or typed code is: one of your containers (with where it is), one of your household's printed codes still waiting to be stuck on something (set it up with create_label and this code), or unknown -- which may be a store-bought sticker or barcode you can claim for a container with claim_code.

|  |  |
|---|---|
| Route | [`GET /api/scan`](api.md#get-apiscan) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `code` | string | no | 1 to 2000 characters | What the camera read: a printed code (XXXXX-XXXXX, with or without the dash), a link ending /l/<code>, any other QR text, or a barcode's digits. |

## search_inventory

Search every household you are in for items and containers, and say where each one is. The question "where is my drill?" is search_inventory with q=drill.

|  |  |
|---|---|
| Route | [`GET /api/search`](api.md#get-apisearch) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | string | no | 1 to 200 characters | What you are looking for, such as "drill" or "christmas lights". A typo is forgiven. |

## set_household_colors

Replace the household's color legend: which room or kind of thing each label color stands for. Owners and editors.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/colors`](api.md#post-apihouseholdshousehold_idcolors) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `colors` | map of string | yes |  | Label color (#rrggbb) to what it means here, such as "#d32f2f": "Kitchen". Replaces the whole list. |
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

## set_household_fields

Replace the custom fields items here can fill in (serial numbers, warranty dates, anything), and how many days before an expiry reminders go out. Owners and editors.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/fields`](api.md#post-apihouseholdshousehold_idfields) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `fields` | array of object | yes |  | The whole list. Each has a name, a type (text, number, date, select or bool), options for a select, and the id it already has, if any, so values survive a rename. |
| `fields[].id` | string | yes |  |  |
| `fields[].name` | string | yes |  |  |
| `fields[].options` | array of string | no |  |  |
| `fields[].type` | string | yes |  |  |
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `reminder_days` | integer or null | yes |  | Days before an expiry its reminder goes out, 0-365; null uses the app's default. |

## set_household_member

Give someone with an account a role here -- owner (everything), editor (change the inventory) or viewer (look only) -- or an empty role to take them out. Owners only. A household always keeps at least one owner.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/members`](api.md#post-apihouseholdshousehold_idmembers) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `role` | string | yes | one of `owner`, `editor`, `viewer`, `` | owner, editor or viewer; empty takes them out. |
| `user_id` | integer | yes | at least 1 | Their account id, from GET /api/team (list_team over MCP). |

## set_shortcuts

Replace your changed keyboard shortcuts: action name to binding, such as "search": "Mod+K" (Mod is Cmd on a Mac, Ctrl elsewhere). Send the whole set; an empty binding switches a key off, and {} puts every key back to its default. whoami shows the current set.

|  |  |
|---|---|
| Route | [`POST /api/account/shortcuts`](api.md#post-apiaccountshortcuts) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `shortcuts` | map of string | yes |  | Action name to binding, such as "search": "Mod+K". Mod is Cmd on a Mac and Ctrl elsewhere. Only changed shortcuts are listed; an empty binding switches that key off. Replaces the whole set; {} puts every key back to its default. |

## set_theme

Switch to another theme from your list (or the defaults), by its id.

|  |  |
|---|---|
| Route | [`POST /api/account/theme`](api.md#post-apiaccounttheme) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `theme_id` | string | yes |  | A theme's id. Empty for the first in the list. |

## set_themes

Replace your whole list of color themes (1 to 40) and pick the one you use. themes: null goes back to following the app's defaults. A theme is the Theme schema in the API reference (config/api.json): id, name, mode light|dark, accent, second, and optional colors as #rrggbb. whoami shows the current list.

|  |  |
|---|---|
| Route | [`POST /api/account/themes`](api.md#post-apiaccountthemes) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `builtins` | integer | no |  | Which built-in themes this list has seen: the highest since in Theme.builtIn (static/js/theme.js) when it was saved. Built-in themes added after that join the list the browser shows. Left out: 0, the first ones. |
| `theme_id` | string | no |  | The theme you use. |
| `themes` | array of object or null | no |  | Your whole list, 1 to 40 themes. null goes back to following the app's defaults. |
| `themes[].accent` | string | yes |  | The main color: what you can press. |
| `themes[].background` | string | no |  | The page. Left out: the mode's own. |
| `themes[].button` | object or null | no |  | How the theme's buttons look. A choice left out is the default (rounded, normal, raised, normal). |
| `themes[].button.case` | string | no | one of `normal`, `upper` |  |
| `themes[].button.depth` | string | no | one of `flat`, `raised`, `glossy` |  |
| `themes[].button.shape` | string | no | one of `square`, `soft`, `rounded`, `pill` |  |
| `themes[].button.weight` | string | no | one of `normal`, `medium`, `bold` |  |
| `themes[].dark` | object or null | no |  | A light theme's colors for dark mode, kept while it is light. Switching the theme to dark swaps them in. A light theme's light (or a dark theme's dark) is dropped. |
| `themes[].dark.accent` | string | yes |  | The main color: what you can press. |
| `themes[].dark.background` | string | no |  | The page. Left out: the mode's own. |
| `themes[].dark.overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `themes[].dark.second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `themes[].dark.surface` | string | no |  | Cards. Left out: the mode's own. |
| `themes[].id` | string | yes |  | Lower-case letters, digits, - and _, up to 32. |
| `themes[].light` | object or null | no |  | A dark theme's colors for light mode, kept while it is dark. Switching the theme to light swaps them in. |
| `themes[].light.accent` | string | yes |  | The main color: what you can press. |
| `themes[].light.background` | string | no |  | The page. Left out: the mode's own. |
| `themes[].light.overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `themes[].light.second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `themes[].light.surface` | string | no |  | Cards. Left out: the mode's own. |
| `themes[].mode` | string | yes | one of `light`, `dark` |  |
| `themes[].name` | string | yes |  | 1 to 60 characters. |
| `themes[].overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `themes[].second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `themes[].surface` | string | no |  | Cards. Left out: the mode's own. |

## update_item

Change anything about an item: name, photos, tags, quantity, low-stock level, expiry, custom fields, value, receipts, who it is lent to. Only what you send changes.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/items/{item_id}`](api.md#post-apihouseholdshousehold_iditemsitem_id) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `barcode` | string or null | no | at most 200 characters |  |
| `brand` | string or null | no | at most 80 characters |  |
| `clear_low_stock` | boolean | no |  | Stop low-stock reminders. |
| `clear_remind_days` | boolean | no |  | Use the household's reminder lead again. |
| `description` | string or null | no | at most 4000 characters |  |
| `expires` | string or null | no |  | Expiry date, YYYY-MM-DD, or empty. |
| `fields` | map of string or null | no |  | Custom field values by field id (see the household's fields). Replaces them all. |
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `item_id` | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |
| `lent_due` | string or null | no |  | When it should come back, YYYY-MM-DD, or empty. |
| `lent_to` | string or null | no | at most 80 characters | Who has it on loan; empty when it is back. |
| `low_stock` | number or null | no | at least 0 | Remind when the quantity drops to this or below. |
| `name` | string or null | no | 1 to 80 characters |  |
| `photos` | array of string or null | no |  | File ids of its photos, in order. |
| `purchase_date` | string or null | no |  | YYYY-MM-DD, or empty. |
| `quantity` | number or null | no | at least 0 | How many there are. A new item has 1. |
| `receipts` | array of string or null | no |  | File ids of receipts or warranty papers. |
| `remind_days` | integer or null | no | 0 to 365 | Days before the expiry to remind, instead of the household's lead. |
| `serial` | string or null | no | at most 120 characters |  |
| `tags` | array of string or null | no |  | Words to find it by. Stored in lower case. |
| `unit` | string or null | no | at most 20 characters | pcs, rolls, kg, ... |
| `value` | number or null | no | at least 0 | What one is worth, for the insurance list. |

## update_label

Change a container's name, color, type, location, outer container, photos, notes, moving state, handling flags or destination. Only what you send changes.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/labels/{label_id}`](api.md#post-apihouseholdshousehold_idlabelslabel_id) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `color` | string or null | no |  | Its label color, #rrggbb, or empty. |
| `container_type` | string or null | no | at most 40 characters | bin, box, drawer, shelf, bag, or anything else. |
| `destination` | string or null | no | at most 80 characters | For a move: the room it goes to. |
| `flags` | array of string or null | no |  | Handling notes printed on its label, such as fragile, this side up, heavy. |
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |
| `location_id` | integer or null | no |  | Where it sits; 0 for nowhere. |
| `move_status` | string or null | no | one of ``, `packed`, `in_transit`, `unpacked` | For a move: packed, in_transit, unpacked, or empty. |
| `name` | string or null | no | 1 to 80 characters |  |
| `notes` | string or null | no | at most 4000 characters |  |
| `parent_id` | integer or null | no |  | The container it sits inside; 0 for none. |
| `photos` | array of string or null | no |  | File ids of its photos, in order (see attach_photo). |

## update_location

Change a location's name, or move it inside another (0 for top-level). Only what you send changes.

|  |  |
|---|---|
| Route | [`POST /api/households/{household_id}/locations/{location_id}`](api.md#post-apihouseholdshousehold_idlocationslocation_id) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `location_id` | integer | yes | at least 1 | The location's id, from list_locations. |
| `name` | string or null | no | 1 to 80 characters |  |
| `parent_id` | integer or null | no |  | Move it inside another location; 0 makes it top-level. |

## whoami

The account this API key belongs to, the role it acts with (role -- never higher than the key's own scope), and the account's own role.

|  |  |
|---|---|
| Route | [`GET /api/me`](api.md#get-apime) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|

## HTTP only

Routes an API key may call that have no tool, and why.

| Route | Why |
|---|---|
| [`GET /api/api.json`](api.md#get-apiapijson) | The API reference itself; GET /api/mcp.json is the MCP one. |
| [`GET /api/api.md`](api.md#get-apiapimd) | The API reference itself, for people; GET /api/api.json is the same as JSON. |
| [`GET /api/households/{household_id}/export.csv`](api.md#get-apihouseholdshousehold_idexportcsv) | A CSV file is not JSON: GET this address with the same Authorization header. |
| [`GET /api/households/{household_id}/photos/{file_id}`](api.md#get-apihouseholdshousehold_idphotosfile_id) | Photo bytes are not JSON: GET this address with the same Authorization header. |
| [`GET /api/mcp.json`](api.md#get-apimcpjson) | The MCP reference itself; tools/list gives the same list. |
| [`GET /api/mcp.md`](api.md#get-apimcpmd) | The MCP reference itself, for people; tools/list gives the same list. |
| [`POST /api/files/{file_id}/chunks/{index}`](api.md#post-apifilesfile_idchunksindex) | File bytes: MCP messages are JSON, and clients cap their size. files_start_upload says where to send each piece, with the same key. |
| [`POST /api/logout`](api.md#post-apilogout) | Ends a browser session; a key has none to end. |
| [`POST /api/logs/events`](api.md#post-apilogsevents) | The browser's own activity reports (track.js), not an action. |
| [`POST /mcp`](api.md#post-mcp) | The MCP endpoint itself. |
