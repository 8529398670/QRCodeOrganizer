// The server's AI key card on Settings, for the owner: Smart Photo
// Extraction sends photos with it, to the service it belongs to (see
// ai_form.js). The key goes to the server once and is never shown again; the
// card only learns whether one is set.
const AiKey = {
  form: null,

  init( me ) {
    if ( !me.is_owner ) return;
    Dom.show( Dom.get( "ai-key-card" ) , true );
    this.form = new AiForm( "ai-key" , "/api/owner/ai" , ( state ) => this.paint( state ) );
    Dom.onSubmit( Dom.get( "ai-key-form" ) , async ( event ) => {
      event.preventDefault();
      const body = this.form.body();
      await this.save( body , body.api_key ? "ai.saved" : "ai.choices_saved" );
    } );
    Dom.onPress( Dom.get( "ai-key-remove" ) , async () => {
      if ( !await Dom.confirm( I18n.get( "ai.remove_confirm" ) ) ) return;
      await this.save( { remove_key: true } , "ai.removed" );
    } );
    this.refresh();
  },

  async refresh() {
    try {
      const state = await Api.request( "/api/owner/ai" );
      this.form.paint( state , state.provider , state.chosen_model , state.chosen_effort );
      this.paint( state );
    } catch ( error ) { /* the rest of Settings works without this */ }
    this.paintScope();
  },

  // Whose photos this key reads; the owner's own Account key, when set,
  // wins for their photos, so say so here where they might expect otherwise.
  async paintScope() {
    let text = I18n.get( "ai.scope" ) , link = "";
    try {
      const mine = await Api.request( "/api/account/ai" );
      if ( mine.own_key_set ) {
        text += " " + I18n.get( "ai.scope_own" ).replace( "{service}" , AiForm.providerName( mine.own_provider , mine ) );
        link = I18n.get( "ai.scope_own_link" );
      }
    } catch ( error ) { /* the note still says what this key is for */ }
    AiForm.scope( Dom.get( "ai-key-scope" ) , text , link , "/account/ai" );
  },

  paint( state ) {
    let text;
    if ( !state.key_set ) text = I18n.get( "ai.state_unset" );
    else {
      text = I18n.get( "ai.state_set" ).replace( "{last4}" , state.last4 ).replace( "{service}" , state.provider_name );
      text += " " + ( state.model ? I18n.get( "ai.state_model" ).replace( "{model}" , AiForm.modelName( state.model , this.form.service( state.provider ) ) ) : I18n.get( "ai.state_no_model" ) );
    }
    if ( state.key_set && !state.enabled ) text += " " + I18n.get( "ai.state_off" );
    Dom.text( Dom.get( "ai-key-state" ) , text );
    Dom.show( Dom.get( "ai-key-remove" ) , state.key_set );
  },

  // Not undoable: a key is a credential, and putting an old one back would
  // need it typed again.
  async save( body , toastKey ) {
    try {
      const state = await Api.post( "/api/owner/ai" , body );
      this.form.paint( state , state.provider , state.chosen_model , state.chosen_effort );
      this.paint( state );
      this.form.clearKey();
      Undo.saved( Dom.get( "ai-key-form" ) );
      Dom.toast( I18n.get( toastKey ) , "ok" );
      AiForm.warn( state );
      if ( MyAi.form ) MyAi.refresh();
    } catch ( error ) { Org.failed( error ); }
  },
};
