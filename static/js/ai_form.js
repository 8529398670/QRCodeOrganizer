// The service / key / model / effort form that both AI cards share: the
// owner's on Settings (ai_key.js) and each person's on Account (my_ai.js).
//
// The server sends every service with its last loaded model list and what
// "default" means on that screen. Picking a service shows its models and the
// efforts it takes; picking a model narrows the efforts to the ones that model
// takes, when its service says. The refresh button loads the service's current
// list, with the key typed in (not stored) or the saved one.
class AiForm {
  // prefix is the cards' id prefix ("ai-key" or "my-ai"); endpoint is the
  // settings route, whose /models refreshes; onState repaints the card.
  constructor( prefix , endpoint , onState ) {
    this.endpoint = endpoint;
    this.onState = onState;
    this.state = null;
    this.field = ( name ) => Dom.get( prefix + "-" + name );
    this.field( "provider" ).addEventListener( "change" , () => this.paintChoices( this.field( "provider" ).value , "" , "" ) );
    this.field( "model" ).addEventListener( "change" , () => this.paintEfforts( this.current() , this.field( "model" ).value , this.field( "effort" ).value ) );
    Dom.onPress( this.field( "refresh" ) , () => this.refresh() );
  }

  service( id ) {
    const providers = this.state ? this.state.providers : [];
    return providers.find( ( provider ) => provider.id === id ) || providers[ 0 ];
  }

  current() { return this.service( this.field( "provider" ).value ); }

  static modelName( id , service ) {
    if ( !id ) return "";
    const known = I18n.get( "ai_form.model_" + id.replace( /[^a-z0-9]+/gi , "_" ) );
    if ( known ) return known;
    const info = service && service.models.find( ( model ) => model.id === id );
    return info && info.name && info.name !== id ? info.name + " (" + id + ")" : id;
  }

  static effortName( effort ) { return I18n.get( "ai_form.effort_" + effort ) || effort; }

  static providerName( id , state ) {
    const found = state && state.providers.find( ( provider ) => provider.id === id );
    return found ? found.name : id;
  }

  // paint fills the form from a reply: provider is the saved service,
  // model and effort the saved choices ("" is the default).
  paint( state , provider , model , effort ) {
    this.state = state;
    Dom.swap( this.field( "provider" ) , ...state.providers.map( ( service ) => Dom.el( "option" , { text: service.name , attrs: { value: service.id } } ) ) );
    this.field( "provider" ).value = provider;
    this.saved = { provider: provider , model: model , effort: effort };
    this.paintChoices( provider , model , effort );
  }

  paintChoices( providerId , model , effort ) {
    const service = this.service( providerId );
    if ( !service ) return;
    // Back on the saved service, its saved choices come back.
    if ( this.saved && providerId === this.saved.provider && !model && !effort ) ( { model , effort } = this.saved );
    this.field( "key" ).placeholder = I18n.get( "ai_form.key_placeholder" ).replace( "{hint}" , service.key_hint );
    const defaultName = service.default_model
      ? I18n.get( "ai_form.model_default" ).replace( "{model}" , AiForm.modelName( service.default_model , service ) )
      : I18n.get( "ai_form.model_none" );
    const ids = service.models.map( ( info ) => info.id );
    if ( model && !ids.includes( model ) ) ids.unshift( model );
    Dom.swap( this.field( "model" ) , Dom.el( "option" , { text: defaultName , attrs: { value: "" } } ) ,
      ...ids.map( ( id ) => Dom.el( "option" , { text: AiForm.modelName( id , service ) , attrs: { value: id } } ) ) );
    this.field( "model" ).value = model || "";
    this.paintEfforts( service , model , effort );
    this.paintNote( service );
  }

  // The efforts the chosen model takes, when its service says; else the
  // service's own list.
  paintEfforts( service , model , effort ) {
    if ( !service ) return;
    const info = service.models.find( ( entry ) => entry.id === ( model || service.default_model ) );
    let efforts = info && Array.isArray( info.efforts ) ? info.efforts : service.efforts;
    if ( effort && !efforts.includes( effort ) ) efforts = [ effort , ...efforts ];
    const defaultName = service.default_effort
      ? I18n.get( "ai_form.effort_default" ).replace( "{effort}" , AiForm.effortName( service.default_effort ) )
      : I18n.get( "ai_form.effort_model" );
    const select = this.field( "effort" );
    Dom.swap( select , Dom.el( "option" , { text: defaultName , attrs: { value: "" } } ) ,
      ...efforts.map( ( level ) => Dom.el( "option" , { text: AiForm.effortName( level ) , attrs: { value: level } } ) ) );
    select.value = effort || "";
    select.disabled = efforts.length === 0;
  }

  paintNote( service ) {
    const note = this.field( "models-note" );
    const parts = [];
    if ( service.refreshed_at ) {
      parts.push( Dom.el( "span" , { text: I18n.get( "ai_form.list_loaded" ).replace( "{count}" , service.models.length ) + " " } ) , Dom.when( service.refreshed_at ) , Dom.el( "span" , { text: ". " } ) );
    } else {
      parts.push( Dom.el( "span" , { text: I18n.get( service.models.length ? "ai_form.list_builtin" : "ai_form.list_empty" ) + " " } ) );
    }
    parts.push( Dom.el( "a" , { text: I18n.get( "ai_form.get_key" ).replace( "{service}" , service.name ) ,
      attrs: { href: service.keys_url , target: "_blank" , rel: "noopener noreferrer" } } ) );
    Dom.swap( note , ...parts );
  }

  // What the form says now, for a save.
  body() {
    const body = { provider: this.field( "provider" ).value , model: this.field( "model" ).value , effort: this.field( "effort" ).value };
    const key = this.field( "key" ).value.trim();
    if ( key ) body.api_key = key;
    return body;
  }

  clearKey() { this.field( "key" ).value = ""; }

  async refresh() {
    const keep = { provider: this.field( "provider" ).value , model: this.field( "model" ).value , effort: this.field( "effort" ).value };
    const body = { provider: keep.provider };
    const key = this.field( "key" ).value.trim();
    if ( key ) body.api_key = key;
    try {
      const state = await Api.post( this.endpoint + "/models" , body );
      this.state = state;
      this.onState( state );
      // The person's unsaved picks stay as they were.
      this.field( "provider" ).value = keep.provider;
      this.paintChoices( keep.provider , keep.model , keep.effort );
      const count = this.current().models.length;
      Dom.toast( I18n.get( "ai_form.refreshed" ).replace( "{count}" , count ).replace( "{service}" , this.current().name ) , "ok" );
    } catch ( error ) { App.showError( error.message ); }
  }

  // A save that loaded the list but could not say so otherwise.
  static warn( state ) {
    if ( state && state.models_error ) App.showError( I18n.get( "ai_form.list_failed" ).replace( "{error}" , state.models_error ) );
  }

  // scope fills a card's "whose key is this" note: what the key is used for,
  // and a link to the other card when that is the one someone may have meant.
  static scope( holder , text , link , href ) {
    const body = Dom.el( "div" , { children: [ Dom.el( "span" , { text: text } ) ] } );
    if ( link ) {
      body.appendChild( document.createTextNode( " " ) );
      body.appendChild( Dom.el( "a" , { class: "alert-link" , text: link , attrs: { href: href } } ) );
    }
    Dom.swap( holder , Dom.el( "i" , { class: "bi bi-info-circle" , attrs: { "aria-hidden": "true" } } ) , body );
    Dom.show( holder , true );
  }
}
