package routes

import (
	bytes "bytes"
	errors "errors"
	io "io"

	fiber "github.com/gofiber/fiber/v3"

	models "qrcodeorganizer/server/models"
	security "qrcodeorganizer/server/security"
)

// Photos and receipts are ordinary uploads (routes/files.go): the page sends
// one up, then names its file id in a container's or item's photos. That
// attaches the file to the household, and from then on every member may see
// it through the route below -- while the Files screen still lists it as its
// uploader's own, and nobody else's.

// maxInlinePhoto is the largest photo served to an <img>: the page shrinks
// photos well under this before sending them.
const maxInlinePhoto = 20 << 20

// checkPhotos makes sure every file id a request names is a finished file the
// caller may attach here, and attaches the ones that are not yet.
func ( handlers *Handlers ) checkPhotos( c fiber.Ctx , household *models.Household , ids *[]string ) ( err error ) {
	if ids == nil { return }
	user_id := security.UserFrom( c ).ID
	for _ , file_id := range *ids {
		file , get_err := models.GetFile( handlers.Store , file_id )
		usable := get_err == nil && file.Deleted() == false && file.State == models.FileReady
		usable = usable && ( file.HouseholdID == household.ID || ( file.HouseholdID == 0 && file.UserID == user_id ) )
		if usable == false {
			err = refuse( fiber.StatusBadRequest , "no such photo: upload it first, then name its file id" )
			return
		}
		if file.HouseholdID == 0 {
			if err = models.AttachFile( handlers.Store , file_id , household.ID ); err != nil { return }
		}
	}
	return
}

type householdPhotoPath struct {
	householdPath
	FileID string `json:"file_id" path:"file_id" jsonschema_description:"A file id from a container's or item's photos or receipts."`
}

// imageType is what the bytes are, by their first few: a photo is served as
// itself only when it really is one, so an upload cannot pass for a page.
func imageType( head []byte ) ( kind string ) {
	switch {
	case bytes.HasPrefix( head , []byte{ 0xFF , 0xD8 , 0xFF } ):
		kind = "image/jpeg"
	case bytes.HasPrefix( head , []byte( "\x89PNG\r\n\x1a\n" ) ):
		kind = "image/png"
	case len( head ) >= 12 && bytes.Equal( head[ :4 ] , []byte( "RIFF" ) ) && bytes.Equal( head[ 8:12 ] , []byte( "WEBP" ) ):
		kind = "image/webp"
	case bytes.HasPrefix( head , []byte( "GIF8" ) ):
		kind = "image/gif"
	}
	return
}

// @route("GET /api/households/{household_id}/photos/{file_id}", summary="A household photo or receipt",
//     tag="Households", errors=(404,), reply_type="image/jpeg",
//     reply_description="The photo, as itself when it is a JPEG, PNG, WebP or GIF, else as an attachment",
//     mcp_http="Photo bytes are not JSON: GET this address with the same Authorization header.",
// )
func ( handlers *Handlers ) householdPhoto( c fiber.Ctx , in *householdPhotoPath ) ( err error ) {
	household , _ , err := handlers.household( c , in.HouseholdID , needRead )
	if err != nil { return }
	file , get_err := models.GetFile( handlers.Store , in.FileID )
	if get_err != nil || file.Deleted() || file.State != models.FileReady || file.HouseholdID != household.ID {
		err = notFound( c , "no such photo" )
		return
	}
	if file.Size > maxInlinePhoto {
		err = handlers.sendFile( c , file )
		return
	}
	reader , open_err := models.OpenFile( handlers.cfg() , file )
	if errors.Is( open_err , models.ErrFileMissing ) {
		err = notFound( c , "no such photo" )
		return
	}
	if open_err != nil {
		err = serverError( c , open_err )
		return
	}
	defer reader.Close()
	data , read_err := io.ReadAll( reader )
	if read_err != nil {
		err = serverError( c , read_err )
		return
	}
	kind := imageType( data )
	if kind == "" {
		c.Set( fiber.HeaderContentType , fiber.MIMEOctetStream )
		c.Set( fiber.HeaderContentDisposition , contentDisposition( file.Name ) )
	} else {
		c.Set( fiber.HeaderContentType , kind )
	}
	// A file id never names other bytes, so the browser may keep it a while.
	c.Set( fiber.HeaderCacheControl , "private, max-age=86400" )
	err = c.Send( data )
	return
}
