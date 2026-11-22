// The Print screen: make brand-new codes and print them as sticker sheets, or
// print a container's own label again. The server only mints codes, so none
// is ever printed twice; the layout is drawn here and printed by the browser,
// which can also save it as a PDF.
//
// Each label carries the QR (an address ending /l/<code>, which any phone's
// camera opens), the code in letters for typing in, a band of its color, and
// its name and handling notes when it has them.
const Print = {
  // Sheet layouts, sized in inches: label width and height, columns, rows,
  // and the gap between labels. The page is Letter unless it says A4.
  layouts: {
    "letter-2.5": { w: 2.5 , h: 2.5 , cols: 3 , rows: 4 , gap: 0 , page: "letter" },
    "letter-2": { w: 2 , h: 2 , cols: 4 , rows: 5 , gap: 0 , page: "letter" },
    "letter-1.5": { w: 1.5 , h: 1.5 , cols: 5 , rows: 7 , gap: 0 , page: "letter" },
    "avery-5160": { w: 2.625 , h: 1 , cols: 3 , rows: 10 , gap: 0.125 , page: "letter" , wide: true },
    "avery-22806": { w: 2 , h: 2 , cols: 3 , rows: 4 , gap: 0.625 , page: "letter" },
    "a4-60": { w: 2.362 , h: 2.362 , cols: 3 , rows: 4 , gap: 0.1 , page: "a4" },
    "a4-40": { w: 1.575 , h: 1.575 , cols: 4 , rows: 6 , gap: 0.15 , page: "a4" },
  },
  pages: { letter: { w: 8.5 , h: 11 } , a4: { w: 8.268 , h: 11.693 } },

  init() {
    const layout = Dom.get( "print-layout" );
    Dom.swap( layout , ...Object.keys( this.layouts ).map( ( key ) => Dom.el( "option" , { text: I18n.get( "print.layout_" + key.replace( /[.-]/g , "_" ) ) , attrs: { value: key } } ) ) );
    const palette = Dom.get( "print-palette" );
    Dom.swap( palette , ...[ "original" , "neon" , "modern" , "mono" ].map( ( key ) => Dom.el( "option" , { text: I18n.get( "print.palette_" + key ) , attrs: { value: key } } ) ) );
    Dom.onSubmit( Dom.get( "print-form" ) , ( event ) => this.make( event ) );
    window.addEventListener( "afterprint" , () => {
      document.body.classList.remove( "printing" );
      Dom.swap( Dom.get( "print-area" ) );
    } );
    window.addEventListener( "household" , () => { if ( Nav.current() === "print" ) this.refresh(); } );
    window.addEventListener( "nav" , ( event ) => { if ( event.detail.screen === "print" ) this.refresh(); } );
  },

  async refresh() {
    if ( !Org.current ) return;
    Dom.show( Dom.get( "print-form" ).closest( ".card" ) , Org.canEdit );
    const page = await Pager.load( "sheets-table" , "sheets" , ( offset ) => Org.get( "/sheets" , { offset: offset } ) , this.renderRow.bind( this ) );
    if ( page ) Dom.show( Dom.get( "sheets-empty" ) , page.total === 0 );
  },

  renderRow( sheet ) {
    const row = Dom.el( "tr" , { children: [
      Dom.el( "td" , { children: [ Dom.when( sheet.created_at ) ] } ),
      Dom.el( "td" , { text: String( sheet.codes.length ) } ),
      Dom.el( "td" , { text: I18n.get( "print.used_of" ).replace( "{used}" , sheet.used ).replace( "{count}" , sheet.codes.length ) } ),
      Dom.el( "td" , { class: "text-nowrap" , children: [ Dom.rowActions( I18n.get( "print.run" ) , [
        Org.button( "print.print_again" , "printer" , "btn-outline-primary" , () => this.sheet( sheet ) ),
        Org.canEdit ? Org.button( "print.forget" , "trash3" , "btn-outline-danger" , async () => {
          if ( !await Dom.confirm( I18n.get( "print.forget_confirm" ) ) ) return;
          try {
            await Org.post( "/sheets/" + sheet.id + "/delete" );
            await this.refresh();
          } catch ( error ) { Org.failed( error ); }
        } ) : null,
      ] ) ] } ),
    ] } );
    Dom.get( "sheets-table" ).querySelector( "tbody" ).appendChild( row );
  },

  async make( event ) {
    event.preventDefault();
    const names = Dom.get( "print-names" ).value.split( /\n|,/ ).map( ( each ) => each.trim() );
    try {
      const sheet = await Org.post( "/sheets" , {
        count: Number( Dom.get( "print-count" ).value ),
        layout: Dom.get( "print-layout" ).value,
        palette: Dom.get( "print-palette" ).value,
        names: names.filter( ( each , index ) => index < names.length && ( each || names.some( Boolean ) ) ),
      } );
      Dom.get( "print-names" ).value = "";
      await this.refresh();
      this.sheet( sheet );
    } catch ( error ) { Org.failed( error ); }
  },

  // Colors for a run: the palette's colors in turn.
  colorFor( palette , index ) {
    const colors = Org.palettes[ palette ];
    return colors ? colors[ index % colors.length ] : "";
  },

  sheet( sheet ) {
    const labels = sheet.codes.map( ( code , index ) => ( { code: code , name: sheet.names[ index ] || "" , color: this.colorFor( sheet.palette , index ) , flags: [] } ) );
    this.print( labels , sheet.layout );
  },

  // Prints containers' own labels again: a lost or worn sticker.
  async reprint( labelIds ) {
    try {
      const answer = await Org.post( "/labels/reprint" , { label_ids: labelIds } );
      const labels = answer.labels.map( ( label ) => ( { code: label.code , name: label.name , color: label.color , flags: label.flags } ) );
      this.print( labels , Dom.get( "print-layout" ).value || "letter-2.5" );
    } catch ( error ) { Org.failed( error ); }
  },

  inches( value ) {
    return value + "in";
  },

  labelNode( label , layout ) {
    const address = window.location.origin + "/l/" + label.code.replace( "-" , "" );
    const canvas = document.createElement( "canvas" );
    Qr.draw( canvas , address );
    const qr = Dom.el( "img" , { class: "print-qr" , attrs: { src: canvas.toDataURL( "image/png" ) , alt: "" } } );
    const words = Dom.el( "div" , { class: "print-words" , children: [
      label.name ? Dom.el( "div" , { class: "print-name" , text: label.name } ) : null,
      Dom.el( "div" , { class: "print-code" , text: label.code } ),
      label.flags.length ? Dom.el( "div" , { class: "print-flags" , text: label.flags.join( " · " ).toUpperCase() } ) : null,
    ].filter( Boolean ) } );
    const node = Dom.el( "div" , { class: "print-label" + ( layout.wide ? " print-label-wide" : "" ) , children: [ qr , words ] } );
    node.style.width = this.inches( layout.w );
    node.style.height = this.inches( layout.h );
    if ( label.color ) node.style.borderColor = label.color;
    else node.classList.add( "print-label-mono" );
    return node;
  },

  print( labels , layoutKey ) {
    const layout = this.layouts[ layoutKey ] || this.layouts[ "letter-2.5" ];
    const page = this.pages[ layout.page ];
    const perPage = layout.cols * layout.rows;
    const sheets = [];
    for ( let start = 0 ; start < labels.length ; start += perPage ) {
      const sheet = Dom.el( "div" , { class: "print-sheet" , children: labels.slice( start , start + perPage ).map( ( label ) => this.labelNode( label , layout ) ) } );
      const usedW = layout.cols * layout.w + ( layout.cols - 1 ) * layout.gap;
      const usedH = layout.rows * layout.h + ( layout.rows - 1 ) * layout.gap;
      sheet.style.width = this.inches( page.w );
      sheet.style.height = this.inches( page.h );
      sheet.style.padding = this.inches( Math.max( 0 , ( page.h - usedH ) / 2 ) ) + " " + this.inches( Math.max( 0 , ( page.w - usedW ) / 2 ) );
      sheet.style.gap = this.inches( layout.gap );
      sheet.style.gridTemplateColumns = "repeat(" + layout.cols + ", " + this.inches( layout.w ) + ")";
      sheet.style.gridAutoRows = this.inches( layout.h );
      sheets.push( sheet );
    }
    Dom.swap( Dom.get( "print-area" ) , ...sheets );
    document.body.classList.add( "printing" );
    // A frame for the images to settle before the print dialog takes a copy.
    window.setTimeout( () => window.print() , 150 );
  },

  // A report (Reports screen) printed the same way, on plain pages.
  document( title , nodes ) {
    const page = Dom.el( "div" , { class: "print-document" , children: [ Dom.el( "h1" , { class: "h3" , text: title } ) , ...nodes ] } );
    Dom.swap( Dom.get( "print-area" ) , page );
    document.body.classList.add( "printing" );
    window.setTimeout( () => window.print() , 150 );
  },
};
