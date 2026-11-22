// The organizer's shared state and helpers: which household is open, calls to
// its routes, photo uploads, colors and pickers. Every organizer screen (scan,
// inventory, labels, container, locations, print, reminders, reports,
// households, trash) builds on this and listens for the "household" event,
// sent whenever a different household is chosen or this one changed.
const Org = {
  me: null,
  households: [],
  current: null,
  storageKey: "household",
  // The label colors each print palette uses, as SmartLabels' packs do.
  palettes: {
    original: [ "#d32f2f" , "#388e3c" , "#1976d2" , "#fbc02d" ],
    neon: [ "#ff6d00" , "#ffea00" , "#00e676" , "#ff4081" ],
    modern: [ "#7b1fa2" , "#00acc1" , "#ec407a" , "#ff7043" ],
  },
  // Photos are shrunk in the browser to this many pixels on the long side
  // before they are sent, and a thumbnail is made beside them.
  photoSide: 1600,

  async init( me ) {
    this.me = me;
    Dom.show( Dom.get( "sidebar-org" ) , true );
    Dom.get( "household-select" ).addEventListener( "change" , ( event ) => this.choose( Number( event.target.value ) ) );
    Dom.onSubmit( Dom.get( "welcome-form" ) , async ( event ) => {
      event.preventDefault();
      await this.create( Dom.get( "welcome-name" ).value );
      Nav.go( "/scan" );
    } );
    await this.refresh();
  },

  // Loads the household list and opens the remembered one (or the first).
  async refresh() {
    const page = await Api.request( "/api/households?limit=1000" );
    this.households = page.households;
    const saved = Number( this.remembered() );
    const pick = this.households.find( ( each ) => each.id === ( this.current ? this.current.id : saved ) ) || this.households[ 0 ] || null;
    this.current = pick;
    this.paintSelect();
    this.showScreens();
    if ( pick ) this.remember( pick.id );
    window.dispatchEvent( new CustomEvent( "household" ) );
  },

  showScreens() {
    const has = Boolean( this.current );
    Dom.show( Dom.get( "welcome-panel" ) , !has );
    [ "scan" , "inventory" , "labels" , "container" , "locations" , "print" , "reminders" , "reports" , "households" , "trash" ].forEach( ( name ) => {
      Dom.show( Dom.get( name + "-panel" ) , has );
    } );
    Nav.fallback = has ? ( window.matchMedia( "(max-width: 767.98px)" ).matches ? "scan" : "inventory" ) : "welcome";
    Dom.show( Dom.get( "household-select" ) , this.households.length > 0 );
  },

  paintSelect() {
    const select = Dom.get( "household-select" );
    const options = this.households.map( ( household ) => Dom.el( "option" , { text: household.name , attrs: { value: String( household.id ) } } ) );
    Dom.swap( select , ...options );
    if ( this.current ) select.value = String( this.current.id );
  },

  choose( id ) {
    const next = this.households.find( ( each ) => each.id === id );
    if ( !next || ( this.current && this.current.id === id ) ) return;
    this.current = next;
    this.remember( id );
    this.paintSelect();
    window.dispatchEvent( new CustomEvent( "household" ) );
  },

  // Replaces the open household's record after a change to it.
  updated( household ) {
    this.households = this.households.map( ( each ) => each.id === household.id ? household : each );
    if ( this.current && this.current.id === household.id ) this.current = household;
    this.paintSelect();
    window.dispatchEvent( new CustomEvent( "household" ) );
  },

  async create( name ) {
    const made = await Api.post( "/api/households" , { name: name } );
    this.current = made;
    this.remember( made.id );
    await this.refresh();
    return made;
  },

  remembered() {
    try { return window.localStorage.getItem( this.storageKey ); } catch ( error ) { return null; }
  },

  remember( id ) {
    try { window.localStorage.setItem( this.storageKey , String( id ) ); } catch ( error ) { /* not remembered */ }
  },

  get id() { return this.current ? this.current.id : 0; },

  // Whether the signed-in person may change things here (owner or editor).
  get canEdit() { return Boolean( this.current ) && this.current.my_role !== "viewer"; },

  path( rest ) { return "/api/households/" + this.id + rest; },

  get( rest , params ) { return Api.request( this.path( rest ) + ( params ? Api.query( params ) : "" ) ); },

  post( rest , body ) { return Api.post( this.path( rest ) , body ); },

  photoUrl( fileId ) { return this.path( "/photos/" + encodeURIComponent( fileId ) ); },

  // ---- Photos ----------------------------------------------------------

  // A picked or taken photo, shrunk to photoSide pixels on its long side,
  // re-encoded as JPEG (which also drops its location data), and uploaded.
  // Answers the new file's id. The server attaches it to this household when
  // a container or item names it.
  async uploadPhoto( file , side ) {
    const blob = await this.shrink( file , side || this.photoSide );
    const name = ( file.name || "photo" ).replace( /\.[^.]*$/ , "" ) + ".jpg";
    return this.upload( new File( [ blob ] , name , { type: "image/jpeg" } ) );
  },

  async upload( file ) {
    const created = await Api.createFile( file.name , file.size , file.type );
    try {
      let next = 0;
      while ( next < created.chunks ) {
        const start = next * created.chunk_bytes;
        const sent = await Api.sendFileChunk( created.id , next , file.slice( start , start + created.chunk_bytes ) );
        next = sent.received;
      }
      await Api.finishFile( created.id );
    } catch ( error ) {
      await Api.deleteFile( created.id ).catch( () => {} );
      throw error;
    }
    return created.id;
  },

  async loadImage( source ) {
    const url = source instanceof Blob ? URL.createObjectURL( source ) : source;
    try {
      const image = new Image();
      image.decoding = "async";
      image.src = url;
      await image.decode();
      return image;
    } finally {
      if ( source instanceof Blob ) window.setTimeout( () => URL.revokeObjectURL( url ) , 1000 );
    }
  },

  async shrink( file , side ) {
    const image = await this.loadImage( file );
    const scale = Math.min( 1 , side / Math.max( image.naturalWidth , image.naturalHeight ) );
    const canvas = document.createElement( "canvas" );
    canvas.width = Math.max( 1 , Math.round( image.naturalWidth * scale ) );
    canvas.height = Math.max( 1 , Math.round( image.naturalHeight * scale ) );
    canvas.getContext( "2d" ).drawImage( image , 0 , 0 , canvas.width , canvas.height );
    return new Promise( ( resolve , reject ) => canvas.toBlob( ( blob ) => blob ? resolve( blob ) : reject( new Error( I18n.get( "org.photo_failed" ) ) ) , "image/jpeg" , 0.85 ) );
  },

  // Cuts one item out of a photo, from a box given as fractions of the
  // photo (x, y, width, height), with a little room around it.
  async crop( photoUrl , box ) {
    const image = await this.loadImage( photoUrl );
    const pad = 0.04;
    const x = Math.max( 0 , box[ 0 ] - pad ) * image.naturalWidth;
    const y = Math.max( 0 , box[ 1 ] - pad ) * image.naturalHeight;
    const width = Math.min( 1 , box[ 2 ] + pad * 2 ) * image.naturalWidth;
    const height = Math.min( 1 , box[ 3 ] + pad * 2 ) * image.naturalHeight;
    const canvas = document.createElement( "canvas" );
    canvas.width = Math.max( 1 , Math.round( Math.min( width , image.naturalWidth - x ) ) );
    canvas.height = Math.max( 1 , Math.round( Math.min( height , image.naturalHeight - y ) ) );
    canvas.getContext( "2d" ).drawImage( image , x , y , canvas.width , canvas.height , 0 , 0 , canvas.width , canvas.height );
    return new Promise( ( resolve ) => canvas.toBlob( resolve , "image/jpeg" , 0.85 ) );
  },

  // An <input type=file> for photos: on a phone it offers the camera.
  photoInput( onFiles , multiple ) {
    const input = Dom.el( "input" , { class: "visually-hidden" , attrs: { type: "file" , accept: "image/*" } } );
    if ( multiple ) input.multiple = true;
    input.addEventListener( "change" , async () => {
      const files = Array.from( input.files );
      input.value = "";
      if ( files.length ) await onFiles( files );
    } );
    return input;
  },

  // A small square photo, which opens full size in a new tab.
  thumb( fileId , size ) {
    const image = Dom.el( "img" , { class: "org-thumb rounded border" , attrs: { src: this.photoUrl( fileId ) , alt: "" , loading: "lazy" } } );
    if ( size === "lg" ) image.classList.add( "org-thumb-lg" );
    return Dom.el( "a" , { attrs: { href: this.photoUrl( fileId ) , target: "_blank" , rel: "noopener" } , children: [ image ] } );
  },

  // ---- Colors, places, containers ---------------------------------------

  // A colored dot for a label color.
  swatch( color ) {
    const dot = Dom.el( "span" , { class: "org-swatch" , attrs: { "aria-hidden": "true" } } );
    if ( color ) dot.style.backgroundColor = color;
    else dot.classList.add( "org-swatch-none" );
    return dot;
  },

  // Every color worth offering: the palettes, then any the household named.
  colors() {
    const seen = new Set();
    const list = [];
    Object.values( this.palettes ).flat().concat( Object.keys( ( this.current && this.current.colors ) || {} ) ).forEach( ( color ) => {
      if ( !seen.has( color ) ) { seen.add( color ); list.push( color ); }
    } );
    return list;
  },

  colorName( color ) {
    if ( !color ) return I18n.get( "org.no_color" );
    const meaning = this.current && this.current.colors[ color ];
    const names = { "#d32f2f": "red" , "#388e3c": "green" , "#1976d2": "blue" , "#fbc02d": "yellow" , "#ff6d00": "orange" , "#ffea00": "yellow" ,
      "#00e676": "green" , "#ff4081": "pink" , "#7b1fa2": "purple" , "#00acc1": "turquoise" , "#ec407a": "pink" , "#ff7043": "orange" };
    const base = names[ color ] ? I18n.get( "org.color_" + names[ color ] ) : color;
    return meaning ? base + " — " + meaning : base;
  },

  fillColors( select , chosen ) {
    const options = [ Dom.el( "option" , { text: I18n.get( "org.no_color" ) , attrs: { value: "" } } ) ]
      .concat( this.colors().map( ( color ) => Dom.el( "option" , { text: this.colorName( color ) , attrs: { value: color } } ) ) );
    Dom.swap( select , ...options );
    select.value = chosen || "";
  },

  // The household's locations, shared by every picker on the page for a few
  // seconds, so opening a screen does not ask for the same list three times.
  locationCache: null,

  locations() {
    const now = Date.now();
    const cache = this.locationCache;
    if ( cache && cache.id === this.id && now - cache.at < 5000 ) return cache.list;
    const list = this.get( "/locations" , { limit: 1000 } ).then( ( page ) => page.locations );
    list.catch( () => { this.locationCache = null; } );
    this.locationCache = { id: this.id , at: now , list: list };
    return list;
  },

  // A location was added, changed or deleted: the next picker asks again.
  forgetLocations() {
    this.locationCache = null;
  },

  async fillLocations( select , chosen , noneKey , skipId ) {
    const locations = await this.locations();
    const options = [ Dom.el( "option" , { text: I18n.get( noneKey || "org.no_location" ) , attrs: { value: "0" } } ) ]
      .concat( locations.filter( ( each ) => each.id !== skipId ).map( ( each ) => Dom.el( "option" , { text: each.path , attrs: { value: String( each.id ) } } ) ) );
    Dom.swap( select , ...options );
    select.value = String( chosen || 0 );
    return locations;
  },

  async labels() {
    const page = await this.get( "/labels" , { limit: 1000 , sort: "path" } );
    return page.labels;
  },

  // Asks which container, in a dialog. Answers its id, or null.
  async pickLabel( title , skipId ) {
    const labels = ( await this.labels() ).filter( ( each ) => each.id !== skipId );
    const select = Dom.el( "select" , { class: "form-select" , attrs: { id: "pick-label-select" } ,
      children: labels.map( ( each ) => Dom.el( "option" , { text: each.path + "  (" + each.printed_code + ")" , attrs: { value: String( each.id ) } } ) ) } );
    if ( labels.length === 0 ) {
      Dom.toast( I18n.get( "org.no_other_containers" ) , "error" );
      return null;
    }
    const chosen = await this.dialog( title , [
      Dom.el( "label" , { class: "form-label" , text: I18n.get( "org.container_label" ) , attrs: { for: "pick-label-select" } } ) , select ,
    ] , () => Number( select.value ) );
    return chosen;
  },

  // Asks for one line of text, in the app's own dialog. Answers it, or null.
  ask( title , labelKey , value ) {
    const input = Dom.el( "input" , { class: "form-control" , attrs: { id: "org-ask-input" , maxlength: "80" , required: "" } } );
    input.value = value || "";
    const asked = this.dialog( title , [
      Dom.el( "label" , { class: "form-label" , text: I18n.get( labelKey ) , attrs: { for: "org-ask-input" } } ) , input ,
    ] , () => input.value.trim() || null );
    window.setTimeout( () => input.focus() , 300 );
    return asked;
  },

  // A small dialog with a body and an OK button. Answers what read() gives
  // when OK is pressed, or null.
  dialog( title , body , read ) {
    return new Promise( ( resolve ) => {
      const ok = Dom.el( "button" , { class: "btn btn-primary" , text: I18n.get( "ui.ok" ) || "OK" , attrs: { type: "submit" } } );
      const form = Dom.el( "form" , { class: "modal-content" , children: [
        Dom.el( "div" , { class: "modal-header" , children: [
          Dom.el( "h2" , { class: "modal-title h5" , text: title } ),
          Dom.el( "button" , { class: "btn-close" , attrs: { type: "button" , "data-bs-dismiss": "modal" , "aria-label": I18n.get( "nav.close" ) } } ),
        ] } ),
        Dom.el( "div" , { class: "modal-body" , children: body } ),
        Dom.el( "div" , { class: "modal-footer" , children: [
          Dom.el( "button" , { class: "btn btn-outline-secondary" , text: I18n.get( "ui.cancel" ) , attrs: { type: "button" , "data-bs-dismiss": "modal" } } ) , ok ,
        ] } ),
      ] } );
      const holder = Dom.el( "div" , { class: "modal fade" , attrs: { tabindex: "-1" } , children: [ Dom.el( "div" , { class: "modal-dialog" , children: [ form ] } ) ] } );
      document.body.appendChild( holder );
      const modal = new bootstrap.Modal( holder );
      let answer = null;
      form.addEventListener( "submit" , ( event ) => {
        event.preventDefault();
        answer = read();
        modal.hide();
      } );
      holder.addEventListener( "hidden.bs.modal" , () => {
        holder.remove();
        resolve( answer );
      } );
      modal.show();
    } );
  },

  // ---- Numbers ----------------------------------------------------------

  money( value ) {
    return new Intl.NumberFormat( undefined , { style: "currency" , currency: I18n.get( "org.currency" ) || "USD" } ).format( value || 0 );
  },

  amount( value ) {
    return new Intl.NumberFormat( undefined , { maximumFractionDigits: 3 } ).format( value || 0 );
  },

  date( value ) {
    if ( !value ) return "";
    const parsed = new Date( value + "T00:00:00" );
    return isNaN( parsed.getTime() ) ? value : parsed.toLocaleDateString();
  },

  // "In 3 days", "Today", "2 days ago", as a badge colored by how soon.
  dueBadge( days ) {
    if ( days === null || days === undefined ) return null;
    const format = new Intl.RelativeTimeFormat( undefined , { numeric: "auto" } );
    const tone = days < 0 ? "danger" : days <= 7 ? "warning" : "secondary";
    return Dom.badge( format.format( days , "day" ) , tone );
  },

  icon( name ) {
    return Dom.el( "i" , { class: "bi bi-" + name , attrs: { "aria-hidden": "true" } } );
  },

  // A button with an icon and a label.
  button( key , icon , kind , onClick ) {
    return Dom.el( "button" , { class: "btn btn-sm " + ( kind || "btn-outline-secondary" ) , attrs: { type: "button" } ,
      children: [ this.icon( icon ) , Dom.el( "span" , { class: "ms-1" , text: I18n.get( key ) } ) ] , on: { click: onClick } } );
  },

  // A link to a container's page.
  labelLink( id , text ) {
    return Dom.el( "a" , { text: text , attrs: { href: "/labels/" + id } } );
  },

  // Reports a failed call, the way every screen does.
  failed( error ) {
    App.showError( error.message );
  },
};
