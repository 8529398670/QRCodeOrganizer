// The Scan screen: point the camera at a label (or type its code) and land on
// what it is. A phone's camera app opening a printed label's address
// (/l/<code>) lands here too, already looked up.
//
// Reading uses the browser's own BarcodeDetector where there is one (Chrome on
// Android), and the vendored ZXing library everywhere else (Safari on iPhone,
// Firefox). Both read QR codes and the usual product barcodes, so a bought
// sticker or a box's UPC can be claimed for a container too.
const Scan = {
  stream: null,
  reader: null,
  detector: null,
  timer: null,
  lastCode: "",

  init() {
    Dom.onPress( Dom.get( "scan-start" ) , () => this.start() );
    Dom.get( "scan-stop" ).addEventListener( "click" , () => this.stop() );
    Dom.onSubmit( Dom.get( "scan-form" ) , async ( event ) => {
      event.preventDefault();
      await this.lookup( Dom.get( "scan-input" ).value );
    } );
    window.addEventListener( "nav" , ( event ) => {
      if ( event.detail.screen !== "scan" ) {
        this.stop();
        return;
      }
      if ( Nav.param ) this.lookup( Nav.param );
    } );
  },

  async start() {
    Dom.show( Dom.get( "scan-camera-error" ) , false );
    if ( !navigator.mediaDevices || !navigator.mediaDevices.getUserMedia ) {
      this.cameraError( I18n.get( "scan.no_camera_https" ) );
      return;
    }
    const video = Dom.get( "scan-video" );
    try {
      if ( "BarcodeDetector" in window ) {
        this.detector = new window.BarcodeDetector( { formats: [ "qr_code" , "ean_13" , "ean_8" , "upc_a" , "upc_e" , "code_128" , "code_39" , "data_matrix" ] } );
        this.stream = await navigator.mediaDevices.getUserMedia( { video: { facingMode: "environment" } , audio: false } );
        video.srcObject = this.stream;
        await video.play();
        this.timer = window.setInterval( () => this.detect( video ) , 250 );
      } else if ( window.ZXing ) {
        this.reader = new ZXing.BrowserMultiFormatReader();
        await this.reader.decodeFromConstraints( { video: { facingMode: "environment" } } , video , ( result ) => {
          if ( result ) this.found( result.getText() );
        } );
      } else {
        this.cameraError( I18n.get( "scan.no_reader" ) );
        return;
      }
    } catch ( error ) {
      this.stop();
      this.cameraError( I18n.get( "scan.camera_refused" ) );
      return;
    }
    Dom.show( Dom.get( "scan-camera-box" ) , true );
    Dom.show( Dom.get( "scan-stop" ) , true );
    Dom.show( Dom.get( "scan-start" ) , false );
  },

  async detect( video ) {
    if ( !this.detector || video.readyState < 2 ) return;
    try {
      const codes = await this.detector.detect( video );
      if ( codes.length ) this.found( codes[ 0 ].rawValue );
    } catch ( error ) { /* a frame it could not read */ }
  },

  // A code seen by the camera. The same one seen again straight away is the
  // same label still in front of it.
  found( text ) {
    if ( !text || text === this.lastCode ) return;
    this.lastCode = text;
    window.setTimeout( () => { if ( this.lastCode === text ) this.lastCode = ""; } , 4000 );
    if ( navigator.vibrate ) navigator.vibrate( 60 );
    this.stop();
    this.lookup( text );
  },

  stop() {
    window.clearInterval( this.timer );
    this.timer = null;
    if ( this.reader ) {
      try { this.reader.reset(); } catch ( error ) { /* already stopped */ }
      this.reader = null;
    }
    if ( this.stream ) {
      this.stream.getTracks().forEach( ( track ) => track.stop() );
      this.stream = null;
    }
    this.detector = null;
    const video = Dom.get( "scan-video" );
    if ( video ) video.srcObject = null;
    Dom.show( Dom.get( "scan-camera-box" ) , false );
    Dom.show( Dom.get( "scan-stop" ) , false );
    Dom.show( Dom.get( "scan-start" ) , true );
  },

  cameraError( text ) {
    Dom.text( Dom.get( "scan-camera-error" ) , text );
    Dom.show( Dom.get( "scan-camera-error" ) , true );
  },

  async lookup( raw ) {
    const code = ( raw || "" ).trim();
    if ( !code ) return;
    try {
      const answer = await Api.request( "/api/scan" + Api.query( { code: code } ) );
      // A code from another household of yours: open that one.
      if ( answer.household_id && answer.household_id !== Org.id ) Org.choose( answer.household_id );
      await this.show( answer , code );
    } catch ( error ) {
      Org.failed( error );
    }
  },

  async show( answer , raw ) {
    let body;
    if ( answer.kind === "label" ) body = this.labelResult( answer );
    else if ( answer.kind === "unassigned" ) body = await this.unassignedResult( answer );
    else body = await this.unknownResult( answer , raw );
    Dom.swap( Dom.get( "scan-result-body" ) , ...body );
    Dom.show( Dom.get( "scan-result" ) , true );
    Dom.reveal( Dom.get( "scan-result" ) , "start" );
  },

  labelResult( answer ) {
    const label = answer.label;
    const facts = Dom.el( "p" , { class: "mb-2" , children: [
      Org.swatch( label.color ),
      Dom.el( "span" , { class: "ms-2" , text: I18n.get( "scan.holds" ).replace( "{count}" , label.item_count ) } ),
      Dom.el( "span" , { class: "ms-2 text-body-secondary" , text: label.printed_code } ),
    ] } );
    const parts = [
      Dom.el( "h3" , { class: "card-title" , text: label.name } ),
      Dom.el( "p" , { class: "text-body-secondary mb-1" , text: answer.household + " › " + ( label.path || label.name ) } ),
      facts,
    ];
    if ( answer.in_trash ) parts.push( Dom.el( "p" , { class: "text-warning" , text: I18n.get( "scan.in_trash" ) } ) );
    parts.push( Dom.el( "div" , { class: "d-flex flex-wrap gap-2" , children: [
      Dom.el( "a" , { class: "btn btn-primary" , attrs: { href: "/labels/" + label.id } , children: [ Org.icon( "box-seam" ) , Dom.el( "span" , { class: "ms-1" , text: I18n.get( "scan.open" ) } ) ] } ),
      answer.can_edit ? Org.button( "scan.add_item" , "plus-lg" , "btn-outline-primary" , async () => {
        if ( await ItemEditor.open( null , label.id ) ) Dom.toast( I18n.get( "scan.item_added" ) , "ok" );
      } ) : null,
    ].filter( Boolean ) } ) );
    return parts;
  },

  // A printed code that is not on anything yet: name the container it is on.
  async unassignedResult( answer ) {
    if ( !answer.can_edit ) return [ Dom.el( "p" , { text: I18n.get( "scan.unassigned_viewer" ) } ) ];
    const name = Dom.el( "input" , { class: "form-control" , attrs: { id: "scan-new-name" , maxlength: "80" , required: "" } } );
    const location = Dom.el( "select" , { class: "form-select" , attrs: { id: "scan-new-location" } } );
    const color = Dom.el( "select" , { class: "form-select" , attrs: { id: "scan-new-color" } } );
    await Org.fillLocations( location );
    Org.fillColors( color );
    const form = Dom.el( "form" , { class: "row g-2 align-items-end" , children: [
      Dom.el( "div" , { class: "col-12 col-md-5" , children: [ Dom.el( "label" , { class: "form-label" , text: I18n.get( "labels.name_label" ) , attrs: { for: "scan-new-name" } } ) , name ] } ),
      Dom.el( "div" , { class: "col-6 col-md-3" , children: [ Dom.el( "label" , { class: "form-label" , text: I18n.get( "labels.location_label" ) , attrs: { for: "scan-new-location" } } ) , location ] } ),
      Dom.el( "div" , { class: "col-6 col-md-2" , children: [ Dom.el( "label" , { class: "form-label" , text: I18n.get( "labels.color_label" ) , attrs: { for: "scan-new-color" } } ) , color ] } ),
      Dom.el( "div" , { class: "col-12 col-md-2 d-grid" , children: [ Dom.el( "button" , { class: "btn btn-primary" , text: I18n.get( "scan.set_up" ) , attrs: { type: "submit" } } ) ] } ),
    ] } );
    Dom.onSubmit( form , async ( event ) => {
      event.preventDefault();
      try {
        const made = await Org.post( "/labels" , { code: answer.code , name: name.value , location_id: Number( location.value ) , color: color.value } );
        Nav.go( "/labels/" + made.id );
      } catch ( error ) { Org.failed( error ); }
    } );
    return [
      Dom.el( "h3" , { class: "card-title" , text: I18n.get( "scan.unassigned_heading" ).replace( "{code}" , answer.code ) } ),
      Dom.el( "p" , { text: I18n.get( "scan.unassigned_text" ) } ),
      form,
    ];
  },

  // Not one of yours: perhaps a store-bought sticker or a barcode to claim.
  async unknownResult( answer , raw ) {
    const parts = [ Dom.el( "h3" , { class: "card-title" , text: I18n.get( "scan.unknown_heading" ) } ) ];
    parts.push( Dom.el( "p" , { class: "text-break" , children: [ Dom.el( "code" , { text: raw } ) ] } ) );
    if ( /^[0-9A-Za-z]{5}-?[0-9A-Za-z]{5}$/.test( raw ) || /\/l\/[^/]+$/.test( raw ) ) {
      parts.push( Dom.el( "p" , { text: I18n.get( "scan.not_yours" ) } ) );
      return parts;
    }
    if ( !Org.canEdit ) return parts;
    parts.push( Dom.el( "p" , { text: I18n.get( "scan.claim_text" ) } ) );
    parts.push( Dom.el( "div" , { class: "d-flex flex-wrap gap-2" , children: [
      Org.button( "scan.claim_new" , "plus-square" , "btn-primary" , async () => {
        const name = await Org.ask( I18n.get( "scan.claim_new" ) , "labels.name_label" );
        if ( !name ) return;
        try {
          const made = await Org.post( "/labels" , { code: raw , name: name } );
          Nav.go( "/labels/" + made.id );
        } catch ( error ) { Org.failed( error ); }
      } ),
      Org.button( "scan.claim_existing" , "link-45deg" , "btn-outline-primary" , async () => {
        const target = await Org.pickLabel( I18n.get( "scan.claim_existing" ) );
        if ( !target ) return;
        try {
          await Org.post( "/labels/" + target + "/codes" , { code: raw } );
          Nav.go( "/labels/" + target );
        } catch ( error ) { Org.failed( error ); }
      } ),
    ] } ) );
    return parts;
  },
};
