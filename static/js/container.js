// One container's page, /labels/<id>: what it is and where, what is in it,
// its photos and codes, and Smart Photo Extraction -- a photo of everything
// going in, read by Claude into a list to tick through before it is saved.
const Container = {
  label: null,
  aiReady: false,
  polling: null,

  init() {
    window.addEventListener( "nav" , ( event ) => {
      if ( event.detail.screen !== "container" ) return;
      if ( !Nav.param ) {
        Nav.go( "/labels" );
        return;
      }
      this.load( Number( Nav.param ) );
    } );
    window.addEventListener( "inventory-changed" , () => { if ( Nav.current() === "container" && this.label ) this.load( this.label.id ); } );
    window.addEventListener( "household" , () => { if ( Nav.current() === "container" && this.label && this.label.household_id !== Org.id ) Nav.go( "/labels" ); } );
  },

  async load( id ) {
    try {
      const [ label , ai ] = await Promise.all( [ Org.get( "/labels/" + id ) , Api.request( "/api/ai" ) ] );
      this.label = label;
      this.aiReady = ai.ready;
      Dom.text( Dom.get( "container-title" ) , label.name );
      Dom.text( Dom.get( "container-path" ) , [ Org.current.name , label.path ].join( " › " ) );
      const heading = document.querySelector( '.view[data-view="container"] h2' );
      if ( heading ) document.title = label.name + " · " + I18n.get( "app.title" );
      Dom.swap( Dom.get( "container-body" ) , this.summaryCard( label ) , this.itemsCard( label ) , this.photoCard( label ) , this.detailsCard( label ) , this.codesCard( label ) );
    } catch ( error ) {
      if ( error.status === 404 ) {
        Dom.text( Dom.get( "container-title" ) , I18n.get( "container.missing" ) );
        Dom.swap( Dom.get( "container-body" ) );
        return;
      }
      Org.failed( error );
    }
  },

  card( title , children ) {
    return Dom.el( "section" , { class: "card" , children: [ Dom.el( "div" , { class: "card-body" , children: [
      title ? Dom.el( "h3" , { class: "card-title" , text: title } ) : null , ...children ,
    ].filter( Boolean ) } ) ] } );
  },

  summaryCard( label ) {
    const facts = [
      [ "container.code" , label.printed_code ],
      [ "container.color" , Org.colorName( label.color ) ],
      [ "container.items" , String( label.item_count ) ],
      [ "container.value" , Org.money( label.value ) ],
    ];
    const list = Dom.el( "dl" , { class: "row mb-2" , children: facts.flatMap( ( [ key , value ] ) => [
      Dom.el( "dt" , { class: "col-5 col-md-3 fw-normal text-body-secondary" , text: I18n.get( key ) } ),
      Dom.el( "dd" , { class: "col-7 col-md-9 mb-1" , text: value } ),
    ] ) } );
    const flags = label.flags.length || label.move_status ? Dom.el( "div" , { class: "d-flex flex-wrap gap-1 mb-2" , children: [
      label.move_status ? Dom.badge( I18n.get( Labels.moving[ label.move_status ] ) + ( label.destination ? " → " + label.destination : "" ) , "warning" ) : null,
      ...label.flags.map( ( flag ) => Dom.badge( flag , "danger" ) ),
    ].filter( Boolean ) } ) : null;
    const buttons = [
      Org.button( "container.print" , "printer" , null , () => Print.reprint( [ label.id ] ) ),
    ];
    if ( Org.canEdit ) {
      buttons.push( Org.button( "container.move_all" , "box-arrow-right" , null , () => this.moveAll( label ) ) );
      buttons.push( Org.button( "container.delete" , "trash3" , "btn-outline-danger" , () => this.remove( label ) ) );
    }
    const children = label.children.length ? Dom.el( "p" , { class: "small mb-2" , children: [
      Dom.el( "span" , { text: I18n.get( "container.inside" ) + " " } ),
      ...label.children.flatMap( ( child , index ) => [ index ? Dom.el( "span" , { text: ", " } ) : null , Org.labelLink( child.id , child.name ) ].filter( Boolean ) ),
    ] } ) : null;
    return this.card( null , [ flags , list , children , label.notes ? Dom.el( "p" , { class: "small" , text: label.notes } ) : null ,
      Dom.el( "div" , { class: "d-flex flex-wrap gap-2" , children: buttons } ) ].filter( Boolean ) );
  },

  itemsCard( label ) {
    const rows = label.items.map( ( item ) => Dom.el( "tr" , { attrs: { "data-key": "item-" + item.id } , children: [
      Dom.el( "td" , { children: [ Inventory.nameCell( item ) ] } ),
      Dom.el( "td" , { class: "text-nowrap" , text: Org.amount( item.quantity ) + ( item.unit ? " " + item.unit : "" ) } ),
      Dom.el( "td" , { class: "text-nowrap" , children: item.expires ? [ Dom.el( "div" , { text: Org.date( item.expires ) } ) , Org.dueBadge( item.expires_in ) ] : [] } ),
      Dom.el( "td" , { class: "text-nowrap" , children: [ ItemEditor.actions( item , () => this.load( label.id ) ) ] } ),
    ] } ) );
    const table = Dom.el( "div" , { class: "table-responsive" , children: [ Dom.el( "table" , { class: "table table-hover align-middle" , children: [
      Dom.el( "thead" , { children: [ Dom.el( "tr" , { children: [ "container.col_item" , "container.col_quantity" , "container.col_expires" , "container.col_actions" ]
        .map( ( key ) => Dom.el( "th" , { text: I18n.get( key ) , attrs: { scope: "col" } } ) ) } ) ] } ),
      Dom.el( "tbody" , { children: rows } ),
    ] } ) ] } );
    const tools = [];
    if ( Org.canEdit ) {
      tools.push( Org.button( "container.add_item" , "plus-lg" , "btn-primary" , async () => { if ( await ItemEditor.open( null , label.id ) ) await this.load( label.id ); } ) );
      if ( this.aiReady ) {
        const input = Org.photoInput( ( files ) => this.extract( label , files[ 0 ] ) );
        tools.push( Dom.el( "label" , { class: "btn btn-sm btn-outline-primary" , children: [ Org.icon( "magic" ) , Dom.el( "span" , { class: "ms-1" , text: I18n.get( "container.ai_button" ) } ) , input ] } ) );
      }
    }
    const holder = Dom.el( "div" , { attrs: { id: "container-ai" } } );
    return this.card( I18n.get( "container.items_heading" ) , [
      Dom.el( "div" , { class: "d-flex flex-wrap gap-2 mb-3" , children: tools } ),
      this.aiReady || !Org.canEdit ? null : Dom.el( "p" , { class: "small text-body-secondary" , text: I18n.get( "container.ai_off" ) } ),
      holder,
      label.items.length ? table : Dom.el( "p" , { class: "text-body-secondary mb-0" , text: I18n.get( "container.empty" ) } ),
    ].filter( Boolean ) );
  },

  photoCard( label ) {
    const strip = Dom.el( "div" , { class: "d-flex flex-wrap gap-2" , children: label.photos.map( ( fileId ) => Org.thumb( fileId , "lg" ) ) } );
    const parts = [ label.photos.length ? strip : Dom.el( "p" , { class: "text-body-secondary small" , text: I18n.get( "container.no_photos" ) } ) ];
    if ( Org.canEdit ) {
      const input = Org.photoInput( async ( files ) => {
        try {
          const added = [];
          for ( const file of files ) added.push( await Org.uploadPhoto( file ) );
          await this.setPhotos( label , label.photos.concat( added ) );
        } catch ( error ) { Org.failed( error ); }
      } , true );
      parts.push( Dom.el( "div" , { class: "d-flex flex-wrap gap-2 mt-2" , children: [
        Dom.el( "label" , { class: "btn btn-sm btn-outline-secondary" , children: [ Org.icon( "camera" ) , Dom.el( "span" , { class: "ms-1" , text: I18n.get( "container.add_photo" ) } ) , input ] } ),
        label.photos.length ? Org.button( "container.clear_photos" , "x-lg" , null , () => this.setPhotos( label , [] ) ) : null,
      ].filter( Boolean ) } ) );
    }
    return this.card( I18n.get( "container.photos_heading" ) , parts );
  },

  async setPhotos( label , photos ) {
    const before = label.photos.slice();
    const send = ( list ) => Org.post( "/labels/" + label.id , { photos: list } );
    try {
      await send( photos );
      Undo.record( {
        label: I18n.get( "undo.photos" ),
        toast: true,
        undo: async () => { await send( before ); await this.load( label.id ); },
        redo: async () => { await send( photos ); await this.load( label.id ); },
      } );
      await this.load( label.id );
    } catch ( error ) { Org.failed( error ); }
  },

  // Name, color, type, place, outer container, notes, moving, flags.
  detailsCard( label ) {
    if ( !Org.canEdit ) return null;
    const id = ( name ) => "container-" + name;
    const input = ( name , value , attrs ) => {
      const node = Dom.el( "input" , { class: "form-control" , attrs: Object.assign( { id: id( name ) } , attrs || {} ) } );
      node.value = value || "";
      return node;
    };
    const field = ( name , key , node , width ) => Dom.el( "div" , { class: width || "col-12 col-md-6" , children: [
      Dom.el( "label" , { class: "form-label" , text: I18n.get( key ) , attrs: { for: id( name ) } } ) , node ,
    ] } );
    const color = Dom.el( "select" , { class: "form-select" , attrs: { id: id( "color" ) } } );
    Org.fillColors( color , label.color );
    const location = Dom.el( "select" , { class: "form-select" , attrs: { id: id( "location" ) } } );
    Org.fillLocations( location , label.location_id );
    const parent = Dom.el( "select" , { class: "form-select" , attrs: { id: id( "parent" ) } } );
    Org.labels().then( ( labels ) => {
      Dom.swap( parent , Dom.el( "option" , { text: I18n.get( "container.no_parent" ) , attrs: { value: "0" } } ) ,
        ...labels.filter( ( each ) => each.id !== label.id ).map( ( each ) => Dom.el( "option" , { text: each.path , attrs: { value: String( each.id ) } } ) ) );
      parent.value = String( label.parent_id || 0 );
    } );
    const moving = Dom.el( "select" , { class: "form-select" , attrs: { id: id( "moving" ) } , children: [
      Dom.el( "option" , { text: I18n.get( "container.not_moving" ) , attrs: { value: "" } } ),
      ...Object.entries( Labels.moving ).map( ( [ value , key ] ) => Dom.el( "option" , { text: I18n.get( key ) , attrs: { value: value } } ) ),
    ] } );
    moving.value = label.move_status || "";
    const notes = Dom.el( "textarea" , { class: "form-control" , text: label.notes || "" , attrs: { id: id( "notes" ) , rows: "2" , maxlength: "4000" } } );
    const form = Dom.el( "form" , { class: "row g-2" , children: [
      field( "name" , "labels.name_label" , input( "name" , label.name , { maxlength: "80" , required: "" } ) ),
      field( "type" , "container.type" , input( "type" , label.container_type , { maxlength: "40" , list: "container-types" } ) , "col-6 col-md-3" ),
      field( "color" , "labels.color_label" , color , "col-6 col-md-3" ),
      field( "location" , "labels.location_label" , location ),
      field( "parent" , "container.parent" , parent ),
      field( "moving" , "container.moving" , moving , "col-6 col-md-3" ),
      field( "destination" , "container.destination" , input( "destination" , label.destination , { maxlength: "80" } ) , "col-6 col-md-3" ),
      field( "flags" , "container.flags" , input( "flags" , label.flags.join( ", " ) , { placeholder: I18n.get( "container.flags_hint" ) } ) ),
      field( "notes" , "container.notes" , notes , "col-12" ),
      Dom.el( "datalist" , { attrs: { id: "container-types" } , children: [ "bin" , "box" , "drawer" , "shelf" , "bag" , "tote" , "cabinet" , "suitcase" ].map( ( each ) => Dom.el( "option" , { attrs: { value: each } } ) ) } ),
      Dom.el( "div" , { class: "col-12" , children: [ Dom.el( "button" , { class: "btn btn-primary" , text: I18n.get( "container.save" ) , attrs: { type: "submit" } } ) ] } ),
    ] } );
    const read = () => ( {
      name: Dom.get( id( "name" ) ).value,
      container_type: Dom.get( id( "type" ) ).value,
      color: color.value,
      location_id: Number( location.value ),
      parent_id: Number( parent.value ),
      move_status: moving.value,
      destination: Dom.get( id( "destination" ) ).value,
      flags: Dom.get( id( "flags" ) ).value.split( "," ).map( ( each ) => each.trim() ).filter( Boolean ),
      notes: notes.value,
    } );
    const before = { name: label.name , container_type: label.container_type || "" , color: label.color || "" , location_id: label.location_id || 0 ,
      parent_id: label.parent_id || 0 , move_status: label.move_status || "" , destination: label.destination || "" , flags: label.flags , notes: label.notes || "" };
    Dom.onSubmit( form , async ( event ) => {
      event.preventDefault();
      const after = read();
      try {
        await Org.post( "/labels/" + label.id , after );
        Undo.saved( form );
        Undo.record( {
          label: I18n.get( "undo.label_edited" ),
          undo: async () => { await Org.post( "/labels/" + label.id , before ); await this.load( label.id ); },
          redo: async () => { await Org.post( "/labels/" + label.id , after ); await this.load( label.id ); },
        } );
        Dom.toast( I18n.get( "container.saved" ) , "ok" );
        await this.load( label.id );
      } catch ( error ) { Org.failed( error ); }
    } );
    return this.card( I18n.get( "container.details_heading" ) , [ form ] );
  },

  // The codes that open this container: its own, and any it claimed.
  codesCard( label ) {
    const rows = [ Dom.el( "li" , { class: "list-group-item d-flex justify-content-between align-items-center" , children: [
      Dom.el( "span" , { class: "font-monospace" , text: label.printed_code } ), Dom.badge( I18n.get( "container.own_code" ) , "secondary" ),
    ] } ) ].concat( label.claimed.map( ( code ) => Dom.el( "li" , { class: "list-group-item d-flex justify-content-between align-items-center gap-2" , children: [
      Dom.el( "span" , { class: "font-monospace text-break small" , text: code } ),
      Org.canEdit ? Org.button( "container.release" , "x-lg" , null , async () => {
        try {
          await Org.post( "/labels/" + label.id + "/codes/release" , { code: code } );
          await this.load( label.id );
        } catch ( error ) { Org.failed( error ); }
      } ) : null,
    ].filter( Boolean ) } ) ) );
    const parts = [ Dom.el( "p" , { class: "small text-body-secondary" , text: I18n.get( "container.codes_text" ) } ) , Dom.el( "ul" , { class: "list-group mb-2" , children: rows } ) ];
    if ( Org.canEdit ) {
      const input = Dom.el( "input" , { class: "form-control" , attrs: { id: "container-claim" , autocomplete: "off" , placeholder: I18n.get( "container.claim_placeholder" ) } } );
      const form = Dom.el( "form" , { class: "row g-2" , children: [
        Dom.el( "div" , { class: "col-12 col-md-8" , children: [ Dom.el( "label" , { class: "visually-hidden" , text: I18n.get( "container.claim_placeholder" ) , attrs: { for: "container-claim" } } ) , input ] } ),
        Dom.el( "div" , { class: "col-12 col-md-4 d-grid" , children: [ Dom.el( "button" , { class: "btn btn-outline-primary" , text: I18n.get( "container.claim" ) , attrs: { type: "submit" } } ) ] } ),
      ] } );
      Dom.onSubmit( form , async ( event ) => {
        event.preventDefault();
        try {
          await Org.post( "/labels/" + label.id + "/codes" , { code: input.value } );
          await this.load( label.id );
        } catch ( error ) { Org.failed( error ); }
      } );
      parts.push( form );
    }
    return this.card( I18n.get( "container.codes_heading" ) , parts );
  },

  async moveAll( label ) {
    const target = await Org.pickLabel( I18n.get( "container.move_all" ) , label.id );
    if ( !target ) return;
    try {
      const moved = await Org.post( "/labels/" + label.id + "/move-contents" , { to_label_id: target } );
      const pairs = ( to ) => moved.item_ids.map( ( itemId ) => ( { item_id: itemId , label_id: to } ) );
      const send = ( to ) => Api.post( "/api/bulk/households/items/move" , { items: pairs( to ) , shared: { household_id: Org.id } } );
      Undo.record( {
        label: I18n.get( "undo.moved" ),
        toast: true,
        undo: async () => { await send( label.id ); await this.load( label.id ); },
        redo: async () => { await send( target ); await this.load( label.id ); },
      } );
      await this.load( label.id );
    } catch ( error ) { Org.failed( error ); }
  },

  async remove( label ) {
    try {
      await Org.post( "/labels/" + label.id + "/delete" );
      Undo.record( {
        label: I18n.get( "undo.label_deleted" ),
        toast: true,
        undo: async () => { await Org.post( "/labels/" + label.id + "/restore" ); Nav.go( "/labels/" + label.id ); },
        redo: async () => { await Org.post( "/labels/" + label.id + "/delete" ); Nav.go( "/labels" ); },
      } );
      Nav.go( "/labels" );
    } catch ( error ) { Org.failed( error ); }
  },

  // ---- Smart Photo Extraction -----------------------------------------

  async extract( label , file ) {
    const holder = Dom.get( "container-ai" );
    const say = ( key ) => Dom.swap( holder , Dom.el( "div" , { class: "alert alert-info d-flex align-items-center gap-2" , children: [
      Dom.el( "span" , { class: "spinner-border spinner-border-sm" , attrs: { "aria-hidden": "true" } } ) , Dom.el( "span" , { text: I18n.get( key ) } ) ,
    ] } ) );
    try {
      say( "container.ai_uploading" );
      const fileId = await Org.uploadPhoto( file );
      say( "container.ai_reading" );
      let draft = await Org.post( "/labels/" + label.id + "/extract" , { file_id: fileId } );
      const started = Date.now();
      while ( draft.status === "working" && Date.now() - started < 300000 ) {
        await new Promise( ( resolve ) => window.setTimeout( resolve , 2000 ) );
        draft = await Org.get( "/extractions/" + draft.id );
      }
      if ( draft.status !== "ready" ) throw new Error( draft.error || I18n.get( "container.ai_failed" ) );
      await this.review( label , draft , holder );
    } catch ( error ) {
      Dom.swap( holder , Dom.el( "div" , { class: "alert alert-danger" , text: error.message } ) );
    }
  },

  // The draft as a checklist: each item ticked, with its cropped picture and
  // editable name and count. Done saves the ticked ones.
  async review( label , draft , holder ) {
    if ( draft.items.length === 0 ) {
      Dom.swap( holder , Dom.el( "div" , { class: "alert alert-warning" , text: I18n.get( "container.ai_nothing" ) } ) );
      return;
    }
    const photoUrl = Org.photoUrl( draft.file_id );
    const rows = await Promise.all( draft.items.map( async ( found , index ) => {
      const tick = Dom.el( "input" , { class: "form-check-input" , attrs: { type: "checkbox" , checked: "" , "aria-label": I18n.get( "container.ai_keep" ) } } );
      const name = Dom.el( "input" , { class: "form-control form-control-sm" , attrs: { maxlength: "80" , "aria-label": I18n.get( "item.name" ) } } );
      name.value = found.name;
      const quantity = Dom.el( "input" , { class: "form-control form-control-sm" , attrs: { type: "number" , min: "0" , step: "any" , "aria-label": I18n.get( "item.quantity" ) } } );
      quantity.value = String( found.quantity || 1 );
      let preview = null;
      let crop = null;
      if ( found.box ) {
        crop = await Org.crop( photoUrl , found.box ).catch( () => null );
        if ( crop ) preview = Dom.el( "img" , { class: "org-thumb rounded border" , attrs: { src: URL.createObjectURL( crop ) , alt: "" } } );
      }
      const row = Dom.el( "tr" , { children: [
        Dom.el( "td" , { children: [ tick ] } ),
        Dom.el( "td" , { children: preview ? [ preview ] : [] } ),
        Dom.el( "td" , { children: [ name , Dom.el( "div" , { class: "small text-body-secondary mt-1" , text: [ found.brand , found.description ].filter( Boolean ).join( " · " ) } ) ] } ),
        Dom.el( "td" , { class: "org-narrow" , children: [ quantity ] } ),
      ] } );
      return { row , tick , name , quantity , found , crop , index };
    } ) );
    const table = Dom.el( "div" , { class: "table-responsive" , children: [ Dom.el( "table" , { class: "table align-middle" , children: [
      Dom.el( "thead" , { children: [ Dom.el( "tr" , { children: [ "container.ai_col_keep" , "container.ai_col_photo" , "container.ai_col_item" , "container.ai_col_quantity" ]
        .map( ( key ) => Dom.el( "th" , { text: I18n.get( key ) , attrs: { scope: "col" } } ) ) } ) ] } ),
      Dom.el( "tbody" , { children: rows.map( ( each ) => each.row ) } ),
    ] } ) ] } );
    const keepPhoto = Dom.el( "input" , { class: "form-check-input" , attrs: { type: "checkbox" , id: "ai-keep-photo" , checked: "" } } );
    const rename = draft.summary && label.name === "Container" ? Dom.el( "p" , { class: "small" , text: I18n.get( "container.ai_summary" ).replace( "{summary}" , draft.summary ) } ) : null;
    const done = Dom.el( "button" , { class: "btn btn-primary" , text: I18n.get( "container.ai_done" ) , attrs: { type: "button" } , on: { click: async () => {
      const kept = rows.filter( ( each ) => each.tick.checked && each.name.value.trim() );
      if ( kept.length === 0 ) {
        Dom.swap( holder );
        return;
      }
      try {
        const items = [];
        for ( const each of kept ) {
          const photos = each.crop ? [ await Org.upload( new File( [ each.crop ] , "item-" + each.index + ".jpg" , { type: "image/jpeg" } ) ) ] : [];
          items.push( { name: each.name.value.trim() , quantity: Number( each.quantity.value || 1 ) , description: each.found.description ,
            brand: each.found.brand , tags: each.found.tags , photos: photos } );
        }
        const saved = await Org.post( "/extractions/" + draft.id + "/accept" , { items: items } );
        if ( keepPhoto.checked ) await Org.post( "/labels/" + label.id , { photos: label.photos.concat( [ draft.file_id ] ) } );
        if ( rename && draft.summary ) await Org.post( "/labels/" + label.id , { name: draft.summary } );
        const ids = saved.items.map( ( item ) => ( { item_id: item.id } ) );
        const send = ( action ) => Api.post( "/api/bulk/households/items/" + action , { items: ids , shared: { household_id: Org.id } } );
        Undo.record( {
          label: I18n.get( "undo.items_added" ),
          toast: true,
          undo: async () => { await send( "delete" ); await this.load( label.id ); },
          redo: async () => { await send( "restore" ); await this.load( label.id ); },
        } );
        Dom.toast( I18n.get( "container.ai_saved" ).replace( "{count}" , saved.items.length ) , "ok" );
        await this.load( label.id );
      } catch ( error ) { Org.failed( error ); }
    } } } );
    Dom.swap( holder , Dom.el( "div" , { class: "border rounded p-3 mb-3" , children: [
      Dom.el( "h4" , { class: "h6" , text: I18n.get( "container.ai_review" ).replace( "{count}" , draft.items.length ) } ),
      rename,
      table,
      Dom.el( "div" , { class: "form-check mb-2" , children: [ keepPhoto , Dom.el( "label" , { class: "form-check-label" , text: I18n.get( "container.ai_keep_photo" ) , attrs: { for: "ai-keep-photo" } } ) ] } ),
      Dom.el( "div" , { class: "d-flex gap-2" , children: [ done , Dom.el( "button" , { class: "btn btn-outline-secondary" , text: I18n.get( "ui.cancel" ) , attrs: { type: "button" } , on: { click: () => Dom.swap( holder ) } } ) ] } ),
    ].filter( Boolean ) } ) );
  },
};
