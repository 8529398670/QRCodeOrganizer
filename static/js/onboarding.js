// Getting started: the Help screen and the checklist a new household sees.
//
// The Help screen's text all lives in language.yaml's help section: the
// getting-started steps and a list of topics, each drawn as one accordion
// item. /help#print opens the print topic, also from a link inside the app
// (Nav drops the #, so the click is noted here first).
//
// The checklist sits at the top of Scan and Find stuff while the open
// household is new, ticking itself off from what the household holds. It goes
// once a container and an item exist, or when Hide is pressed; the Help screen
// brings it back. Both choices are per household, in this browser. Viewers
// cannot add anything, so they never see it.
const Onboarding = {
  storagePrefix: "onboarding-",
  topic: "",
  // What the last count found, per household, so a screen change does not
  // ask the server again unless something may have changed.
  counted: null,

  init() {
    this.renderHelp();
    this.watchEmpty();
    // Before Nav's own listener (Nav.init runs later), which would drop the #.
    document.addEventListener( "click" , ( event ) => {
      const link = event.target.closest ? event.target.closest( "a[href^='/help#']" ) : null;
      if ( link ) this.topic = link.getAttribute( "href" ).split( "#" )[ 1 ] || "";
    } );
    Dom.onPress( Dom.get( "help-show-checklist" ) , () => {
      this.setState( "shown" );
      this.paintHelpButton();
      Dom.toast( I18n.get( "help.checklist_shown" ) , "ok" );
      return this.refresh();
    } );
    window.addEventListener( "household" , () => {
      this.counted = null;
      this.paintHelpButton();
      this.refresh();
    } );
    window.addEventListener( "inventory-changed" , () => { this.counted = null; } );
    window.addEventListener( "nav" , ( event ) => {
      const screen = event.detail.screen;
      if ( screen === "help" ) this.openTopic();
      // Coming back from a screen where things get made: count again.
      if ( screen === "scan" || screen === "inventory" ) {
        if ( this.left ) this.counted = null;
        this.left = false;
        this.refresh();
      } else {
        this.left = true;
      }
    } );
  },

  // ---- The Help screen ---------------------------------------------------

  renderHelp() {
    const help = ( I18n.data && I18n.data.help ) || {};
    const steps = Array.isArray( help.start_steps ) ? help.start_steps : [];
    Dom.swap( Dom.get( "help-start" ) , ...steps.map( ( step ) => Dom.el( "li" , { class: "list-group-item" , text: step } ) ) );
    Dom.show( Dom.get( "help-start" ) , steps.length > 0 );
    const topics = Array.isArray( help.topics ) ? help.topics : [];
    Dom.swap( Dom.get( "help-topics" ) , ...topics.filter( ( topic ) => topic && topic.id && topic.title ).map( ( topic ) => this.topicItem( topic ) ) );
  },

  topicItem( topic ) {
    const id = "help-topic-" + topic.id;
    const button = Dom.el( "button" , {
      class: "accordion-button collapsed" ,
      attrs: { type: "button" , "data-bs-toggle": "collapse" , "data-bs-target": "#" + id , "aria-expanded": "false" , "aria-controls": id } ,
      children: [
        Dom.el( "i" , { class: "bi bi-" + ( topic.icon || "question-circle" ) + " me-2 text-primary" , attrs: { "aria-hidden": "true" } } ),
        Dom.el( "span" , { text: topic.title } ),
      ] ,
    } );
    const steps = ( Array.isArray( topic.steps ) ? topic.steps : [] ).map( ( step ) => Dom.el( "p" , { text: step } ) );
    if ( topic.link && topic.link_label ) {
      steps.push( Dom.el( "a" , { class: "btn btn-outline-primary btn-sm" , text: topic.link_label , attrs: { href: topic.link } } ) );
    }
    return Dom.el( "div" , { class: "accordion-item" , children: [
      Dom.el( "h4" , { class: "accordion-header" , children: [ button ] } ),
      Dom.el( "div" , { class: "accordion-collapse collapse" , attrs: { id: id , "data-bs-parent": "#help-topics" } , children: [
        Dom.el( "div" , { class: "accordion-body" , children: steps } ),
      ] } ),
    ] } );
  },

  // Opens the topic a link (or the address) named, and brings it into view.
  openTopic() {
    const name = this.topic || window.location.hash.replace( /^#/ , "" );
    this.topic = "";
    if ( !name ) return;
    const panel = document.getElementById( "help-topic-" + name );
    if ( !panel || !window.bootstrap ) return;
    bootstrap.Collapse.getOrCreateInstance( panel , { toggle: false } ).show();
    window.setTimeout( () => Dom.reveal( panel.closest( ".accordion-item" ) , "start" ) , 0 );
  },

  paintHelpButton() {
    Dom.show( Dom.get( "help-show-checklist" ) , Org.canEdit && this.state() !== "shown" );
  },

  // An empty table's "How does this work?" link shows whenever its message does.
  watchEmpty() {
    Dom.all( "[data-help-for]" ).forEach( ( link ) => {
      const message = Dom.get( link.getAttribute( "data-help-for" ) );
      if ( !message ) return;
      const sync = () => Dom.show( link , !message.hidden );
      new MutationObserver( sync ).observe( message , { attributes: true , attributeFilter: [ "hidden" ] } );
      sync();
    } );
  },

  // ---- The checklist -----------------------------------------------------

  state() {
    try { return window.localStorage.getItem( this.storagePrefix + Org.id ) || ""; } catch ( error ) { return ""; }
  },

  setState( value ) {
    try { window.localStorage.setItem( this.storagePrefix + Org.id , value ); } catch ( error ) { /* not remembered */ }
  },

  async count() {
    if ( this.counted && this.counted.id === Org.id ) return this.counted;
    const total = ( rest , list ) => Org.get( rest , { limit: 1 } ).then( ( page ) => typeof page.total === "number" ? page.total : ( page[ list ] || [] ).length );
    const [ locations , containers , items ] = await Promise.all( [
      total( "/locations" , "locations" ),
      total( "/labels" , "labels" ),
      total( "/items" , "items" ),
    ] );
    this.counted = { id: Org.id , locations: locations , containers: containers , items: items };
    return this.counted;
  },

  async refresh() {
    const slots = Dom.all( ".onboarding-slot" );
    const state = this.state();
    if ( !Org.current || !Org.canEdit || state === "hidden" ) {
      slots.forEach( ( slot ) => Dom.clear( slot ) );
      return;
    }
    let found;
    try {
      found = await this.count();
    } catch ( error ) {
      return;
    }
    const steps = this.steps( found );
    const required = steps.filter( ( step ) => !step.optional );
    if ( state !== "shown" && required.every( ( step ) => step.done ) ) {
      slots.forEach( ( slot ) => Dom.clear( slot ) );
      return;
    }
    slots.forEach( ( slot ) => Dom.swap( slot , this.card( steps ) ) );
  },

  steps( found ) {
    const household = Org.current;
    const steps = [
      { key: "household" , done: true },
      { key: "container" , done: found.containers > 0 , link: "/labels" , hint: true },
      { key: "item" , done: found.items > 0 , link: "/labels" , hint: true },
      { key: "location" , done: found.locations > 0 , link: "/locations" , optional: true },
    ];
    if ( household.my_role === "owner" ) {
      const members = Array.isArray( household.members ) ? household.members.length : 1;
      steps.push( { key: "invite" , done: members > 1 , link: "/households" , optional: true } );
    }
    return steps;
  },

  card( steps ) {
    const done = steps.filter( ( step ) => step.done ).length;
    const percent = Math.round( 100 * done / steps.length );
    const progress = I18n.get( "onboarding.progress" ).replace( "{done}" , String( done ) ).replace( "{total}" , String( steps.length ) );
    const hide = Dom.el( "button" , { class: "btn btn-sm btn-outline-secondary" , text: I18n.get( "onboarding.hide" ) , attrs: { type: "button" } , on: { click: () => {
      this.setState( "hidden" );
      this.paintHelpButton();
      Dom.toast( I18n.get( "onboarding.hidden_toast" ) , "ok" );
      return this.refresh();
    } } } );
    return Dom.el( "section" , { class: "card border-primary-subtle" , children: [ Dom.el( "div" , { class: "card-body" , children: [
      Dom.el( "div" , { class: "d-flex align-items-start gap-2 mb-2" , children: [
        Dom.el( "div" , { class: "me-auto" , children: [
          Dom.el( "h3" , { class: "card-title mb-0" , text: I18n.get( "onboarding.heading" ) } ),
          Dom.el( "p" , { class: "small text-body-secondary mb-0" , text: progress } ),
        ] } ),
        hide,
      ] } ),
      Dom.el( "div" , { class: "progress mb-3" , attrs: { role: "progressbar" , "aria-label": progress , "aria-valuenow": String( percent ) , "aria-valuemin": "0" , "aria-valuemax": "100" } ,
        children: [ Dom.el( "div" , { class: "progress-bar" , attrs: { style: "width: " + percent + "%" } } ) ] } ),
      Dom.el( "ul" , { class: "list-group list-group-flush" , children: steps.map( ( step ) => this.stepRow( step ) ) } ),
      Dom.el( "a" , { class: "btn btn-link btn-sm px-0 mt-2" , attrs: { href: "/help" } , children: [
        Dom.el( "i" , { class: "bi bi-question-circle me-1" , attrs: { "aria-hidden": "true" } } ),
        Dom.el( "span" , { text: I18n.get( "onboarding.guide" ) } ),
      ] } ),
    ] } ) ] } );
  },

  stepRow( step ) {
    const icon = Dom.el( "i" , { class: "bi fs-5 " + ( step.done ? "bi-check-circle-fill text-success" : "bi-circle text-body-tertiary" ) , attrs: { "aria-hidden": "true" } } );
    const title = Dom.el( "span" , { class: step.done ? "text-body-secondary text-decoration-line-through" : "fw-semibold" , text: I18n.get( "onboarding.step_" + step.key ) } );
    const text = [ title ];
    if ( step.optional ) text.push( Dom.el( "span" , { class: "badge text-bg-light border ms-2" , text: I18n.get( "onboarding.optional" ) } ) );
    if ( step.hint && !step.done ) text.push( Dom.el( "div" , { class: "small text-body-secondary" , text: I18n.get( "onboarding.step_" + step.key + "_hint" ) } ) );
    const children = [ icon , Dom.el( "div" , { class: "me-auto" , children: text } ) ];
    if ( !step.done && step.link ) children.push( Dom.el( "a" , { class: "btn btn-sm btn-primary" , text: I18n.get( "onboarding.go" ) , attrs: { href: step.link } } ) );
    return Dom.el( "li" , { class: "list-group-item d-flex align-items-start gap-2 px-0" , children: children } );
  },
};
