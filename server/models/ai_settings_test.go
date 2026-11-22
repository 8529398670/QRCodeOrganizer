package models

import (
	testing "testing"

	catalog "qrcodeorganizer/server/ai/catalog"
)

func strp( value string ) ( pointer *string ) { pointer = &value; return }

const testSecret = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestChangingServiceNeedsAKeyForIt( t *testing.T ) {
	store := testStore( t )
	if _ , err := SetUserAI( store , testSecret , 7 , AIChange{ APIKey: strp( "sk-ant-api03-abcdefghijklmnop" ) , Model: strp( "claude-opus-5-5" ) , Effort: strp( "high" ) } ) ; err != nil { t.Fatal( err ) }
	if _ , err := SetUserAI( store , testSecret , 7 , AIChange{ Provider: strp( catalog.OpenAI ) } ) ; err == nil { t.Fatal( "switched service without a key" ) }
	settings , err := SetUserAI( store , testSecret , 7 , AIChange{ Provider: strp( catalog.OpenAI ) , APIKey: strp( "sk-proj-abcdefghijklmnopqrst" ) } )
	if err != nil { t.Fatal( err ) }
	if settings.Service() != catalog.OpenAI || settings.Model != "" || settings.Effort != "" || settings.Last4 != "qrst" { t.Fatalf( "%+v" , settings ) }
	key , err := UserAPIKey( store , testSecret , 7 )
	if err != nil || key != "sk-proj-abcdefghijklmnopqrst" { t.Fatalf( "key %q %v" , key , err ) }
	if _ , err = SetUserAI( store , testSecret , 7 , AIChange{ Effort: strp( "max" ) } ) ; err == nil { t.Fatal( "OpenAI took effort max" ) }
}

func TestAnOldRecordIsAnthropic( t *testing.T ) {
	store := testStore( t )
	if err := store.Put( "app_secrets" , []byte( "ai" ) , map[string]any{ "anthropic_key": "sealed" , "last4" : "abcd" } ) ; err != nil { t.Fatal( err ) }
	server := ServerAI( store )
	if server.KeySet() == false || server.Service() != catalog.Anthropic { t.Fatalf( "%+v" , server ) }
}
