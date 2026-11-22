// The Reports screen: totals at a glance, value by location, the inventory as
// a spreadsheet or a printable (PDF) list, an insurance list with values,
// serial numbers and receipt photos, and adding items from a spreadsheet.
const Reports = {
  init() {
    Dom.onPress( Dom.get( "reports-pdf" ) , () => this.printList( false ) );
    Dom.onPress( Dom.get( "reports-insurance" ) , () => this.printList( true ) );
    Dom.get( "reports-import" ).addEventListener( "change" , ( event ) => this.importFile( event ) );
    window.addEventListener( "household" , () => { if ( Nav.current() === "reports" ) this.refresh(); } );
    window.addEventListener( "nav" , ( event ) => { if ( event.detail.screen === "reports" ) this.refresh(); } );
  },

  async everything() {
    const [ items , labels , locations ] = await Promise.all( [
      Org.get( "/items" , { limit: 1000 , sort: "path" } ),
      Org.labels(),
      Org.locations(),
    ] );
    return { items: items.items , total: items.total , value: items.total_value , labels: labels , locations: locations };
  },

  stat( key , value , icon ) {
    return Dom.el( "div" , { class: "col" , children: [ Dom.el( "div" , { class: "card h-100 mb-0" , children: [ Dom.el( "div" , { class: "card-body" , children: [
      Dom.el( "div" , { class: "small text-body-secondary" , children: [ Org.icon( icon ) , Dom.el( "span" , { class: "ms-1" , text: I18n.get( key ) } ) ] } ),
      Dom.el( "div" , { class: "fs-4 fw-semibold" , text: value } ),
    ] } ) ] } ) ] } );
  },

  async refresh() {
    if ( !Org.current ) return;
    Dom.get( "reports-csv" ).setAttribute( "href" , Org.path( "/export.csv" ) );
    Dom.show( Dom.get( "reports-import" ).closest( ".card" ) , Org.canEdit );
    try {
      const all = await this.everything();
      const expiring = all.items.filter( ( item ) => item.expires_in !== null && item.expires_in <= 30 ).length;
      const low = all.items.filter( ( item ) => item.low_stock_now ).length;
      Dom.swap( Dom.get( "reports-stats" ) ,
        this.stat( "reports.stat_items" , Org.amount( all.total ) , "box2" ),
        this.stat( "reports.stat_containers" , Org.amount( all.labels.length ) , "box-seam" ),
        this.stat( "reports.stat_value" , Org.money( all.value ) , "cash-stack" ),
        this.stat( "reports.stat_attention" , Org.amount( expiring + low ) , "exclamation-triangle" ) );
      const byPlace = new Map();
      const place = ( label ) => {
        const parts = label.path.split( " › " );
        return parts.length > 1 ? parts.slice( 0 , -1 ).join( " › " ) : I18n.get( "reports.no_location" );
      };
      all.labels.forEach( ( label ) => {
        const key = place( label );
        const entry = byPlace.get( key ) || { containers: 0 , items: 0 , value: 0 };
        entry.containers += 1;
        entry.items += label.item_count;
        entry.value += label.value;
        byPlace.set( key , entry );
      } );
      const rows = Array.from( byPlace.entries() ).sort( ( a , b ) => a[ 0 ].localeCompare( b[ 0 ] ) ).map( ( [ name , entry ] ) => Dom.el( "tr" , { children: [
        Dom.el( "td" , { text: name } ),
        Dom.el( "td" , { text: String( entry.containers ) } ),
        Dom.el( "td" , { text: String( entry.items ) } ),
        Dom.el( "td" , { text: Org.money( entry.value ) } ),
      ] } ) );
      Dom.swap( Dom.get( "reports-table" ).querySelector( "tbody" ) , ...rows );
    } catch ( error ) { Org.failed( error ); }
  },

  // The inventory as a printable list, grouped by container; the browser's
  // print dialog saves it as a PDF. The insurance list adds what each thing
  // is worth, its serial and purchase date, and its receipt photos.
  async printList( insurance ) {
    try {
      const all = await this.everything();
      const groups = new Map();
      all.items.forEach( ( item ) => {
        const list = groups.get( item.path ) || [];
        list.push( item );
        groups.set( item.path , list );
      } );
      const heads = insurance
        ? [ "reports.col_item" , "reports.col_quantity" , "reports.col_serial" , "reports.col_bought" , "reports.col_value" ]
        : [ "reports.col_item" , "reports.col_quantity" , "reports.col_tags" , "reports.col_expires" ];
      const nodes = [ Dom.el( "p" , { text: Org.current.name + " · " + new Date().toLocaleDateString() + " · " +
        I18n.get( "reports.summary" ).replace( "{count}" , all.total ).replace( "{value}" , Org.money( all.value ) ) } ) ];
      groups.forEach( ( items , path ) => {
        nodes.push( Dom.el( "h2" , { class: "h6 mt-3" , text: path } ) );
        nodes.push( Dom.el( "table" , { class: "table table-sm" , children: [
          Dom.el( "thead" , { children: [ Dom.el( "tr" , { children: heads.map( ( key ) => Dom.el( "th" , { text: I18n.get( key ) } ) ) } ) ] } ),
          Dom.el( "tbody" , { children: items.map( ( item ) => Dom.el( "tr" , { children: ( insurance ? [
            [ item.name , item.brand ].filter( Boolean ).join( " · " ),
            Org.amount( item.quantity ),
            item.serial || "",
            Org.date( item.purchase_date ),
            item.value ? Org.money( item.value * Math.max( item.quantity , 1 ) ) : "",
          ] : [
            item.name,
            Org.amount( item.quantity ) + ( item.unit ? " " + item.unit : "" ),
            item.tags.join( ", " ),
            Org.date( item.expires ),
          ] ).map( ( text ) => Dom.el( "td" , { text: text } ) ) } ) ) } ),
        ] } ) );
        if ( insurance ) {
          const receipts = items.flatMap( ( item ) => item.receipts.concat( item.photos.slice( 0 , 1 ) ) );
          if ( receipts.length ) nodes.push( Dom.el( "div" , { class: "d-flex flex-wrap gap-2" , children: receipts.map( ( fileId ) =>
            Dom.el( "img" , { class: "print-receipt" , attrs: { src: Org.photoUrl( fileId ) , alt: "" } } ) ) } ) );
        }
      } );
      Print.document( I18n.get( insurance ? "reports.insurance_title" : "reports.list_title" ) , nodes );
    } catch ( error ) { Org.failed( error ); }
  },

  async importFile( event ) {
    const file = event.target.files[ 0 ];
    event.target.value = "";
    if ( !file ) return;
    const result = Dom.get( "reports-import-result" );
    try {
      const fileId = await Org.upload( file );
      const done = await Org.post( "/import" , { file_id: fileId } );
      let text = I18n.get( "reports.imported" ).replace( "{items}" , done.items ).replace( "{containers}" , done.containers ).replace( "{locations}" , done.locations );
      if ( done.skipped.length ) text += " " + I18n.get( "reports.import_skipped" ).replace( "{rows}" , done.skipped.join( "; " ) );
      Dom.text( result , text );
      Dom.show( result , true );
      await Api.deleteFile( fileId ).catch( () => {} );
      await this.refresh();
    } catch ( error ) { Org.failed( error ); }
  },
};
