// The Containers screen: every bin, box and drawer with a label, where it is
// and how full, and a form to set up a new one without a scan.
const Labels = {
  moving: { packed: "labels.move_packed" , in_transit: "labels.move_in_transit" , unpacked: "labels.move_unpacked" },

  init() {
    Pager.search( "labels-table" , "labels-search" , () => this.refresh() );
    Pager.sortable( "labels-table" , () => this.refresh() );
    Dom.get( "labels-move" ).addEventListener( "change" , () => this.refresh() );
    Dom.onSubmit( Dom.get( "label-form" ) , ( event ) => this.create( event ) );
    window.addEventListener( "household" , () => this.reset() );
    window.addEventListener( "nav" , ( event ) => { if ( event.detail.screen === "labels" ) this.refresh(); } );
  },

  async reset() {
    this.filled = null;
    if ( Org.current && Nav.current() === "labels" ) await this.refresh();
  },

  async fill() {
    if ( this.filled === Org.id ) return;
    this.filled = Org.id;
    Dom.show( Dom.get( "label-form" ).closest( ".card" ) , Org.canEdit );
    await Org.fillLocations( Dom.get( "label-location" ) );
    Org.fillColors( Dom.get( "label-color" ) );
  },

  async create( event ) {
    event.preventDefault();
    try {
      const made = await Org.post( "/labels" , {
        name: Dom.get( "label-name" ).value,
        location_id: Number( Dom.get( "label-location" ).value ),
        color: Dom.get( "label-color" ).value,
      } );
      Dom.get( "label-name" ).value = "";
      Undo.saved( Dom.get( "label-form" ) );
      Undo.record( {
        label: I18n.get( "undo.label_created" ),
        toast: true,
        undo: async () => { await Org.post( "/labels/" + made.id + "/delete" ); await this.refresh(); },
        redo: async () => { await Org.post( "/labels/" + made.id + "/restore" ); await this.refresh(); },
      } );
      await this.refresh();
    } catch ( error ) { Org.failed( error ); }
  },

  async refresh() {
    if ( !Org.current ) return;
    await this.fill();
    const params = Pager.params( "labels-table" );
    const moving = Dom.get( "labels-move" ).value;
    if ( moving ) params.move_status = moving;
    const page = await Pager.load( "labels-table" , "labels" , ( offset ) => Org.get( "/labels" , { offset: offset , ...params } ) , this.renderRow.bind( this ) );
    if ( page ) Dom.show( Dom.get( "labels-empty" ) , page.total === 0 );
  },

  renderRow( label ) {
    const name = Dom.el( "div" , { class: "d-flex align-items-center gap-2" , children: [
      label.photos.length ? Org.thumb( label.photos[ 0 ] ) : Org.swatch( label.color ),
      Dom.el( "div" , { children: [
        Org.labelLink( label.id , label.name ),
        label.container_type ? Dom.el( "div" , { class: "small text-body-secondary" , text: label.container_type } ) : null,
      ].filter( Boolean ) } ),
    ] } );
    const moving = label.move_status ? Dom.badge( I18n.get( this.moving[ label.move_status ] ) + ( label.destination ? " → " + label.destination : "" ) , label.move_status === "unpacked" ? "success" : "warning" ) : null;
    const row = Dom.el( "tr" , { attrs: { "data-key": "label-" + label.id } , children: [
      Dom.el( "td" , { children: [ name ] } ),
      Dom.el( "td" , { class: "text-nowrap font-monospace small" , text: label.printed_code } ),
      Dom.el( "td" , { class: "small" , text: label.path } ),
      Dom.el( "td" , { class: "text-nowrap" , text: String( label.item_count ) } ),
      Dom.el( "td" , { children: moving ? [ moving ] : [] } ),
      Dom.el( "td" , { class: "text-nowrap" , children: [ Dom.rowActions( label.name , [
        Dom.el( "a" , { class: "btn btn-sm btn-outline-primary" , text: I18n.get( "labels.open" ) , attrs: { href: "/labels/" + label.id } } ),
        Org.button( "labels.print_one" , "printer" , null , () => Print.reprint( [ label.id ] ) ),
        Org.canEdit ? Org.button( "labels.delete" , "trash3" , "btn-outline-danger" , () => this.remove( label ) ) : null,
      ] ) ] } ),
    ] } );
    Dom.get( "labels-table" ).querySelector( "tbody" ).appendChild( row );
  },

  async remove( label ) {
    try {
      await Org.post( "/labels/" + label.id + "/delete" );
      Undo.record( {
        label: I18n.get( "undo.label_deleted" ),
        toast: true,
        undo: async () => { await Org.post( "/labels/" + label.id + "/restore" ); await this.refresh(); },
        redo: async () => { await Org.post( "/labels/" + label.id + "/delete" ); await this.refresh(); },
      } );
      await this.refresh();
    } catch ( error ) { Org.failed( error ); }
  },
};
