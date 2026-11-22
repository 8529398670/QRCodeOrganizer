// Reminders, without email: the list of what is expiring, low or overdue, a
// badge on the sidebar, a heads-up from the browser while a tab is open, and
// a calendar feed a phone's calendar can subscribe to.
const Reminders = {
  seen: 0,
  browserKey: "reminders-browser",
  checkEvery: 5 * 60 * 1000,

  init() {
    Dom.onPress( Dom.get( "reminders-read-all" ) , async () => {
      try {
        await Api.post( "/api/notifications/read" , {} );
        await this.refresh();
      } catch ( error ) { Org.failed( error ); }
    } );
    Dom.onPress( Dom.get( "feed-make" ) , () => this.makeFeed() );
    Dom.onPress( Dom.get( "feed-off" ) , () => this.feedOff() );
    Dom.get( "feed-copy" ).addEventListener( "click" , () => Dom.copy( Dom.get( "feed-value" ) ) );
    const browser = Dom.get( "reminders-browser" );
    browser.checked = this.browserOn();
    browser.addEventListener( "change" , async () => {
      if ( browser.checked && "Notification" in window && Notification.permission !== "granted" ) {
        browser.checked = ( await Notification.requestPermission() ) === "granted";
      }
      try { window.localStorage.setItem( this.browserKey , browser.checked ? "1" : "" ); } catch ( error ) { /* not remembered */ }
    } );
    Dom.show( browser.closest( ".card" ) , "Notification" in window );
    window.addEventListener( "nav" , ( event ) => { if ( event.detail.screen === "reminders" ) this.refresh(); } );
    this.badge();
    window.setInterval( () => this.badge() , this.checkEvery );
  },

  browserOn() {
    try { return window.localStorage.getItem( this.browserKey ) === "1" && "Notification" in window && Notification.permission === "granted"; } catch ( error ) { return false; }
  },

  async badge() {
    try {
      const page = await Api.request( "/api/notifications?unread=true&limit=5" );
      const badge = Dom.get( "reminders-badge" );
      Dom.text( badge , String( page.unread ) );
      Dom.show( badge , page.unread > 0 );
      if ( page.unread > this.seen && this.seen !== 0 && this.browserOn() && page.notifications.length ) {
        const notice = page.notifications[ 0 ];
        new Notification( I18n.get( "app.title" ) , { body: this.sentence( notice ) } );
      }
      this.seen = Math.max( page.unread , 1 );
    } catch ( error ) { /* the badge is a nicety */ }
  },

  sentence( notice ) {
    return I18n.get( "reminders.kind_" + notice.kind ).replace( "{item}" , notice.item_name ).replace( "{due}" , Org.date( notice.due ) );
  },

  async refresh() {
    const page = await Pager.load( "notices-table" , "notifications" , ( offset ) => Api.request( "/api/notifications" + Api.query( { offset: offset } ) ) , ( notice ) => {
      const tone = { expired: "danger" , overdue: "danger" , expiring: "warning" , low_stock: "info" }[ notice.kind ];
      const row = Dom.el( "tr" , { class: notice.read ? "text-body-secondary" : "fw-semibold" , children: [
        Dom.el( "td" , { children: [ Dom.badge( I18n.get( "reminders.label_" + notice.kind ) , tone ) ] } ),
        Dom.el( "td" , { children: [ Org.labelLink( notice.label_id , notice.item_name ) ] } ),
        Dom.el( "td" , { text: Org.date( notice.due ) } ),
        Dom.el( "td" , { children: [ Dom.when( notice.created_at ) ] } ),
      ] } );
      Dom.get( "notices-table" ).querySelector( "tbody" ).appendChild( row );
    } );
    if ( page ) Dom.show( Dom.get( "notices-empty" ) , page.total === 0 );
    const feed = await Api.request( "/api/account/calendar-feed" );
    Dom.show( Dom.get( "feed-off" ) , feed.on );
    Dom.text( Dom.get( "feed-make" ) , I18n.get( feed.on ? "reminders.feed_remake" : "reminders.feed_make" ) );
    await this.badge();
  },

  async makeFeed() {
    try {
      const feed = await Api.post( "/api/account/calendar-feed" );
      // Calendar apps want webcal:// to subscribe rather than download.
      Dom.text( Dom.get( "feed-value" ) , window.location.origin.replace( /^http/ , "webcal" ) + feed.path );
      Dom.show( Dom.get( "feed-box" ) , true );
      await this.refresh();
    } catch ( error ) { Org.failed( error ); }
  },

  async feedOff() {
    try {
      await Api.post( "/api/account/calendar-feed/off" );
      Dom.show( Dom.get( "feed-box" ) , false );
      await this.refresh();
    } catch ( error ) { Org.failed( error ); }
  },
};
