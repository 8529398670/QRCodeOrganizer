// The "Reading photos with AI" card on Account: your own Anthropic key (your
// photos then go on your own bill), and the model and effort it uses. Without
// a key of your own, an admin decides whether you may use the server's.
const MyAi = {
  state: null,

  init( me ) {
    Dom.show( Dom.get( "my-ai-panel" ) , true );
    Dom.onSubmit( Dom.get( "my-ai-form" ) , ( event ) => this.save( event ) );
    Dom.onPress( Dom.get( "my-ai-remove" ) , async () => {
      if ( !await Dom.confirm( I18n.get( "my_ai.remove_confirm" ) ) ) return;
      await this.send( { remove_key: true } , "my_ai.removed" );
    } );
    this.refresh();
  },

  async refresh() {
    try {
      this.paint( await Api.request( "/api/account/ai" ) );
    } catch ( error ) { /* the rest of Account works without this */ }
  },

  modelName( model ) {
    return I18n.get( "my_ai.model_" + model.replace( /^claude-/ , "" ).replace( /-/g , "_" ) ) || model;
  },

  paint( state ) {
    this.state = state;
    const select = Dom.get( "my-ai-model" );
    const appName = I18n.get( "my_ai.model_app" ).replace( "{model}" , this.modelName( state.app_model ) );
    Dom.swap( select , Dom.el( "option" , { text: appName , attrs: { value: "" } } ) ,
      ...state.models.map( ( model ) => Dom.el( "option" , { text: this.modelName( model ) , attrs: { value: model } } ) ) );
    select.value = state.own_model || "";
    Dom.get( "my-ai-effort" ).value = state.own_effort || "";
    let text;
    if ( !state.enabled ) text = I18n.get( "my_ai.state_off" );
    else if ( state.using === "own" ) text = I18n.get( "my_ai.state_own" ).replace( "{last4}" , state.last4 ).replace( "{model}" , this.modelName( state.model ) );
    else if ( state.using === "shared" ) text = I18n.get( "my_ai.state_shared" ).replace( "{model}" , this.modelName( state.model ) );
    else if ( state.shared_allowed && !state.key_set ) text = I18n.get( "my_ai.state_no_server_key" );
    else text = I18n.get( "my_ai.state_none" );
    Dom.text( Dom.get( "my-ai-state" ) , text );
    Dom.show( Dom.get( "my-ai-remove" ) , state.own_key_set );
  },

  async save( event ) {
    event.preventDefault();
    const body = { model: Dom.get( "my-ai-model" ).value , effort: Dom.get( "my-ai-effort" ).value };
    const key = Dom.get( "my-ai-key" ).value.trim();
    if ( key ) body.api_key = key;
    const before = this.state ? { model: this.state.own_model || "" , effort: this.state.own_effort || "" } : null;
    await this.send( body , key ? "my_ai.key_saved" : "my_ai.saved" );
    // The model and effort can be put back; a key is a credential and is not.
    if ( before && !key ) {
      Undo.record( {
        label: I18n.get( "undo.ai_settings" ),
        undo: () => this.send( before ),
        redo: () => this.send( { model: body.model , effort: body.effort } ),
      } );
    }
  },

  async send( body , toastKey ) {
    try {
      this.paint( await Api.post( "/api/account/ai" , body ) );
      Dom.get( "my-ai-key" ).value = "";
      Undo.saved( Dom.get( "my-ai-form" ) );
      if ( toastKey ) Dom.toast( I18n.get( toastKey ) , "ok" );
    } catch ( error ) { App.showError( error.message ); }
  },
};
