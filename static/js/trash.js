// The Trash screen: deleted items, containers and locations, until the hourly
// purge removes them for good (inventory.trash_days). Restoring a container
// brings back the items deleted with it.
const Trash = {
  init() {
    window.addEventListener( "household" , () => { if ( Nav.current() === "trash" ) this.refresh(); } );
    window.addEventListener( "nav" , ( event ) => { if ( event.detail.screen === "trash" ) this.refresh(); } );
  },

  async refresh() {
    if ( !Org.current ) return;
    try {
      const [ items , labels , locations ] = await Promise.all( [
        Org.get( "/items" , { trash: true , limit: 1000 } ),
        Org.get( "/labels" , { trash: true , limit: 1000 } ),
        Org.get( "/locations" , { trash: true , limit: 1000 } ),
      ] );
      // An item deleted with its container comes back with it: list only the
      // ones that can be restored on their own.
      const trashedLabels = new Set( labels.labels.map( ( label ) => label.id ) );
      const entries = [
        ...labels.labels.map( ( label ) => ( { kind: "label" , id: label.id , name: label.name , detail: label.path , deleted: label.deleted_at } ) ),
        ...items.items.filter( ( item ) => !trashedLabels.has( item.label_id ) ).map( ( item ) => ( { kind: "item" , id: item.id , name: item.name , detail: item.path , deleted: item.deleted_at } ) ),
        ...locations.locations.map( ( location ) => ( { kind: "location" , id: location.id , name: location.name , detail: "" , deleted: location.deleted_at } ) ),
      ].sort( ( a , b ) => String( b.deleted ).localeCompare( String( a.deleted ) ) );
      const paths = { label: "/labels/" , item: "/items/" , location: "/locations/" };
      const rows = entries.map( ( entry ) => Dom.el( "tr" , { children: [
        Dom.el( "td" , { children: [ Dom.el( "div" , { text: entry.name } ) , entry.detail ? Dom.el( "div" , { class: "small text-body-secondary" , text: entry.detail } ) : null ].filter( Boolean ) } ),
        Dom.el( "td" , { text: I18n.get( "trash.kind_" + entry.kind ) } ),
        Dom.el( "td" , { children: [ Dom.when( entry.deleted ) ] } ),
        Dom.el( "td" , { children: Org.canEdit ? [ Org.button( "trash.restore" , "arrow-counterclockwise" , "btn-outline-primary" , async () => {
          try {
            await Org.post( paths[ entry.kind ] + entry.id + "/restore" );
            Undo.record( {
              label: I18n.get( "undo.restored" ),
              toast: true,
              undo: async () => { await Org.post( paths[ entry.kind ] + entry.id + "/delete" ); await this.refresh(); },
              redo: async () => { await Org.post( paths[ entry.kind ] + entry.id + "/restore" ); await this.refresh(); },
            } );
            await this.refresh();
          } catch ( error ) { Org.failed( error ); }
        } ) ] : [] } ),
      ] } ) );
      Dom.swap( Dom.get( "trash-table" ).querySelector( "tbody" ) , ...rows );
      Dom.show( Dom.get( "trash-empty" ) , entries.length === 0 );
    } catch ( error ) { Org.failed( error ); }
  },
};
