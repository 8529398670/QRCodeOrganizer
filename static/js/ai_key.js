// The Anthropic API key card on Settings, for the owner: Smart Photo
// Extraction sends photos to Claude with it. The key goes to the server once
// and is never shown again; the card only learns whether one is set.
const AiKey = {
  init( me ) {
    if ( !me.is_owner ) return;
    Dom.show( Dom.get( "ai-key-card" ) , true );
    Dom.onSubmit( Dom.get( "ai-key-form" ) , async ( event ) => {
      event.preventDefault();
      await this.save( Dom.get( "ai-key-input" ).value.trim() );
    } );
    Dom.onPress( Dom.get( "ai-key-remove" ) , async () => {
      if ( !await Dom.confirm( I18n.get( "ai.remove_confirm" ) ) ) return;
      await this.save( "" );
    } );
    this.refresh();
  },

  async refresh() {
    try {
      this.paint( await Api.request( "/api/owner/ai" ) );
    } catch ( error ) { /* the rest of Settings works without this */ }
  },

  paint( state ) {
    let text = state.key_set ? I18n.get( "ai.state_set" ).replace( "{last4}" , state.last4 ) : I18n.get( "ai.state_unset" );
    if ( state.key_set && !state.enabled ) text += " " + I18n.get( "ai.state_off" );
    Dom.text( Dom.get( "ai-key-state" ) , text );
    Dom.show( Dom.get( "ai-key-remove" ) , state.key_set );
  },

  // Not undoable: a key is a credential, and putting an old one back would
  // need it typed again.
  async save( key ) {
    try {
      this.paint( await Api.post( "/api/owner/ai" , { api_key: key } ) );
      Dom.get( "ai-key-input" ).value = "";
      Undo.saved( Dom.get( "ai-key-form" ) );
      Dom.toast( I18n.get( key ? "ai.saved" : "ai.removed" ) , "ok" );
    } catch ( error ) { Org.failed( error ); }
  },
};
