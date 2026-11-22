// The "Reading photos with AI" card on Account: your own key for one of the
// AI services (your photos then go on your own bill), and the model and
// effort it uses (see ai_form.js). Without a key of your own, an admin
// decides whether you may use the server's.
const MyAi = {
  state: null,
  form: null,
  owner: false,

  init( me ) {
    this.owner = !!me.is_owner;
    Dom.show( Dom.get( "my-ai-panel" ) , true );
    this.form = new AiForm( "my-ai" , "/api/account/ai" , ( state ) => this.paint( state ) );
    Dom.onSubmit( Dom.get( "my-ai-form" ) , ( event ) => this.save( event ) );
    Dom.onPress( Dom.get( "my-ai-remove" ) , async () => {
      if ( !await Dom.confirm( I18n.get( "my_ai.remove_confirm" ) ) ) return;
      await this.send( { remove_key: true } , "my_ai.removed" );
    } );
    this.refresh();
  },

  async refresh() {
    try {
      const state = await Api.request( "/api/account/ai" );
      this.form.paint( state , state.own_provider , state.own_model , state.own_effort );
      this.paint( state );
    } catch ( error ) { /* the rest of Account works without this */ }
  },

  paint( state ) {
    this.state = state;
    const model = AiForm.modelName( state.model , this.form.service( state.provider ) );
    let text;
    if ( !state.enabled ) text = I18n.get( "my_ai.state_off" );
    else if ( state.using !== "none" && !state.model ) text = I18n.get( "my_ai.state_no_model" ).replace( "{service}" , state.provider_name );
    else if ( state.using === "own" ) text = I18n.get( "my_ai.state_own" ).replace( "{last4}" , state.last4 ).replace( "{service}" , state.provider_name ).replace( "{model}" , model );
    else if ( state.using === "shared" ) text = I18n.get( "my_ai.state_shared" ).replace( "{service}" , state.provider_name ).replace( "{model}" , model );
    else if ( state.shared_allowed && !state.key_set ) text = I18n.get( "my_ai.state_no_server_key" );
    else text = I18n.get( "my_ai.state_none" );
    Dom.text( Dom.get( "my-ai-state" ) , text );
    Dom.show( Dom.get( "my-ai-remove" ) , state.own_key_set );
    let scope = I18n.get( "my_ai.scope" ) + " ";
    scope += state.key_set ? I18n.get( "my_ai.scope_server" ).replace( "{service}" , AiForm.providerName( state.server_provider , state ) ) : I18n.get( "my_ai.scope_server_none" );
    AiForm.scope( Dom.get( "my-ai-scope" ) , scope , this.owner ? I18n.get( "my_ai.scope_owner_link" ) : "" , "/admin/settings/ai" );
  },

  async save( event ) {
    event.preventDefault();
    const body = this.form.body();
    const before = this.state ? { provider: this.state.own_provider , model: this.state.own_model || "" , effort: this.state.own_effort || "" } : null;
    await this.send( body , body.api_key ? "my_ai.key_saved" : "my_ai.saved" );
    // The service, model and effort can be put back; a key is a credential
    // and is not. Without a new key, the service only changes when there was
    // no key, so putting it back needs none either.
    if ( before && !body.api_key ) {
      Undo.record( {
        label: I18n.get( "undo.ai_settings" ),
        undo: () => this.send( before ),
        redo: () => this.send( { provider: body.provider , model: body.model , effort: body.effort } ),
      } );
    }
  },

  async send( body , toastKey ) {
    try {
      const state = await Api.post( "/api/account/ai" , body );
      this.form.paint( state , state.own_provider , state.own_model , state.own_effort );
      this.paint( state );
      this.form.clearKey();
      Undo.saved( Dom.get( "my-ai-form" ) );
      if ( toastKey ) Dom.toast( I18n.get( toastKey ) , "ok" );
      AiForm.warn( state );
      if ( AiKey.form ) AiKey.paintScope();
    } catch ( error ) { App.showError( error.message ); }
  },
};
