// The "Who may use the server's AI key" card on Users, for admins: one switch
// for everyone, and per person "follow the default", "may use it" or "must
// bring their own". Someone with their own key always uses it first.
const AiAccess = {
  init( me ) {
    if ( me.role !== "admin" ) return;
    Dom.show( Dom.get( "ai-access-panel" ) , true );
    Dom.get( "ai-access-default" ).addEventListener( "change" , ( event ) => this.setDefault( event.target.checked , true ) );
    window.addEventListener( "nav" , ( event ) => { if ( event.detail.screen === "users" ) this.refresh(); } );
  },

  async refresh() {
    try {
      this.paint( await Api.request( "/api/admin/ai-access?limit=1000" ) );
    } catch ( error ) { App.showError( error.message ); }
  },

  paint( state ) {
    Dom.get( "ai-access-default" ).checked = state.default_shared;
    Dom.show( Dom.get( "ai-access-nokey" ) , !state.server_key_set );
    const rows = state.people.map( ( person ) => {
      let choice;
      if ( person.owner ) {
        choice = Dom.el( "span" , { class: "text-body-secondary" , text: I18n.get( "ai_access.owner" ) } );
      } else {
        choice = Dom.el( "select" , { class: "form-select form-select-sm" , attrs: { "aria-label": I18n.get( "ai_access.choice_for" ).replace( "{name}" , person.display_name ) } ,
          children: [ "default" , "shared" , "own" ].map( ( value ) => Dom.el( "option" , { text: I18n.get( "ai_access.access_" + value ) , attrs: { value: value } } ) ) } );
        choice.value = person.access;
        choice.addEventListener( "change" , () => this.setAccess( person , person.access , choice.value , true ) );
      }
      let result = I18n.get( person.may_share ? "ai_access.result_shared" : "ai_access.result_own" );
      if ( person.own_key_set ) result = I18n.get( "ai_access.result_has_own" );
      return Dom.el( "tr" , { attrs: { "data-key": "ai-" + person.user_id } , children: [
        Dom.el( "td" , { text: person.display_name } ),
        Dom.el( "td" , { children: [ choice ] } ),
        Dom.el( "td" , { class: "small" , text: result } ),
      ] } );
    } );
    Dom.swap( Dom.get( "ai-access-table" ).querySelector( "tbody" ) , ...rows );
  },

  async setDefault( shared , record ) {
    try {
      this.paint( await Api.post( "/api/admin/ai-access" , { default_shared: shared } ) );
      if ( record ) Undo.record( {
        label: I18n.get( "undo.ai_access" ),
        undo: () => this.setDefault( !shared , false ),
        redo: () => this.setDefault( shared , false ),
      } );
    } catch ( error ) {
      App.showError( error.message );
      await this.refresh();
    }
  },

  async setAccess( person , before , after , record ) {
    try {
      this.paint( await Api.post( "/api/admin/ai-access/" + person.user_id , { access: after } ) );
      if ( record ) Undo.record( {
        label: I18n.get( "undo.ai_access" ),
        undo: () => this.setAccess( person , after , before , false ),
        redo: () => this.setAccess( person , before , after , false ),
      } );
    } catch ( error ) {
      App.showError( error.message );
      await this.refresh();
    }
  },
};
