// The Locations screen: the places containers sit, as a tree (House › Garage
// › Shelf 2), with rename, move and delete on each.
const Locations = {
  init() {
    Dom.onSubmit( Dom.get( "location-form" ) , ( event ) => this.create( event ) );
    window.addEventListener( "household" , () => { if ( Nav.current() === "locations" ) this.refresh(); } );
    window.addEventListener( "nav" , ( event ) => { if ( event.detail.screen === "locations" ) this.refresh(); } );
  },

  async create( event ) {
    event.preventDefault();
    try {
      const made = await Org.post( "/locations" , { name: Dom.get( "location-name" ).value , parent_id: Number( Dom.get( "location-parent" ).value ) } );
      Dom.get( "location-name" ).value = "";
      Undo.saved( Dom.get( "location-form" ) );
      Undo.record( {
        label: I18n.get( "undo.location_created" ),
        undo: async () => { await Org.post( "/locations/" + made.id + "/delete" ); await this.refresh(); },
        redo: async () => { await Org.post( "/locations/" + made.id + "/restore" ); await this.refresh(); },
      } );
      await this.refresh();
    } catch ( error ) { Org.failed( error ); }
  },

  async refresh() {
    if ( !Org.current ) return;
    Org.forgetLocations();
    Dom.show( Dom.get( "location-form" ).closest( ".card" ) , Org.canEdit );
    const locations = await Org.fillLocations( Dom.get( "location-parent" ) , 0 , "locations.top_level" );
    const rows = locations.map( ( location ) => this.row( location ) );
    Dom.swap( Dom.get( "locations-table" ).querySelector( "tbody" ) , ...rows );
    Dom.show( Dom.get( "locations-empty" ) , locations.length === 0 );
  },

  row( location ) {
    const depth = location.path.split( " › " ).length - 1;
    const name = Dom.el( "div" , { class: "org-indent-" + Math.min( depth , 5 ) , children: [
      Org.icon( depth ? "arrow-return-right" : "geo-alt" ) , Dom.el( "span" , { class: "ms-2" , text: location.name } ),
      depth ? Dom.el( "div" , { class: "small text-body-secondary" , text: location.path } ) : null,
    ].filter( Boolean ) } );
    const actions = Org.canEdit ? [
      Org.button( "locations.rename" , "pencil" , null , async () => {
        const next = await Org.ask( I18n.get( "locations.rename" ) , "locations.name_label" , location.name );
        if ( !next || next === location.name ) return;
        await this.update( location , { name: next } , { name: location.name } );
      } ),
      Org.button( "locations.move" , "arrows-move" , null , async () => {
        const select = Dom.el( "select" , { class: "form-select" , attrs: { id: "location-move-select" } } );
        await Org.fillLocations( select , location.parent_id , "locations.top_level" , location.id );
        const parent = await Org.dialog( I18n.get( "locations.move" ) , [
          Dom.el( "label" , { class: "form-label" , text: I18n.get( "locations.parent_label" ) , attrs: { for: "location-move-select" } } ) , select ,
        ] , () => Number( select.value ) );
        if ( parent === null || parent === ( location.parent_id || 0 ) ) return;
        await this.update( location , { parent_id: parent } , { parent_id: location.parent_id || 0 } );
      } ),
      Org.button( "locations.delete" , "trash3" , "btn-outline-danger" , async () => {
        try {
          await Org.post( "/locations/" + location.id + "/delete" );
          Undo.record( {
            label: I18n.get( "undo.location_deleted" ),
            toast: true,
            undo: async () => { await Org.post( "/locations/" + location.id + "/restore" ); await this.refresh(); },
            redo: async () => { await Org.post( "/locations/" + location.id + "/delete" ); await this.refresh(); },
          } );
          await this.refresh();
        } catch ( error ) { Org.failed( error ); }
      } ),
    ] : [];
    return Dom.el( "tr" , { attrs: { "data-key": "location-" + location.id } , children: [
      Dom.el( "td" , { children: [ name ] } ),
      Dom.el( "td" , { text: String( location.label_count ) } ),
      Dom.el( "td" , { class: "text-nowrap" , children: [ Dom.rowActions( location.name , actions ) ] } ),
    ] } );
  },

  async update( location , after , before ) {
    try {
      await Org.post( "/locations/" + location.id , after );
      Undo.record( {
        label: I18n.get( "undo.location_edited" ),
        undo: async () => { await Org.post( "/locations/" + location.id , before ); await this.refresh(); },
        redo: async () => { await Org.post( "/locations/" + location.id , after ); await this.refresh(); },
      } );
      await this.refresh();
    } catch ( error ) { Org.failed( error ); }
  },
};
