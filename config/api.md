# QRCodeOrganizer API

Every route this server answers, apart from the pages at /login/*, /join/* and /docs/*. Generated from the code by scripts/api_docs.sh: an app's own routes get their docs the same way, and a test fails until they do.

This file is generated: do not edit it. Change the docs next to the handlers and run `scripts/api_docs.sh`.

## Contents

- [Calling the API](#calling-the-api)
- [Errors](#errors)
- [Account](#account)
- [Sign-in](#sign-in)
- [API keys](#api-keys)
- [Files](#files)
- [Admin](#admin)
- [Invite links](#invite-links)
- [Owner](#owner)
- [Webhooks](#webhooks)
- [Monitoring](#monitoring)
- [Containers](#containers)
- [Households](#households)
- [Items](#items)
- [Labels](#labels)
- [Locations](#locations)
- [Photo extraction](#photo-extraction)
- [Reminders](#reminders)
- [Reports](#reports)
- [Scan](#scan)
- [Schemas](#schemas)

## Calling the API

| Way in | How |
|---|---|
| Browser session | A browser session. Its POSTs carry csrf_token in the body. The cookie's name is server.cookie_name, with the __Host- prefix while server.secure_cookies is on. |
| API key | An API key, as Authorization: Bearer <id>.<secret>. |

## Errors

An error reply is a JSON object whose `error` says what went wrong.

| Status | Meaning | Body |
|---|---|---|
| 400 | The request was refused as it stands. | [`Error`](#error) |
| 401 | Not signed in. | [`Error`](#error) |
| 403 | Not allowed, or a bad csrf_token. | [`Error`](#error) |
| 404 | No such thing, or this feature is switched off. | [`Error`](#error) |
| 409 | The record has moved on: an undo too late, or a stale edit. | [`Error`](#error) |
| 429 | Too many requests. | [`Error`](#error) |
| 503 | In maintenance mode. Retry-After says when to try again. | [`Maintenance`](#maintenance) |

## Account

| Method | Path | Summary | Who may call |
|---|---|---|---|
| POST | [`/api/account/login-link`](#post-apiaccountlogin-link) | A login link for yourself | browser session or API key, anyone signed in |
| POST | [`/api/account/login-link/cancel`](#post-apiaccountlogin-linkcancel) | Stop your own login link working | browser session or API key, anyone signed in |
| POST | [`/api/account/password`](#post-apiaccountpassword) | Set or clear your password | browser session only, anyone signed in |
| POST | [`/api/account/pin`](#post-apiaccountpin) | Set or clear your PIN | browser session only, anyone signed in |
| POST | [`/api/account/rename`](#post-apiaccountrename) | Rename your account | browser session or API key, anyone signed in |
| GET | [`/api/account/sessions`](#get-apiaccountsessions) | Your signed-in devices | browser session or API key, anyone signed in |
| POST | [`/api/account/sessions/revoke-others`](#post-apiaccountsessionsrevoke-others) | Sign out every other device | browser session or API key, anyone signed in |
| POST | [`/api/account/sessions/{session_id}/label`](#post-apiaccountsessionssession_idlabel) | Name one of your devices | browser session or API key, anyone signed in |
| POST | [`/api/account/sessions/{session_id}/revoke`](#post-apiaccountsessionssession_idrevoke) | Sign out one of your devices | browser session or API key, anyone signed in |
| POST | [`/api/account/shortcuts`](#post-apiaccountshortcuts) | Change your keyboard shortcuts | browser session or API key, anyone signed in |
| POST | [`/api/account/theme`](#post-apiaccounttheme) | Pick your theme | browser session or API key, anyone signed in |
| POST | [`/api/account/themes`](#post-apiaccountthemes) | Replace your own themes | browser session or API key, anyone signed in |
| POST | [`/api/account/two-factor/confirm`](#post-apiaccounttwo-factorconfirm) | Turn two-factor on | browser session only, anyone signed in |
| POST | [`/api/account/two-factor/disable`](#post-apiaccounttwo-factordisable) | Turn two-factor off | browser session only, anyone signed in |
| POST | [`/api/account/two-factor/start`](#post-apiaccounttwo-factorstart) | Begin two-factor setup | browser session only, anyone signed in |
| POST | [`/api/bulk/account/sessions/revoke`](#post-apibulkaccountsessionsrevoke) | Sign out one of your devices, for many at once | browser session or API key, anyone signed in |
| GET | [`/api/language`](#get-apilanguage) | The UI's text | anyone |
| POST | [`/api/logout`](#post-apilogout) | Sign out | browser session or API key, anyone signed in |
| POST | [`/api/logs/events`](#post-apilogsevents) | Record what happened on the page | browser session or API key, anyone signed in |
| GET | [`/api/me`](#get-apime) | Who am I | browser session or API key, anyone signed in |
| GET | [`/api/team`](#get-apiteam) | List team | browser session or API key, anyone signed in |

### POST /api/account/login-link

A login link for yourself

A link that signs your own account in on a browser or another device. An API key may ask only if your account let it, from a browser session (POST /api/keys/{key_id}/sign-in); its link works once, for at most 15 minutes.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`create_my_login_link`](mcp.md#create_my_login_link) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expires_in_minutes` | integer or null | no | 0 to 525600 | 0 never expires. Default auth.login_token_ttl_seconds for a link that works once, auth.multi_use_login_token_ttl_seconds otherwise. |
| `max_uses` | integer or null | no | 0 to 1000 | Sign-ins the link allows: 1-1000, or 0 for unlimited. Default 1. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expires_at` | string (date-time) or null | yes |  | null when it never expires. |
| `expires_in_seconds` | integer | yes |  | 0 when it never expires. |
| `login_path` | string | yes |  | Path of the login link. A credential: treat it like a password. |
| `max_uses` | integer | yes |  | 0 when unlimited. |
| `unlimited` | boolean | yes |  |  |

### POST /api/account/login-link/cancel

Stop your own login link working

Void your account's live login link without making a new one. Devices it already signed in stay signed in. Answers with removed: 0 or 1.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`cancel_my_login_link`](mcp.md#cancel_my_login_link) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `removed` | integer | yes |  | Links removed: 0 or 1. |

### POST /api/account/password

Set or clear your password

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `code` | string | no |  |  |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `password_digest` | string | yes |  |  |
| `username` | string | yes |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `has_password` | boolean | yes |  |  |
| `ok` | boolean | yes |  |  |

### POST /api/account/pin

Set or clear your PIN

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `pin` | string | yes |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `has_pin` | boolean | yes |  |  |
| `ok` | boolean | yes |  |  |

### POST /api/account/rename

Rename your account

Change your own display name (1-80 characters). expect, when given, is the name you think it has now; if someone changed it meanwhile the rename is refused with the current name, so nothing is overwritten unseen.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`rename_account`](mcp.md#rename_account) |
| Errors | 400, 401, 403, 409, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `display_name` | string | yes | 1 to 80 characters |  |
| `expect` | string or null | no |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `display_name` | string | yes |  |  |
| `ok` | boolean | yes |  |  |

### GET /api/account/sessions

Your signed-in devices

Where your account is signed in, a page at a time: id, when it started and was last seen, the browser, and the name you gave it. Through an API key no device is current. next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`list_my_sessions`](mcp.md#list_my_sessions) |
| Errors | 400, 401, 403, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `sessions` | array of [`Session`](#session) | yes |  |  |
| `total` | integer | yes |  |  |

### POST /api/account/sessions/revoke-others

Sign out every other device

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`revoke_my_other_sessions`](mcp.md#revoke_my_other_sessions) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `removed` | integer | yes |  |  |

### POST /api/account/sessions/{session_id}/label

Name one of your devices

Give one of your devices a name (at most 80 characters; empty clears it). expect, when given, is the name you think it has now.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`label_my_session`](mcp.md#label_my_session) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `session_id` | path | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expect` | string or null | no |  | The name you think it has. Answers 409 if it has another. |
| `label` | string | yes | at most 80 characters | At most 80 characters. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `label` | string | yes |  |  |
| `ok` | boolean | yes |  |  |

### POST /api/account/sessions/{session_id}/revoke

Sign out one of your devices

End one of your signed-in sessions; that browser has to sign in again.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`revoke_my_session`](mcp.md#revoke_my_session) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `session_id` | path | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/account/shortcuts

Change your keyboard shortcuts

Replace your changed keyboard shortcuts: action name to binding, such as "search": "Mod+K" (Mod is Cmd on a Mac, Ctrl elsewhere). Send the whole set; an empty binding switches a key off, and {} puts every key back to its default.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`set_shortcuts`](mcp.md#set_shortcuts) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `shortcuts` | map of string | yes |  | Action name to binding, such as "search": "Mod+K". Mod is Cmd on a Mac and Ctrl elsewhere. Only changed shortcuts are listed; an empty binding switches that key off. Replaces the whole set; {} puts every key back to its default. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `shortcuts` | map of string | yes |  |  |

### POST /api/account/theme

Pick your theme

Switch to another theme from your list (or the defaults), by its id.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`set_theme`](mcp.md#set_theme) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `theme_id` | string | yes |  | A theme's id. Empty for the first in the list. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `theme_id` | string | yes |  |  |

### POST /api/account/themes

Replace your own themes

Replace your whole list of color themes (1 to 40) and pick the one you use. themes: null goes back to following the app's defaults. A theme is the Theme schema in the API reference (config/api.json): id, name, mode light|dark, accent, second, and optional colors as #rrggbb.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`set_themes`](mcp.md#set_themes) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `builtins` | integer | no |  | Which built-in themes this list has seen: the highest since in Theme.builtIn (static/js/theme.js) when it was saved. Built-in themes added after that join the list the browser shows. Left out: 0, the first ones. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `theme_id` | string | no |  | The theme you use. |
| `themes` | array of [`Theme`](#theme) or null | no |  | Your whole list, 1 to 40 themes. null goes back to following the app's defaults. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `theme_id` | string | yes |  |  |
| `themes` | array of [`Theme`](#theme) or null | yes |  | What you now see; null while the built-in themes apply. |
| `themes_builtins` | integer | yes |  | Which built-in themes that list has seen (see builtins on POST /api/account/themes). |
| `themes_custom` | boolean | yes |  |  |

### POST /api/account/two-factor/confirm

Turn two-factor on

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `code` | string | yes |  |  |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `recovery_codes` | array of string | yes |  |  |

### POST /api/account/two-factor/disable

Turn two-factor off

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `code` | string | yes |  |  |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/account/two-factor/start

Begin two-factor setup

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `secret` | string | yes |  |  |

### POST /api/bulk/account/sessions/revoke

Sign out one of your devices, for many at once

Runs POST /api/account/sessions/{session_id}/revoke once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`revoke_my_session_bulk`](mcp.md#revoke_my_session_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].session_id` | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### GET /api/language

The UI's text

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 429, 503 |

**Reply 200**: Success

An empty object.

### POST /api/logout

Sign out

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | None, HTTP only: Ends a browser session; a key has none to end. |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/logs/events

Record what happened on the page

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | None, HTTP only: The browser's own activity reports (track.js), not an action. |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `events` | array of object | yes |  |  |
| `tab` | string | yes |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `accepted` | integer | yes |  |  |

### GET /api/me

Who am I

The account this API key belongs to, the role it acts with (role -- never higher than the key's own scope), and the account's own role.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`whoami`](mcp.md#whoami) |
| Errors | 401, 429, 503 |

**Reply 200**: Success

[`Me`](#me).

### GET /api/team

List team

People with an active account, a page at a time: id and display name.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`list_team`](mcp.md#list_team) |
| Errors | 400, 401, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `members` | array of [`Member`](#member) | yes |  |  |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |

## Sign-in

| Method | Path | Summary | Who may call |
|---|---|---|---|
| POST | [`/api/login`](#post-apilogin) | Redeem a login link | anyone |
| GET | [`/api/login/options`](#get-apiloginoptions) | What sign-in offers | anyone |
| POST | [`/api/login/password`](#post-apiloginpassword) | Sign in with a password | anyone |
| POST | [`/api/login/two-factor`](#post-apilogintwo-factor) | Finish a sign-in with a code | anyone |

### POST /api/login

Redeem a login link

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 400, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `credential` | string | yes |  |  |

**Reply 200**: Success

[`SignIn`](#signin).

### GET /api/login/options

What sign-in offers

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `passwords` | boolean | yes |  |  |
| `salt` | string | no |  |  |
| `two_factor` | boolean | yes |  |  |

### POST /api/login/password

Sign in with a password

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 400, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `password_digest` | string | yes |  |  |
| `username` | string | yes |  |  |

**Reply 200**: Success

[`SignIn`](#signin).

### POST /api/login/two-factor

Finish a sign-in with a code

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 400, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `challenge` | string | yes |  |  |
| `code` | string | yes |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

## API keys

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/admin/keys`](#get-apiadminkeys) | Every key in the app | browser session only, admins |
| POST | [`/api/admin/keys`](#post-apiadminkeys) | Create a key that acts as someone else | browser session only, admins |
| POST | [`/api/admin/keys/{key_id}/delete`](#post-apiadminkeyskey_iddelete) | Delete anyone's key | browser session only, admins |
| POST | [`/api/admin/keys/{key_id}/expiry`](#post-apiadminkeyskey_idexpiry) | Change when anyone's key expires | browser session only, admins |
| POST | [`/api/admin/keys/{key_id}/rename`](#post-apiadminkeyskey_idrename) | Rename anyone's key | browser session only, admins |
| POST | [`/api/admin/keys/{key_id}/restore`](#post-apiadminkeyskey_idrestore) | Undo deleting anyone's key | browser session only, admins |
| POST | [`/api/admin/keys/{key_id}/revoke`](#post-apiadminkeyskey_idrevoke) | Revoke anyone's key | browser session only, admins |
| GET | [`/api/keys`](#get-apikeys) | Your API keys | browser session only, anyone signed in |
| POST | [`/api/keys`](#post-apikeys) | Create an API key | browser session only, anyone signed in |
| POST | [`/api/keys/{key_id}/delete`](#post-apikeyskey_iddelete) | Delete one of your keys | browser session only, anyone signed in |
| POST | [`/api/keys/{key_id}/expiry`](#post-apikeyskey_idexpiry) | Change when one of your keys expires | browser session only, anyone signed in |
| POST | [`/api/keys/{key_id}/rename`](#post-apikeyskey_idrename) | Rename one of your keys | browser session only, anyone signed in |
| POST | [`/api/keys/{key_id}/restore`](#post-apikeyskey_idrestore) | Undo deleting one of your keys | browser session only, anyone signed in |
| POST | [`/api/keys/{key_id}/revoke`](#post-apikeyskey_idrevoke) | Revoke one of your keys | browser session only, anyone signed in |
| POST | [`/api/keys/{key_id}/sign-in`](#post-apikeyskey_idsign-in) | Let one of your keys make sign-in links | browser session only, anyone signed in |

### GET /api/admin/keys

Every key in the app

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no |  | Keep keys whose name or owner contains this, in any case. |
| `live` | query | string | no | one of `true` | Leave out revoked and expired keys. |
| `sort` | query | string | no | one of `display_name`, `name`, `role`, `last_used_at`, `expires_at`, `created_at` | Field to sort the whole list by. Rows without it go last. |
| `dir` | query | string | no | one of `asc`, `desc` | Default asc. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `keys` | array of [`APIKey`](#apikey) | yes |  |  |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |

### POST /api/admin/keys

Create a key that acts as someone else

Any admin may do this for an ordinary user, and only the owner for another admin. Nobody may do it for the owner. The key records who made it and shows that to the person it acts as.

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expires_in_days` | integer or null | no | 0 to 3650 | Absent or null: the configured default. 0: never expires. |
| `name` | string | yes | 1 to 80 characters |  |
| `role` | string | no | one of `admin`, `user`; default `user` |  |
| `user_id` | integer | yes | at least 1 |  |

**Reply 201**: Success

[`NewAPIKey`](#newapikey).

### POST /api/admin/keys/{key_id}/delete

Delete anyone's key

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/keys/{key_id}/expiry

Change when anyone's key expires

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expires_in_days` | integer | yes | 0 to 3650 | Counted from now. 0 means it never expires. |

**Reply 200**: Success

[`APIKey`](#apikey).

### POST /api/admin/keys/{key_id}/rename

Rename anyone's key

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expect` | string or null | no |  | The name you think it has. Answers 409 if it has another. |
| `name` | string | yes | 1 to 80 characters | 1-80 characters. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `name` | string | yes |  |  |
| `ok` | boolean | yes |  |  |

### POST /api/admin/keys/{key_id}/restore

Undo deleting anyone's key

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`APIKey`](#apikey).

### POST /api/admin/keys/{key_id}/revoke

Revoke anyone's key

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### GET /api/keys

Your API keys

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no |  | Keep keys whose name contains this, in any case. |
| `live` | query | string | no | one of `true` | Leave out revoked and expired keys. |
| `sort` | query | string | no | one of `name`, `role`, `last_used_at`, `expires_at`, `created_at` | Field to sort the whole list by. Rows without it go last. |
| `dir` | query | string | no | one of `asc`, `desc` | Default asc. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `keys` | array of [`APIKey`](#apikey) | yes |  |  |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |

### POST /api/keys

Create an API key

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expires_in_days` | integer or null | no | 0 to 3650 | Absent or null: the configured default. 0: never expires. |
| `name` | string | yes | 1 to 80 characters |  |
| `role` | string | no | one of `admin`, `user`; default `user` |  |

**Reply 201**: Success

[`NewAPIKey`](#newapikey).

### POST /api/keys/{key_id}/delete

Delete one of your keys

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/keys/{key_id}/expiry

Change when one of your keys expires

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expires_in_days` | integer | yes | 0 to 3650 | Counted from now. 0 means it never expires. |

**Reply 200**: Success

[`APIKey`](#apikey).

### POST /api/keys/{key_id}/rename

Rename one of your keys

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expect` | string or null | no |  | The name you think it has. Answers 409 if it has another. |
| `name` | string | yes | 1 to 80 characters | 1-80 characters. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `name` | string | yes |  |  |
| `ok` | boolean | yes |  |  |

### POST /api/keys/{key_id}/restore

Undo deleting one of your keys

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`APIKey`](#apikey).

### POST /api/keys/{key_id}/revoke

Revoke one of your keys

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/keys/{key_id}/sign-in

Let one of your keys make sign-in links

Off by default. When on, the key may ask POST /api/account/login-link for a link that signs your account in on a browser, which then can do everything you can, including what keys may not. The link works once, for at most 15 minutes.

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `may_sign_in` | boolean | yes |  | true lets the key make a login link for your account: one that works once, for at most 15 minutes. |

**Reply 200**: Success

[`APIKey`](#apikey).

## Files

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/account/storage`](#get-apiaccountstorage) | Your storage | browser session or API key, anyone signed in |
| GET | [`/api/admin/files`](#get-apiadminfiles) | List everyone's files | browser session or API key, admins |
| POST | [`/api/admin/files/{file_id}/delete`](#post-apiadminfilesfile_iddelete) | Delete anyone's file | browser session only, admins |
| GET | [`/api/admin/files/{file_id}/download`](#get-apiadminfilesfile_iddownload) | Download anyone's file | browser session only, admins |
| POST | [`/api/admin/files/{file_id}/rename`](#post-apiadminfilesfile_idrename) | Rename any file | browser session only, admins |
| POST | [`/api/admin/files/{file_id}/restore`](#post-apiadminfilesfile_idrestore) | Put back a deleted file | browser session only, admins |
| GET | [`/api/admin/storage`](#get-apiadminstorage) | Storage for everyone | browser session or API key, admins |
| POST | [`/api/bulk/files/delete`](#post-apibulkfilesdelete) | Delete a file, for many at once | browser session or API key, anyone signed in |
| POST | [`/api/bulk/files/restore`](#post-apibulkfilesrestore) | Restore a deleted file, for many at once | browser session or API key, anyone signed in |
| GET | [`/api/files`](#get-apifiles) | List your files | browser session or API key, anyone signed in |
| POST | [`/api/files`](#post-apifiles) | Start an upload | browser session or API key, anyone signed in |
| GET | [`/api/files/{file_id}`](#get-apifilesfile_id) | One of your files | browser session or API key, anyone signed in |
| POST | [`/api/files/{file_id}/chunks/{index}`](#post-apifilesfile_idchunksindex) | Send the next piece of an upload | browser session or API key, anyone signed in |
| POST | [`/api/files/{file_id}/delete`](#post-apifilesfile_iddelete) | Delete a file | browser session or API key, anyone signed in |
| GET | [`/api/files/{file_id}/download`](#get-apifilesfile_iddownload) | Download one of your files | browser session or API key, anyone signed in |
| POST | [`/api/files/{file_id}/finish`](#post-apifilesfile_idfinish) | Finish an upload | browser session or API key, anyone signed in |
| POST | [`/api/files/{file_id}/rename`](#post-apifilesfile_idrename) | Rename a file | browser session or API key, anyone signed in |
| POST | [`/api/files/{file_id}/restore`](#post-apifilesfile_idrestore) | Restore a deleted file | browser session or API key, anyone signed in |

### GET /api/account/storage

Your storage

How many bytes your files use, and your quota (0 is no limit).

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`my_storage`](mcp.md#my_storage) |
| Errors | 401, 404, 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `quota_bytes` | integer | yes |  |  |
| `used_bytes` | integer | yes |  |  |

### GET /api/admin/files

List everyone's files

Every account's files, a page at a time, with whose each one is (user_id, user_name). next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_list_files`](mcp.md#admin_list_files) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no |  | Keep files whose name or owner contains this, in any case. |
| `sort` | query | string | no | one of `name`, `user_name`, `size`, `created_at` | Field to sort the whole list by. Rows without it go last. |
| `dir` | query | string | no | one of `asc`, `desc` | Default asc. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `files` | array of object | yes |  |  |
| `files[].chunks` | integer | yes |  |  |
| `files[].created_at` | string (date-time) | yes |  |  |
| `files[].id` | string | yes |  |  |
| `files[].name` | string | yes |  |  |
| `files[].received` | integer | yes |  |  |
| `files[].size` | integer | yes |  |  |
| `files[].state` | string | yes | one of `uploading`, `ready` |  |
| `files[].type` | string | yes |  |  |
| `files[].user_id` | integer | yes |  |  |
| `files[].user_name` | string | yes |  |  |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |

### POST /api/admin/files/{file_id}/delete

Delete anyone's file

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `file_id` | path | string | yes |  | The file's id, from GET /api/admin/files (admin_list_files over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### GET /api/admin/files/{file_id}/download

Download anyone's file

Written to the activity log when the file is someone else's.

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `file_id` | path | string | yes |  | The file's id, from GET /api/admin/files (admin_list_files over MCP). |

**Reply 200**: The file, always as an attachment

Raw bytes, as `application/octet-stream`.

### POST /api/admin/files/{file_id}/rename

Rename any file

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `file_id` | path | string | yes |  | The file's id, from GET /api/admin/files (admin_list_files over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expect` | string or null | no |  | The name you think it has. Answers 409 if it has another. |
| `name` | string | yes |  | Cleaned like an upload's name. 1-255 bytes. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `name` | string | yes |  | As stored. |
| `ok` | boolean | yes |  |  |

### POST /api/admin/files/{file_id}/restore

Put back a deleted file

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `file_id` | path | string | yes |  | The file's id, from GET /api/admin/files (admin_list_files over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`File`](#file).

### GET /api/admin/storage

Storage for everyone

How many bytes everyone's files use together, and the disk's free and total bytes where the system says.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_storage`](mcp.md#admin_storage) |
| Errors | 401, 403, 404, 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `disk_free_bytes` | integer | no |  | Left out where the system will not say. |
| `disk_total_bytes` | integer | no |  | Left out where the system will not say. |
| `used_bytes` | integer | yes |  |  |

### POST /api/bulk/files/delete

Delete a file, for many at once

Runs POST /api/files/{file_id}/delete once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`files_delete_bulk`](mcp.md#files_delete_bulk) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].file_id` | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/files/restore

Restore a deleted file, for many at once

Runs POST /api/files/{file_id}/restore once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`files_restore_bulk`](mcp.md#files_restore_bulk) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].file_id` | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### GET /api/files

List your files

Your uploaded files, a page at a time: id, name, type, size, state (uploading or ready) and when. next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`files_list`](mcp.md#files_list) |
| Errors | 400, 401, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no |  | Keep files whose name contains this, in any case. |
| `sort` | query | string | no | one of `name`, `size`, `created_at` | Field to sort the whole list by. Rows without it go last. |
| `dir` | query | string | no | one of `asc`, `desc` | Default asc. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `files` | array of [`File`](#file) | yes |  |  |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |

### POST /api/files

Start an upload

Then send each piece in order and finish. A file over files.max_file_bytes, or one that would pass files.quota_bytes_per_user, is a 409.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`files_start_upload`](mcp.md#files_start_upload) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `name` | string | yes |  |  |
| `size` | integer | yes | at least 0 |  |
| `type` | string | no |  | Shown in the list. Downloads are always application/octet-stream. |

**Reply 201**: Created

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `chunk_bytes` | integer | yes |  | Every piece but the last is exactly this long. |
| `chunks` | integer | yes |  |  |
| `created_at` | string (date-time) | yes |  |  |
| `id` | string | yes |  |  |
| `name` | string | yes |  |  |
| `received` | integer | yes |  |  |
| `size` | integer | yes |  |  |
| `state` | string | yes | one of `uploading`, `ready` |  |
| `type` | string | yes |  |  |

### GET /api/files/{file_id}

One of your files

One file: its details, and for an unfinished upload how many pieces have arrived (received), which is where to resume.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`files_get`](mcp.md#files_get) |
| Errors | 400, 401, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `file_id` | path | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `chunk_bytes` | integer | yes |  | Every piece but the last is exactly this long. |
| `chunks` | integer | yes |  |  |
| `created_at` | string (date-time) | yes |  |  |
| `id` | string | yes |  |  |
| `name` | string | yes |  |  |
| `received` | integer | yes |  |  |
| `size` | integer | yes |  |  |
| `state` | string | yes | one of `uploading`, `ready` |  |
| `type` | string | yes |  |  |

### POST /api/files/{file_id}/chunks/{index}

Send the next piece of an upload

The body is the raw bytes. Pieces go in order, from 0. A browser session sends its csrf token in the X-CSRF-Token header. A 409 carries received, the piece to send next. Counts against server.rate_limit_upload_max.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | None, HTTP only: File bytes: MCP messages are JSON, and clients cap their size. files_start_upload says where to send each piece, with the same key. |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `file_id` | path | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |
| `index` | path | integer | yes | at least 0 |  |
| `X-CSRF-Token` | header | string | no |  | From GET /api/me. Browser sessions only. |

**Request body**

Raw bytes, as `application/octet-stream`.

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `received` | integer | yes |  |  |

### POST /api/files/{file_id}/delete

Delete a file

Move one of your files to the bin. It can be restored for ten minutes; after that it is gone.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`files_delete`](mcp.md#files_delete) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `file_id` | path | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### GET /api/files/{file_id}/download

Download one of your files

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`files_download_url`](mcp.md#files_download_url) |
| Errors | 400, 401, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `file_id` | path | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |

**Reply 200**: The file, always as an attachment

Raw bytes, as `application/octet-stream`.

### POST /api/files/{file_id}/finish

Finish an upload

Seal an upload once every piece has arrived; the file is then ready to download. Asked again, it answers the same. If pieces are missing it says how many arrived (received).

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`files_finish_upload`](mcp.md#files_finish_upload) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `file_id` | path | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`File`](#file).

### POST /api/files/{file_id}/rename

Rename a file

Give one of your files a new name. expect, when given, is the name you think it has now; a rename made meanwhile is refused with the current name.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`files_rename`](mcp.md#files_rename) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `file_id` | path | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expect` | string or null | no |  | The name you think it has. Answers 409 if it has another. |
| `name` | string | yes |  | Cleaned like an upload's name. 1-255 bytes. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `name` | string | yes |  | As stored. |
| `ok` | boolean | yes |  |  |

### POST /api/files/{file_id}/restore

Restore a deleted file

Bring back a file you deleted, while it is still in the bin and your storage quota has room.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`files_restore`](mcp.md#files_restore) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `file_id` | path | string | yes |  | The file's id, from GET /api/files (files_list over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`File`](#file).

## Admin

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/admin/theme-defaults`](#get-apiadmintheme-defaults) | Default themes | browser session or API key, admins |
| POST | [`/api/admin/theme-defaults`](#post-apiadmintheme-defaults) | Replace the default themes | browser session or API key, admins |
| GET | [`/api/admin/users`](#get-apiadminusers) | List users | browser session or API key, admins |
| POST | [`/api/admin/users`](#post-apiadminusers) | Create an account | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/cancel-login-link`](#post-apiadminusersuser_idcancel-login-link) | Stop an account's login link working | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/delete`](#post-apiadminusersuser_iddelete) | Delete an account | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/disabled`](#post-apiadminusersuser_iddisabled) | Disable or enable a user | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/reissue-login`](#post-apiadminusersuser_idreissue-login) | A new login link for an account | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/rename`](#post-apiadminusersuser_idrename) | Rename someone | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/reset-two-factor`](#post-apiadminusersuser_idreset-two-factor) | Turn off an account's two-factor | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/revoke-access`](#post-apiadminusersuser_idrevoke-access) | Revoke a user's access | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/role`](#post-apiadminusersuser_idrole) | Change an account's role | browser session or API key, admins |
| GET | [`/api/admin/users/{user_id}/sessions`](#get-apiadminusersuser_idsessions) | An account's signed-in devices | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/sessions/{session_id}/revoke`](#post-apiadminusersuser_idsessionssession_idrevoke) | Sign out one of an account's devices | browser session or API key, admins |
| POST | [`/api/bulk/admin/users/cancel-login-link`](#post-apibulkadminuserscancel-login-link) | Stop an account's login link working, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/users/delete`](#post-apibulkadminusersdelete) | Delete an account, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/users/disabled`](#post-apibulkadminusersdisabled) | Disable or enable a user, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/users/reset-two-factor`](#post-apibulkadminusersreset-two-factor) | Turn off an account's two-factor, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/users/revoke-access`](#post-apibulkadminusersrevoke-access) | Revoke a user's access, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/users/role`](#post-apibulkadminusersrole) | Change an account's role, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/users/sessions/revoke`](#post-apibulkadminuserssessionsrevoke) | Sign out one of an account's devices, for many at once | browser session or API key, admins |

### GET /api/admin/theme-defaults

Default themes

The theme list everyone without their own list sees. themes is null while the built-in ones apply.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_get_theme_defaults`](mcp.md#admin_get_theme_defaults) |
| Errors | 401, 403, 404, 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `builtins` | integer | yes |  | Which built-in themes that list has seen (see builtins on POST /api/account/themes). |
| `themes` | array of [`Theme`](#theme) or null | yes |  | null while the built-in themes apply. |

### POST /api/admin/theme-defaults

Replace the default themes

Replace the theme list everyone without their own list sees, 1 to 40 themes in the Theme schema of the API reference. null puts the built-in themes back.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_theme_defaults`](mcp.md#admin_set_theme_defaults) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `builtins` | integer | no |  | Which built-in themes this list has seen: the highest since in Theme.builtIn (static/js/theme.js) when it was saved. Built-in themes added after that join the list the browser shows. Left out: 0, the first ones. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `themes` | array of [`Theme`](#theme) or null | no |  | 1 to 40 themes. null goes back to the built-in ones. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `builtins` | integer | yes |  | Which built-in themes that list has seen (see builtins on POST /api/account/themes). |
| `ok` | boolean | yes |  |  |
| `themes` | array of [`Theme`](#theme) or null | yes |  |  |

### GET /api/admin/users

List users

Accounts, a page at a time, including disabled ones, with each one's role and which is the owner. next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_list_users`](mcp.md#admin_list_users) |
| Errors | 400, 401, 403, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no |  | Keep accounts whose name or username contains this, in any case. |
| `sort` | query | string | no | one of `display_name`, `role`, `disabled`, `created_at` | Field to sort the whole list by. Rows without it go last. |
| `dir` | query | string | no | one of `asc`, `desc` | Default asc. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |
| `users` | array of [`User`](#user) | yes |  |  |

### POST /api/admin/users

Create an account

Make a new account and its first login link, to hand to the person. role is admin or user. Answers with user_id and the link.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_create_user`](mcp.md#admin_create_user) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `display_name` | string | yes | 1 to 80 characters |  |
| `expires_in_minutes` | integer or null | no | 0 to 525600 | 0 never expires. Default auth.login_token_ttl_seconds for a link that works once, auth.multi_use_login_token_ttl_seconds otherwise. |
| `max_uses` | integer or null | no | 0 to 1000 | Sign-ins the link allows: 1-1000, or 0 for unlimited. Default 1. |
| `role` | string | yes | one of `admin`, `user` |  |

**Reply 201**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expires_at` | string (date-time) or null | yes |  | null when it never expires. |
| `expires_in_seconds` | integer | yes |  | 0 when it never expires. |
| `login_path` | string | yes |  | Path of the login link. A credential: treat it like a password. |
| `max_uses` | integer | yes |  | 0 when unlimited. |
| `unlimited` | boolean | yes |  |  |
| `user_id` | integer | yes |  |  |

### POST /api/admin/users/{user_id}/cancel-login-link

Stop an account's login link working

Void an account's live login link without making a new one: a link that went somewhere it should not have.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_cancel_login_link`](mcp.md#admin_cancel_login_link) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `removed` | integer | yes |  | Links removed: 0 or 1. |

### POST /api/admin/users/{user_id}/delete

Delete an account

Remove an account for good, with its sessions, login link and API keys.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_delete_user`](mcp.md#admin_delete_user) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `removed` | [`Revoked`](#revoked) | yes |  |  |

### POST /api/admin/users/{user_id}/disabled

Disable or enable a user

Switch an account off (disabled: true) or back on (false). Disabling signs it out everywhere and voids its live login link; its API keys stop working while it is off. Reversible. The owner and your own account cannot be disabled.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_user_disabled`](mcp.md#admin_set_user_disabled) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `disabled` | boolean | yes |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/users/{user_id}/reissue-login

A new login link for an account

For someone who lost their device or their link: a fresh login link for their account.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_reissue_login`](mcp.md#admin_reissue_login) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expires_in_minutes` | integer or null | no | 0 to 525600 | 0 never expires. Default auth.login_token_ttl_seconds for a link that works once, auth.multi_use_login_token_ttl_seconds otherwise. |
| `max_uses` | integer or null | no | 0 to 1000 | Sign-ins the link allows: 1-1000, or 0 for unlimited. Default 1. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expires_at` | string (date-time) or null | yes |  | null when it never expires. |
| `expires_in_seconds` | integer | yes |  | 0 when it never expires. |
| `login_path` | string | yes |  | Path of the login link. A credential: treat it like a password. |
| `max_uses` | integer | yes |  | 0 when unlimited. |
| `unlimited` | boolean | yes |  |  |

### POST /api/admin/users/{user_id}/rename

Rename someone

Change another account's display name (1-80 characters). Only the owner renames an admin. expect, when given, is the name you think it has now; if it changed meanwhile the rename is refused with the current name.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_rename_user`](mcp.md#admin_rename_user) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `display_name` | string | yes | 1 to 80 characters | 1-80 characters. |
| `expect` | string or null | no |  | The name you think they have. Answers 409 if it is another. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `display_name` | string | yes |  |  |
| `ok` | boolean | yes |  |  |

### POST /api/admin/users/{user_id}/reset-two-factor

Turn off an account's two-factor

For someone who lost both their phone and their recovery codes: switches their two-factor sign-in off so a login link lets them in again. Only the owner resets an admin's; the owner's is reset from a shell.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_reset_two_factor`](mcp.md#admin_reset_two_factor) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/users/{user_id}/revoke-access

Revoke a user's access

For an account that may be compromised: ends every session and voids its live login link and every API key it holds, at once.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_revoke_access`](mcp.md#admin_revoke_access) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `revoked` | [`Revoked`](#revoked) | yes |  |  |

### POST /api/admin/users/{user_id}/role

Change an account's role

Make an account an admin or an ordinary user. Only the owner takes an admin's rights away; the owner is always an admin, and nobody changes their own role. Becoming an admin ends the account's sessions and live login link.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_user_role`](mcp.md#admin_set_user_role) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `role` | string | yes | one of `admin`, `user` |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `role` | string | yes |  |  |

### GET /api/admin/users/{user_id}/sessions

An account's signed-in devices

Where an account is signed in: every session, with when it started and was last seen. Addresses are shown to the owner only, and only the owner sees the owner's.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_list_user_sessions`](mcp.md#admin_list_user_sessions) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `sessions` | array of [`Session`](#session) | yes |  |  |
| `total` | integer | yes |  |  |

### POST /api/admin/users/{user_id}/sessions/{session_id}/revoke

Sign out one of an account's devices

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_revoke_user_session`](mcp.md#admin_revoke_user_session) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `session_id` | path | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/bulk/admin/users/cancel-login-link

Stop an account's login link working, for many at once

Runs POST /api/admin/users/{user_id}/cancel-login-link once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_cancel_login_link_bulk`](mcp.md#admin_cancel_login_link_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/users/delete

Delete an account, for many at once

Runs POST /api/admin/users/{user_id}/delete once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_delete_user_bulk`](mcp.md#admin_delete_user_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/users/disabled

Disable or enable a user, for many at once

Runs POST /api/admin/users/{user_id}/disabled once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_user_disabled_bulk`](mcp.md#admin_set_user_disabled_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].disabled` | boolean | no |  |  |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |
| `shared.disabled` | boolean | no |  |  |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/users/reset-two-factor

Turn off an account's two-factor, for many at once

Runs POST /api/admin/users/{user_id}/reset-two-factor once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_reset_two_factor_bulk`](mcp.md#admin_reset_two_factor_bulk) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/users/revoke-access

Revoke a user's access, for many at once

Runs POST /api/admin/users/{user_id}/revoke-access once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_revoke_access_bulk`](mcp.md#admin_revoke_access_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/users/role

Change an account's role, for many at once

Runs POST /api/admin/users/{user_id}/role once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_user_role_bulk`](mcp.md#admin_set_user_role_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].role` | string | no | one of `admin`, `user` |  |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |
| `shared.role` | string | no | one of `admin`, `user` |  |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/users/sessions/revoke

Sign out one of an account's devices, for many at once

Runs POST /api/admin/users/{user_id}/sessions/{session_id}/revoke once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_revoke_user_session_bulk`](mcp.md#admin_revoke_user_session_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].session_id` | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

## Invite links

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/admin/invite-links`](#get-apiadmininvite-links) | List invite links | browser session or API key, admins |
| POST | [`/api/admin/invite-links`](#post-apiadmininvite-links) | Create an invite link | browser session or API key, admins |
| POST | [`/api/admin/invite-links/{link_id}/delete`](#post-apiadmininvite-linkslink_iddelete) | Delete an invite link | browser session or API key, admins |
| POST | [`/api/admin/invite-links/{link_id}/limits`](#post-apiadmininvite-linkslink_idlimits) | Change an invite link's places and expiry | browser session or API key, admins |
| POST | [`/api/admin/invite-links/{link_id}/rename`](#post-apiadmininvite-linkslink_idrename) | Rename an invite link | browser session or API key, admins |
| POST | [`/api/admin/invite-links/{link_id}/restore`](#post-apiadmininvite-linkslink_idrestore) | Undo deleting an invite link | browser session or API key, admins |
| POST | [`/api/admin/invite-links/{link_id}/revoke`](#post-apiadmininvite-linkslink_idrevoke) | Switch off an invite link | browser session or API key, admins |
| POST | [`/api/admin/invite-links/{link_id}/switch-on`](#post-apiadmininvite-linkslink_idswitch-on) | Switch an invite link back on | browser session or API key, admins |
| POST | [`/api/bulk/admin/invite-links/delete`](#post-apibulkadmininvite-linksdelete) | Delete an invite link, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/invite-links/restore`](#post-apibulkadmininvite-linksrestore) | Undo deleting an invite link, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/invite-links/revoke`](#post-apibulkadmininvite-linksrevoke) | Switch off an invite link, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/invite-links/switch-on`](#post-apibulkadmininvite-linksswitch-on) | Switch an invite link back on, for many at once | browser session or API key, admins |
| POST | [`/api/join/claim`](#post-apijoinclaim) | Sign in as an existing account | anyone |
| POST | [`/api/join/create`](#post-apijoincreate) | Join as a new account | anyone |
| POST | [`/api/join/info`](#post-apijoininfo) | What an invite link is | anyone |
| POST | [`/api/join/match`](#post-apijoinmatch) | Accounts matching a name | anyone |

### GET /api/admin/invite-links

List invite links

Group invite links, a page at a time: label, places used and left, expiry, and whether each still works. The links themselves are never shown again after creation. next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_list_invite_links`](mcp.md#admin_list_invite_links) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no |  | Keep links whose label or maker contains this, in any case. |
| `sort` | query | string | no | one of `label`, `used`, `expires_at` | Field to sort the whole list by. Rows without it go last. |
| `dir` | query | string | no | one of `asc`, `desc` | Default asc. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `invite_links` | array of [`InviteLink`](#invitelink) | yes |  |  |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |

### POST /api/admin/invite-links

Create an invite link

One link for a group: each person who opens it joins as a new user, up to max_uses people, for days days.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_create_invite_link`](mcp.md#admin_create_invite_link) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `days` | integer | yes | 1 to 365 |  |
| `label` | string | yes | 1 to 80 characters |  |
| `max_uses` | integer | yes | 1 to 1000 |  |

**Reply 201**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_by` | integer | yes |  |  |
| `created_by_name` | string | no |  |  |
| `expires_at` | string (date-time) | yes |  |  |
| `id` | string | yes |  |  |
| `join_path` | string | yes |  | The link to hand out, shown in this response only. A credential: treat it like a password. |
| `label` | string | yes |  |  |
| `live` | boolean | yes |  |  |
| `max_uses` | integer | yes |  |  |
| `remaining` | integer | yes |  |  |
| `revoked` | boolean | yes |  |  |
| `used` | integer | yes |  |  |

### POST /api/admin/invite-links/{link_id}/delete

Delete an invite link

Take an invite link off the list and stop it working. The accounts it made stay.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_delete_invite_link`](mcp.md#admin_delete_invite_link) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `link_id` | path | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/invite-links/{link_id}/limits

Change an invite link's places and expiry

Set how many people the link may take in (never fewer than have joined) and, when days is given, a new expiry counted from now. The link then counts as this key's: revoking the key switches it off.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_invite_link_limits`](mcp.md#admin_set_invite_link_limits) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `link_id` | path | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `days` | integer or null | no | 1 to 365 | Counted from now. Leave out to keep the expiry. |
| `max_uses` | integer | yes | 1 to 1000 | Not fewer than have joined. |

**Reply 200**: Success

[`InviteLink`](#invitelink).

### POST /api/admin/invite-links/{link_id}/rename

Rename an invite link

Change the label people see on an invite link's join page (1-80 characters). expect, when given, is the label you think it has now.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_rename_invite_link`](mcp.md#admin_rename_invite_link) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `link_id` | path | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expect` | string or null | no |  | The label you think it has. Answers 409 if it has another. |
| `label` | string | yes | 1 to 80 characters | 1-80 characters. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `label` | string | yes |  |  |
| `ok` | boolean | yes |  |  |

### POST /api/admin/invite-links/{link_id}/restore

Undo deleting an invite link

Bring back an invite link deleted in the last few minutes, working as before. The link then counts as this key's: revoking the key switches it off.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_restore_invite_link`](mcp.md#admin_restore_invite_link) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `link_id` | path | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`InviteLink`](#invitelink).

### POST /api/admin/invite-links/{link_id}/revoke

Switch off an invite link

Stop an invite link from letting anyone else join or sign in through it. Accounts it already made keep working. Cannot be switched back on.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_revoke_invite_link`](mcp.md#admin_revoke_invite_link) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `link_id` | path | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/invite-links/{link_id}/switch-on

Switch an invite link back on

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_switch_on_invite_link`](mcp.md#admin_switch_on_invite_link) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `link_id` | path | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`InviteLink`](#invitelink).

### POST /api/bulk/admin/invite-links/delete

Delete an invite link, for many at once

Runs POST /api/admin/invite-links/{link_id}/delete once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_delete_invite_link_bulk`](mcp.md#admin_delete_invite_link_bulk) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/invite-links/restore

Undo deleting an invite link, for many at once

Runs POST /api/admin/invite-links/{link_id}/restore once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_restore_invite_link_bulk`](mcp.md#admin_restore_invite_link_bulk) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/invite-links/revoke

Switch off an invite link, for many at once

Runs POST /api/admin/invite-links/{link_id}/revoke once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_revoke_invite_link_bulk`](mcp.md#admin_revoke_invite_link_bulk) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/invite-links/switch-on

Switch an invite link back on, for many at once

Runs POST /api/admin/invite-links/{link_id}/switch-on once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_switch_on_invite_link_bulk`](mcp.md#admin_switch_on_invite_link_bulk) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/join/claim

Sign in as an existing account

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 400, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `credential` | string | yes |  |  |
| `display_name` | string | no |  |  |
| `pin` | string | no |  |  |
| `user_id` | integer | no |  |  |

**Reply 200**: Success

[`Joined`](#joined).

### POST /api/join/create

Join as a new account

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 400, 403, 404, 409, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `credential` | string | yes |  |  |
| `display_name` | string | no |  |  |
| `pin` | string | no |  |  |
| `user_id` | integer | no |  |  |

**Reply 201**: Success

[`Joined`](#joined).

### POST /api/join/info

What an invite link is

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 400, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `credential` | string | yes |  |  |
| `display_name` | string | no |  |  |
| `pin` | string | no |  |  |
| `user_id` | integer | no |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `label` | string | yes |  |  |
| `remaining` | integer | yes |  |  |

### POST /api/join/match

Accounts matching a name

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 400, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `credential` | string | yes |  |  |
| `display_name` | string | no |  |  |
| `pin` | string | no |  |  |
| `user_id` | integer | no |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `can_create` | boolean | yes |  |  |
| `matches` | array of object | yes |  |  |
| `matches[].display_name` | string | yes |  |  |
| `matches[].has_pin` | boolean | yes |  |  |
| `matches[].id` | integer | yes |  |  |

## Owner

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/owner/jobs`](#get-apiownerjobs) | Background jobs | browser session only, the owner |
| POST | [`/api/owner/jobs/{name}/run`](#post-apiownerjobsnamerun) | Run a job now | browser session only, the owner |
| GET | [`/api/owner/logs/days`](#get-apiownerlogsdays) | The days on record | browser session only, the owner |
| GET | [`/api/owner/logs/days/{date}`](#get-apiownerlogsdaysdate) | One day's entries | browser session only, the owner |
| POST | [`/api/owner/logs/days/{date}/delete`](#post-apiownerlogsdaysdatedelete) | Delete one day | browser session only, the owner |
| GET | [`/api/owner/logs/days/{date}/download`](#get-apiownerlogsdaysdatedownload) | One day's entries as a file | browser session only, the owner |
| GET | [`/api/owner/logs/ip-names`](#get-apiownerlogsip-names) | What the owner calls addresses | browser session only, the owner |
| POST | [`/api/owner/logs/ip-names`](#post-apiownerlogsip-names) | Name an address or a range, or rename one | browser session only, the owner |
| POST | [`/api/owner/logs/ip-names/delete`](#post-apiownerlogsip-namesdelete) | Take a name off an address or a range | browser session only, the owner |
| POST | [`/api/owner/logs/prune`](#post-apiownerlogsprune) | Apply the retention rules now | browser session only, the owner |
| GET | [`/api/owner/logs/stream`](#get-apiownerlogsstream) | New log entries as they happen | browser session only, the owner |
| GET | [`/api/owner/settings`](#get-apiownersettings) | config.yaml, as a form | browser session only, the owner |
| POST | [`/api/owner/settings`](#post-apiownersettings) | Save settings | browser session only, the owner |
| POST | [`/api/owner/users/{user_id}/password-access`](#post-apiownerusersuser_idpassword-access) | Allow or block an account's password | browser session only, the owner |

### GET /api/owner/jobs

Background jobs

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 404, 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `jobs` | array of [`Job`](#job) | yes |  |  |

### POST /api/owner/jobs/{name}/run

Run a job now

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `name` | path | string | yes |  | The job's name, from GET /api/owner/jobs. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 202**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### GET /api/owner/logs/days

The days on record

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `q` | query | string | no |  | The Logs screen's search. |

**Reply 200**: Success

[`LogDays`](#logdays).

### GET /api/owner/logs/days/{date}

One day's entries

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `date` | path | string | yes |  | The day's date, from GET /api/owner/logs/days. |
| `q` | query | string | no |  | The Logs screen's search. |
| `before` | query | integer or null | no | at least 0 |  |
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |

**Reply 200**: Success

[`LogDay`](#logday).

### POST /api/owner/logs/days/{date}/delete

Delete one day

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `date` | path | string | yes |  | The day's date, from GET /api/owner/logs/days. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `deleted` | boolean | yes |  |  |

### GET /api/owner/logs/days/{date}/download

One day's entries as a file

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `date` | path | string | yes |  | The day's date, from GET /api/owner/logs/days. |

**Reply 200**: Success

Text, as `application/x-ndjson`.

### GET /api/owner/logs/ip-names

What the owner calls addresses

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 403, 404, 429, 503 |

**Reply 200**: Success

[`IPNames`](#ipnames).

### POST /api/owner/logs/ip-names

Name an address or a range, or rename one

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `name` | string | yes | 1 to 80 characters |  |
| `target` | string | yes | e.g. `10.0.0.0/24` | An address, or a range such as 10.0.0.0/24. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `name` | string | yes |  |  |
| `target` | string | yes |  |  |

### POST /api/owner/logs/ip-names/delete

Take a name off an address or a range

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `target` | string | yes | e.g. `10.0.0.0/24` | An address, or a range such as 10.0.0.0/24. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `deleted` | boolean | yes |  |  |

### POST /api/owner/logs/prune

Apply the retention rules now

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`PruneResult`](#pruneresult).

### GET /api/owner/logs/stream

New log entries as they happen

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `q` | query | string | no |  | The Logs screen's search. |

**Reply 200**: Success

Text, as `text/event-stream`.

### GET /api/owner/settings

config.yaml, as a form

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 404, 429, 503 |

**Reply 200**: Success

[`SettingsSnapshot`](#settingssnapshot).

### POST /api/owner/settings

Save settings

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 409, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `changes` | object | yes |  | Setting paths and their new values. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `version` | string | yes |  |  |

**Reply 200**: Success

[`SettingsResult`](#settingsresult).

### POST /api/owner/users/{user_id}/password-access

Allow or block an account's password

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `allowed` | boolean | yes |  |  |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `password_blocked` | boolean | yes |  |  |

## Webhooks

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/owner/webhooks`](#get-apiownerwebhooks) | Webhooks | browser session only, the owner |
| POST | [`/api/owner/webhooks`](#post-apiownerwebhooks) | Create a webhook | browser session only, the owner |
| GET | [`/api/owner/webhooks/events`](#get-apiownerwebhooksevents) | Events a webhook can ask for | browser session only, the owner |
| POST | [`/api/owner/webhooks/{webhook_id}/delete`](#post-apiownerwebhookswebhook_iddelete) | Delete a webhook | browser session only, the owner |
| GET | [`/api/owner/webhooks/{webhook_id}/deliveries`](#get-apiownerwebhookswebhook_iddeliveries) | Recent deliveries | browser session only, the owner |
| POST | [`/api/owner/webhooks/{webhook_id}/deliveries/{delivery_id}/resend`](#post-apiownerwebhookswebhook_iddeliveriesdelivery_idresend) | Send a failed delivery again | browser session only, the owner |
| POST | [`/api/owner/webhooks/{webhook_id}/enabled`](#post-apiownerwebhookswebhook_idenabled) | Switch a webhook on or off | browser session only, the owner |
| POST | [`/api/owner/webhooks/{webhook_id}/rename`](#post-apiownerwebhookswebhook_idrename) | Rename a webhook | browser session only, the owner |
| POST | [`/api/owner/webhooks/{webhook_id}/restore`](#post-apiownerwebhookswebhook_idrestore) | Undo deleting a webhook | browser session only, the owner |
| POST | [`/api/owner/webhooks/{webhook_id}/rotate-secret`](#post-apiownerwebhookswebhook_idrotate-secret) | Replace a webhook's signing secret | browser session only, the owner |
| POST | [`/api/owner/webhooks/{webhook_id}/target`](#post-apiownerwebhookswebhook_idtarget) | Change a webhook's address and events | browser session only, the owner |
| POST | [`/api/owner/webhooks/{webhook_id}/test`](#post-apiownerwebhookswebhook_idtest) | Send a test event | browser session only, the owner |

### GET /api/owner/webhooks

Webhooks

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no |  | Keep webhooks whose name or address contains this, in any case. |
| `sort` | query | string | no | one of `name`, `url`, `enabled` | Field to sort the whole list by. Rows without it go last. |
| `dir` | query | string | no | one of `asc`, `desc` | Default asc. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |
| `webhooks` | array of [`Webhook`](#webhook) | yes |  |  |

### POST /api/owner/webhooks

Create a webhook

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `events` | array of string | yes |  |  |
| `name` | string | yes | 1 to 80 characters |  |
| `url` | string | yes | 1 to 2000 characters |  |

**Reply 201**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_at` | string (date-time) | yes |  |  |
| `enabled` | boolean | yes |  |  |
| `events` | array of string | yes |  |  |
| `id` | string | yes |  |  |
| `name` | string | yes |  |  |
| `secret` | string | yes |  | Shown in this response only. |
| `url` | string | yes |  |  |

### GET /api/owner/webhooks/events

Events a webhook can ask for

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 404, 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `events` | array of string | yes |  |  |

### POST /api/owner/webhooks/{webhook_id}/delete

Delete a webhook

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `webhook_id` | path | string | yes |  | The webhook's id, from GET /api/owner/webhooks. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### GET /api/owner/webhooks/{webhook_id}/deliveries

Recent deliveries

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `webhook_id` | path | string | yes |  | The webhook's id, from GET /api/owner/webhooks. |
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `deliveries` | array of [`WebhookDelivery`](#webhookdelivery) | yes |  |  |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |

### POST /api/owner/webhooks/{webhook_id}/deliveries/{delivery_id}/resend

Send a failed delivery again

A new delivery with the same body. The failed one stays listed.

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `webhook_id` | path | string | yes |  | The webhook's id, from GET /api/owner/webhooks. |
| `delivery_id` | path | string | yes |  | The delivery's id, from GET /api/owner/webhooks/{webhook_id}/deliveries. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 202**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/owner/webhooks/{webhook_id}/enabled

Switch a webhook on or off

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `webhook_id` | path | string | yes |  | The webhook's id, from GET /api/owner/webhooks. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `enabled` | boolean | yes |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/owner/webhooks/{webhook_id}/rename

Rename a webhook

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `webhook_id` | path | string | yes |  | The webhook's id, from GET /api/owner/webhooks. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expect` | string or null | no |  | The name you think it has. Answers 409 if it has another. |
| `name` | string | yes | 1 to 80 characters | 1-80 characters. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `name` | string | yes |  |  |
| `ok` | boolean | yes |  |  |

### POST /api/owner/webhooks/{webhook_id}/restore

Undo deleting a webhook

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `webhook_id` | path | string | yes |  | The webhook's id, from GET /api/owner/webhooks. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`Webhook`](#webhook).

### POST /api/owner/webhooks/{webhook_id}/rotate-secret

Replace a webhook's signing secret

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `webhook_id` | path | string | yes |  | The webhook's id, from GET /api/owner/webhooks. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `secret` | string | yes |  | Shown only here. The old one has stopped working. |

### POST /api/owner/webhooks/{webhook_id}/target

Change a webhook's address and events

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `webhook_id` | path | string | yes |  | The webhook's id, from GET /api/owner/webhooks. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `events` | array of string | yes |  |  |
| `expect` | object or null | no |  | What you think it has now. Answers 409 if it has something else. |
| `expect.events` | array of string | yes |  |  |
| `expect.url` | string | yes |  |  |
| `url` | string | yes | 1 to 2000 characters | Checked like a new webhook's. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/owner/webhooks/{webhook_id}/test

Send a test event

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `webhook_id` | path | string | yes |  | The webhook's id, from GET /api/owner/webhooks. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 202**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

## Monitoring

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/api.json`](#get-apiapijson) | This document | browser session or API key, anyone signed in |
| GET | [`/api/api.md`](#get-apiapimd) | The API reference as Markdown tables for people | browser session or API key, anyone signed in |
| GET | [`/api/health`](#get-apihealth) | Is the process up | anyone |
| GET | [`/api/mcp.json`](#get-apimcpjson) | The MCP tools: every tool /mcp has, its arguments, and the route it stands for | browser session or API key, anyone signed in |
| GET | [`/api/mcp.md`](#get-apimcpmd) | The MCP tools as Markdown tables for people | browser session or API key, anyone signed in |
| GET | [`/api/ready`](#get-apiready) | Can the server do real work | anyone |
| GET | [`/mcp`](#get-mcp) | Refused: this server has no stream | anyone |
| POST | [`/mcp`](#post-mcp) | Model Context Protocol, at the key's role | API key only, anyone signed in |
| DELETE | [`/mcp`](#delete-mcp) | Refused: this server keeps no sessions | anyone |

### GET /api/api.json

This document

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | None, HTTP only: The API reference itself; GET /api/mcp.json is the MCP one. |
| Errors | 401, 404, 429, 503 |

**Reply 200**: Success

An empty object.

### GET /api/api.md

The API reference as Markdown tables for people

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | None, HTTP only: The API reference itself, for people; GET /api/api.json is the same as JSON. |
| Errors | 401, 404, 429, 503 |

**Reply 200**: Success

Text, as `text/markdown`.

### GET /api/health

Is the process up

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### GET /api/mcp.json

The MCP tools: every tool /mcp has, its arguments, and the route it stands for

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | None, HTTP only: The MCP reference itself; tools/list gives the same list. |
| Errors | 401, 404, 429, 503 |

**Reply 200**: Success

An empty object.

### GET /api/mcp.md

The MCP tools as Markdown tables for people

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | None, HTTP only: The MCP reference itself, for people; tools/list gives the same list. |
| Errors | 401, 404, 429, 503 |

**Reply 200**: Success

Text, as `text/markdown`.

### GET /api/ready

Can the server do real work

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### GET /mcp

Refused: this server has no stream

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 404, 429, 503 |

**Reply 405**: Success

### POST /mcp

Model Context Protocol, at the key's role

|  |  |
|---|---|
| Who may call | API key only, anyone signed in |
| MCP tool | None, HTTP only: The MCP endpoint itself. |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

An empty object.

**Reply 200**: Success

An empty object.

### DELETE /mcp

Refused: this server keeps no sessions

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 404, 429, 503 |

**Reply 405**: Success

## Containers

| Method | Path | Summary | Who may call |
|---|---|---|---|
| POST | [`/api/bulk/households/labels/delete`](#post-apibulkhouseholdslabelsdelete) | Delete a container, for many at once | browser session or API key, anyone signed in |
| POST | [`/api/bulk/households/labels/restore`](#post-apibulkhouseholdslabelsrestore) | Restore a container, for many at once | browser session or API key, anyone signed in |
| GET | [`/api/households/{household_id}/labels`](#get-apihouseholdshousehold_idlabels) | List containers | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/labels`](#post-apihouseholdshousehold_idlabels) | Set up a container | browser session or API key, anyone signed in |
| GET | [`/api/households/{household_id}/labels/{label_id}`](#get-apihouseholdshousehold_idlabelslabel_id) | One container and what is in it | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/labels/{label_id}`](#post-apihouseholdshousehold_idlabelslabel_id) | Change a container | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/labels/{label_id}/codes`](#post-apihouseholdshousehold_idlabelslabel_idcodes) | Add a code to a container | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/labels/{label_id}/codes/release`](#post-apihouseholdshousehold_idlabelslabel_idcodesrelease) | Take a code off a container | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/labels/{label_id}/delete`](#post-apihouseholdshousehold_idlabelslabel_iddelete) | Delete a container | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/labels/{label_id}/move-contents`](#post-apihouseholdshousehold_idlabelslabel_idmove-contents) | Move everything to another container | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/labels/{label_id}/restore`](#post-apihouseholdshousehold_idlabelslabel_idrestore) | Restore a container | browser session or API key, anyone signed in |

### POST /api/bulk/households/labels/delete

Delete a container, for many at once

Runs POST /api/households/{household_id}/labels/{label_id}/delete once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`delete_label_bulk`](mcp.md#delete_label_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `items[].label_id` | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/households/labels/restore

Restore a container, for many at once

Runs POST /api/households/{household_id}/labels/{label_id}/restore once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`restore_label_bulk`](mcp.md#restore_label_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `items[].label_id` | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### GET /api/households/{household_id}/labels

List containers

Containers in a household, each with its code, where it is, how many items it holds and what they are worth. Narrow by words, location, outer container, color or moving state.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`list_labels`](mcp.md#list_labels) |
| Errors | 400, 401, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no |  | Keep containers whose name, code, notes or place has these words. |
| `location_id` | query | integer | no |  | Only containers at this location or inside it. |
| `parent_id` | query | integer | no |  | Only containers sitting directly inside this one. |
| `color` | query | string | no |  | Only this label color, #rrggbb. |
| `move_status` | query | string | no | one of `packed`, `in_transit`, `unpacked` | Only containers in this moving state. |
| `trash` | query | boolean | no |  | List the ones in Trash instead. |
| `sort` | query | string | no | one of `name`, `path`, `item_count`, `value`, `updated_at`, `created_at` | Field to sort by. |
| `dir` | query | string | no | one of `asc`, `desc` | Default asc. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `labels` | array of [`Container`](#container) | yes |  |  |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |

### POST /api/households/{household_id}/labels

Set up a container

Make a container: give it a name, and optionally the code on its sticker, a color, where it is, and the container it sits in.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`create_label`](mcp.md#create_label) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `code` | string | no | at most 2000 characters | A printed code that is not on anything yet (from a print run), or any other QR or barcode value to claim, such as a bought sticker. Leave out to make a fresh code. |
| `color` | string or null | no |  | Its label color, #rrggbb, or empty. |
| `container_type` | string or null | no | at most 40 characters | bin, box, drawer, shelf, bag, or anything else. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `destination` | string or null | no | at most 80 characters | For a move: the room it goes to. |
| `flags` | array of string or null | no |  | Handling notes printed on its label, such as fragile, this side up, heavy. |
| `location_id` | integer or null | no |  | Where it sits; 0 for nowhere. |
| `move_status` | string or null | no | one of ``, `packed`, `in_transit`, `unpacked` | For a move: packed, in_transit, unpacked, or empty. |
| `name` | string or null | no | 1 to 80 characters |  |
| `notes` | string or null | no | at most 4000 characters |  |
| `parent_id` | integer or null | no |  | The container it sits inside; 0 for none. |
| `photos` | array of string or null | no |  | File ids of its photos, in order (see attach_photo). |

**Reply 201**: Success

[`Container`](#container).

### GET /api/households/{household_id}/labels/{label_id}

One container and what is in it

A container with its items and the containers inside it.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`get_label`](mcp.md#get_label) |
| Errors | 400, 401, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | path | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `child_count` | integer | yes |  | Containers sitting inside this one. |
| `children` | array of [`Container`](#container) | yes |  | Containers sitting inside it. |
| `claimed` | array of string | yes |  |  |
| `code` | string | yes |  |  |
| `color` | string | no |  |  |
| `color_means` | string | yes |  | What its color stands for in this household. |
| `container_type` | string | no |  |  |
| `created_at` | string (date-time) | yes |  |  |
| `created_by` | integer | yes |  |  |
| `deleted_at` | string (date-time) or null | no |  |  |
| `deleted_by` | integer | no |  |  |
| `destination` | string | no |  |  |
| `flags` | array of string | yes |  |  |
| `household_id` | integer | yes |  |  |
| `id` | integer | yes |  |  |
| `item_count` | integer | yes |  |  |
| `items` | array of [`Item`](#item) | yes |  | What is in it, not counting what is in Trash. |
| `location_id` | integer | no |  |  |
| `move_status` | string | no |  |  |
| `name` | string | yes |  |  |
| `notes` | string | no |  |  |
| `parent_id` | integer | no |  |  |
| `path` | string | yes |  | Where it is: location › outer containers › this one. |
| `photos` | array of string | yes |  |  |
| `printed_code` | string | yes |  | The code as printed under its QR, XXXXX-XXXXX. |
| `updated_at` | string (date-time) | yes |  |  |
| `value` | number | yes |  | What its items are worth together. |

### POST /api/households/{household_id}/labels/{label_id}

Change a container

Change a container's name, color, type, location, outer container, photos, notes, moving state, handling flags or destination. Only what you send changes.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`update_label`](mcp.md#update_label) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | path | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `color` | string or null | no |  | Its label color, #rrggbb, or empty. |
| `container_type` | string or null | no | at most 40 characters | bin, box, drawer, shelf, bag, or anything else. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `destination` | string or null | no | at most 80 characters | For a move: the room it goes to. |
| `flags` | array of string or null | no |  | Handling notes printed on its label, such as fragile, this side up, heavy. |
| `location_id` | integer or null | no |  | Where it sits; 0 for nowhere. |
| `move_status` | string or null | no | one of ``, `packed`, `in_transit`, `unpacked` | For a move: packed, in_transit, unpacked, or empty. |
| `name` | string or null | no | 1 to 80 characters |  |
| `notes` | string or null | no | at most 4000 characters |  |
| `parent_id` | integer or null | no |  | The container it sits inside; 0 for none. |
| `photos` | array of string or null | no |  | File ids of its photos, in order (see attach_photo). |

**Reply 200**: Success

[`Container`](#container).

### POST /api/households/{household_id}/labels/{label_id}/codes

Add a code to a container

Make another code open this container too: a store-bought QR sticker, a product's barcode, or one of this household's printed codes.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`claim_code`](mcp.md#claim_code) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | path | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `code` | string | yes | 1 to 2000 characters | The scanned value: a QR code's text or a barcode's digits. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`Container`](#container).

### POST /api/households/{household_id}/labels/{label_id}/codes/release

Take a code off a container

Stop an added code opening this container. Its own code cannot be taken off.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`release_code`](mcp.md#release_code) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | path | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `code` | string | yes | 1 to 2000 characters | The scanned value: a QR code's text or a barcode's digits. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`Container`](#container).

### POST /api/households/{household_id}/labels/{label_id}/delete

Delete a container

Put a container and the items in it in Trash. restore_label brings both back. Its codes keep pointing at it until it is purged.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`delete_label`](mcp.md#delete_label) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | path | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`Container`](#container).

### POST /api/households/{household_id}/labels/{label_id}/move-contents

Move everything to another container

Move every item in this container into another one in the same household, in one go.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`move_contents`](mcp.md#move_contents) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | path | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `to_label_id` | integer | yes | at least 1 | The container everything goes into. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `item_ids` | array of integer | yes |  |  |
| `moved` | integer | yes |  |  |

### POST /api/households/{household_id}/labels/{label_id}/restore

Restore a container

Take a container back out of Trash, with the items that were deleted along with it.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`restore_label`](mcp.md#restore_label) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | path | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`Container`](#container).

## Households

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/households`](#get-apihouseholds) | List your households | browser session or API key, anyone signed in |
| POST | [`/api/households`](#post-apihouseholds) | Create a household | browser session or API key, anyone signed in |
| GET | [`/api/households/{household_id}`](#get-apihouseholdshousehold_id) | One household | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/colors`](#post-apihouseholdshousehold_idcolors) | Say what label colors mean | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/delete`](#post-apihouseholdshousehold_iddelete) | Delete a household | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/fields`](#post-apihouseholdshousehold_idfields) | Set custom fields and reminder lead | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/leave`](#post-apihouseholdshousehold_idleave) | Leave a household | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/members`](#post-apihouseholdshousehold_idmembers) | Add, change or remove a member | browser session or API key, anyone signed in |
| GET | [`/api/households/{household_id}/photos/{file_id}`](#get-apihouseholdshousehold_idphotosfile_id) | A household photo or receipt | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/rename`](#post-apihouseholdshousehold_idrename) | Rename a household | browser session or API key, anyone signed in |

### GET /api/households

List your households

The households you are a member of, each with your role, its members, color meanings and custom fields. Every other inventory tool takes one of these ids.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`list_households`](mcp.md#list_households) |
| Errors | 400, 401, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `households` | array of [`Household`](#household) | yes |  |  |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |

### POST /api/households

Create a household

Make a new shared inventory. You are its first owner.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`create_household`](mcp.md#create_household) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `name` | string | yes | 1 to 80 characters |  |

**Reply 201**: Success

[`Household`](#household).

### GET /api/households/{household_id}

One household

A household you are in: your role, its members, color meanings and custom fields.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`get_household`](mcp.md#get_household) |
| Errors | 400, 401, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

**Reply 200**: Success

[`Household`](#household).

### POST /api/households/{household_id}/colors

Say what label colors mean

Replace the household's color legend: which room or kind of thing each label color stands for. Owners and editors.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`set_household_colors`](mcp.md#set_household_colors) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `colors` | map of string | yes |  | Label color (#rrggbb) to what it means here, such as "#d32f2f": "Kitchen". Replaces the whole list. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`Household`](#household).

### POST /api/households/{household_id}/delete

Delete a household

Delete a household and everything in it, for good: its locations, containers, codes, print runs and items. Owners only. Photos stay in their uploaders' Files.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`delete_household`](mcp.md#delete_household) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/households/{household_id}/fields

Set custom fields and reminder lead

Replace the custom fields items here can fill in (serial numbers, warranty dates, anything), and how many days before an expiry reminders go out. Owners and editors.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`set_household_fields`](mcp.md#set_household_fields) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `fields` | array of object | yes |  | The whole list. Each has a name, a type (text, number, date, select or bool), options for a select, and the id it already has, if any, so values survive a rename. |
| `fields[].id` | string | yes |  |  |
| `fields[].name` | string | yes |  |  |
| `fields[].options` | array of string | no |  |  |
| `fields[].type` | string | yes |  |  |
| `reminder_days` | integer or null | yes |  | Days before an expiry its reminder goes out, 0-365; null uses the app's default. |

**Reply 200**: Success

[`Household`](#household).

### POST /api/households/{household_id}/leave

Leave a household

Stop being a member. The last owner cannot leave: make someone else an owner, or delete the household.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`leave_household`](mcp.md#leave_household) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/households/{household_id}/members

Add, change or remove a member

Give someone with an account a role here -- owner (everything), editor (change the inventory) or viewer (look only) -- or an empty role to take them out. Owners only. A household always keeps at least one owner.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`set_household_member`](mcp.md#set_household_member) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `role` | string | yes | one of `owner`, `editor`, `viewer`, `` | owner, editor or viewer; empty takes them out. |
| `user_id` | integer | yes | at least 1 | Their account id, from GET /api/team (list_team over MCP). |

**Reply 200**: Success

[`Household`](#household).

### GET /api/households/{household_id}/photos/{file_id}

A household photo or receipt

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | None, HTTP only: Photo bytes are not JSON: GET this address with the same Authorization header. |
| Errors | 400, 401, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `file_id` | path | string | yes |  | A file id from a container's or item's photos or receipts. |

**Reply 200**: The photo, as itself when it is a JPEG, PNG, WebP or GIF, else as an attachment

Text, as `image/jpeg`.

### POST /api/households/{household_id}/rename

Rename a household

Give a household a new name. Owners and editors.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`rename_household`](mcp.md#rename_household) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `name` | string | yes | 1 to 80 characters |  |

**Reply 200**: Success

[`Household`](#household).

## Items

| Method | Path | Summary | Who may call |
|---|---|---|---|
| POST | [`/api/bulk/households/items/delete`](#post-apibulkhouseholdsitemsdelete) | Delete an item, for many at once | browser session or API key, anyone signed in |
| POST | [`/api/bulk/households/items/move`](#post-apibulkhouseholdsitemsmove) | Move an item to another container, for many at once | browser session or API key, anyone signed in |
| POST | [`/api/bulk/households/items/restore`](#post-apibulkhouseholdsitemsrestore) | Restore an item, for many at once | browser session or API key, anyone signed in |
| GET | [`/api/households/{household_id}/items`](#get-apihouseholdshousehold_iditems) | List and search items | browser session or API key, anyone signed in |
| GET | [`/api/households/{household_id}/items/{item_id}`](#get-apihouseholdshousehold_iditemsitem_id) | One item | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/items/{item_id}`](#post-apihouseholdshousehold_iditemsitem_id) | Change an item | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/items/{item_id}/adjust`](#post-apihouseholdshousehold_iditemsitem_idadjust) | Add to or take from an item's count | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/items/{item_id}/delete`](#post-apihouseholdshousehold_iditemsitem_iddelete) | Delete an item | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/items/{item_id}/move`](#post-apihouseholdshousehold_iditemsitem_idmove) | Move an item to another container | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/items/{item_id}/restore`](#post-apihouseholdshousehold_iditemsitem_idrestore) | Restore an item | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/labels/{label_id}/items`](#post-apihouseholdshousehold_idlabelslabel_iditems) | Put items in a container | browser session or API key, anyone signed in |

### POST /api/bulk/households/items/delete

Delete an item, for many at once

Runs POST /api/households/{household_id}/items/{item_id}/delete once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`delete_item_bulk`](mcp.md#delete_item_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `items[].item_id` | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/households/items/move

Move an item to another container, for many at once

Runs POST /api/households/{household_id}/items/{item_id}/move once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`move_item_bulk`](mcp.md#move_item_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `items[].item_id` | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |
| `items[].label_id` | integer | no | at least 1 | The container it goes into. |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |
| `shared.label_id` | integer | no | at least 1 | The container it goes into. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/households/items/restore

Restore an item, for many at once

Runs POST /api/households/{household_id}/items/{item_id}/restore once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`restore_item_bulk`](mcp.md#restore_item_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `items[].item_id` | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### GET /api/households/{household_id}/items

List and search items

Items in a household, each with the container it is in and where that is. Search by words (a typo is forgiven) and narrow by container, location, tag, color, expiry, low stock or loans.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`list_items`](mcp.md#list_items) |
| Errors | 400, 401, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no |  | Words to find: in names, descriptions, tags, brands, barcodes, serials, custom fields, and where it is. Forgives a typo. |
| `label_id` | query | integer | no |  | Only items in this container or the containers inside it. |
| `location_id` | query | integer | no |  | Only items at this location or anywhere inside it. |
| `tag` | query | string | no |  | Only items with this tag. |
| `color` | query | string | no |  | Only items in containers of this label color. |
| `expiring_within` | query | integer or null | no | 0 to 3650 | Only items expiring within this many days (or already expired). |
| `low_stock` | query | boolean | no |  | Only items at or below their low-stock level. |
| `lent` | query | boolean | no |  | Only items lent out. |
| `trash` | query | boolean | no |  | List the ones in Trash instead. |
| `sort` | query | string | no | one of `name`, `path`, `quantity`, `expires`, `value`, `updated_at`, `created_at` | Field to sort by. |
| `dir` | query | string | no | one of `asc`, `desc` | Default asc. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of [`Item`](#item) | yes |  |  |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |
| `total_value` | number | yes |  | What every item matched is worth, value times quantity. |

### GET /api/households/{household_id}/items/{item_id}

One item

An item with everything known about it, where it is, and its history.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`get_item`](mcp.md#get_item) |
| Errors | 400, 401, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `item_id` | path | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |

**Reply 200**: Success

[`Item`](#item).

### POST /api/households/{household_id}/items/{item_id}

Change an item

Change anything about an item: name, photos, tags, quantity, low-stock level, expiry, custom fields, value, receipts, who it is lent to. Only what you send changes.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`update_item`](mcp.md#update_item) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `item_id` | path | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `barcode` | string or null | no | at most 200 characters |  |
| `brand` | string or null | no | at most 80 characters |  |
| `clear_low_stock` | boolean | no |  | Stop low-stock reminders. |
| `clear_remind_days` | boolean | no |  | Use the household's reminder lead again. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `description` | string or null | no | at most 4000 characters |  |
| `expires` | string or null | no |  | Expiry date, YYYY-MM-DD, or empty. |
| `fields` | map of string or null | no |  | Custom field values by field id (see the household's fields). Replaces them all. |
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

**Reply 200**: Success

[`Item`](#item).

### POST /api/households/{household_id}/items/{item_id}/adjust

Add to or take from an item's count

Change an item's quantity by an amount, such as -1 when one is used up. Two people doing this at once both count.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`adjust_quantity`](mcp.md#adjust_quantity) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `item_id` | path | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `delta` | number | yes |  | How many to add; negative to take some out. Never goes below 0. |

**Reply 200**: Success

[`Item`](#item).

### POST /api/households/{household_id}/items/{item_id}/delete

Delete an item

Put an item in Trash; restore_item brings it back until it is purged.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`delete_item`](mcp.md#delete_item) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `item_id` | path | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`Item`](#item).

### POST /api/households/{household_id}/items/{item_id}/move

Move an item to another container

Put an item in another container of the same household.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`move_item`](mcp.md#move_item) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `item_id` | path | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `label_id` | integer | yes | at least 1 | The container it goes into. |

**Reply 200**: Success

[`Item`](#item).

### POST /api/households/{household_id}/items/{item_id}/restore

Restore an item

Take an item back out of Trash. Its container has to be out of Trash first.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`restore_item`](mcp.md#restore_item) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `item_id` | path | integer | yes | at least 1 | The item's id, from list_items, search_inventory or get_label. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`Item`](#item).

### POST /api/households/{household_id}/labels/{label_id}/items

Put items in a container

Add up to 100 items to a container at once, each with at least a name. A new item's quantity is 1 unless you say.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`create_items`](mcp.md#create_items) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | path | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
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

**Reply 201**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of [`Item`](#item) | yes |  |  |

## Labels

| Method | Path | Summary | Who may call |
|---|---|---|---|
| POST | [`/api/households/{household_id}/labels/reprint`](#post-apihouseholdshousehold_idlabelsreprint) | Codes to print again | browser session or API key, anyone signed in |
| GET | [`/api/households/{household_id}/sheets`](#get-apihouseholdshousehold_idsheets) | List print runs | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/sheets`](#post-apihouseholdshousehold_idsheets) | Make codes to print | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/sheets/{sheet_id}/delete`](#post-apihouseholdshousehold_idsheetssheet_iddelete) | Forget a print run | browser session or API key, anyone signed in |

### POST /api/households/{household_id}/labels/reprint

Codes to print again

The codes, names, colors and handling flags of containers you already have, to print replacement labels.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`reprint_labels`](mcp.md#reprint_labels) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `label_ids` | array of integer | yes |  | Containers whose own labels to print again. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `labels` | array of object | yes |  |  |
| `labels[].code` | string | yes |  | As printed, XXXXX-XXXXX. |
| `labels[].color` | string | yes |  |  |
| `labels[].flags` | array of string | yes |  |  |
| `labels[].label_id` | integer | yes |  |  |
| `labels[].name` | string | yes |  |  |

### GET /api/households/{household_id}/sheets

List print runs

Sheets of codes made for printing, newest first, with how many are on containers.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`list_sheets`](mcp.md#list_sheets) |
| Errors | 400, 401, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `sheets` | array of [`PrintRun`](#printrun) | yes |  |  |
| `total` | integer | yes |  |  |

### POST /api/households/{household_id}/sheets

Make codes to print

Mint 1-300 brand-new codes for printing. Each one, once stuck on something and scanned, offers to set up that container.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`create_sheet`](mcp.md#create_sheet) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `count` | integer | yes | 1 to 300 | How many labels. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `layout` | string | no | at most 40 characters | Which sheet it is printed on, as the Print screen names it. |
| `names` | array of string | no |  | Names to print under the codes, in order (optional). |
| `palette` | string | no | at most 40 characters | original, neon, modern, or a single #rrggbb. |

**Reply 201**: Success

[`PrintRun`](#printrun).

### POST /api/households/{household_id}/sheets/{sheet_id}/delete

Forget a print run

Forget a print run. Its codes already on containers keep working; the rest stop, so stickers thrown away cannot be scanned later.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`delete_sheet`](mcp.md#delete_sheet) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `sheet_id` | path | integer | yes | at least 1 | The print run's id, from list_sheets. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

## Locations

| Method | Path | Summary | Who may call |
|---|---|---|---|
| POST | [`/api/bulk/households/locations/delete`](#post-apibulkhouseholdslocationsdelete) | Delete a location, for many at once | browser session or API key, anyone signed in |
| POST | [`/api/bulk/households/locations/restore`](#post-apibulkhouseholdslocationsrestore) | Restore a location, for many at once | browser session or API key, anyone signed in |
| GET | [`/api/households/{household_id}/locations`](#get-apihouseholdshousehold_idlocations) | List locations | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/locations`](#post-apihouseholdshousehold_idlocations) | Add a location | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/locations/{location_id}`](#post-apihouseholdshousehold_idlocationslocation_id) | Rename or move a location | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/locations/{location_id}/delete`](#post-apihouseholdshousehold_idlocationslocation_iddelete) | Delete a location | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/locations/{location_id}/restore`](#post-apihouseholdshousehold_idlocationslocation_idrestore) | Restore a location | browser session or API key, anyone signed in |

### POST /api/bulk/households/locations/delete

Delete a location, for many at once

Runs POST /api/households/{household_id}/locations/{location_id}/delete once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`delete_location_bulk`](mcp.md#delete_location_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `items[].location_id` | integer | yes | at least 1 | The location's id, from list_locations. |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/households/locations/restore

Restore a location, for many at once

Runs POST /api/households/{household_id}/locations/{location_id}/restore once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`restore_location_bulk`](mcp.md#restore_location_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].household_id` | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `items[].location_id` | integer | yes | at least 1 | The location's id, from list_locations. |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### GET /api/households/{household_id}/locations

List locations

The places containers sit, such as House › Garage › Shelf 2, each with its full path, in path order.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`list_locations`](mcp.md#list_locations) |
| Errors | 400, 401, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `trash` | query | boolean | no |  | List the ones in Trash instead. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `locations` | array of [`Location`](#location) | yes |  |  |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |

### POST /api/households/{household_id}/locations

Add a location

Add a place containers can sit, optionally inside another (Garage inside House).

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`create_location`](mcp.md#create_location) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `name` | string | yes | 1 to 80 characters |  |
| `parent_id` | integer | no |  | The location it is inside; leave out for a top-level one. |

**Reply 201**: Success

[`Location`](#location).

### POST /api/households/{household_id}/locations/{location_id}

Rename or move a location

Change a location's name, or move it inside another (0 for top-level). Only what you send changes.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`update_location`](mcp.md#update_location) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `location_id` | path | integer | yes | at least 1 | The location's id, from list_locations. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `name` | string or null | no | 1 to 80 characters |  |
| `parent_id` | integer or null | no |  | Move it inside another location; 0 makes it top-level. |

**Reply 200**: Success

[`Location`](#location).

### POST /api/households/{household_id}/locations/{location_id}/delete

Delete a location

Put a location in Trash. Containers there keep it until it is purged, and show no location meanwhile; restore_location puts it back.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`delete_location`](mcp.md#delete_location) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `location_id` | path | integer | yes | at least 1 | The location's id, from list_locations. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`Location`](#location).

### POST /api/households/{household_id}/locations/{location_id}/restore

Restore a location

Take a location back out of Trash.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`restore_location`](mcp.md#restore_location) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `location_id` | path | integer | yes | at least 1 | The location's id, from list_locations. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`Location`](#location).

## Photo extraction

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/account/ai`](#get-apiaccountai) | Your photo-reading settings | browser session or API key, anyone signed in |
| POST | [`/api/account/ai`](#post-apiaccountai) | Set your own key, model and effort | browser session only, anyone signed in |
| GET | [`/api/admin/ai-access`](#get-apiadminai-access) | Who may use the server's Anthropic key | browser session or API key, admins |
| POST | [`/api/admin/ai-access`](#post-apiadminai-access) | Set the default for the server's Anthropic key | browser session or API key, admins |
| POST | [`/api/admin/ai-access/{user_id}`](#post-apiadminai-accessuser_id) | Set one person's use of the server's Anthropic key | browser session or API key, admins |
| GET | [`/api/ai`](#get-apiai) | Whether Smart Photo Extraction is ready for you | browser session or API key, anyone signed in |
| POST | [`/api/bulk/admin/ai-access`](#post-apibulkadminai-access) | Set one person's use of the server's Anthropic key, for many at once | browser session or API key, admins |
| GET | [`/api/households/{household_id}/extractions/{extraction_id}`](#get-apihouseholdshousehold_idextractionsextraction_id) | A photo draft | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/extractions/{extraction_id}/accept`](#post-apihouseholdshousehold_idextractionsextraction_idaccept) | Save a photo draft's items | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/labels/{label_id}/extract`](#post-apihouseholdshousehold_idlabelslabel_idextract) | Read a photo of a container's contents | browser session or API key, anyone signed in |
| GET | [`/api/owner/ai`](#get-apiownerai) | The Anthropic API key, as far as it can be shown | browser session only, the owner |
| POST | [`/api/owner/ai`](#post-apiownerai) | Set or remove the Anthropic API key | browser session only, the owner |

### GET /api/account/ai

Your photo-reading settings

Whether you have your own Anthropic key (never the key itself), the model and effort you picked, whether you may use the server's key, and whose key your photos would go on now.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`my_ai_settings`](mcp.md#my_ai_settings) |
| Errors | 401, 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `app_effort` | string | yes |  |  |
| `app_model` | string | yes |  | The app's model, used with the server's key. |
| `daily_limit` | integer | yes |  | Photos one household may send a day with the server's key. |
| `effort` | string | yes |  |  |
| `enabled` | boolean | yes |  | ai.enabled in the settings. |
| `key_set` | boolean | yes |  | The server has a key of its own. |
| `last4` | string | yes |  | Your key's last four characters, to tell keys apart. |
| `model` | string | yes |  | The model your photos would be read with. |
| `models` | array of string | yes |  | The models you may pick, cheapest first. |
| `own_effort` | string | yes |  | The effort you picked for your own key; empty uses the app's. |
| `own_key_set` | boolean | yes |  | You have a key of your own set. |
| `own_model` | string | yes |  | The model you picked for your own key; empty uses the app's. |
| `ready` | boolean | yes |  | Your photos can be read: switched on, and there is a key you may use. |
| `set_at` | string (date-time) or null | yes |  |  |
| `shared_allowed` | boolean | yes |  | You may use the server's key when you have none of your own. |
| `using` | string | yes | one of `own`, `shared`, `none` | Whose key your photos would go on: your own, the server's, or none you may use. |

### POST /api/account/ai

Set your own key, model and effort

Store your own Anthropic API key, sealed and never shown again, or remove it; and pick the model and effort your own key uses. Only what you send changes.

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `api_key` | string or null | no | at most 300 characters | Your own Anthropic API key (sk-ant-...). Leave out to keep the one you have. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `effort` | string or null | no | one of ``, `low`, `medium`, `high` | The effort for your own key; empty uses the app's. |
| `model` | string or null | no | one of ``, `claude-haiku-5-5`, `claude-sonnet-5-5`, `claude-opus-5-5` | The model for your own key; empty uses the app's. |
| `remove_key` | boolean | no |  | Take your own key away. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `app_effort` | string | yes |  |  |
| `app_model` | string | yes |  | The app's model, used with the server's key. |
| `daily_limit` | integer | yes |  | Photos one household may send a day with the server's key. |
| `effort` | string | yes |  |  |
| `enabled` | boolean | yes |  | ai.enabled in the settings. |
| `key_set` | boolean | yes |  | The server has a key of its own. |
| `last4` | string | yes |  | Your key's last four characters, to tell keys apart. |
| `model` | string | yes |  | The model your photos would be read with. |
| `models` | array of string | yes |  | The models you may pick, cheapest first. |
| `own_effort` | string | yes |  | The effort you picked for your own key; empty uses the app's. |
| `own_key_set` | boolean | yes |  | You have a key of your own set. |
| `own_model` | string | yes |  | The model you picked for your own key; empty uses the app's. |
| `ready` | boolean | yes |  | Your photos can be read: switched on, and there is a key you may use. |
| `set_at` | string (date-time) or null | yes |  |  |
| `shared_allowed` | boolean | yes |  | You may use the server's key when you have none of your own. |
| `using` | string | yes | one of `own`, `shared`, `none` | Whose key your photos would go on: your own, the server's, or none you may use. |

### GET /api/admin/ai-access

Who may use the server's Anthropic key

The default for everyone, and each person's choice: follow the default, may use the server's key, or must bring their own. Someone with their own key always uses it.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_list_ai_access`](mcp.md#admin_list_ai_access) |
| Errors | 400, 401, 403, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `default_shared` | boolean | yes |  | Everyone without their own key may use the server's, unless their row says otherwise. |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `people` | array of [`AIAccess`](#aiaccess) | yes |  |  |
| `server_key_set` | boolean | yes |  |  |
| `total` | integer | yes |  |  |

### POST /api/admin/ai-access

Set the default for the server's Anthropic key

Whether people without their own key may use the server's by default.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_ai_default`](mcp.md#admin_set_ai_default) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `default_shared` | boolean | yes |  | true lets everyone without their own key use the server's; false makes everyone bring their own, unless their row says otherwise. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `default_shared` | boolean | yes |  | Everyone without their own key may use the server's, unless their row says otherwise. |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `people` | array of [`AIAccess`](#aiaccess) | yes |  |  |
| `server_key_set` | boolean | yes |  |  |
| `total` | integer | yes |  |  |

### POST /api/admin/ai-access/{user_id}

Set one person's use of the server's Anthropic key

Let one person use the server's key, make them bring their own, or have them follow the default.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_ai_access`](mcp.md#admin_set_ai_access) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `access` | string | yes | one of `default`, `shared`, `own` | default follows the default; shared may use the server's key; own must bring their own. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `default_shared` | boolean | yes |  | Everyone without their own key may use the server's, unless their row says otherwise. |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `people` | array of [`AIAccess`](#aiaccess) | yes |  |  |
| `server_key_set` | boolean | yes |  |  |
| `total` | integer | yes |  |  |

### GET /api/ai

Whether Smart Photo Extraction is ready for you

Whether your photos can be read, whose key they would go on (your own or the server's), and with which model.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`ai_status`](mcp.md#ai_status) |
| Errors | 401, 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `daily_limit` | integer | yes |  | Photos one household may send a day with the server's key. |
| `effort` | string | yes |  |  |
| `enabled` | boolean | yes |  | ai.enabled in the settings. |
| `key_set` | boolean | yes |  | The server has a key of its own. |
| `model` | string | yes |  | The model your photos would be read with. |
| `ready` | boolean | yes |  | Your photos can be read: switched on, and there is a key you may use. |
| `using` | string | yes | one of `own`, `shared`, `none` | Whose key your photos would go on: your own, the server's, or none you may use. |

### POST /api/bulk/admin/ai-access

Set one person's use of the server's Anthropic key, for many at once

Runs POST /api/admin/ai-access/{user_id} once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_ai_access_bulk`](mcp.md#admin_set_ai_access_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].access` | string | no | one of `default`, `shared`, `own` | default follows the default; shared may use the server's key; own must bring their own. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |
| `shared.access` | string | no | one of `default`, `shared`, `own` | default follows the default; shared may use the server's key; own must bring their own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### GET /api/households/{household_id}/extractions/{extraction_id}

A photo draft

What Claude found in a photo so far: status working, then ready (with items) or failed (with error).

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`get_extraction`](mcp.md#get_extraction) |
| Errors | 400, 401, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `extraction_id` | path | integer | yes | at least 1 | The draft's id, from extract_photo. |

**Reply 200**: Success

[`PhotoDraft`](#photodraft).

### POST /api/households/{household_id}/extractions/{extraction_id}/accept

Save a photo draft's items

Put the items you kept from a draft into its container, and mark the draft done.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`accept_extraction`](mcp.md#accept_extraction) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `extraction_id` | path | integer | yes | at least 1 | The draft's id, from extract_photo. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
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

**Reply 201**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of [`Item`](#item) | yes |  |  |

### POST /api/households/{household_id}/labels/{label_id}/extract

Read a photo of a container's contents

Ask Claude to list the items in an uploaded top-down photo, with your own API key if you set one, else the server's if an admin lets you. Answers at once with a draft whose status is working; ask get_extraction until it is ready, then accept_extraction with the items that are right.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`extract_photo`](mcp.md#extract_photo) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |
| `label_id` | path | integer | yes | at least 1 | The container's id, from list_labels, search_inventory or scan_code. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `file_id` | string | yes | at least 1 characters | The photo, already uploaded (files_start_upload): JPEG, PNG, WebP or GIF, taken from above. |

**Reply 202**: Success

[`PhotoDraft`](#photodraft).

### GET /api/owner/ai

The Anthropic API key, as far as it can be shown

Whether an API key is set, its last four characters and when it was set. The key itself is never shown again.

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 404, 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `daily_limit` | integer | yes |  | Photos one household may send a day with the server's key. |
| `effort` | string | yes |  |  |
| `enabled` | boolean | yes |  | ai.enabled in the settings. |
| `key_set` | boolean | yes |  | The server has a key of its own. |
| `last4` | string | yes |  | The key's last four characters, to tell keys apart. |
| `model` | string | yes |  | The model your photos would be read with. |
| `ready` | boolean | yes |  | Your photos can be read: switched on, and there is a key you may use. |
| `set_at` | string (date-time) or null | yes |  |  |
| `using` | string | yes | one of `own`, `shared`, `none` | Whose key your photos would go on: your own, the server's, or none you may use. |

### POST /api/owner/ai

Set or remove the Anthropic API key

Store the API key Smart Photo Extraction uses, sealed in the database, or remove it with an empty value. It is never shown again.

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `api_key` | string | yes | at most 300 characters | An Anthropic API key (sk-ant-...), or empty to remove it. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `daily_limit` | integer | yes |  | Photos one household may send a day with the server's key. |
| `effort` | string | yes |  |  |
| `enabled` | boolean | yes |  | ai.enabled in the settings. |
| `key_set` | boolean | yes |  | The server has a key of its own. |
| `last4` | string | yes |  | The key's last four characters, to tell keys apart. |
| `model` | string | yes |  | The model your photos would be read with. |
| `ready` | boolean | yes |  | Your photos can be read: switched on, and there is a key you may use. |
| `set_at` | string (date-time) or null | yes |  |  |
| `using` | string | yes | one of `own`, `shared`, `none` | Whose key your photos would go on: your own, the server's, or none you may use. |

## Reminders

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/account/calendar-feed`](#get-apiaccountcalendar-feed) | Whether your calendar feed is on | browser session or API key, anyone signed in |
| POST | [`/api/account/calendar-feed`](#post-apiaccountcalendar-feed) | Make a calendar feed address | browser session only, anyone signed in |
| POST | [`/api/account/calendar-feed/off`](#post-apiaccountcalendar-feedoff) | Switch your calendar feed off | browser session only, anyone signed in |
| GET | [`/api/feed/{token}`](#get-apifeedtoken) | A calendar feed | anyone |
| GET | [`/api/notifications`](#get-apinotifications) | Your reminders | browser session or API key, anyone signed in |
| POST | [`/api/notifications/read`](#post-apinotificationsread) | Mark reminders read | browser session or API key, anyone signed in |

### GET /api/account/calendar-feed

Whether your calendar feed is on

Whether you have a calendar feed address switched on.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`calendar_feed_status`](mcp.md#calendar_feed_status) |
| Errors | 401, 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `on` | boolean | yes |  |  |
| `path` | string | yes |  | The feed's path, only right after it is made: it is not stored and cannot be shown again. A credential: treat it like a password. |

### POST /api/account/calendar-feed

Make a calendar feed address

Make a new address for a calendar app to subscribe to, listing expiry dates and loan due dates. Any address you had before stops working.

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `on` | boolean | yes |  |  |
| `path` | string | yes |  | The feed's path, only right after it is made: it is not stored and cannot be shown again. A credential: treat it like a password. |

### POST /api/account/calendar-feed/off

Switch your calendar feed off

Stop your calendar feed address working.

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `on` | boolean | yes |  |  |
| `path` | string | yes |  | The feed's path, only right after it is made: it is not stored and cannot be shown again. A credential: treat it like a password. |

### GET /api/feed/{token}

A calendar feed

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 400, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `token` | path | string | yes |  | The feed's credential, ending .ics. |

**Reply 200**: An iCalendar file: one all-day event per expiry date and loan due date

Text, as `text/calendar`.

### GET /api/notifications

Your reminders

Reminders about items in your households, newest first: expiring or expired, run low, or overdue back from a loan.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`list_notifications`](mcp.md#list_notifications) |
| Errors | 400, 401, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `unread` | query | boolean | no |  | Only the ones not read yet. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `notifications` | array of [`Notification`](#notification) | yes |  |  |
| `total` | integer | yes |  |  |
| `unread` | integer | yes |  | How many are not read yet, in all. |

### POST /api/notifications/read

Mark reminders read

Mark one reminder read, or all of them.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`mark_notifications_read`](mcp.md#mark_notifications_read) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `notification_id` | integer | no |  | The one to mark read; leave out to mark them all. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `marked` | integer | yes |  |  |

## Reports

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/households/{household_id}/export.csv`](#get-apihouseholdshousehold_idexportcsv) | Download the inventory as a spreadsheet | browser session or API key, anyone signed in |
| POST | [`/api/households/{household_id}/import`](#post-apihouseholdshousehold_idimport) | Add items from a spreadsheet | browser session or API key, anyone signed in |

### GET /api/households/{household_id}/export.csv

Download the inventory as a spreadsheet

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | None, HTTP only: A CSV file is not JSON: GET this address with the same Authorization header. |
| Errors | 400, 401, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

**Reply 200**: Every item not in Trash, one row each, as CSV

Text, as `text/csv`.

### POST /api/households/{household_id}/import

Add items from a spreadsheet

Add every row of an uploaded CSV as an item. A header row names the columns (Item, Quantity, Container, Location, Tags, Expires, Value ... as in an export, and the usual names other apps use). Containers and locations it does not know are made; nothing is changed or removed.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`import_csv`](mcp.md#import_csv) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `household_id` | path | integer | yes | at least 1 | The household's id, from GET /api/households (list_households over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `file_id` | string | yes | at least 1 characters | A CSV file you uploaded (files_start_upload), with a header row. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `containers` | integer | yes |  | Containers made for them. |
| `items` | integer | yes |  | Items added. |
| `locations` | integer | yes |  | Locations made for them. |
| `skipped` | array of string | yes |  | Rows left out, and why. |

## Scan

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/scan`](#get-apiscan) | Look up a scanned code | browser session or API key, anyone signed in |
| GET | [`/api/search`](#get-apisearch) | Find where something is | browser session or API key, anyone signed in |

### GET /api/scan

Look up a scanned code

What a scanned or typed code is: one of your containers (with where it is), one of your household's printed codes still waiting to be stuck on something (set it up with create_label and this code), or unknown -- which may be a store-bought sticker or barcode you can claim for a container with claim_code.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`scan_code`](mcp.md#scan_code) |
| Errors | 400, 401, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `code` | query | string | no | 1 to 2000 characters | What the camera read: a printed code (XXXXX-XXXXX, with or without the dash), a link ending /l/<code>, any other QR text, or a barcode's digits. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `can_edit` | boolean | yes |  | Whether you may set it up or change it. |
| `code` | string | yes |  | The code as printed, for one of ours; what was scanned, otherwise. |
| `household` | string | yes |  | Its name; empty when unknown. |
| `household_id` | integer | yes |  | 0 when unknown. |
| `in_trash` | boolean | yes |  | The container is in Trash. |
| `kind` | string | yes | one of `label`, `unassigned`, `unknown` |  |
| `label` | [`Container`](#container) or null | yes |  | The container, when kind is label. |

### GET /api/search

Find where something is

Search every household you are in for items and containers, and say where each one is. The question "where is my drill?" is search_inventory with q=drill.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`search_inventory`](mcp.md#search_inventory) |
| Errors | 400, 401, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no | 1 to 200 characters | What you are looking for, such as "drill" or "christmas lights". A typo is forgiven. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `hits` | array of object | yes |  |  |
| `hits[].household` | string | yes |  |  |
| `hits[].household_id` | integer | yes |  |  |
| `hits[].item` | [`Item`](#item) or null | yes |  | When kind is item. |
| `hits[].kind` | string | yes | one of `item`, `label` |  |
| `hits[].label` | [`Container`](#container) or null | yes |  | When kind is label. |
| `hits[].where` | string | yes |  | The whole answer in words: household › location › containers. |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |

## Schemas

The shapes the routes above refer to by name.

### AIAccess

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `access` | string | yes | one of `default`, `shared`, `own` | default follows default_shared; shared may use the server's key; own must bring their own. |
| `display_name` | string | yes |  |  |
| `may_share` | boolean | yes |  | What it comes to: they may use the server's key. |
| `own_key_set` | boolean | yes |  | They have a key of their own, which they always use first. |
| `owner` | boolean | yes |  | The owner, who may always use the server's key. |
| `user_id` | integer | yes |  |  |

### APIKey

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_at` | string (date-time) | yes |  |  |
| `created_by` | integer | no |  | The admin who made this key for its owner. Absent if the owner made it. |
| `created_by_name` | string | no |  |  |
| `display_name` | string | no |  | The key's owner. On the admin list only. |
| `expired` | boolean | yes |  |  |
| `expires_at` | string (date-time) | no |  |  |
| `id` | string | yes |  |  |
| `last_used_at` | string (date-time) | no |  |  |
| `live` | boolean | yes |  |  |
| `may_sign_in` | boolean | yes |  | Whether this key may make a login link for its own account. |
| `name` | string | yes |  |  |
| `revoked` | boolean | yes |  |  |
| `revoked_at` | string (date-time) | no |  |  |
| `role` | string | yes | one of `admin`, `user` |  |
| `user_id` | integer | yes |  |  |

### BulkReply

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `done` | integer | yes |  | How many items worked. |
| `failed` | integer | yes |  | How many did not. |
| `results` | array of [`BulkResult`](#bulkresult) | yes |  | One per item, in the order they were sent. |

### BulkResult

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `error` | string | no |  | When it failed: the route's words for why. |
| `index` | integer | yes |  | The item's place in items, from 0. |
| `ok` | boolean | yes |  |  |
| `reply` | object | no |  | When it worked: the route's own reply, if it has one. |
| `request_id` | string | no |  | On a 5xx: the reference that finds this failure in the log. |
| `status` | integer | no |  | When it failed: the status the route would have answered. |

### Container

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `child_count` | integer | yes |  | Containers sitting inside this one. |
| `claimed` | array of string | yes |  |  |
| `code` | string | yes |  |  |
| `color` | string | no |  |  |
| `color_means` | string | yes |  | What its color stands for in this household. |
| `container_type` | string | no |  |  |
| `created_at` | string (date-time) | yes |  |  |
| `created_by` | integer | yes |  |  |
| `deleted_at` | string (date-time) or null | no |  |  |
| `deleted_by` | integer | no |  |  |
| `destination` | string | no |  |  |
| `flags` | array of string | yes |  |  |
| `household_id` | integer | yes |  |  |
| `id` | integer | yes |  |  |
| `item_count` | integer | yes |  |  |
| `location_id` | integer | no |  |  |
| `move_status` | string | no |  |  |
| `name` | string | yes |  |  |
| `notes` | string | no |  |  |
| `parent_id` | integer | no |  |  |
| `path` | string | yes |  | Where it is: location › outer containers › this one. |
| `photos` | array of string | yes |  |  |
| `printed_code` | string | yes |  | The code as printed under its QR, XXXXX-XXXXX. |
| `updated_at` | string (date-time) | yes |  |  |
| `value` | number | yes |  | What its items are worth together. |

### Error

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `error` | string | yes |  |  |
| `request_id` | string | no |  | On a 5xx: the reference that finds this failure in the log. |

### File

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `chunks` | integer | yes |  |  |
| `created_at` | string (date-time) | yes |  |  |
| `id` | string | yes |  |  |
| `name` | string | yes |  |  |
| `received` | integer | yes |  |  |
| `size` | integer | yes |  |  |
| `state` | string | yes | one of `uploading`, `ready` |  |
| `type` | string | yes |  |  |

### Household

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `colors` | map of string | yes |  | Label color (#rrggbb) to what it means here. |
| `created_at` | string (date-time) | yes |  |  |
| `fields` | array of object | yes |  | Custom fields every item here can fill in. |
| `fields[].id` | string | yes |  |  |
| `fields[].name` | string | yes |  |  |
| `fields[].options` | array of string | no |  |  |
| `fields[].type` | string | yes |  |  |
| `id` | integer | yes |  |  |
| `members` | array of object | yes |  |  |
| `members[].display_name` | string | yes |  |  |
| `members[].role` | string | yes | one of `owner`, `editor`, `viewer` |  |
| `members[].user_id` | integer | yes |  |  |
| `my_role` | string | yes | one of `owner`, `editor`, `viewer` |  |
| `name` | string | yes |  |  |
| `reminder_days` | integer or null | yes |  | Days before an expiry its reminder goes out; null uses the app's default. |

### IPNames

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `names` | array of object | yes |  |  |
| `names[].created_at` | string (date-time) | yes |  |  |
| `names[].name` | string | yes |  |  |
| `names[].target` | string | yes |  |  |

### InviteLink

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_by` | integer | yes |  |  |
| `created_by_name` | string | no |  |  |
| `expires_at` | string (date-time) | yes |  |  |
| `id` | string | yes |  |  |
| `label` | string | yes |  |  |
| `live` | boolean | yes |  |  |
| `max_uses` | integer | yes |  |  |
| `remaining` | integer | yes |  |  |
| `revoked` | boolean | yes |  |  |
| `used` | integer | yes |  |  |

### Item

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `barcode` | string | no |  |  |
| `brand` | string | no |  |  |
| `created_at` | string (date-time) | yes |  |  |
| `created_by` | integer | yes |  |  |
| `deleted_at` | string (date-time) or null | no |  |  |
| `deleted_by` | integer | no |  |  |
| `description` | string | no |  |  |
| `expires` | string | no |  |  |
| `expires_in` | integer or null | yes |  | Days until it expires (negative once it has); null without a date. |
| `fields` | map of string | yes |  |  |
| `history` | array of object | yes |  |  |
| `history[].at` | string (date-time) | yes |  |  |
| `history[].detail` | string | no |  |  |
| `history[].kind` | string | yes |  |  |
| `history[].user_id` | integer | yes |  |  |
| `household_id` | integer | yes |  |  |
| `id` | integer | yes |  |  |
| `label_code` | string | yes |  | Its container's printed code. |
| `label_color` | string | yes |  |  |
| `label_id` | integer | yes |  |  |
| `label_name` | string | yes |  |  |
| `lent_due` | string | no |  |  |
| `lent_to` | string | no |  |  |
| `low_stock` | number or null | no |  |  |
| `low_stock_now` | boolean | yes |  | At or below its low-stock level. |
| `name` | string | yes |  |  |
| `overdue` | boolean | yes |  | Lent out and past its due date. |
| `path` | string | yes |  | Where it is: location › containers. |
| `photos` | array of string | yes |  |  |
| `purchase_date` | string | no |  |  |
| `quantity` | number | yes |  |  |
| `receipts` | array of string | yes |  |  |
| `remind_days` | integer or null | no |  |  |
| `serial` | string | no |  |  |
| `tags` | array of string | yes |  |  |
| `unit` | string | no |  |  |
| `updated_at` | string (date-time) | yes |  |  |
| `value` | number | no |  |  |

### Job

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `interrupted` | boolean | yes |  |  |
| `interval_seconds` | integer | yes |  |  |
| `last_error` | string | no |  |  |
| `last_finished` | string (date-time) | no |  |  |
| `last_note` | string | no |  |  |
| `last_ok` | boolean or null | no |  |  |
| `last_started` | string (date-time) | no |  |  |
| `name` | string | yes |  |  |
| `next_due` | string (date-time) | yes |  |  |
| `running` | boolean | yes |  |  |
| `runs` | integer | yes |  |  |

### Joined

One of these:

1. An object:

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `display_name` | string | yes |  |  |
| `ok` | boolean | yes |  |  |

2. An object:

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `challenge` | string | yes |  |  |
| `two_factor` | boolean | yes |  |  |

### Location

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_at` | string (date-time) | yes |  |  |
| `deleted_at` | string (date-time) or null | no |  |  |
| `deleted_by` | integer | no |  |  |
| `household_id` | integer | yes |  |  |
| `id` | integer | yes |  |  |
| `label_count` | integer | yes |  | Containers directly here. |
| `name` | string | yes |  |  |
| `parent_id` | integer | no |  |  |
| `path` | string | yes |  |  |

### LogDay

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `date` | string | yes |  |  |
| `entries` | array of object | yes |  |  |
| `next` | integer or null | yes |  |  |

### LogDays

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `days` | array of object | yes |  |  |
| `dropped` | integer | yes |  |  |
| `settings` | object | yes |  |  |
| `total_bytes` | integer | yes |  |  |

### LoginLink

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_at` | string (date-time) | yes |  |  |
| `expires_at` | string (date-time) or null | yes |  | null when it never expires. |
| `last_used_at` | string (date-time) or null | yes |  |  |
| `max_uses` | integer | yes |  | Sign-ins it allows. 0 when unlimited. |
| `unlimited` | boolean | yes |  |  |
| `uses` | integer | yes |  | Sign-ins so far. |

### Maintenance

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `error` | string | yes |  |  |
| `maintenance` | string | yes | one of `read_only`, `full` |  |

### Me

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `account_role` | string | yes |  |  |
| `announcement` | string | yes |  | server.announcement. Empty when there is none. |
| `api_doc_url` | string | yes |  | Where to read the API reference as a page for people: config/api.md in app.repo, or /docs/api when there is no repo. |
| `api_key_id` | string | no |  | API keys only. |
| `api_key_name` | string | no |  | API keys only. |
| `api_keys_enabled` | boolean | yes |  |  |
| `auth` | string | yes |  |  |
| `authenticated` | boolean | yes |  |  |
| `csrf_token` | string | no |  | Browser sessions only. |
| `description` | string | yes |  | app.description. Empty when there is none. |
| `display_name` | string | yes |  |  |
| `files_enabled` | boolean | yes |  | files.enabled: whether the Files screen and routes are on. |
| `has_password` | boolean | yes |  |  |
| `has_pin` | boolean | yes |  |  |
| `id` | integer | yes |  |  |
| `invite_links_enabled` | boolean | yes |  | Whether invite links are on. |
| `is_owner` | boolean | yes |  |  |
| `login_link` | [`LoginLink`](#loginlink) or null | yes |  | Your own live login link, or null. |
| `logs_enabled` | boolean | yes |  |  |
| `logs_ui_clicks` | boolean | yes |  |  |
| `maintenance` | string | yes | one of `none`, `read_only`, `full` |  |
| `mcp_doc_url` | string | yes |  | Where to read the MCP tools as a page for people: config/mcp.md in app.repo, or /docs/mcp when there is no repo. |
| `mcp_enabled` | boolean | yes |  | server.mcp: whether POST /mcp answers. |
| `mcp_url` | string | yes |  | Where to read the MCP reference: config/mcp.json in app.repo, or /api/mcp.json when there is no repo. Only with server.mcp on. |
| `openapi_url` | string | yes |  | Where to read this API's reference: config/api.json in app.repo, or /api/api.json when there is no repo. |
| `password_blocked` | boolean | yes |  |  |
| `passwords_enabled` | boolean | yes |  |  |
| `repo` | string | yes |  | app.repo without a trailing / or .git: where the source lives. Empty when it is not set. |
| `role` | string | yes |  | The effective role: never higher than an API key's own. |
| `shortcuts` | map of string | yes |  | The keyboard shortcuts you changed, by action. Mod is Cmd on a Mac and Ctrl elsewhere. |
| `show_about` | boolean | yes |  | app.show_about: whether the menu has About. |
| `theme_id` | string | yes |  | The theme you use. Empty: the first. |
| `themes` | array of [`Theme`](#theme) or null | yes |  | Your themes (see the Theme schema): your own list, else the app's defaults, else null for the built-in ones. |
| `themes_builtins` | integer | yes |  | Which built-in themes that list has seen (see builtins on POST /api/account/themes). |
| `themes_custom` | boolean | yes |  | Whether you have your own list rather than following the defaults. |
| `two_factor` | boolean | yes |  |  |
| `two_factor_enabled` | boolean | yes |  |  |
| `username` | string | yes |  |  |
| `version` | string | yes |  | The running build. dev when none was stamped. |
| `webhooks_enabled` | boolean | yes |  | webhooks.enabled: whether webhooks are sent. |

### Member

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `display_name` | string | yes |  |  |
| `id` | integer | yes |  |  |

### NewAPIKey

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_at` | string (date-time) | yes |  |  |
| `created_by` | integer | no |  | The admin who made this key for its owner. Absent if the owner made it. |
| `created_by_name` | string | no |  |  |
| `credential` | string | yes |  | Shown in this response only, and never recoverable. |
| `display_name` | string | no |  | The key's owner. On the admin list only. |
| `expired` | boolean | yes |  |  |
| `expires_at` | string (date-time) | no |  |  |
| `id` | string | yes |  |  |
| `last_used_at` | string (date-time) | no |  |  |
| `live` | boolean | yes |  |  |
| `may_sign_in` | boolean | yes |  | Whether this key may make a login link for its own account. |
| `name` | string | yes |  |  |
| `revoked` | boolean | yes |  |  |
| `revoked_at` | string (date-time) | no |  |  |
| `role` | string | yes | one of `admin`, `user` |  |
| `user_id` | integer | yes |  |  |

### Notification

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_at` | string (date-time) | yes |  |  |
| `due` | string | yes |  |  |
| `household_id` | integer | yes |  |  |
| `id` | integer | yes |  |  |
| `item_id` | integer | yes |  |  |
| `item_name` | string | yes |  |  |
| `kind` | string | yes | one of `expiring`, `expired`, `low_stock`, `overdue` |  |
| `label_id` | integer | yes |  |  |
| `read` | boolean | yes |  |  |

### PhotoDraft

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_at` | string (date-time) | yes |  |  |
| `error` | string | yes |  |  |
| `file_id` | string | yes |  |  |
| `id` | integer | yes |  |  |
| `items` | array of object | yes |  | What was found. box is where, as fractions of the photo: x, y, width, height. |
| `items[].box` | array of number | no |  |  |
| `items[].brand` | string | no |  |  |
| `items[].category` | string | no |  |  |
| `items[].description` | string | no |  |  |
| `items[].name` | string | yes |  |  |
| `items[].quantity` | number | yes |  |  |
| `items[].tags` | array of string | no |  |  |
| `label_id` | integer | yes |  |  |
| `status` | string | yes | one of `working`, `ready`, `accepted`, `failed` |  |
| `summary` | string | yes |  | A name for the whole lot, to suggest for the container. |

### PrintRun

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `codes` | array of string | yes |  | Each code as printed, XXXXX-XXXXX. |
| `created_at` | string (date-time) | yes |  |  |
| `id` | integer | yes |  |  |
| `layout` | string | yes |  |  |
| `names` | array of string | yes |  |  |
| `palette` | string | yes |  |  |
| `used` | integer | yes |  | How many are on containers by now. |

### PruneResult

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `bytes_after` | integer | yes |  |  |
| `bytes_before` | integer | yes |  |  |
| `compressed` | integer | yes |  |  |
| `deleted` | integer | yes |  |  |

### Revoked

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `api_keys` | integer | yes |  |  |
| `login_links` | integer | yes |  |  |
| `passwords` | integer | yes |  |  |
| `sessions` | integer | yes |  |  |

### Session

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_at` | string (date-time) | yes |  |  |
| `current` | boolean | yes |  |  |
| `expires_at` | string (date-time) | yes |  |  |
| `id` | string | yes |  |  |
| `ip` | string | no |  |  |
| `ip_name` | string | no |  | What the owner calls this address on the Logs screen. The owner's own list only. |
| `label` | string | no |  | The name you gave this device. |
| `last_seen_at` | string (date-time) | no |  |  |
| `user_agent` | string | no |  |  |

### SettingsResult

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `applied` | array of string or null | yes |  |  |
| `changes` | array of object or null | yes |  |  |
| `restart` | array of string or null | yes |  |  |
| `scripts` | array of string or null | yes |  |  |
| `unchanged` | array of string or null | yes |  |  |
| `version` | string | yes |  |  |

### SettingsSnapshot

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `error` | object or null | yes |  |  |
| `error.at` | string (date-time) | yes |  |  |
| `error.message` | string | yes |  |  |
| `error.origin` | string | yes |  |  |
| `file` | string | yes |  |  |
| `in_container` | boolean | yes |  |  |
| `secret_file` | string | yes |  |  |
| `sections` | array of object | yes |  |  |
| `sections[].help` | string | yes |  |  |
| `sections[].keys` | array of object | yes |  |  |
| `sections[].keys[].apply` | string | yes |  |  |
| `sections[].keys[].default` | any | yes |  |  |
| `sections[].keys[].file_only` | boolean | yes |  | A path: change it in config.yaml, never here. |
| `sections[].keys[].help` | string | yes |  |  |
| `sections[].keys[].in_file` | boolean | yes |  |  |
| `sections[].keys[].key` | string | yes |  |  |
| `sections[].keys[].locked` | boolean | yes |  |  |
| `sections[].keys[].name` | string | yes |  |  |
| `sections[].keys[].pending_restart` | boolean | yes |  |  |
| `sections[].keys[].running` | any | yes |  |  |
| `sections[].keys[].type` | string | yes |  |  |
| `sections[].keys[].value` | any | yes |  |  |
| `sections[].name` | string | yes |  |  |
| `stale` | boolean | yes |  |  |
| `version` | string | yes |  | Send it back when saving: a changed file is refused with a 409. |
| `writable` | boolean | yes |  |  |

### SignIn

One of these:

1. An object:

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

2. An object:

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `challenge` | string | yes |  |  |
| `two_factor` | boolean | yes |  |  |

### Theme

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `accent` | string | yes |  | The main color: what you can press. |
| `background` | string | no |  | The page. Left out: the mode's own. |
| `button` | object or null | no |  | How the theme's buttons look. A choice left out is the default (rounded, normal, raised, normal). |
| `button.case` | string | no | one of `normal`, `upper` |  |
| `button.depth` | string | no | one of `flat`, `raised`, `glossy` |  |
| `button.shape` | string | no | one of `square`, `soft`, `rounded`, `pill` |  |
| `button.weight` | string | no | one of `normal`, `medium`, `bold` |  |
| `dark` | object or null | no |  | A light theme's colors for dark mode, kept while it is light. Switching the theme to dark swaps them in. A light theme's light (or a dark theme's dark) is dropped. |
| `dark.accent` | string | yes |  | The main color: what you can press. |
| `dark.background` | string | no |  | The page. Left out: the mode's own. |
| `dark.overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `dark.second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `dark.surface` | string | no |  | Cards. Left out: the mode's own. |
| `id` | string | yes |  | Lower-case letters, digits, - and _, up to 32. |
| `light` | object or null | no |  | A dark theme's colors for light mode, kept while it is dark. Switching the theme to light swaps them in. |
| `light.accent` | string | yes |  | The main color: what you can press. |
| `light.background` | string | no |  | The page. Left out: the mode's own. |
| `light.overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `light.second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `light.surface` | string | no |  | Cards. Left out: the mode's own. |
| `mode` | string | yes | one of `light`, `dark` |  |
| `name` | string | yes |  | 1 to 60 characters. |
| `overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `surface` | string | no |  | Cards. Left out: the mode's own. |

### User

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_at` | string (date-time) | yes |  |  |
| `disabled` | boolean | yes |  |  |
| `display_name` | string | yes |  |  |
| `has_password` | boolean | yes |  |  |
| `id` | integer | yes |  |  |
| `login_link` | [`LoginLink`](#loginlink) or null | yes |  | The account's live login link, or null. |
| `owner` | boolean | yes |  |  |
| `password_blocked` | boolean | yes |  |  |
| `role` | string | yes | one of `admin`, `user` |  |
| `two_factor` | boolean | yes |  |  |
| `username` | string | yes |  | Empty until a password is set. |

### Webhook

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_at` | string (date-time) | yes |  |  |
| `enabled` | boolean | yes |  |  |
| `events` | array of string | yes |  |  |
| `id` | string | yes |  |  |
| `name` | string | yes |  |  |
| `url` | string | yes |  |  |

### WebhookDelivery

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `attempts` | integer | yes |  |  |
| `created_at` | string (date-time) | yes |  |  |
| `event` | string | yes |  |  |
| `finished_at` | string (date-time) | no |  |  |
| `id` | string | yes |  |  |
| `last_error` | string | no |  |  |
| `last_status` | integer | no |  |  |
| `next_attempt_at` | string (date-time) | yes |  |  |
| `status` | string | yes |  |  |
