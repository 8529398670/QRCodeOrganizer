// The item editor: one dialog, used by every screen that shows items, for
// adding an item to a container or changing one. ItemEditor.open answers the
// saved item, or null if it was closed without saving. Saving is undoable.
const ItemEditor = {
  modal: null,
  resolve: null,
  item: null,
  labelId: 0,
  photos: [],
  receipts: [],

  init() {
    this.modal = new bootstrap.Modal( Dom.get( "item-modal" ) );
    Dom.get( "item-modal" ).addEventListener( "hidden.bs.modal" , () => this.finish( null ) );
    Dom.onSubmit( Dom.get( "item-form" ) , ( event ) => this.save( event ) );
  },

  open( item , labelId ) {
    this.item = item;
    this.labelId = labelId || ( item && item.label_id ) || 0;
    this.photos = item ? item.photos.slice() : [];
    this.receipts = item ? item.receipts.slice() : [];
    Dom.text( Dom.get( "item-modal-title" ) , I18n.get( item ? "item.edit_heading" : "item.add_heading" ) );
    Dom.swap( Dom.get( "item-modal-body" ) , this.build( item || { quantity: 1 , tags: [] , fields: {} } ) );
    this.modal.show();
    return new Promise( ( resolve ) => { this.resolve = resolve; } );
  },

  finish( value ) {
    if ( !this.resolve ) return;
    const resolve = this.resolve;
    this.resolve = null;
    resolve( value );
  },

  field( id , key , input , width ) {
    input.id = "item-" + id;
    return Dom.el( "div" , { class: width || "col-12" , children: [
      Dom.el( "label" , { class: "form-label" , text: I18n.get( "item." + key ) , attrs: { for: input.id } } ) , input ,
    ] } );
  },

  input( value , attrs ) {
    const input = Dom.el( "input" , { class: "form-control" , attrs: attrs || {} } );
    input.value = value === undefined || value === null ? "" : String( value );
    return input;
  },

  build( item ) {
    const rows = [
      this.field( "name" , "name" , this.input( item.name , { maxlength: "80" , required: "" } ) , "col-12 col-md-8" ),
      this.field( "quantity" , "quantity" , this.input( item.quantity , { type: "number" , min: "0" , step: "any" } ) , "col-6 col-md-2" ),
      this.field( "unit" , "unit" , this.input( item.unit , { maxlength: "20" } ) , "col-6 col-md-2" ),
      this.field( "description" , "description" , Dom.el( "textarea" , { class: "form-control" , text: item.description || "" , attrs: { rows: "2" , maxlength: "4000" } } ) ),
      this.field( "tags" , "tags" , this.input( ( item.tags || [] ).join( ", " ) , { placeholder: I18n.get( "item.tags_hint" ) } ) ),
      this.photoField( "photos" , this.photos ),
      Dom.el( "h3" , { class: "h6 mt-3 mb-0 col-12" , text: I18n.get( "item.reminders_heading" ) } ),
      this.field( "expires" , "expires" , this.input( item.expires , { type: "date" } ) , "col-6 col-md-4" ),
      this.field( "remind_days" , "remind_days" , this.input( item.remind_days , { type: "number" , min: "0" , max: "365" , placeholder: I18n.get( "item.remind_default" ) } ) , "col-6 col-md-4" ),
      this.field( "low_stock" , "low_stock" , this.input( item.low_stock , { type: "number" , min: "0" , step: "any" } ) , "col-6 col-md-4" ),
      this.field( "lent_to" , "lent_to" , this.input( item.lent_to , { maxlength: "80" } ) , "col-6 col-md-6" ),
      this.field( "lent_due" , "lent_due" , this.input( item.lent_due , { type: "date" } ) , "col-6 col-md-6" ),
      Dom.el( "h3" , { class: "h6 mt-3 mb-0 col-12" , text: I18n.get( "item.details_heading" ) } ),
      this.field( "brand" , "brand" , this.input( item.brand , { maxlength: "80" } ) , "col-6 col-md-4" ),
      this.field( "barcode" , "barcode" , this.input( item.barcode , { maxlength: "200" } ) , "col-6 col-md-4" ),
      this.field( "serial" , "serial" , this.input( item.serial , { maxlength: "120" } ) , "col-6 col-md-4" ),
      this.field( "purchase_date" , "purchase_date" , this.input( item.purchase_date , { type: "date" } ) , "col-6 col-md-4" ),
      this.field( "value" , "value" , this.input( item.value || "" , { type: "number" , min: "0" , step: "0.01" } ) , "col-6 col-md-4" ),
      this.photoField( "receipts" , this.receipts ),
    ];
    const custom = ( Org.current.fields || [] );
    if ( custom.length ) rows.push( Dom.el( "h3" , { class: "h6 mt-3 mb-0 col-12" , text: I18n.get( "item.custom_heading" ) } ) );
    custom.forEach( ( definition ) => rows.push( this.customField( definition , ( item.fields || {} )[ definition.id ] ) ) );
    if ( item.history && item.history.length ) rows.push( this.history( item.history ) );
    return Dom.el( "div" , { class: "row g-2" , children: rows } );
  },

  customField( definition , value ) {
    let input;
    if ( definition.type === "select" ) {
      input = Dom.el( "select" , { class: "form-select" , children: [ Dom.el( "option" , { text: "" , attrs: { value: "" } } ) ]
        .concat( definition.options.map( ( option ) => Dom.el( "option" , { text: option , attrs: { value: option } } ) ) ) } );
      input.value = value || "";
    } else if ( definition.type === "bool" ) {
      input = Dom.el( "select" , { class: "form-select" , children: [
        Dom.el( "option" , { text: "" , attrs: { value: "" } } ),
        Dom.el( "option" , { text: I18n.get( "ui.yes" ) || "Yes" , attrs: { value: "yes" } } ),
        Dom.el( "option" , { text: I18n.get( "ui.no" ) || "No" , attrs: { value: "no" } } ),
      ] } );
      input.value = value || "";
    } else {
      input = this.input( value , { type: definition.type === "number" ? "number" : definition.type === "date" ? "date" : "text" , step: "any" } );
    }
    input.id = "item-custom-" + definition.id;
    input.dataset.field = definition.id;
    return Dom.el( "div" , { class: "col-6 col-md-4" , children: [
      Dom.el( "label" , { class: "form-label" , text: definition.name , attrs: { for: input.id } } ) , input ,
    ] } );
  },

  // A row of thumbnails with remove buttons, and an add button.
  photoField( key , list ) {
    const strip = Dom.el( "div" , { class: "d-flex flex-wrap gap-2 align-items-center" } );
    const paint = () => {
      const tiles = list.map( ( fileId , index ) => Dom.el( "div" , { class: "position-relative" , children: [
        Org.thumb( fileId ),
        Dom.el( "button" , { class: "btn btn-sm btn-light org-thumb-remove" , attrs: { type: "button" , "aria-label": I18n.get( "item.remove_photo" ) } ,
          children: [ Org.icon( "x-lg" ) ] , on: { click: () => { list.splice( index , 1 ); paint(); } } } ),
      ] } ) );
      const input = Org.photoInput( async ( files ) => {
        for ( const file of files ) list.push( await Org.uploadPhoto( file ) );
        paint();
      } , true );
      const add = Dom.el( "label" , { class: "btn btn-sm btn-outline-secondary" , children: [ Org.icon( "camera" ) , Dom.el( "span" , { class: "ms-1" , text: I18n.get( "item.add_" + key ) } ) , input ] } );
      Dom.swap( strip , ...tiles , add );
    };
    paint();
    return Dom.el( "div" , { class: "col-12" , children: [ Dom.el( "div" , { class: "form-label" , text: I18n.get( "item." + key ) } ) , strip ] } );
  },

  history( events ) {
    const list = Dom.el( "ul" , { class: "list-unstyled small text-body-secondary mb-0" ,
      children: events.slice().reverse().slice( 0 , 20 ).map( ( event ) => Dom.el( "li" , { children: [
        Dom.when( event.at ) , Dom.el( "span" , { text: " — " + ( I18n.get( "item.event_" + event.kind ) || event.kind ) + ( event.detail ? ": " + event.detail : "" ) } ),
      ] } ) ) } );
    return Dom.el( "details" , { class: "col-12 mt-3" , children: [ Dom.el( "summary" , { text: I18n.get( "item.history" ) } ) , list ] } );
  },

  value( id ) {
    return Dom.get( "item-" + id ).value.trim();
  },

  // What the form says, in the shape the API takes.
  read() {
    const number = ( id ) => this.value( id ) === "" ? null : Number( this.value( id ) );
    const body = {
      name: this.value( "name" ),
      description: this.value( "description" ),
      quantity: number( "quantity" ) === null ? 1 : number( "quantity" ),
      unit: this.value( "unit" ),
      tags: this.value( "tags" ).split( "," ).map( ( tag ) => tag.trim() ).filter( Boolean ),
      photos: this.photos.slice(),
      receipts: this.receipts.slice(),
      expires: this.value( "expires" ),
      brand: this.value( "brand" ),
      barcode: this.value( "barcode" ),
      serial: this.value( "serial" ),
      purchase_date: this.value( "purchase_date" ),
      value: number( "value" ) || 0,
      lent_to: this.value( "lent_to" ),
      lent_due: this.value( "lent_due" ),
      fields: {},
    };
    if ( number( "low_stock" ) === null ) body.clear_low_stock = true; else body.low_stock = number( "low_stock" );
    if ( number( "remind_days" ) === null ) body.clear_remind_days = true; else body.remind_days = number( "remind_days" );
    Dom.all( "[data-field]" , Dom.get( "item-modal-body" ) ).forEach( ( input ) => {
      if ( input.value.trim() !== "" ) body.fields[ input.dataset.field ] = input.value.trim();
    } );
    return body;
  },

  // The fields of a saved item, to put it back as it was.
  snapshot( item ) {
    const body = {
      name: item.name , description: item.description || "" , quantity: item.quantity , unit: item.unit || "" ,
      tags: item.tags , photos: item.photos , receipts: item.receipts , expires: item.expires || "" ,
      brand: item.brand || "" , barcode: item.barcode || "" , serial: item.serial || "" , purchase_date: item.purchase_date || "" ,
      value: item.value || 0 , lent_to: item.lent_to || "" , lent_due: item.lent_due || "" , fields: item.fields || {} ,
    };
    if ( item.low_stock === undefined || item.low_stock === null ) body.clear_low_stock = true; else body.low_stock = item.low_stock;
    if ( item.remind_days === undefined || item.remind_days === null ) body.clear_remind_days = true; else body.remind_days = item.remind_days;
    return body;
  },

  async save( event ) {
    event.preventDefault();
    const body = this.read();
    try {
      let saved;
      if ( this.item ) {
        const before = this.snapshot( this.item );
        const id = this.item.id;
        saved = await Org.post( "/items/" + id , body );
        Undo.record( {
          label: I18n.get( "undo.item_edited" ),
          toast: true,
          undo: async () => { await Org.post( "/items/" + id , before ); window.dispatchEvent( new CustomEvent( "inventory-changed" ) ); },
          redo: async () => { await Org.post( "/items/" + id , body ); window.dispatchEvent( new CustomEvent( "inventory-changed" ) ); },
        } );
      } else {
        const made = await Org.post( "/labels/" + this.labelId + "/items" , { items: [ body ] } );
        saved = made.items[ 0 ];
        const id = saved.id;
        Undo.record( {
          label: I18n.get( "undo.item_added" ),
          toast: true,
          undo: async () => { await Org.post( "/items/" + id + "/delete" ); window.dispatchEvent( new CustomEvent( "inventory-changed" ) ); },
          redo: async () => { await Org.post( "/items/" + id + "/restore" ); window.dispatchEvent( new CustomEvent( "inventory-changed" ) ); },
        } );
      }
      const resolve = this.resolve;
      this.resolve = null;
      this.modal.hide();
      if ( resolve ) resolve( saved );
      window.dispatchEvent( new CustomEvent( "inventory-changed" ) );
    } catch ( error ) {
      Org.failed( error );
    }
  },

  // The shared row actions for an item, on any screen: edit, +1, -1, move,
  // delete. after() runs once something changed.
  actions( item , after ) {
    if ( !Org.canEdit ) return Dom.rowActions( item.name , [] );
    const adjust = async ( delta ) => {
      try {
        await Org.post( "/items/" + item.id + "/adjust" , { delta: delta } );
        Undo.record( {
          label: I18n.get( "undo.quantity" ),
          merge: "quantity-" + item.id,
          undo: async () => { await Org.post( "/items/" + item.id + "/adjust" , { delta: -delta } ); await after(); },
          redo: async () => { await Org.post( "/items/" + item.id + "/adjust" , { delta: delta } ); await after(); },
        } );
        await after();
      } catch ( error ) { Org.failed( error ); }
    };
    return Dom.rowActions( item.name , [
      Org.button( "item.edit" , "pencil" , "btn-outline-primary" , async () => { if ( await this.open( item ) ) await after(); } ),
      Dom.el( "button" , { class: "btn btn-sm btn-outline-secondary" , attrs: { type: "button" , "aria-label": I18n.get( "item.minus_one" ) , title: I18n.get( "item.minus_one" ) } , children: [ Org.icon( "dash-lg" ) ] , on: { click: () => adjust( -1 ) } } ),
      Dom.el( "button" , { class: "btn btn-sm btn-outline-secondary" , attrs: { type: "button" , "aria-label": I18n.get( "item.plus_one" ) , title: I18n.get( "item.plus_one" ) } , children: [ Org.icon( "plus-lg" ) ] , on: { click: () => adjust( 1 ) } } ),
      Org.button( "item.move" , "box-arrow-right" , null , () => this.move( [ item ] , after ) ),
      Org.button( "item.delete" , "trash3" , "btn-outline-danger" , () => this.remove( [ item ] , after ) ),
    ] , 3 );
  },

  async move( items , after ) {
    const target = await Org.pickLabel( I18n.get( "item.move_heading" ) , items.length === 1 ? items[ 0 ].label_id : 0 );
    if ( !target ) return;
    const from = items.map( ( item ) => ( { item_id: item.id , label_id: item.label_id } ) );
    const send = ( pairs ) => Api.post( "/api/bulk/households/items/move" , { items: pairs , shared: { household_id: Org.id } } );
    try {
      await send( items.map( ( item ) => ( { item_id: item.id , label_id: target } ) ) );
      Undo.record( {
        label: I18n.get( "undo.moved" ),
        toast: true,
        undo: async () => { await send( from ); await after(); },
        redo: async () => { await send( items.map( ( item ) => ( { item_id: item.id , label_id: target } ) ) ); await after(); },
      } );
      await after();
    } catch ( error ) { Org.failed( error ); }
  },

  async remove( items , after ) {
    const pairs = items.map( ( item ) => ( { item_id: item.id } ) );
    const send = ( action ) => Api.post( "/api/bulk/households/items/" + action , { items: pairs , shared: { household_id: Org.id } } );
    try {
      await send( "delete" );
      Undo.record( {
        label: I18n.get( items.length === 1 ? "undo.item_deleted" : "undo.items_deleted" ),
        toast: true,
        undo: async () => { await send( "restore" ); await after(); },
        redo: async () => { await send( "delete" ); await after(); },
      } );
      await after();
    } catch ( error ) { Org.failed( error ); }
  },
};
