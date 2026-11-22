// The Households screen: the open household's name, members and their roles,
// what its label colors mean, its custom item fields and reminder lead; and
// making another household, or leaving or deleting this one.
const Households = {
  init() {
    Dom.onSubmit( Dom.get( "household-form" ) , async ( event ) => {
      event.preventDefault();
      try {
        await Org.create( Dom.get( "household-name" ).value );
        Dom.get( "household-name" ).value = "";
        Dom.toast( I18n.get( "households.created" ) , "ok" );
      } catch ( error ) { Org.failed( error ); }
    } );
    window.addEventListener( "household" , () => { if ( Nav.current() === "households" ) this.refresh(); } );
    window.addEventListener( "nav" , ( event ) => { if ( event.detail.screen === "households" ) this.refresh(); } );
  },

  card( title , children ) {
    return Dom.el( "section" , { class: "card" , children: [ Dom.el( "div" , { class: "card-body" , children: [ Dom.el( "h3" , { class: "card-title" , text: title } ) , ...children ] } ) ] } );
  },

  async refresh() {
    const household = Org.current;
    if ( !household ) return;
    const owner = household.my_role === "owner";
    const cards = [ this.nameCard( household , owner ) , await this.membersCard( household , owner ) , this.colorsCard( household ) , this.fieldsCard( household ) ];
    Dom.swap( Dom.get( "households-body" ) , ...cards );
  },

  async save( path , body ) {
    try {
      const household = await Org.post( path , body );
      Org.updated( household );
      return household;
    } catch ( error ) {
      Org.failed( error );
      return null;
    }
  },

  nameCard( household , owner ) {
    const parts = [];
    if ( Org.canEdit ) {
      parts.push( Dom.editable( household.name , {
        label: I18n.get( "households.name_label" ),
        maxLength: 80,
        save: async ( value ) => {
          const before = household.name;
          const saved = await this.save( "/rename" , { name: value } );
          if ( !saved ) return before;
          Undo.record( {
            label: I18n.get( "undo.household_renamed" ),
            undo: () => this.save( "/rename" , { name: before } ),
            redo: () => this.save( "/rename" , { name: value } ),
          } );
          return saved.name;
        },
      } ) );
    } else {
      parts.push( Dom.el( "p" , { class: "fs-5" , text: household.name } ) );
    }
    parts.push( Dom.el( "p" , { class: "small text-body-secondary mt-2" , text: I18n.get( "households.your_role" ).replace( "{role}" , I18n.get( "households.role_" + household.my_role ) ) } ) );
    const buttons = [ Org.button( "households.leave" , "box-arrow-left" , null , async () => {
      if ( !await Dom.confirm( I18n.get( "households.leave_confirm" ) ) ) return;
      try {
        await Org.post( "/leave" );
        Org.current = null;
        await Org.refresh();
      } catch ( error ) { Org.failed( error ); }
    } ) ];
    if ( owner ) buttons.push( Org.button( "households.delete" , "trash3" , "btn-outline-danger" , async () => {
      if ( !await Dom.confirm( I18n.get( "households.delete_confirm" ).replace( "{name}" , household.name ) ) ) return;
      try {
        await Org.post( "/delete" );
        Org.current = null;
        await Org.refresh();
      } catch ( error ) { Org.failed( error ); }
    } ) );
    parts.push( Dom.el( "div" , { class: "d-flex flex-wrap gap-2" , children: buttons } ) );
    return this.card( I18n.get( "households.this_heading" ) , parts );
  },

  async membersCard( household , owner ) {
    const roleSelect = ( value , id ) => {
      const select = Dom.el( "select" , { class: "form-select form-select-sm" , attrs: { id: id } , children: [ "owner" , "editor" , "viewer" ]
        .map( ( role ) => Dom.el( "option" , { text: I18n.get( "households.role_" + role ) , attrs: { value: role } } ) ) } );
      select.value = value;
      return select;
    };
    const rows = household.members.map( ( member ) => {
      const isMe = member.user_id === Org.me.id;
      let role = Dom.el( "span" , { text: I18n.get( "households.role_" + member.role ) } );
      if ( owner && !isMe ) {
        role = roleSelect( member.role , "member-role-" + member.user_id );
        role.setAttribute( "aria-label" , I18n.get( "households.role_for" ).replace( "{name}" , member.display_name ) );
        role.addEventListener( "change" , async () => {
          const before = member.role;
          const after = role.value;
          if ( await this.save( "/members" , { user_id: member.user_id , role: after } ) ) {
            Undo.record( {
              label: I18n.get( "undo.member_role" ),
              undo: () => this.save( "/members" , { user_id: member.user_id , role: before } ),
              redo: () => this.save( "/members" , { user_id: member.user_id , role: after } ),
            } );
          } else role.value = before;
        } );
      }
      return Dom.el( "tr" , { children: [
        Dom.el( "td" , { text: member.display_name + ( isMe ? " " + I18n.get( "households.you" ) : "" ) } ),
        Dom.el( "td" , { children: [ role ] } ),
        Dom.el( "td" , { children: owner && !isMe ? [ Org.button( "households.remove_member" , "person-dash" , "btn-outline-danger" , async () => {
          if ( await this.save( "/members" , { user_id: member.user_id , role: "" } ) ) {
            Undo.record( {
              label: I18n.get( "undo.member_removed" ),
              toast: true,
              undo: () => this.save( "/members" , { user_id: member.user_id , role: member.role } ),
              redo: () => this.save( "/members" , { user_id: member.user_id , role: "" } ),
            } );
          }
        } ) ] : [] } ),
      ] } );
    } );
    const parts = [ Dom.el( "div" , { class: "table-responsive" , children: [ Dom.el( "table" , { class: "table align-middle" , children: [
      Dom.el( "thead" , { children: [ Dom.el( "tr" , { children: [ "households.col_member" , "households.col_role" , "households.col_actions" ].map( ( key ) => Dom.el( "th" , { text: I18n.get( key ) , attrs: { scope: "col" } } ) ) } ) ] } ),
      Dom.el( "tbody" , { children: rows } ),
    ] } ) ] } ) ];
    if ( owner ) {
      const team = await Api.request( "/api/team?limit=1000" ).catch( () => ( { members: [] } ) );
      const candidates = team.members.filter( ( person ) => !household.members.some( ( member ) => member.user_id === person.id ) );
      const person = Dom.el( "select" , { class: "form-select" , attrs: { id: "member-add-person" } , children: candidates.map( ( each ) => Dom.el( "option" , { text: each.display_name , attrs: { value: String( each.id ) } } ) ) } );
      const role = roleSelect( "editor" , "member-add-role" );
      role.classList.remove( "form-select-sm" );
      const form = Dom.el( "form" , { class: "row g-2 align-items-end" , children: [
        Dom.el( "div" , { class: "col-12 col-md-5" , children: [ Dom.el( "label" , { class: "form-label" , text: I18n.get( "households.add_person" ) , attrs: { for: "member-add-person" } } ) , person ] } ),
        Dom.el( "div" , { class: "col-6 col-md-4" , children: [ Dom.el( "label" , { class: "form-label" , text: I18n.get( "households.col_role" ) , attrs: { for: "member-add-role" } } ) , role ] } ),
        Dom.el( "div" , { class: "col-6 col-md-3 d-grid" , children: [ Dom.el( "button" , { class: "btn btn-outline-primary" , text: I18n.get( "households.add_button" ) , attrs: { type: "submit" } } ) ] } ),
      ] } );
      Dom.onSubmit( form , async ( event ) => {
        event.preventDefault();
        const userId = Number( person.value );
        if ( await this.save( "/members" , { user_id: userId , role: role.value } ) ) {
          Undo.record( {
            label: I18n.get( "undo.member_added" ),
            undo: () => this.save( "/members" , { user_id: userId , role: "" } ),
            redo: () => this.save( "/members" , { user_id: userId , role: role.value } ),
          } );
        }
      } );
      parts.push( candidates.length ? form : Dom.el( "p" , { class: "small text-body-secondary mb-0" , text: I18n.get( "households.no_candidates" ) } ) );
      parts.push( Dom.el( "p" , { class: "form-text mb-0" , text: I18n.get( "households.invite_hint" ) } ) );
    }
    return this.card( I18n.get( "households.members_heading" ) , parts );
  },

  colorsCard( household ) {
    const inputs = [];
    const rows = Org.colors().map( ( color , index ) => {
      const input = Dom.el( "input" , { class: "form-control form-control-sm" , attrs: { id: "color-meaning-" + index , maxlength: "80" , placeholder: I18n.get( "households.color_placeholder" ) } } );
      input.value = household.colors[ color ] || "";
      input.disabled = !Org.canEdit;
      inputs.push( [ color , input ] );
      return Dom.el( "div" , { class: "col-12 col-md-6 col-xl-4" , children: [ Dom.el( "div" , { class: "input-group input-group-sm" , children: [
        Dom.el( "label" , { class: "input-group-text" , attrs: { for: "color-meaning-" + index } , children: [ Org.swatch( color ) , Dom.el( "span" , { class: "ms-2" , text: Org.colorName( color ).split( " — " )[ 0 ] } ) ] } ),
        input,
      ] } ) ] } );
    } );
    const form = Dom.el( "form" , { children: [ Dom.el( "div" , { class: "row g-2 mb-3" , children: rows } ) ] } );
    if ( Org.canEdit ) {
      form.appendChild( Dom.el( "button" , { class: "btn btn-primary" , text: I18n.get( "households.save_colors" ) , attrs: { type: "submit" } } ) );
      Dom.onSubmit( form , async ( event ) => {
        event.preventDefault();
        const before = Object.assign( {} , household.colors );
        const after = {};
        inputs.forEach( ( [ color , input ] ) => { if ( input.value.trim() ) after[ color ] = input.value.trim(); } );
        if ( await this.save( "/colors" , { colors: after } ) ) {
          Undo.saved( form );
          Undo.record( { label: I18n.get( "undo.colors" ) , undo: () => this.save( "/colors" , { colors: before } ) , redo: () => this.save( "/colors" , { colors: after } ) } );
          Dom.toast( I18n.get( "households.saved" ) , "ok" );
        }
      } );
    }
    return this.card( I18n.get( "households.colors_heading" ) , [ Dom.el( "p" , { class: "small text-body-secondary" , text: I18n.get( "households.colors_text" ) } ) , form ] );
  },

  fieldsCard( household ) {
    const list = household.fields.map( ( field ) => Object.assign( {} , field , { options: field.options.slice() } ) );
    const holder = Dom.el( "div" );
    const types = [ "text" , "number" , "date" , "select" , "bool" ];
    const paint = () => {
      const rows = list.map( ( field , index ) => {
        const name = Dom.el( "input" , { class: "form-control form-control-sm" , attrs: { maxlength: "40" , "aria-label": I18n.get( "households.field_name" ) } } );
        name.value = field.name;
        name.addEventListener( "input" , () => { field.name = name.value; } );
        const type = Dom.el( "select" , { class: "form-select form-select-sm" , attrs: { "aria-label": I18n.get( "households.field_type" ) } ,
          children: types.map( ( each ) => Dom.el( "option" , { text: I18n.get( "households.type_" + each ) , attrs: { value: each } } ) ) } );
        type.value = field.type;
        const options = Dom.el( "input" , { class: "form-control form-control-sm" , attrs: { "aria-label": I18n.get( "households.field_options" ) , placeholder: I18n.get( "households.field_options" ) } } );
        options.value = field.options.join( ", " );
        Dom.show( options , field.type === "select" );
        type.addEventListener( "change" , () => { field.type = type.value; Dom.show( options , field.type === "select" ); } );
        options.addEventListener( "input" , () => { field.options = options.value.split( "," ).map( ( each ) => each.trim() ).filter( Boolean ); } );
        return Dom.el( "div" , { class: "row g-2 mb-2" , children: [
          Dom.el( "div" , { class: "col-12 col-md-4" , children: [ name ] } ),
          Dom.el( "div" , { class: "col-6 col-md-3" , children: [ type ] } ),
          Dom.el( "div" , { class: "col-6 col-md-3" , children: [ options ] } ),
          Dom.el( "div" , { class: "col-12 col-md-2" , children: [ Org.button( "households.remove_field" , "x-lg" , null , () => { list.splice( index , 1 ); paint(); } ) ] } ),
        ] } );
      } );
      Dom.swap( holder , ...rows );
    };
    paint();
    const lead = Dom.el( "input" , { class: "form-control" , attrs: { id: "household-lead" , type: "number" , min: "0" , max: "365" , placeholder: I18n.get( "households.lead_default" ) } } );
    lead.value = household.reminder_days === null || household.reminder_days === undefined ? "" : String( household.reminder_days );
    const parts = [ Dom.el( "p" , { class: "small text-body-secondary" , text: I18n.get( "households.fields_text" ) } ) , holder ];
    if ( Org.canEdit ) {
      parts.push( Org.button( "households.add_field" , "plus-lg" , null , () => { list.push( { name: "" , type: "text" , options: [] } ); paint(); } ) );
      parts.push( Dom.el( "div" , { class: "row mt-3" , children: [ Dom.el( "div" , { class: "col-12 col-md-6" , children: [
        Dom.el( "label" , { class: "form-label" , text: I18n.get( "households.lead_label" ) , attrs: { for: "household-lead" } } ) , lead ,
      ] } ) ] } ) );
      parts.push( Dom.el( "button" , { class: "btn btn-primary mt-3" , text: I18n.get( "households.save_fields" ) , attrs: { type: "button" } , on: { click: async () => {
        const before = { fields: household.fields , reminder_days: household.reminder_days === undefined ? null : household.reminder_days };
        const after = { fields: list.filter( ( field ) => field.name.trim() ) , reminder_days: lead.value === "" ? null : Number( lead.value ) };
        if ( await this.save( "/fields" , after ) ) {
          Undo.record( { label: I18n.get( "undo.fields" ) , undo: () => this.save( "/fields" , before ) , redo: () => this.save( "/fields" , after ) } );
          Dom.toast( I18n.get( "households.saved" ) , "ok" );
        }
      } } } ) );
    }
    return this.card( I18n.get( "households.fields_heading" ) , parts );
  },
};
