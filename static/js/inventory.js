// The Inventory screen: every item in the open household, searchable (a typo
// is forgiven), narrowed by location, tag, or what needs attention, sortable,
// with the item's container and place on every row. Tick rows to move or
// delete several at once.
const Inventory = {
  selected: new Map(),

  init() {
    Pager.search( "inventory-table" , "inventory-search" , () => this.refresh() );
    Pager.sortable( "inventory-table" , () => this.refresh() );
    Dom.typed( Dom.get( "inventory-tag" ) , () => this.refresh() );
    Dom.get( "inventory-location" ).addEventListener( "change" , () => this.refresh() );
    Dom.get( "inventory-filter" ).addEventListener( "change" , () => this.refresh() );
    Dom.get( "inventory-select-all" ).addEventListener( "change" , ( event ) => {
      Dom.all( "[data-pick]" , Dom.get( "inventory-table" ) ).forEach( ( box ) => {
        box.checked = event.target.checked;
        box.dispatchEvent( new Event( "change" ) );
      } );
    } );
    Dom.onPress( Dom.get( "inventory-bulk-move" ) , () => ItemEditor.move( Array.from( this.selected.values() ) , () => this.refresh() ) );
    Dom.onPress( Dom.get( "inventory-bulk-delete" ) , () => ItemEditor.remove( Array.from( this.selected.values() ) , () => this.refresh() ) );
    window.addEventListener( "household" , () => this.reset() );
    window.addEventListener( "inventory-changed" , () => { if ( Nav.current() === "inventory" ) this.refresh(); } );
    window.addEventListener( "nav" , ( event ) => { if ( event.detail.screen === "inventory" ) this.refresh(); } );
  },

  async reset() {
    this.filled = null;
    if ( Org.current && Nav.current() === "inventory" ) await this.refresh();
  },

  // The location picker, filled once per household, when the screen opens.
  async fill() {
    if ( this.filled === Org.id ) return;
    this.filled = Org.id;
    await Org.fillLocations( Dom.get( "inventory-location" ) , 0 , "inventory.everywhere" );
  },

  params() {
    const params = Pager.params( "inventory-table" );
    const location = Number( Dom.get( "inventory-location" ).value || 0 );
    if ( location ) params.location_id = location;
    params.tag = Dom.get( "inventory-tag" ).value.trim();
    const filter = Dom.get( "inventory-filter" ).value;
    if ( filter === "expiring" ) params.expiring_within = 30;
    if ( filter === "low" ) params.low_stock = true;
    if ( filter === "lent" ) params.lent = true;
    return params;
  },

  async refresh() {
    if ( !Org.current ) return;
    await this.fill();
    this.selected.clear();
    this.paintBulk();
    Dom.get( "inventory-select-all" ).checked = false;
    const params = this.params();
    let total = 0;
    const page = await Pager.load( "inventory-table" , "items" , ( offset ) => Org.get( "/items" , { offset: offset , ...params } ).then( ( answer ) => {
      total = answer.total_value;
      return answer;
    } ) , this.renderRow.bind( this ) );
    if ( !page ) return;
    const narrowed = Pager.narrowed( "inventory-table" ) || params.location_id || params.tag || Dom.get( "inventory-filter" ).value;
    Dom.show( Dom.get( "inventory-empty" ) , page.total === 0 );
    Dom.text( Dom.get( "inventory-empty" ) , I18n.get( narrowed ? "inventory.no_match" : "inventory.empty" ) );
    Dom.text( Dom.get( "inventory-total" ) , page.total ? I18n.get( "inventory.total" ).replace( "{count}" , page.total ).replace( "{value}" , Org.money( total ) ) : "" );
  },

  paintBulk() {
    const count = this.selected.size;
    Dom.show( Dom.get( "inventory-bulk" ) , count > 0 && Org.canEdit );
    Dom.text( Dom.get( "inventory-bulk-count" ) , I18n.get( "inventory.selected" ).replace( "{count}" , count ) );
  },

  // The item column: its photo, name, tags and the badges that need noticing.
  nameCell( item ) {
    const badges = [];
    if ( item.low_stock_now ) badges.push( Dom.badge( I18n.get( "inventory.badge_low" ) , "warning" ) );
    if ( item.lent_to ) badges.push( Dom.badge( I18n.get( "inventory.badge_lent" ).replace( "{who}" , item.lent_to ) , item.overdue ? "danger" : "info" ) );
    const text = Dom.el( "div" , { children: [
      Dom.el( "button" , { class: "btn btn-link p-0 text-start fw-semibold" , text: item.name , attrs: { type: "button" } , on: { click: async () => {
        if ( await ItemEditor.open( item ) ) await this.refresh();
      } } } ),
      item.tags.length ? Dom.el( "div" , { class: "small text-body-secondary" , text: item.tags.join( ", " ) } ) : null,
      badges.length ? Dom.el( "div" , { class: "d-flex flex-wrap gap-1 mt-1" , children: badges } ) : null,
    ].filter( Boolean ) } );
    return Dom.el( "div" , { class: "d-flex align-items-center gap-2" , children: [ item.photos.length ? Org.thumb( item.photos[ 0 ] ) : null , text ].filter( Boolean ) } );
  },

  renderRow( item ) {
    const pick = Dom.el( "input" , { class: "form-check-input" , attrs: { type: "checkbox" , "data-pick": "" , "aria-label": I18n.get( "inventory.select_one" ).replace( "{name}" , item.name ) } } );
    pick.addEventListener( "change" , () => {
      if ( pick.checked ) this.selected.set( item.id , item );
      else this.selected.delete( item.id );
      this.paintBulk();
    } );
    const where = Dom.el( "div" , { children: [
      Dom.el( "div" , { class: "d-flex align-items-center gap-1" , children: [ Org.swatch( item.label_color ) , Org.labelLink( item.label_id , item.label_name ) ] } ),
      Dom.el( "div" , { class: "small text-body-secondary" , text: item.path } ),
    ] } );
    const expires = Dom.el( "td" , { class: "text-nowrap" , children: item.expires ? [ Dom.el( "div" , { text: Org.date( item.expires ) } ) , Org.dueBadge( item.expires_in ) ] : [] } );
    const row = Dom.el( "tr" , { attrs: { "data-key": "item-" + item.id } , children: [
      Dom.el( "td" , { children: [ pick ] } ),
      Dom.el( "td" , { children: [ this.nameCell( item ) ] } ),
      Dom.el( "td" , { class: "text-nowrap" , text: Org.amount( item.quantity ) + ( item.unit ? " " + item.unit : "" ) } ),
      Dom.el( "td" , { children: [ where ] } ),
      expires,
      Dom.el( "td" , { class: "text-nowrap" , text: item.value ? Org.money( item.value * Math.max( item.quantity , 1 ) ) : "" } ),
      Dom.el( "td" , { class: "text-nowrap" , children: [ ItemEditor.actions( item , () => this.refresh() ) ] } ),
    ] } );
    Dom.get( "inventory-table" ).querySelector( "tbody" ).appendChild( row );
  },
};
