/*
 * Copyright (c) 2023-2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст.
 */

"use strict";

// SiteApp - turns any multi-page static site (first of all the pages the
// Figma parser in gozaurus exports) into one single-page application.
//
// Nothing here knows about a particular site. Per-site settings come from a
// JSON block the builder writes into every page from the project's
// site.json, and everything not configured is detected from the pages:
//
//   <script type="application/json" id="site-config">
//   {"mode": "site", "layout": "native", "title": "...", "lang": "en"}
//   </script>
//
// (the older <meta name="app-mode" content="site"> + <meta name="app-title">
// pair is still understood). Config keys - all optional:
//   layout   "native" (default) - every page keeps its own exported design;
//            the router swaps it in place: no reloads, View Transitions,
//            prefetch, history, scroll restore, head styles/scripts in sync.
//            "reader" - pages are read as content (title, text, images in
//            visual order, the site menu) and rendered into one responsive
//            shell with header, drawer, search and footer. Meant for frames
//            drawn with absolute positions that cannot adapt to a phone.
//   root     selector of the swapped region (native), default "body > main"
//   home     home page file, default "index.html"
//   title    site name (reader header/footer, <title> suffix)
//   lang     UI language for the shell: en, ru, uk (default: <html lang>)
//   strings  overrides for any UI string, e.g. {"search": "Find"}
//   menu     selector of the menu to build navigation from (reader);
//            default: found automatically - the element whose children
//            hold the most links to different pages
//   logo     image URL or selector of the logo (reader); default: an image
//            that links to the home page
//   mark     extra brand image for the footer (reader), URL
//   exclude  selector(s) of page parts that are not content (reader)
//   theme    CSS variables, e.g. {"accent": "#c8102e"} -> --sa-accent;
//            default: colors and fonts taken from the exported design
//   search   false to hide the site search (reader)
//
// Pages without the config / meta tag are left alone, so the admin apps
// built on this file behave exactly as before.
const SiteApp = (function () {
    "use strict";

    // ---- styles ----------------------------------------------------------
    // Router styles (progress bar, page transitions) apply in both layouts;
    // everything under html.sa-reader is the reader shell. A site re-themes
    // it with --sa-* variables (config "theme" or its own CSS).
    const STYLES = `
html.sa-reader{--sa-paper:#f7f6f2;--sa-paper-2:#ebe8e1;--sa-ink:#1d1f23;--sa-muted:#5f646d;--sa-deep:#1f2a37;--sa-night:#111827;--sa-accent:#2563eb;--sa-highlight:#d4a72c;--sa-line:rgba(29,31,35,.14);--sa-font-ui:system-ui,-apple-system,'Segoe UI',Roboto,sans-serif;--sa-font-text:Georgia,'Times New Roman',serif;--sa-wrap:1200px;--sa-pad:clamp(16px,4vw,40px);--sa-header-h:72px;--sa-ease:cubic-bezier(.2,.7,.2,1);scroll-padding-top:calc(var(--sa-header-h) + 16px);background:var(--sa-paper)}
html.sa-reader body{margin:0;background:var(--sa-paper);color:var(--sa-ink);font-family:var(--sa-font-text);-webkit-font-smoothing:antialiased;text-rendering:optimizeLegibility;overflow-x:hidden}
html.sa-lock{overflow:hidden}
html.sa-reader.sa-boot body>main,html.sa-wait body>main{visibility:hidden}
.sa-root *,.sa-root *::before,.sa-root *::after{box-sizing:border-box}
.sa-root a{color:inherit}
.sa-root img{max-width:100%;display:block}
.sa-root svg{fill:none;stroke:currentColor;stroke-width:1.8;stroke-linecap:round;stroke-linejoin:round}
.sa-wrap{width:100%;max-width:var(--sa-wrap);margin:0 auto;padding:0 var(--sa-pad)}
.sa-skip{position:fixed;left:12px;top:-60px;z-index:100;background:var(--sa-ink);color:#fff;padding:10px 16px;border-radius:8px;font:600 14px var(--sa-font-ui);text-decoration:none;transition:top .2s}
.sa-skip:focus{top:12px}
.sa-root :focus-visible{outline:2px solid var(--sa-highlight);outline-offset:3px;border-radius:4px}
.sa-progress{position:fixed;inset:0 0 auto 0;height:3px;z-index:90;background:linear-gradient(90deg,var(--sa-accent),var(--sa-highlight));transform:scaleX(0);transform-origin:left;opacity:0;pointer-events:none}
html.sa-loading .sa-progress{opacity:1;animation:sa-progress 1.2s var(--sa-ease) forwards}
@keyframes sa-progress{0%{transform:scaleX(0)}60%{transform:scaleX(.7)}100%{transform:scaleX(.92)}}
html.sa-loading{cursor:progress}

/* header */
.sa-header{position:fixed;inset:0 0 auto 0;z-index:50;height:var(--sa-header-h);color:#fff;transition:background .3s,box-shadow .3s,height .3s;view-transition-name:sa-header}
.sa-header::before{content:"";position:absolute;inset:0 0 -40px 0;background:linear-gradient(rgba(5,15,12,.55),transparent);pointer-events:none;transition:opacity .3s}
.sa-header.is-scrolled{background:rgba(10,26,21,.86);-webkit-backdrop-filter:blur(14px) saturate(1.4);backdrop-filter:blur(14px) saturate(1.4);box-shadow:0 1px 0 rgba(255,255,255,.06),0 10px 30px -18px rgba(0,0,0,.6)}
.sa-header.is-scrolled::before{opacity:0}
.sa-header .sa-wrap{max-width:1600px}
.sa-header-row{position:relative;height:100%;display:flex;align-items:center;gap:24px}
.sa-brand{display:inline-flex;align-items:center;gap:12px;text-decoration:none;font:700 18px/1 var(--sa-font-ui);letter-spacing:.02em;white-space:nowrap;flex:none}
.sa-brand img{width:40px;height:40px;object-fit:contain;filter:drop-shadow(0 2px 6px rgba(0,0,0,.35))}
.sa-nav{margin-left:auto}
.sa-nav>ul{display:flex;align-items:center;gap:2px;list-style:none;margin:0;padding:0}
.sa-nav li{position:relative;list-style:none;flex:none}
.sa-nav .sa-group-btn span{white-space:nowrap}
.sa-nav .sa-link,.sa-nav .sa-group-btn{display:inline-flex;align-items:center;gap:6px;min-height:44px;padding:0 11px;border:0;border-radius:999px;background:none;color:rgba(255,255,255,.88);font:500 14px/1 var(--sa-font-ui);text-decoration:none;cursor:pointer;white-space:nowrap;transition:background .2s,color .2s}
.sa-nav .sa-link:hover,.sa-nav .sa-group-btn:hover,.sa-nav .sa-group-btn[aria-expanded="true"]{background:rgba(255,255,255,.12);color:#fff}
.sa-nav .sa-link.is-active,.sa-nav .sa-group.is-active>.sa-group-btn{color:#fff;box-shadow:inset 0 -2px 0 var(--sa-highlight);border-radius:0}
.sa-group-btn svg{width:11px;height:8px;transition:transform .25s var(--sa-ease)}
.sa-group-btn[aria-expanded="true"] svg{transform:rotate(180deg)}
.sa-nav .sa-sub{position:absolute;top:calc(100% + 8px);left:50%;min-width:240px;margin:0;padding:8px;list-style:none;background:#fff;color:var(--sa-ink);border-radius:14px;box-shadow:0 20px 50px -12px rgba(0,0,0,.35),0 0 0 1px rgba(0,0,0,.05);opacity:0;visibility:hidden;transform:translate(-50%,-6px);transition:opacity .18s,transform .22s var(--sa-ease),visibility .18s}
.sa-nav .sa-group-btn[aria-expanded="true"]+.sa-sub{opacity:1;visibility:visible;transform:translate(-50%,0)}
@media (hover:hover){.sa-nav .sa-group:hover>.sa-sub{opacity:1;visibility:visible;transform:translate(-50%,0)}.sa-nav .sa-group:hover>.sa-group-btn svg{transform:rotate(180deg)}}
.sa-nav .sa-sub::before{content:"";position:absolute;inset:-10px 0 auto 0;height:10px}
.sa-nav .sa-sub .sa-link{display:flex;width:100%;min-height:42px;padding:0 14px;border-radius:10px;color:var(--sa-ink);font-weight:500}
.sa-nav .sa-sub .sa-link:hover{background:var(--sa-paper)}
.sa-nav .sa-sub .sa-link.is-active{box-shadow:none;border-radius:10px;background:var(--sa-paper);color:var(--sa-accent)}
.sa-actions{display:flex;gap:4px;margin-left:8px}
.sa-icon{display:inline-grid;place-items:center;width:44px;height:44px;border:0;border-radius:999px;background:none;color:inherit;cursor:pointer;transition:background .2s}
.sa-icon:hover{background:rgba(255,255,255,.14)}
.sa-icon svg{width:22px;height:22px}
.sa-burger{display:none;gap:5px;align-content:center}
.sa-burger span{display:block;width:22px;height:2px;border-radius:2px;background:currentColor;transition:transform .3s var(--sa-ease),opacity .2s}
.sa-burger[aria-expanded="true"] span:nth-child(1){transform:translateY(7px) rotate(45deg)}
.sa-burger[aria-expanded="true"] span:nth-child(2){opacity:0}
.sa-burger[aria-expanded="true"] span:nth-child(3){transform:translateY(-7px) rotate(-45deg)}

/* drawer (mobile menu) */
.sa-drawer{position:fixed;inset:var(--sa-header-h) 0 0 0;z-index:45;background:rgba(5,12,10,.5);opacity:0;transition:opacity .25s}
.sa-drawer.is-open{opacity:1}
.sa-drawer-panel{height:100%;overflow:auto;background:var(--sa-night);color:#fff;padding:12px var(--sa-pad) 40px;transform:translateY(-12px);transition:transform .3s var(--sa-ease)}
.sa-drawer.is-open .sa-drawer-panel{transform:none}
.sa-drawer-list,.sa-drawer-list ul{list-style:none;margin:0;padding:0}
.sa-drawer-list>li{border-bottom:1px solid rgba(255,255,255,.08)}
.sa-drawer-list .sa-link,.sa-drawer-list .sa-group-btn{display:flex;align-items:center;justify-content:space-between;width:100%;min-height:56px;padding:0 4px;border:0;background:none;color:#fff;font:600 17px var(--sa-font-ui);text-align:left;text-decoration:none;cursor:pointer}
.sa-drawer-list .sa-group-btn svg{width:13px;height:9px;transition:transform .25s var(--sa-ease)}
.sa-drawer-list .sa-group-btn[aria-expanded="true"] svg{transform:rotate(180deg)}
.sa-drawer-list .sa-sub{display:block;max-height:0;overflow:hidden;visibility:hidden;transition:max-height .35s var(--sa-ease),visibility .35s}
.sa-drawer-list .sa-group-btn[aria-expanded="true"]+.sa-sub{max-height:600px;visibility:visible;padding-bottom:10px}
.sa-drawer-list .sa-sub .sa-link{min-height:46px;padding-left:16px;font-weight:500;font-size:16px;color:rgba(255,255,255,.78);border-left:2px solid rgba(255,255,255,.12)}
.sa-drawer-list .sa-link.is-active{color:var(--sa-highlight);border-left-color:var(--sa-highlight)}

/* search */
.sa-search{position:fixed;inset:0;z-index:80;display:flex;justify-content:center;align-items:flex-start;padding:min(14vh,120px) 16px 16px;background:rgba(5,12,10,.55);-webkit-backdrop-filter:blur(4px);backdrop-filter:blur(4px);opacity:0;transition:opacity .2s}
.sa-search.is-open{opacity:1}
.sa-search-panel{width:100%;max-width:640px;max-height:76vh;display:flex;flex-direction:column;background:#fff;border-radius:18px;box-shadow:0 30px 80px -20px rgba(0,0,0,.5);overflow:hidden;transform:translateY(-8px) scale(.98);transition:transform .25s var(--sa-ease)}
.sa-search.is-open .sa-search-panel{transform:none}
.sa-search-row{display:flex;align-items:center;gap:10px;padding:8px 8px 8px 18px;border-bottom:1px solid var(--sa-line);color:var(--sa-muted)}
.sa-search-row>svg{width:20px;height:20px;flex:none}
.sa-search input{flex:1;min-width:0;height:48px;padding:0;border:0!important;outline:0!important;box-shadow:none!important;background:none!important;-webkit-appearance:none;appearance:none;font:500 17px var(--sa-font-ui);color:var(--sa-ink)}
.sa-search input::-webkit-search-cancel-button{display:none}
.sa-search .sa-icon:hover{background:var(--sa-paper)}
.sa-results{list-style:none;margin:0;padding:8px;overflow:auto}
.sa-result{display:block;padding:12px 14px;border-radius:12px;text-decoration:none}
.sa-result:hover,.sa-result:focus-visible{background:var(--sa-paper)}
.sa-result strong{display:block;font:600 15px var(--sa-font-ui);margin-bottom:4px}
.sa-result span{display:block;font-size:14px;line-height:1.5;color:var(--sa-muted)}
.sa-result mark{background:color-mix(in srgb,var(--sa-highlight) 35%,transparent);color:inherit;border-radius:2px}
.sa-noresult{padding:18px 14px;color:var(--sa-muted);font:500 15px var(--sa-font-ui)}

/* page */
.sa-main{display:block;min-height:70vh;outline:0;view-transition-name:sa-main}
.sa-hero{position:relative;display:flex;align-items:flex-end;min-height:clamp(380px,62vh,640px);padding:calc(var(--sa-header-h) + 40px) 0 clamp(32px,6vw,64px);color:#fff;background:var(--sa-deep);overflow:hidden;isolation:isolate}
.sa-hero-plain{min-height:clamp(280px,40vh,420px);background:radial-gradient(120% 140% at 0% 0%,color-mix(in srgb,var(--sa-deep) 70%,#fff) 0%,var(--sa-deep) 45%,var(--sa-night) 100%)}
.sa-hero-img{position:absolute;inset:0;z-index:-2;width:100%;height:100%!important;max-width:none!important;object-fit:cover;animation:sa-kenburns 18s ease-out both}
.sa-hero::after{content:"";position:absolute;inset:0;z-index:-1;background:linear-gradient(180deg,rgba(5,15,12,.35) 0%,rgba(5,15,12,0) 30%,rgba(5,15,12,.25) 55%,rgba(5,15,12,.82) 100%)}
@keyframes sa-kenburns{from{transform:scale(1.08)}to{transform:scale(1)}}
.sa-hero-text{position:relative}
.sa-eyebrow{display:inline-flex;align-items:center;gap:10px;margin:0 0 14px;font:600 13px/1.2 var(--sa-font-ui);letter-spacing:.14em;text-transform:uppercase;color:var(--sa-highlight)}
.sa-eyebrow::before{content:"";width:28px;height:2px;background:currentColor}
.sa-hero h1{margin:0;max-width:16ch;font:800 clamp(38px,7vw,84px)/1.02 var(--sa-font-ui);letter-spacing:-.02em;text-wrap:balance;text-shadow:0 2px 30px rgba(0,0,0,.35);outline:0}
.sa-content{padding-top:clamp(40px,7vw,88px);padding-bottom:clamp(48px,8vw,104px)}
.sa-prose{max-width:68ch;margin:0 auto}
.sa-content>.sa-prose+.sa-prose,.sa-content>*+*{margin-top:clamp(32px,5vw,64px)}
.sa-prose p{margin:0 0 1.1em;font-size:clamp(17px,1.25vw,19px);line-height:1.72;text-align:left;hyphens:auto;-webkit-hyphens:auto}
.sa-prose p:last-child{margin-bottom:0}
.sa-prose .sa-lead{font:500 clamp(19px,1.6vw,23px)/1.6 var(--sa-font-ui);color:var(--sa-deep);letter-spacing:-.005em}
.sa-h2{max-width:68ch;margin-left:auto!important;margin-right:auto;font:700 clamp(24px,2.6vw,34px)/1.2 var(--sa-font-ui);color:var(--sa-deep);text-wrap:balance}
.sa-h2::after{content:"";display:block;width:48px;height:3px;margin-top:14px;background:var(--sa-accent)}
.sa-split{display:grid;grid-template-columns:minmax(0,1fr) minmax(0,1fr);gap:clamp(28px,5vw,72px);align-items:center}
.sa-split .sa-prose{margin:0;max-width:60ch}
.sa-split.sa-rev .sa-prose{order:2}
.sa-figure{margin-left:0;margin-right:0;margin-bottom:0;border-radius:18px;overflow:hidden;background:var(--sa-paper-2);box-shadow:0 30px 60px -30px rgba(0,0,0,.35)}
.sa-figure img{width:100%;height:auto;object-fit:cover;transition:transform .8s var(--sa-ease)}
.sa-figure:hover img{transform:scale(1.03)}
.sa-figure.sa-wide{max-width:1000px;margin-left:auto!important;margin-right:auto}
.sa-list{display:grid;grid-template-columns:repeat(auto-fill,minmax(240px,1fr));gap:12px;max-width:1000px;margin-left:auto!important;margin-right:auto;padding:0;list-style:none}
.sa-list li{position:relative;padding:16px 18px 16px 40px;border:1px solid var(--sa-line);border-radius:14px;background:#fff;font:600 16px/1.35 var(--sa-font-ui);color:var(--sa-deep)}
.sa-list li::before{content:"";position:absolute;left:18px;top:23px;width:8px;height:8px;border-radius:50%;background:var(--sa-accent)}
.sa-ornament{display:flex;justify-content:center}
.sa-ornament img{max-height:90px;width:auto;opacity:.9}
.sa-pager{display:grid;grid-template-columns:1fr 1fr;gap:16px;max-width:900px;margin-left:auto!important;margin-right:auto;padding-top:clamp(32px,5vw,56px);border-top:1px solid var(--sa-line)}
.sa-page-link{display:flex;flex-direction:column;gap:6px;padding:18px 22px;border:1px solid var(--sa-line);border-radius:16px;background:#fff;text-decoration:none;transition:border-color .2s,transform .25s var(--sa-ease),box-shadow .25s}
.sa-page-link:hover{border-color:var(--sa-deep);transform:translateY(-2px);box-shadow:0 14px 30px -18px rgba(0,0,0,.35)}
.sa-page-link small{font:600 12px var(--sa-font-ui);letter-spacing:.1em;text-transform:uppercase;color:var(--sa-muted)}
.sa-page-link span{font:700 18px/1.3 var(--sa-font-ui);color:var(--sa-deep)}
.sa-next{text-align:right;grid-column:2}
.sa-empty{text-align:center;color:var(--sa-muted);font-family:var(--sa-font-ui)}

/* footer */
.sa-footer{position:relative;background:var(--sa-night);color:rgba(255,255,255,.78);font-family:var(--sa-font-ui);view-transition-name:sa-footer}
.sa-footer::before{content:"";position:absolute;inset:0 0 auto 0;height:1px;background:linear-gradient(90deg,transparent,var(--sa-highlight),transparent);opacity:.6}
.sa-footer-grid{display:grid;grid-template-columns:minmax(220px,.8fr) 3.2fr;gap:clamp(32px,5vw,72px);padding-top:clamp(48px,6vw,80px);padding-bottom:clamp(32px,4vw,56px)}
.sa-fbrand .sa-brand{color:#fff;font-size:22px}
.sa-fbrand .sa-brand img{width:56px;height:56px}
.sa-flag{width:84px;height:auto;margin:22px 0 0;border-radius:4px;box-shadow:0 0 0 1px rgba(255,255,255,.1)}
.sa-fquick{display:flex;flex-wrap:wrap;gap:8px;margin-top:24px}
.sa-pill{display:inline-flex;align-items:center;min-height:40px;padding:0 16px;border:1px solid rgba(255,255,255,.18);border-radius:999px;color:#fff;font-size:14px;font-weight:500;text-decoration:none;transition:background .2s,border-color .2s}
.sa-pill:hover{background:rgba(255,255,255,.08);border-color:rgba(255,255,255,.4)}
.sa-pill.is-active{background:var(--sa-highlight);border-color:var(--sa-highlight);color:var(--sa-night)}
.sa-fcols{display:grid;grid-template-columns:repeat(auto-fit,minmax(170px,1fr));gap:clamp(24px,3vw,40px)}
.sa-fcol summary{display:flex;align-items:center;justify-content:space-between;list-style:none;cursor:default}
.sa-fcol summary::-webkit-details-marker{display:none}
.sa-fcol summary::before,.sa-fcol summary::after{content:none!important;display:none!important}
.sa-fcol summary svg{display:none;width:13px;height:9px;transition:transform .25s var(--sa-ease)}
.sa-fcol h2{margin:0 0 14px;font-size:12px;font-weight:700;letter-spacing:.14em;text-transform:uppercase;color:var(--sa-highlight)}
.sa-fcol ul{list-style:none;margin:0;padding:0}
.sa-fcol .sa-link{position:relative;display:inline-flex;align-items:center;min-height:36px;color:rgba(255,255,255,.82);font-size:15px;text-decoration:none;transition:color .2s}
.sa-fcol .sa-link::after{content:"";position:absolute;left:0;right:0;bottom:6px;height:1px;background:currentColor;transform:scaleX(0);transform-origin:left;transition:transform .3s var(--sa-ease)}
.sa-fcol .sa-link:hover{color:#fff}
.sa-fcol .sa-link:hover::after{transform:scaleX(1)}
.sa-fcol .sa-link.is-active{color:var(--sa-highlight)}
.sa-fcol .sa-link.is-active::before{content:"";position:absolute;left:-14px;top:50%;width:6px;height:6px;margin-top:-3px;border-radius:50%;background:var(--sa-highlight)}
.sa-fbottom{display:flex;align-items:center;justify-content:space-between;gap:16px;padding-top:20px;padding-bottom:24px;border-top:1px solid rgba(255,255,255,.08);font-size:13px;color:rgba(255,255,255,.5)}
.sa-totop{display:inline-flex;align-items:center;gap:8px;min-height:40px;padding:0 14px;border:1px solid rgba(255,255,255,.18);border-radius:999px;background:none;color:#fff;font:500 13px var(--sa-font-ui);cursor:pointer;transition:background .2s}
.sa-totop:hover{background:rgba(255,255,255,.08)}
.sa-totop svg{width:11px;height:8px}
.sa-live{position:absolute;width:1px;height:1px;overflow:hidden;clip:rect(0 0 0 0);white-space:nowrap}

/* responsive */
.sa-header.sa-compact .sa-nav{visibility:hidden;position:absolute;pointer-events:none}.sa-header.sa-compact .sa-burger{display:inline-grid}.sa-header.sa-compact .sa-actions{margin-left:auto}
@media (max-width:900px){.sa-nav{display:none}.sa-burger{display:inline-grid}.sa-actions{margin-left:auto}}
@media (max-width:820px){.sa-split{grid-template-columns:1fr}.sa-split.sa-rev .sa-prose{order:0}.sa-split .sa-prose{max-width:68ch;margin:0 auto}.sa-footer-grid{grid-template-columns:1fr}}
@media (max-width:720px){
 html.sa-reader{--sa-header-h:60px}
 .sa-brand{font-size:16px}.sa-brand img{width:34px;height:34px}
 .sa-hero h1{max-width:none}
 .sa-figure{border-radius:12px}
 .sa-pager{grid-template-columns:1fr}.sa-next{grid-column:1}
 .sa-fcols{grid-template-columns:1fr;gap:0}
 .sa-fcol{border-bottom:1px solid rgba(255,255,255,.08)}
 .sa-fcol summary{min-height:54px;cursor:pointer}
 .sa-fcol summary svg{display:block}
 .sa-fcol[open] summary svg{transform:rotate(180deg)}
 .sa-fcol h2{margin:0;font-size:13px}
 .sa-fcol ul{padding:0 0 12px}
 .sa-fcol .sa-link{min-height:44px;font-size:16px}
 .sa-fbottom{flex-direction:row}
}

/* page transitions */
::view-transition-old(sa-main){animation:sa-vt-out .22s var(--sa-ease) both}
::view-transition-new(sa-main){animation:sa-vt-in .42s var(--sa-ease) both}
html[data-sa-dir="back"]::view-transition-old(sa-main){animation-name:sa-vt-out-back}
html[data-sa-dir="back"]::view-transition-new(sa-main){animation-name:sa-vt-in-back}
::view-transition-old(sa-header),::view-transition-new(sa-header),::view-transition-old(sa-footer),::view-transition-new(sa-footer){animation:none}
@keyframes sa-vt-out{to{opacity:0;transform:translateY(-14px)}}
@keyframes sa-vt-in{from{opacity:0;transform:translateY(22px)}}
@keyframes sa-vt-out-back{to{opacity:0;transform:translateY(14px)}}
@keyframes sa-vt-in-back{from{opacity:0;transform:translateY(-22px)}}
.sa-out{opacity:0;transform:translateY(-10px);transition:opacity .16s,transform .16s}
.sa-in{animation:sa-vt-in .4s var(--sa-ease) both}
.sa-page .sa-content>*{animation:sa-rise .7s var(--sa-ease) both}
.sa-page .sa-content>*:nth-child(2){animation-delay:.06s}
.sa-page .sa-content>*:nth-child(3){animation-delay:.12s}
.sa-page .sa-content>*:nth-child(n+4){animation-delay:.18s}
@keyframes sa-rise{from{opacity:0;transform:translateY(18px)}}
@media (prefers-reduced-motion:reduce){.sa-root *,.sa-root *::before,.sa-root *::after{animation:none!important;transition:none!important}::view-transition-group(*),::view-transition-old(*),::view-transition-new(*){animation:none!important}}
`;

    // ---- UI strings ------------------------------------------------------
    const I18N = {
        en: {
            skip: 'Skip to content', menu: 'Menu', mainMenu: 'Main menu', search: 'Search',
            searchSite: 'Search the site…', close: 'Close', noResults: 'Nothing found',
            prev: '← Previous', next: 'Next →', top: 'Back to top', sitemap: 'Site map',
            sections: 'Sections', empty: 'This section is being prepared.',
        },
        ru: {
            skip: 'К содержанию', menu: 'Меню', mainMenu: 'Главное меню', search: 'Поиск',
            searchSite: 'Поиск по сайту…', close: 'Закрыть', noResults: 'Ничего не найдено',
            prev: '← Назад', next: 'Далее →', top: 'Наверх', sitemap: 'Карта сайта',
            sections: 'Разделы', empty: 'Раздел в разработке.',
        },
        uk: {
            skip: 'До змісту', menu: 'Меню', mainMenu: 'Головне меню', search: 'Пошук',
            searchSite: 'Пошук по сайту…', close: 'Закрити', noResults: 'Нічого не знайдено',
            prev: '← Назад', next: 'Далі →', top: 'Нагору', sitemap: 'Мапа сайту',
            sections: 'Розділи', empty: 'Розділ у розробці.',
        },
    };

    // ---- Figma export conventions ----------------------------------------
    // Pages link with hx-on:click="navigateToPage(event, '<frame name>.html')"
    // or window.open('<frame name>.html'); files are saved under
    // ConvertFileName(frame name) - same table as gozaurus
    // fabric/figma/translitMap.go, so links resolve to existing files.
    const ATTR_CLICK = 'hx-on\\:click';
    const LINK_SEL = `[${ATTR_CLICK}*="navigateToPage"], [${ATTR_CLICK}*="window.open"], a[href]`;
    const CHROME_TAGS = 'header, footer, nav, template, dialog, form';
    const CHROME_NAME = /(^|[\s_/-])(header|footer|top ?bar|nav(bar|igation)?|menu|burger|search|pop ?up|overlay|cookie)/i;
    const TRANSLIT = {
        'а': 'a', 'б': 'b', 'в': 'v', 'г': 'h', 'ґ': 'g', 'д': 'd', 'е': 'e', 'є': 'ye', 'ж': 'zh',
        'з': 'z', 'и': 'y', 'і': 'i', 'ї': 'yi', 'й': 'y', 'к': 'k', 'л': 'l', 'м': 'm', 'н': 'n',
        'о': 'o', 'п': 'p', 'р': 'r', 'с': 's', 'т': 't', 'у': 'u', 'ф': 'f', 'х': 'kh', 'ц': 'ts',
        'ч': 'ch', 'ш': 'sh', 'щ': 'shch', 'ы': 'y', 'э': 'e', 'ю': 'yu', 'я': 'ya', 'ё': 'yo',
        'ь': '', 'ъ': '', '’': '', "'": '',
    };

    function convertFileName(name) {
        let out = '';
        for (const ch of name) {
            const low = ch.toLowerCase();
            if (Object.prototype.hasOwnProperty.call(TRANSLIT, low)) {
                const t = TRANSLIT[low];
                out += (ch !== low && t) ? t[0].toUpperCase() + t.slice(1) : t;
            } else {
                out += ch;
            }
        }
        return out.replace(/[^a-zA-Z0-9_.]/g, '_');
    }

    const state = {
        cfg: null,
        t: I18N.en,
        base: '',          // directory of the site, e.g. '/Kareliia/'
        home: 'index.html',
        current: '',
        navToken: 0,
        raw: new Map(),    // file -> Promise<html text>
        pages: new Map(),  // file -> Promise<page model> (reader)
        started: false,
        // reader
        nav: [],
        files: new Map(),  // file -> {label, group}
        brand: {title: '', logo: '', mark: ''},
        // native
        keep: new Set(),       // our own body nodes (progress bar, live region)
        pageLinks: new Set(),  // stylesheets that belong to pages
    };

    const $ = (sel, root = document) => root.querySelector(sel);
    const esc = s => String(s).replace(/[&<>"']/g, c => ({
        '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'
    }[c]));
    const norm = s => String(s || '').replace(/\s+/g, ' ').trim();
    const reduceMotion = () => window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    const urlOf = file => state.base + file;
    const absUrl = file => new URL(urlOf(file), location.href).href;

    // ---- config ----------------------------------------------------------

    function readConfig() {
        let cfg = null;
        const node = document.getElementById('site-config');
        if (node) {
            try {
                cfg = JSON.parse(node.textContent || '{}');
            } catch (e) {
                console.error('SiteApp: bad #site-config JSON', e);
            }
        }
        const meta = document.querySelector('meta[name="app-mode"]');
        if (!cfg && meta && meta.content === 'site') cfg = {mode: 'site', layout: 'reader'};
        if (!cfg || (cfg.mode && cfg.mode !== 'site')) return null;
        const title = document.querySelector('meta[name="app-title"]');
        if (!cfg.title && title) cfg.title = title.content;
        cfg.layout = cfg.layout === 'reader' ? 'reader' : 'native';
        return cfg;
    }

    function pickStrings(cfg) {
        const lang = String(cfg.lang || document.documentElement.lang || navigator.language || 'en').toLowerCase();
        const base = I18N[lang.slice(0, 2)] || I18N.en;
        return Object.assign({}, base, cfg.strings || {});
    }

    // ---- links -----------------------------------------------------------

    // Target of a link element as written by the exporter (or a plain href).
    function targetOf(el) {
        if (el.tagName === 'A' && el.hasAttribute('href')) return {raw: el.getAttribute('href'), url: true};
        const code = el.getAttribute('hx-on:click') || '';
        const m = code.match(/(?:navigateToPage\(\s*event\s*,|window\.open\()\s*'([^']*)'/);
        return m ? {raw: m[1], url: /^(https?:)?\/\//.test(m[1]) || m[1].startsWith('/')} : null;
    }

    // Raw target -> local page file, or null when it is not a page of this
    // site. Frame names ('Read News / Project Page.html') are converted as a
    // whole; URLs by their last path segment. Absolute URLs of another host
    // only count when `known` (pages discovered in the menu) has the file.
    function resolveFile(raw, {url = true, known = null} = {}) {
        if (raw === undefined || raw === null) return null;
        raw = String(raw).trim().split('#')[0].split('?')[0];
        if (raw === '' || raw === '/' || raw === './') return state.home;

        if (!url) {
            let name = raw;
            try {
                name = decodeURIComponent(raw);
            } catch (e) { /* keep as is */
            }
            return /\.html?$/i.test(name) ? convertFileName(name.replace(/\.html?$/i, '')) + '.html' : null;
        }

        let u;
        try {
            u = new URL(raw, location.href);
        } catch (e) {
            return null;
        }
        if (!/^https?:$/.test(u.protocol)) return null;
        let file = u.pathname.split('/').pop() || state.home;
        try {
            file = decodeURIComponent(file);
        } catch (e) { /* keep */
        }
        if (!/\.html?$/i.test(file)) return null;
        file = convertFileName(file.replace(/\.html?$/i, '')) + '.html';

        if (u.origin !== location.origin) return known && known.has(file) ? file : null;
        const dir = decodeURIComponent(u.pathname.slice(0, u.pathname.lastIndexOf('/') + 1));
        return dir === decodeURIComponent(state.base) ? file : null;
    }

    function fileOfElement(el, known) {
        const t = targetOf(el);
        return t ? resolveFile(t.raw, {url: t.url, known}) : null;
    }

    // ---- loading ---------------------------------------------------------

    function loadRaw(file) {
        if (!state.raw.has(file)) {
            const p = fetch(urlOf(file), {credentials: 'same-origin'}).then(r => {
                if (!r.ok) throw new Error(`${r.status} ${file}`);
                return r.text();
            });
            p.catch(() => state.raw.delete(file));
            state.raw.set(file, p);
        }
        return state.raw.get(file);
    }

    const parseDoc = html => new DOMParser().parseFromString(html, 'text/html');

    // =====================================================================
    //  Router core - shared by both layouts
    // =====================================================================

    const renderer = {
        prepare: null,  // file -> Promise<page>    (load + parse, no DOM changes)
        swap: null,     // page -> void              (sync DOM update, runs inside the view transition)
        before: null,   // page -> Promise           (async work before the swap, e.g. stylesheets)
        title: null,    // page -> string
        root: null,     // () -> Element             (region the fallback animation runs on)
    };

    function progress(on) {
        document.documentElement.classList.toggle('sa-loading', on);
    }

    async function go(file, {push = true, scroll = 0, hash = ''} = {}) {
        const token = ++state.navToken;
        closePanels();
        if (push && history.state && history.state.saFile) {
            history.replaceState({...history.state, scroll: window.scrollY}, '');
        }
        progress(true);
        let page;
        try {
            page = await renderer.prepare(file);
            if (renderer.before) await renderer.before(page);
        } catch (e) {
            progress(false);
            console.error('SiteApp:', e);
            location.assign(urlOf(file) + hash);  // let the server answer
            return;
        }
        if (token !== state.navToken) return;    // a newer click won

        const swap = () => {
            renderer.swap(page);
            document.title = renderer.title(page);
            state.current = file;
            if (push) history.pushState({saFile: file, scroll: 0}, '', urlOf(file) + hash);
            const target = hash && document.getElementById(decodeURIComponent(hash.slice(1)));
            if (target) target.scrollIntoView();
            else window.scrollTo({top: scroll, behavior: 'instant'});
        };

        const finish = () => {
            progress(false);
            const live = $('.sa-live');
            if (live) live.textContent = document.title;
            if (push) {
                const h = (renderer.root() || document).querySelector('h1, [role="heading"]');
                if (h) {
                    if (!h.hasAttribute('tabindex')) h.setAttribute('tabindex', '-1');
                    h.focus({preventScroll: true});
                }
            }
            document.dispatchEvent(new CustomEvent('site:load', {detail: {file, url: urlOf(file)}}));
        };

        if (document.startViewTransition && !reduceMotion()) {
            document.documentElement.dataset.saDir = push ? 'forward' : 'back';
            const vt = document.startViewTransition(swap);
            vt.finished.finally(finish);
        } else {
            const old = renderer.root();
            if (!reduceMotion() && old) {
                old.classList.add('sa-out');
                await new Promise(r => setTimeout(r, 160));
                if (token !== state.navToken) return;
            }
            swap();
            const now = renderer.root();
            if (now) {
                now.classList.remove('sa-out');
                if (!reduceMotion()) {
                    now.classList.add('sa-in');
                    setTimeout(() => now.classList.remove('sa-in'), 400);
                }
            }
            finish();
        }
    }

    // Link elements of this site -> file, or null (left to the browser).
    function internalFile(el) {
        if (!el) return null;
        if (el.dataset && el.dataset.file) return el.dataset.file;
        if (el.tagName === 'A' && (el.target === '_blank' || el.hasAttribute('download'))) return null;
        return fileOfElement(el, state.files);
    }

    function bindRouter() {
        document.addEventListener('click', evt => {
            if (evt.defaultPrevented || evt.button !== 0 || evt.metaKey || evt.ctrlKey || evt.shiftKey || evt.altKey) return;
            const a = evt.target.closest('a[href]');
            const file = a && internalFile(a);
            if (!file) return;
            evt.preventDefault();
            const hash = new URL(a.href, location.href).hash;
            if (file === state.current && !hash) {
                closePanels();
                window.scrollTo({top: 0, behavior: reduceMotion() ? 'instant' : 'smooth'});
                return;
            }
            go(file, {hash});
        });

        // Exported pages navigate through this global from hx-on:click.
        window.navigateToPage = function (event, url) {
            const file = resolveFile(url, {
                url: /^(https?:)?\/\//.test(url) || String(url).startsWith('/'),
                known: state.files
            });
            const newTab = event && (event.ctrlKey || event.metaKey || event.button === 1);
            const newWindow = event && event.shiftKey && !newTab;
            if (newWindow) window.open(file ? urlOf(file) : url, '_blank', 'popup=yes');
            else if (newTab) window.open(file ? urlOf(file) : url, '_blank', 'noopener');
            else if (file) go(file);
            else window.location.assign(url);
        };

        // Prefetch on intent: hover, keyboard focus, touch.
        const prefetch = evt => {
            const el = evt.target.closest && evt.target.closest(LINK_SEL);
            const file = el && internalFile(el);
            if (file && file !== state.current) renderer.prepare(file).catch(() => {
            });
        };
        document.addEventListener('pointerover', prefetch, {passive: true});
        document.addEventListener('focusin', prefetch);
        document.addEventListener('touchstart', prefetch, {passive: true});

        window.addEventListener('popstate', evt => {
            const st = evt.state || {};
            const file = st.saFile || resolveFile(location.pathname) || state.home;
            go(file, {push: false, scroll: st.scroll || 0, hash: location.hash});
        });

        document.addEventListener('keydown', evt => {
            if (evt.key === 'Escape') closePanels();
        });
    }

    function closePanels(except) {
        if (state.cfg && state.cfg.layout === 'reader') readerClosePanels(except);
    }

    // =====================================================================
    //  Native layout - every page keeps its own design
    // =====================================================================

    const native = {
        rootSel: 'body > main',

        prepare(file) {
            return loadRaw(file).then(html => ({file, doc: parseDoc(html)}));
        },

        // New page stylesheets must be loaded before the swap, or the page
        // flashes unstyled in the middle of the transition.
        before(page) {
            const have = new Set([...document.querySelectorAll('link[rel="stylesheet"]')].map(l => l.href));
            const loads = [];
            page.doc.querySelectorAll('head link[rel="stylesheet"]').forEach(l => {
                const href = new URL(l.getAttribute('href'), absUrl(page.file)).href;
                if (have.has(href)) return;
                const link = document.createElement('link');
                link.rel = 'stylesheet';
                link.href = href;
                state.pageLinks.add(link);
                loads.push(new Promise(res => {
                    link.onload = link.onerror = res;
                    setTimeout(res, 3000);
                }));
                document.head.append(link);
            });
            // page-specific head scripts (e.g. animation libraries) - once
            const scripts = new Set([...document.scripts].map(sc => sc.src).filter(Boolean));
            page.doc.querySelectorAll('head script[src]').forEach(sc => {
                const url = new URL(sc.getAttribute('src'), absUrl(page.file)).href;
                if (scripts.has(url)) return;
                const n = document.createElement('script');
                n.src = url;
                n.async = false;
                loads.push(new Promise(res => {
                    n.onload = n.onerror = res;
                    setTimeout(res, 5000);
                }));
                document.head.append(n);
            });
            if (!page.doc.querySelector(native.rootSel)) {
                return Promise.reject(new Error(`no "${native.rootSel}" in ${page.file}`));
            }
            return Promise.all(loads);
        },

        swap(page) {
            const doc = page.doc;
            const pageUrl = absUrl(page.file);

            // stylesheets the new page doesn't use any more (only page ones -
            // links other scripts add at runtime are not ours to remove)
            const wanted = new Set([...doc.querySelectorAll('head link[rel="stylesheet"]')]
                .map(l => new URL(l.getAttribute('href'), pageUrl).href));
            state.pageLinks.forEach(l => {
                if (!wanted.has(l.href)) {
                    l.remove();
                    state.pageLinks.delete(l);
                }
            });

            // the old page, with whatever it appended to <body> at runtime
            // (overlays, popups ...), goes away completely
            [...document.body.children].forEach(n => {
                if (!state.keep.has(n)) n.remove();
            });

            // body attributes of the new page
            [...document.body.attributes].forEach(a => document.body.removeAttribute(a.name));
            [...doc.body.attributes].forEach(a => document.body.setAttribute(a.name, a.value));

            // the new page's body, in document order, before our own nodes
            const anchor = document.body.firstChild;
            const rootTag = native.rootSel.replace(/^body\s*>\s*/, '');
            let root = null;
            [...doc.body.children].forEach(n => {
                const node = document.importNode(n, true);
                document.body.insertBefore(node, anchor);
                if (!root) root = node.matches(rootTag) ? node : node.querySelector(rootTag);
            });
            if (doc.documentElement.lang) document.documentElement.lang = doc.documentElement.lang;

            // imported <script>s don't run - re-create them
            document.body.querySelectorAll('script').forEach(s => {
                if (state.keep.has(s) || s.closest('template')) return;
                const n = document.createElement('script');
                [...s.attributes].forEach(a => n.setAttribute(a.name, a.value));
                n.text = s.text;
                s.replaceWith(n);
            });

            if (root) {
                root.style.viewTransitionName = 'sa-main';
                adaptTouch(root);
                if (window.htmx) htmx.process(root);
            }
        },

        title(page) {
            return page.doc.title || state.cfg.title || document.title;
        },

        root() {
            return $(native.rootSel);
        },
    };

    // Exported hover interactions (hx-on:mouseenter) also on touch screens.
    function adaptTouch(root) {
        if (!window.matchMedia('(hover: none)').matches) return;
        root.querySelectorAll('[hx-on\\:mouseenter]').forEach(el => {
            if (!el.hasAttribute('hx-on:touchstart')) el.setAttribute('hx-on:touchstart', el.getAttribute('hx-on:mouseenter'));
        });
    }

    function startNative() {
        native.rootSel = state.cfg.root || 'body > main';
        const root = $(native.rootSel);
        if (!root) {
            console.warn(`SiteApp: no "${native.rootSel}" on the page - left as is`);
            return false;
        }
        document.querySelectorAll('head link[rel="stylesheet"]').forEach(l => state.pageLinks.add(l));
        root.style.viewTransitionName = 'sa-main';

        const live = document.createElement('div');
        live.className = 'sa-live';
        live.setAttribute('aria-live', 'polite');
        const bar = document.createElement('div');
        bar.className = 'sa-progress';
        bar.setAttribute('aria-hidden', 'true');
        document.body.prepend(bar, live);
        state.keep.add(bar).add(live);

        Object.assign(renderer, native);
        state.raw.set(state.current, Promise.resolve(document.documentElement.outerHTML));
        return true;
    }

    // =====================================================================
    //  Reader layout - content rendered into one responsive shell
    // =====================================================================

    // ---- menu discovery ----

    // The menu is the element whose *children* lead to the most different
    // pages: a header row, a footer column set or a burger popup all score
    // high, while <main> (links scattered in a few big sections) does not.
    function findMenu(docs) {
        const cfgSel = state.cfg.menu;
        const roots = [];
        docs.forEach(doc => {
            roots.push(doc);
            doc.querySelectorAll('template').forEach(t => roots.push(t.content));
        });
        if (cfgSel) {
            for (const r of roots) {
                const m = r.querySelector(cfgSel);
                if (m) return m;
            }
            console.warn(`SiteApp: menu "${cfgSel}" not found - detecting`);
        }

        const score = new Map(); // element -> {children:Set, files:Set}
        roots.forEach(r => r.querySelectorAll(LINK_SEL).forEach(link => {
            const file = fileOfElement(link, null) || (link.tagName === 'A' && foreignPage(link));
            if (!file) return;
            let child = link;
            for (let a = link.parentElement, depth = 0; a && depth < 10; child = a, a = a.parentElement, depth++) {
                if (a === r.body || a.tagName === 'HTML') break;
                let s = score.get(a);
                if (!s) score.set(a, s = {children: new Set(), files: new Set()});
                s.children.add(child);
                s.files.add(file);
            }
        }));

        let best = null, bs = null;
        score.forEach((s, el) => {
            if (s.children.size < 3) return;
            const better = !best ||
                s.children.size > bs.children.size ||
                (s.children.size === bs.children.size && s.files.size > bs.files.size) ||
                (s.children.size === bs.children.size && s.files.size === bs.files.size &&
                    el.getElementsByTagName('*').length < best.getElementsByTagName('*').length);
            if (better) {
                best = el;
                bs = s;
            }
        });
        return best;
    }

    // a[href] to the host the design was once published on: the frame name
    // is the file name. Accepted only inside the menu (see buildNav).
    function foreignPage(a) {
        try {
            const u = new URL(a.getAttribute('href'), location.href);
            if (u.origin === location.origin) return null;
            const name = decodeURIComponent(u.pathname.split('/').pop() || '');
            return /\.html?$/i.test(name) ? convertFileName(name.replace(/\.html?$/i, '')) + '.html' : null;
        } catch (e) {
            return null;
        }
    }

    function linkItem(el) {
        const file = fileOfElement(el, null) || (el.tagName === 'A' && foreignPage(el));
        const label = norm(el.textContent || el.dataset.name);
        return file && label ? {label, file} : null;
    }

    function buildNav(menu) {
        const nav = [];
        if (!menu) return nav;
        for (const child of menu.children) {
            if (child.matches(LINK_SEL)) {
                const it = linkItem(child);
                if (it) nav.push(it);
                continue;
            }
            const items = [];
            const seen = new Set();
            child.querySelectorAll(LINK_SEL).forEach(a => {
                const it = linkItem(a);
                if (it && !seen.has(it.file)) {
                    seen.add(it.file);
                    items.push(it);
                }
            });
            if (!items.length) continue;
            if (items.length === 1) {
                nav.push(items[0]);
                continue;
            }
            // group caption: plain (non-link) text of the group
            const caption = [...child.querySelectorAll('span, div, p, h1, h2, h3, h4, h5, h6, strong')].find(el =>
                !el.closest(LINK_SEL) && !el.querySelector(LINK_SEL) && norm(el.textContent));
            nav.push({label: norm(caption ? caption.textContent : child.dataset.name || items[0].label), items});
        }
        return nav;
    }

    function indexNav() {
        state.files.clear();
        state.nav.forEach(n => {
            if (n.items) n.items.forEach(i => state.files.set(i.file, {label: i.label, group: n}));
            else state.files.set(n.file, {label: n.label, group: null});
        });
    }

    // ---- design -> theme ----

    const hexOf = c => {
        c = String(c || '').trim();
        let m = c.match(/^#([0-9a-f]{6})([0-9a-f]{2})?$/i);
        if (m) return m[2] === '00' ? null : '#' + m[1].toLowerCase();
        m = c.match(/^rgba?\((\d+),\s*(\d+),\s*(\d+)(?:,\s*([\d.]+))?\)$/i);
        if (m && (m[4] === undefined || +m[4] > 0.5)) {
            return '#' + [m[1], m[2], m[3]].map(v => (+v).toString(16).padStart(2, '0')).join('');
        }
        return null;
    };
    const lum = hex => {
        const v = [1, 3, 5].map(i => parseInt(hex.slice(i, i + 2), 16) / 255)
            .map(c => (c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4));
        return 0.2126 * v[0] + 0.7152 * v[1] + 0.0722 * v[2];
    };
    const sat = hex => {
        const v = [1, 3, 5].map(i => parseInt(hex.slice(i, i + 2), 16));
        const mx = Math.max(...v), mn = Math.min(...v);
        return mx ? (mx - mn) / mx : 0;
    };
    const classColors = (el, prefix) => {
        const out = [];
        const re = new RegExp(`(?:^|\\s)${prefix}-\\[([^\\]]+)\\]`, 'g');
        let m;
        while ((m = re.exec(el.getAttribute('class') || ''))) {
            const h = hexOf(m[1].replace(/_/g, ' '));
            if (h) out.push(h);
        }
        return out;
    };
    const top = counts => [...counts.entries()].sort((a, b) => b[1] - a[1]).map(e => e[0]);

    // Palette and fonts of the exported design, as --sa-* defaults.
    function themeFromDesign(doc, menu) {
        const main = doc.querySelector('main') || doc.body;
        const vars = {};
        const paper = classColors(main, 'bg')[0];
        if (paper && lum(paper) > 0.6) vars.paper = paper;

        const darks = new Map();
        main.querySelectorAll('[class*="bg-["]').forEach(el => {
            if (el.closest('footer, [data-name="Footer"]')) return;
            classColors(el, 'bg').forEach(c => {
                if (lum(c) < 0.06) darks.set(c, (darks.get(c) || 0) + 1);
            });
        });
        const deep = top(darks).find(c => c !== '#000000') || top(darks)[0];
        if (deep) vars.deep = deep;

        const foot = doc.querySelector('footer, [data-name="Footer"]');
        const night = foot && classColors(foot, 'bg').find(c => lum(c) < 0.05);
        if (night) vars.night = night === '#000000' && deep ? `color-mix(in srgb, ${deep} 35%, #000)` : night;
        else if (deep) vars.night = `color-mix(in srgb, ${deep} 55%, #000)`;

        const accents = new Map();
        (menu ? [menu, ...menu.querySelectorAll('*')] : []).concat([...main.querySelectorAll('.hyperlink, [class*="text-["]')])
            .forEach(el => classColors(el, 'text').forEach(c => {
                if (sat(c) > 0.45 && lum(c) > 0.04 && lum(c) < 0.7) accents.set(c, (accents.get(c) || 0) + 1);
            }));
        const accent = top(accents)[0];
        if (accent) vars.accent = vars.highlight = accent;

        // fonts: family of the biggest text for headings, most used for text
        const fonts = new Map();
        let big = null, bigSize = 0;
        main.querySelectorAll('.text, p, h1, h2, h3').forEach(el => {
            const cls = el.getAttribute('class') || '';
            const f = cls.match(/font-\[(?:&#39;|')?([^\]'&]+)/);
            if (!f) return;
            const fam = f[1].split(',').map(s => s.replace(/_/g, ' ').trim()).filter(Boolean)
                .map(s => /\s/.test(s) ? `'${s}'` : s).join(', ');
            const len = norm(el.textContent).length;
            fonts.set(fam, (fonts.get(fam) || 0) + len);
            const sz = parseFloat((cls.match(/text-\[(\d+(?:\.\d+)?)px\]/) || [])[1]) || ({
                'text-3xl': 30,
                'text-4xl': 36,
                'text-5xl': 48,
                'text-6xl': 60
            }[(cls.match(/text-[3-6]xl/) || [])[0]] || 0);
            if (sz > bigSize) {
                bigSize = sz;
                big = fam;
            }
        });
        const textFont = top(fonts)[0];
        if (textFont) vars['font-text'] = `${textFont}, Georgia, serif`;
        if (big || textFont) vars['font-ui'] = `${big || textFont}, system-ui, sans-serif`;
        return vars;
    }

    const cssVars = vars => Object.entries(vars).map(([k, v]) => `--sa-${k}:${v}`).join(';');

    // ---- page model ----

    function isChrome(el, main) {
        for (let n = el; n && n !== main; n = n.parentElement) {
            if (n.matches(CHROME_TAGS)) return true;
            if (n.dataset && CHROME_NAME.test(n.dataset.name || '')) return true;
            if (state.cfg.exclude && n.matches(state.cfg.exclude)) return true;
        }
        return false;
    }

    function imgHref(el) {
        if (el.tagName.toLowerCase() === 'img') return el.getAttribute('src');
        const im = el.querySelector('image');
        return im && (im.getAttribute('href') || im.getAttribute('xlink:href'));
    }

    function bgImage(el) {
        const m = (el.getAttribute('class') || '').match(/bg-\[url\((?:&#39;|')?\.\.\/(img\/[^'&)]+)/);
        return m ? m[1] : null;
    }

    function posOf(el, main, mainH) {
        for (let n = el; n && n !== main; n = n.parentElement) {
            const cls = n.getAttribute('class') || '';
            const m = cls.match(/(?:^|\s)top-\[(-?[\d.]+)(%|px)\]/);
            if (m) {
                const v = parseFloat(m[1]);
                const left = cls.match(/(?:^|\s)left-\[(-?[\d.]+)px\]/);
                return {top: m[2] === '%' ? v : (mainH ? v / mainH * 100 : NaN), left: left ? parseFloat(left[1]) : 0};
            }
        }
        return null;
    }

    function paragraphsOf(el) {
        const clone = el.cloneNode(true);
        clone.querySelectorAll('br').forEach(br => br.replaceWith('\n'));
        return clone.textContent.split(/\n+/).map(norm).filter(Boolean);
    }

    function parsePage(html, file) {
        const doc = parseDoc(html);
        const main = doc.querySelector('main') || doc.body;
        const info = state.files.get(file) || {};
        let title = info.label || norm(main.dataset.name || doc.title);
        if (file === state.home || title === 'index') title = info.label || state.brand.title;
        const mainH = parseFloat(((main.getAttribute('class') || '').match(/(?:^|\s)h-\[([\d.]+)px\]/) || [])[1]) || 0;

        // menus inside the page are navigation, not content: elements with
        // at least 3 children that lead to different pages
        const menuKids = new Map();
        main.querySelectorAll(LINK_SEL).forEach(a => {
            if (!fileOfElement(a, state.files) && !(a.tagName === 'A' && foreignPage(a))) return;
            let child = a;
            for (let p = a.parentElement; p && p !== main.parentElement; child = p, p = p.parentElement) {
                if (!menuKids.has(p)) menuKids.set(p, new Set());
                menuKids.get(p).add(child);
            }
        });
        const menus = [...menuKids].filter(([, kids]) => kids.size >= 3).map(([el]) => el);
        const inMenu = el => menus.some(m => m.contains(el));

        const items = [];
        const seenText = new Set();
        let order = 0;
        main.querySelectorAll('.text, p, h1, h2, h3, h4, li, svg, img').forEach(el => {
            if (isChrome(el, main) || inMenu(el)) return;
            const tag = el.tagName.toLowerCase();
            if (tag === 'svg' || tag === 'img') {
                if (el.parentElement && el.parentElement.closest('svg')) return;
                const href = imgHref(el);
                if (!href || href === state.brand.logo || href === state.brand.mark) return;
                const w = parseFloat(el.getAttribute('width')) || el.naturalWidth || 0;
                const h = parseFloat(el.getAttribute('height')) || el.naturalHeight || 0;
                if (w < 120 || h < 80) return;
                items.push({type: 'img', src: href, w, h, pos: posOf(el, main, mainH), order: order++});
                return;
            }
            if (el.parentElement.closest('.text, p, h1, h2, h3, h4, li')) return;
            const paras = paragraphsOf(el);
            if (!paras.length) return;
            const text = norm(paras.join(' ')).toLowerCase();
            if ((text.match(/\p{L}/gu) || []).length < 3) return;    // "/ 01", "→", "×"
            if (seenText.has(text)) return;                          // same component twice
            seenText.add(text);
            if (text === norm(title).toLowerCase() || text === norm(main.dataset.name).toLowerCase()) return;
            const isHeading = /^h[1-4]$/.test(tag) ||
                (paras.length === 1 && text.length <= 70 && !/[.!?…]$/.test(text));
            items.push({type: isHeading ? 'h' : 'p', paras, pos: posOf(el, main, mainH), order: order++});
        });

        // visual order: absolutely placed frames by top/left, auto-layout by DOM
        const positioned = items.filter(i => i.pos && !isNaN(i.pos.top)).length;
        if (positioned >= items.length * 0.6) {
            items.forEach(i => {
                if (!i.pos || isNaN(i.pos.top)) i.pos = {top: 1e6 + i.order, left: 0};
            });
            items.sort((a, b) => (Math.abs(a.pos.top - b.pos.top) < 3 ? a.pos.left - b.pos.left : a.pos.top - b.pos.top));
        }

        // unique images; the biggest wide one becomes the hero
        const seen = new Set();
        const blocks = items.filter(i => i.type !== 'img' || (!seen.has(i.src) && seen.add(i.src)));
        let hero = null;
        blocks.filter(i => i.type === 'img' && i.w >= 600 && i.w / i.h > 0.9)
            .forEach(i => {
                if (!hero || i.w * i.h > hero.w * hero.h) hero = i;
            });
        if (hero) blocks.splice(blocks.indexOf(hero), 1);
        const heroSrc = hero ? hero.src
            : [main, ...main.querySelectorAll('[class*="bg-[url"]')].map(bgImage).find(Boolean) || null;

        const words = blocks.filter(b => b.type !== 'img').map(b => b.paras.join(' ')).join(' ');
        return {file, title, heroSrc, blocks, text: words};
    }

    function loadPage(file) {
        if (!state.pages.has(file)) {
            const p = loadRaw(file).then(html => parsePage(html, file));
            p.catch(() => state.pages.delete(file));
            state.pages.set(file, p);
        }
        return state.pages.get(file);
    }

    // ---- rendering ----

    const src = s => esc(s && !/^(https?:|\/|data:)/.test(s) ? state.base + s : s);
    const linkHTML = (it, cls = 'sa-link') =>
        `<a class="${cls}" href="${esc(urlOf(it.file))}" data-file="${esc(it.file)}">${esc(it.label)}</a>`;
    const chevron = (d = 'M1 1.5l5 5 5-5') => `<svg aria-hidden="true" viewBox="0 0 12 8"><path d="${d}"/></svg>`;

    function navHTML(kind) {
        return state.nav.map((n, i) => {
            if (!n.items) return kind === 'top' && n.file === state.home ? '' : `<li>${linkHTML(n)}</li>`;
            const id = `sa-${kind}-g${i}`;
            return `<li class="sa-group">
                <button class="sa-group-btn" type="button" aria-expanded="false" aria-controls="${id}">
                    <span>${esc(n.label)}</span>${chevron()}
                </button>
                <ul class="sa-sub" id="${id}">${n.items.map(it => `<li>${linkHTML(it)}</li>`).join('')}</ul>
            </li>`;
        }).join('');
    }

    function footerHTML() {
        const t = state.t;
        const singles = state.nav.filter(n => !n.items);
        const groups = state.nav.filter(n => n.items);
        const col = g => `
            <details class="sa-fcol" open>
                <summary><h2>${esc(g.label)}</h2>${chevron()}</summary>
                <ul>${g.items.map(it => `<li>${linkHTML(it)}</li>`).join('')}</ul>
            </details>`;
        return `
        <footer class="sa-footer">
            <div class="sa-wrap sa-footer-grid">
                <div class="sa-fbrand">
                    ${brandHTML(56)}
                    ${state.brand.mark ? `<img class="sa-flag" src="${src(state.brand.mark)}" alt="" loading="lazy">` : ''}
                    ${singles.length ? `<nav class="sa-fquick" aria-label="${esc(t.sections)}">${singles.map(n => linkHTML(n, 'sa-link sa-pill')).join('')}</nav>` : ''}
                </div>
                ${groups.length ? `<nav class="sa-fcols" aria-label="${esc(t.sitemap)}">${groups.map(col).join('')}</nav>` : ''}
            </div>
            <div class="sa-wrap sa-fbottom">
                <span>© ${new Date().getFullYear()} ${esc(state.brand.title)}</span>
                <button class="sa-totop" type="button" data-sa="top">${esc(t.top)}${chevron('M1 6.5l5-5 5 5')}</button>
            </div>
        </footer>`;
    }

    function brandHTML(size) {
        return `<a class="sa-brand" href="${esc(urlOf(state.home))}" data-file="${esc(state.home)}">
            ${state.brand.logo ? `<img src="${src(state.brand.logo)}" alt="" width="${size}" height="${size}">` : ''}
            <span>${esc(state.brand.title)}</span></a>`;
    }

    const searchIcon = '<svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="10.5" cy="10.5" r="6.5"/><path d="M15.5 15.5L21 21"/></svg>';

    function shellHTML() {
        const t = state.t;
        const search = state.cfg.search !== false;
        return `
        <a class="sa-skip" href="#sa-main">${esc(t.skip)}</a>
        <div class="sa-progress" aria-hidden="true"></div>
        <header class="sa-header">
            <div class="sa-wrap sa-header-row">
                ${brandHTML(40)}
                <nav class="sa-nav" aria-label="${esc(t.mainMenu)}"><ul>${navHTML('top')}</ul></nav>
                <div class="sa-actions">
                    ${search ? `<button class="sa-icon" type="button" data-sa="search" aria-label="${esc(t.search)}">${searchIcon}</button>` : ''}
                    <button class="sa-icon sa-burger" type="button" data-sa="menu" aria-label="${esc(t.menu)}" aria-expanded="false" aria-controls="sa-drawer">
                        <span></span><span></span><span></span>
                    </button>
                </div>
            </div>
        </header>
        <div class="sa-drawer" id="sa-drawer" hidden>
            <div class="sa-drawer-panel" role="dialog" aria-modal="true" aria-label="${esc(t.menu)}">
                <nav aria-label="${esc(t.menu)}"><ul class="sa-drawer-list">${navHTML('drawer')}</ul></nav>
            </div>
        </div>
        ${search ? `<div class="sa-search" id="sa-search" hidden>
            <div class="sa-search-panel" role="dialog" aria-modal="true" aria-label="${esc(t.search)}">
                <div class="sa-search-row">${searchIcon}
                    <input type="search" placeholder="${esc(t.searchSite)}" aria-label="${esc(t.searchSite)}" autocomplete="off">
                    <button class="sa-icon" type="button" data-sa="close" aria-label="${esc(t.close)}">
                        <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6 6l12 12M18 6L6 18"/></svg></button>
                </div>
                <ol class="sa-results" aria-live="polite"></ol>
            </div>
        </div>` : ''}
        <main id="sa-main" class="sa-main" tabindex="-1"></main>
        ${footerHTML()}
        <div class="sa-live" aria-live="polite" aria-atomic="true"></div>`;
    }

    function figure(b, cls = '') {
        const ratio = b.w && b.h ? ` style="aspect-ratio:${(b.w / b.h).toFixed(3)}"` : '';
        return `<figure class="sa-figure ${cls}"><img src="${src(b.src)}" alt="" loading="lazy" decoding="async"${ratio}></figure>`;
    }

    function pageHTML(page) {
        const t = state.t;
        const info = state.files.get(page.file) || {};
        const group = info.group;
        let title = page.title;
        if (group && title.toLowerCase().startsWith(group.label.toLowerCase())) {
            title = norm(title.slice(group.label.length).replace(/^[:\s–-]+/, '')) || title;
        }

        const parts = [];
        let firstP = true, splits = 0;
        const bs = page.blocks;
        for (let i = 0; i < bs.length; i++) {
            const b = bs[i];
            if (b.type === 'img') {
                if (b.w < 320 && b.h < 260) parts.push(`<div class="sa-ornament"><img src="${src(b.src)}" alt="" loading="lazy"></div>`);
                else parts.push(figure(b, 'sa-wide'));
                continue;
            }
            if (b.type === 'h') {
                if (group && norm(b.paras[0]).toLowerCase() === group.label.toLowerCase()) continue;
                let j = i;
                while (j < bs.length && bs[j].type === 'h') j++;
                if (j - i >= 3) {   // feature lists, tags, steps - not a pile of headings
                    parts.push(`<ul class="sa-list">${bs.slice(i, j).map(h => `<li>${esc(h.paras[0])}</li>`).join('')}</ul>`);
                    i = j - 1;
                    continue;
                }
                parts.push(`<h2 class="sa-h2">${esc(b.paras[0])}</h2>`);
                continue;
            }
            const text = b.paras.map(p => {
                const cls = firstP ? ' class="sa-lead"' : '';
                firstP = false;
                return `<p${cls}>${esc(p)}</p>`;
            }).join('');
            const next = bs[i + 1];
            if (next && next.type === 'img' && next.w >= 320) {
                parts.push(`<section class="sa-split${splits++ % 2 ? ' sa-rev' : ''}"><div class="sa-prose">${text}</div>${figure(next)}</section>`);
                i++;
            } else {
                parts.push(`<section class="sa-prose">${text}</section>`);
            }
        }

        let pager = '';
        if (group) {
            const k = group.items.findIndex(it => it.file === page.file);
            const a = (it, dir) => it ? `<a class="sa-page-link sa-${dir}" href="${esc(urlOf(it.file))}" data-file="${esc(it.file)}">
                <small>${esc(dir === 'prev' ? t.prev : t.next)}</small><span>${esc(it.label)}</span></a>` : '<span></span>';
            pager = `<nav class="sa-pager" aria-label="${esc(group.label)}">${a(group.items[k - 1], 'prev')}${a(group.items[k + 1], 'next')}</nav>`;
        }

        const hero = page.heroSrc ? `<img class="sa-hero-img" src="${src(page.heroSrc)}" alt="" fetchpriority="high" decoding="async">` : '';
        return `
        <article class="sa-page" data-file="${esc(page.file)}">
            <header class="sa-hero${hero ? '' : ' sa-hero-plain'}">
                ${hero}
                <div class="sa-wrap sa-hero-text">
                    ${group ? `<p class="sa-eyebrow">${esc(group.label)}</p>` : ''}
                    <h1 tabindex="-1">${esc(title)}</h1>
                </div>
            </header>
            <div class="sa-wrap sa-content">${parts.join('\n') || `<p class="sa-empty">${esc(t.empty)}</p>`}${pager}</div>
        </article>`;
    }

    function markActive(file) {
        document.querySelectorAll('.sa-link, .sa-brand').forEach(a => {
            const on = a.dataset.file === file;
            a.classList.toggle('is-active', on);
            if (on && a.classList.contains('sa-link')) a.setAttribute('aria-current', 'page');
            else a.removeAttribute('aria-current');
        });
        document.querySelectorAll('.sa-group').forEach(g =>
            g.classList.toggle('is-active', !!g.querySelector(`[data-file="${CSS.escape(file)}"]`)));
    }

    const reader = {
        prepare: loadPage,
        before: null,
        swap(page) {
            $('#sa-main').innerHTML = pageHTML(page);
            markActive(page.file);
        },
        title(page) {
            return page.title === state.brand.title || !state.brand.title ? page.title : `${page.title} — ${state.brand.title}`;
        },
        root() {
            return $('#sa-main');
        },
    };

    // ---- panels ----

    function readerClosePanels(except) {
        if (except !== 'menu') toggleDrawer(false);
        if (except !== 'search') toggleSearch(false);
        document.querySelectorAll('.sa-nav .sa-group-btn[aria-expanded="true"]').forEach(b => b.setAttribute('aria-expanded', 'false'));
    }

    function toggleDrawer(open) {
        const d = $('#sa-drawer'), btn = $('[data-sa="menu"]');
        if (!d) return;
        if (open === undefined) open = d.hidden;
        if (open === !d.hidden) return;
        btn.setAttribute('aria-expanded', String(open));
        document.documentElement.classList.toggle('sa-lock', open);
        if (open) {
            d.hidden = false;
            requestAnimationFrame(() => d.classList.add('is-open'));
            const first = d.querySelector('a, button');
            if (first) first.focus();
        } else {
            d.classList.remove('is-open');
            setTimeout(() => {
                if (!d.classList.contains('is-open')) d.hidden = true;
            }, 250);
        }
    }

    let searchReady = null;

    function toggleSearch(open) {
        const s = $('#sa-search');
        if (!s) return;
        if (open === undefined) open = s.hidden;
        if (open === !s.hidden) return;
        document.documentElement.classList.toggle('sa-lock', open);
        if (open) {
            s.hidden = false;
            requestAnimationFrame(() => s.classList.add('is-open'));
            const input = s.querySelector('input');
            input.focus();
            input.select();
            searchReady = searchReady || Promise.allSettled([...state.files.keys()].map(loadPage));
            searchReady.then(() => runSearch(input.value));
        } else {
            s.classList.remove('is-open');
            setTimeout(() => {
                if (!s.classList.contains('is-open')) s.hidden = true;
            }, 200);
        }
    }

    async function runSearch(q) {
        const list = $('.sa-results');
        q = norm(q).toLowerCase();
        if (q.length < 2) {
            list.innerHTML = '';
            return;
        }
        const pages = (await Promise.allSettled([...state.files.keys()].map(loadPage)))
            .filter(r => r.status === 'fulfilled').map(r => r.value);
        const hits = [];
        for (const p of pages) {
            const inTitle = p.title.toLowerCase().includes(q);
            const at = p.text.toLowerCase().indexOf(q);
            if (!inTitle && at < 0) continue;
            let snip = '';
            if (at >= 0) {
                const s = Math.max(0, at - 60), e = Math.min(p.text.length, at + q.length + 90);
                snip = (s ? '…' : '') + esc(p.text.slice(s, at)) + '<mark>' + esc(p.text.slice(at, at + q.length)) + '</mark>' +
                    esc(p.text.slice(at + q.length, e)) + (e < p.text.length ? '…' : '');
            }
            hits.push({p, score: inTitle ? 0 : 1, snip});
        }
        hits.sort((a, b) => a.score - b.score);
        list.innerHTML = hits.length ? hits.map(h =>
                `<li><a class="sa-result" href="${esc(urlOf(h.p.file))}" data-file="${esc(h.p.file)}">
                <strong>${esc(h.p.title)}</strong>${h.snip ? `<span>${h.snip}</span>` : ''}</a></li>`).join('')
            : `<li class="sa-noresult">${esc(state.t.noResults)}</li>`;
    }

    function bindReader() {
        document.addEventListener('click', evt => {
            const btn = evt.target.closest('[data-sa], .sa-group-btn');
            if (btn) {
                const act = btn.dataset.sa;
                if (act === 'menu') toggleDrawer();
                else if (act === 'search') {
                    readerClosePanels('search');
                    toggleSearch(true);
                } else if (act === 'close') readerClosePanels();
                else if (act === 'top') window.scrollTo({top: 0, behavior: reduceMotion() ? 'instant' : 'smooth'});
                else if (btn.classList.contains('sa-group-btn')) {
                    const open = btn.getAttribute('aria-expanded') !== 'true';
                    if (btn.closest('.sa-nav')) readerClosePanels();
                    btn.setAttribute('aria-expanded', String(open));
                }
                return;
            }
            if (evt.target.matches('.sa-drawer, .sa-search')) readerClosePanels();
            else if (!evt.target.closest('.sa-nav')) {
                document.querySelectorAll('.sa-nav .sa-group-btn[aria-expanded="true"]').forEach(b => b.setAttribute('aria-expanded', 'false'));
            }
        });

        document.addEventListener('keydown', evt => {
            if (!$('#sa-search')) return;
            if ((evt.key === '/' && !evt.target.closest('input, textarea, [contenteditable]')) ||
                (evt.key.toLowerCase() === 'k' && (evt.metaKey || evt.ctrlKey))) {
                evt.preventDefault();
                readerClosePanels('search');
                toggleSearch(true);
            }
        });

        const input = $('#sa-search input');
        if (input) {
            let t;
            input.addEventListener('input', evt => {
                clearTimeout(t);
                t = setTimeout(() => runSearch(evt.target.value), 120);
            });
        }

        const header = $('.sa-header');
        const onScroll = () => header.classList.toggle('is-scrolled', window.scrollY > 24);
        window.addEventListener('scroll', onScroll, {passive: true});
        onScroll();

        // the menu comes from the design and can be any length:
        // switch to the burger as soon as the row doesn't fit
        const nav = $('.sa-nav'), row = $('.sa-header-row');
        const fit = () => {
            header.classList.remove('sa-compact');
            const used = [...row.children].reduce((w, el) => w + el.getBoundingClientRect().width, 0);
            const gap = parseFloat(getComputedStyle(row).columnGap) || 0;
            if (nav.scrollWidth > nav.clientWidth + 1 || used + gap * 3 > row.clientWidth) header.classList.add('sa-compact');
        };
        new ResizeObserver(fit).observe(row);
        if (document.fonts) document.fonts.ready.then(fit);

        const mq = window.matchMedia('(max-width: 720px)');
        const syncCols = () => document.querySelectorAll('.sa-fcol').forEach(d => {
            d.open = !mq.matches;
        });
        mq.addEventListener('change', syncCols);
        syncCols();
    }

    // Some frames carry an older or shorter menu instance, so the home
    // page is the site map; this page is the fallback.
    async function homeDoc() {
        if (state.current === state.home) return document;
        try {
            const html = await loadRaw(state.home);
            return parseDoc(html);
        } catch (e) {
            return document;
        }
    }

    function findLogo(docs) {
        const cfgLogo = state.cfg.logo;
        if (cfgLogo && /\.(png|jpe?g|gif|svg|webp|avif)(\?|$)/i.test(cfgLogo)) return cfgLogo;
        for (const doc of docs) {
            if (cfgLogo) {
                const el = doc.querySelector(cfgLogo);
                if (el) return imgHref(el.matches('svg, img') ? el : el.querySelector('svg, img') || el);
            }
            // an image that takes you home
            for (const el of doc.querySelectorAll(LINK_SEL)) {
                if (fileOfElement(el, null) !== state.home) continue;
                const img = el.matches('svg, img') ? el : el.querySelector('svg, img');
                const href = img && imgHref(img);
                if (href) return href;
            }
        }
        return '';
    }

    async function startReader() {
        const docs = [await homeDoc(), document].filter((d, i, a) => a.indexOf(d) === i);
        const menu = findMenu(docs);
        state.nav = buildNav(menu);
        if (!state.nav.length) {
            console.warn('SiteApp: no menu found - reader layout needs one ("menu" in site.json)');
            return false;
        }
        indexNav();

        const home = state.files.get(state.home);
        const folder = decodeURIComponent(state.base.split('/').filter(Boolean).pop() || '').replace(/[-_]+/g, ' ');
        state.brand.title = state.cfg.title || (home && home.label) || norm(folder) || location.hostname;
        state.brand.logo = findLogo(docs);
        state.brand.mark = state.cfg.mark || '';

        // theme: design-derived defaults < site CSS < config "theme"
        const auto = document.createElement('style');
        auto.id = 'sa-theme';
        auto.textContent = `html.sa-reader{${cssVars(themeFromDesign(docs[0], menu))}}`;
        $('#sa-style').after(auto);
        if (state.cfg.theme) {
            Object.entries(state.cfg.theme).forEach(([k, v]) =>
                document.documentElement.style.setProperty(k.startsWith('--') ? k : `--sa-${k}`, v));
        }

        state.pages.set(state.current, Promise.resolve(parsePage(document.documentElement.outerHTML, state.current)));

        // exported markup stays in the DOM (no-JS / crawlers) but hidden
        [...document.body.children].forEach(n => {
            if (n.tagName !== 'SCRIPT') {
                n.hidden = true;
                n.inert = true;
            }
        });
        const shell = document.createElement('div');
        shell.className = 'sa-root';
        shell.innerHTML = shellHTML();
        document.body.prepend(shell);
        document.documentElement.classList.add('sa-reader');

        Object.assign(renderer, reader);
        bindReader();
        return true;
    }

    // =====================================================================

    function enabled() {
        return !!readConfig();
    }

    function start() {
        if (state.started) return false;
        const cfg = readConfig();
        if (!cfg) return false;
        state.started = true;
        state.cfg = cfg;
        state.t = pickStrings(cfg);
        state.home = cfg.home || 'index.html';
        const path = location.pathname;
        state.base = path.slice(0, path.lastIndexOf('/') + 1);
        state.current = resolveFile(location.href) || state.home;
        const root = document.documentElement;

        const st = document.createElement('style');
        st.id = 'sa-style';
        st.textContent = STYLES;
        document.head.prepend(st);
        if (cfg.layout === 'native') root.classList.remove('sa-wait');
        else root.classList.add('sa-boot');

        const done = ok => {
            root.classList.remove('sa-boot', 'sa-wait');
            if (!ok) return;
            bindRouter();
            history.replaceState({saFile: state.current, scroll: window.scrollY}, '', location.href);
            if ('scrollRestoration' in history) history.scrollRestoration = 'manual';
            if (cfg.layout === 'reader') {
                if (!root.lang) root.lang = cfg.lang || 'en';
                go(state.current, {push: false, scroll: 0, hash: location.hash});
            }
        };

        if (cfg.layout === 'reader') startReader().then(done, e => {
            console.error('SiteApp:', e);
            done(false);
        });
        else done(startNative());
        return true;
    }

    return {
        start, enabled, go, resolveFile, convertFileName,
        loadPage: file => renderer.prepare(file),
        get state() {
            return state;
        },
    };
})();

// App is the page-lifecycle/content-loading singleton: one instance, built
// once at parse time, holding what used to be a dozen unrelated top-level
// functions. Everything below the IIFE is a same-named global wrapper (e.g.
// `function loadContent(url){ return App.loadContent(url); }`) so every
// existing caller - forms.js, observer.js, set_clicks.js, setFlatPickr.js,
// table-row.js, and any onclick="..."/hx-on="..." attribute already baked into
// rendered HTML - keeps working unchanged. New code should call App.* directly.
//
// This file no longer touches page regions by hand. Content, sidebar,
// breadcrumbs and <title> are swapped by htmx from the response (swap target +
// out-of-band elements), history/back/forward are htmx's own (hx-push-url),
// and every request - hx-* element, boosted link/form, or script - gets its
// auth/language headers from Auth.headers() via observer.js's
// htmx:configRequest. See navigate()/loadContent() below.
const App = (function () {

    function fancyOpen(data) {
        $.fancybox.open({
            'autoScale': true,
            'transitionIn': 'elastic',
            'transitionOut': 'elastic',
            'speedIn': 500,
            'speedOut': 300,
            'type': 'html',
            'autoDimensions': true,
            'centerOnScroll': true,
            'content': data
        });
        setTimeout(function () {
            processAll('.fancybox-inner');
        }, 1000);
    }

    // replace special symbols
    function replMacros(url) {
        return url.replace(/{page}/, getPageLines() * 2)
    }

    // get lines for table according to windows height
    function getPageLines() {
        return Math.round((window.innerHeight - 60) / 22)
    }

    function loadStyles(id, styles) {
        let $head = $('head > style#' + id);
        if ($head.length === 0) {
            $('head').append('<style title="themes" id="' + id + '">' + styles + '</style>');
        } else {
            $head.html(styles);
        }
    }

    function loadCSSOnce(url) {
        if (document.querySelector(`link[href="${url}"]`)) return;

        const link = document.createElement("link");
        link.rel = "stylesheet";
        link.href = url;
        document.head.appendChild(link);
    }

    function addStyles(name, css) {
        $('head').append(`<style title="${name}">${css}</style>`);
    }

    function addScript(name, js) {
        $('head').append(`<script title="${name}">${js}</script>`);
    }

    // Script-initiated navigation, done the htmx way - the single replacement
    // for the old setContent()/putContent()/findAndReplaceElem()/
    // setDocumentHash()/onPopState() set, which fetched a page with a custom
    // handler and then copied pieces of it (#content, breadcrumbs, sidebar,
    // topline, <title>) into the DOM by hand and kept their own history.
    // Now htmx does all of that from the response itself:
    //   * the main body is swapped into `target` (innerHTML);
    //   * every other region the page owns arrives as an out-of-band element
    //     in the same response (<div id="catalog_sidebar" hx-swap-oob="true">,
    //     <breadcrumbs id="breadcrumbs" hx-swap-oob="true">, ...) - see
    //     clearStaleSidebar() in observer.js for the "no sidebar" case;
    //   * <title> in the response sets document.title;
    //   * history, back/forward and the page snapshot are htmx's own
    //     (hx-push-url) - no popstate handler of ours.
    // htmx reads hx-push-url from the request's source element, so each call
    // gets a throw-away hidden one (a shared element would make htmx drop a
    // second request while the first is still in flight). Being a normal
    // htmx request it picks up Auth.headers() in htmx:configRequest, and a
    // 401 goes through htmx:responseError -> Auth.relogin(evt).
    function navigate(url, {target = '#content', select, push = true} = {}) {
        const source = document.createElement('span');
        source.hidden = true;
        if (push) source.setAttribute('hx-push-url', 'true');
        document.body.append(source);

        const done = () => setTimeout(() => source.remove(), 0);
        return htmx.ajax('GET', url, {source, target, select, swap: 'innerHTML'}).then(done, done);
    }

    function loadContent(url) {
        return navigate(url);
    }

    function loadJScript(url, asyncS, cacheS, successFunc, completeFunc) {
        $.ajax({
            type: "GET",
            async: asyncS,
            cache: cacheS,
            url: url,
            global: false,
            dataType: "script",
            success: successFunc,
            complete: completeFunc,
            error: errorLoadResource
        });
    }

    // `cacheS` is kept only for call-signature compatibility (changeTheme()
    // in user.js still passes it) - htmx.ajax has no cache-busting option to
    // forward it to, and `true` (the only value ever passed) is the normal
    // browser-cache behavior anyway. Errors fall back to the shared
    // htmx:responseError handling (observer.js) instead of errorLoadResource
    // - htmx.ajax's `handler` option only runs on success, so a per-call
    // error callback isn't available here any more; see loadJScript() below
    // for the one remaining $.ajax call that still uses errorLoadResource
    // directly (loading a <script>, not data - out of scope for this
    // htmx conversion).
    function loadCSS(url, cacheS, successFunc) {
        htmx.ajax('GET', url, {
            swap: 'none',
            handler: (elt, info) => successFunc(info.xhr.responseText, info.xhr.statusText, info.xhr)
        });
    }

    function errorLoadResource(xhr, textStatus, errorThrown) {
        if (errorThrown !== undefined) {
            console.error(`%s from '${xhr}'! (${textStatus}). Pls, reload page!`, errorThrown);
        } else {
            console.error(`Can't load resource from '${xhr}'! (${textStatus}). Pls, reload page! %s`, textStatus);
        }
    }

    function isScrollableY(node) {
        const overflowY = window.getComputedStyle(node)['overflow-y'];
        return (overflowY === 'scroll' || overflowY === 'auto');
    }

    function isScrollableX(node) {
        const overflowX = window.getComputedStyle(node)['overflow-x'];
        return (overflowX === 'scroll' || overflowX === 'auto');
    }

    function setTableEvents() {
        if (window.setSortedClasses === undefined) {
            loadJScript("/js/table.js", false, true, setSortedClasses);
        }
    }

    // init runs once, on document ready - everything the old bare
    // `$(function () {...})` block at the bottom of app.js used to do.
    function init() {
        console.log("v1.4.0");
        // Static site exported by the Figma parser -> single-page app.
        // A public site must not ask "leave this page?" on every exit, so
        // the beforeunload guard below is for the admin apps only.
        const isSite = SiteApp.start() || SiteApp.enabled();
        if (!isSite) window.addEventListener("beforeunload", evt => {
            evt = evt || window.event;

            if (evt) {
                var y = evt.pageY || evt.clientY;
                if (y === undefined) {
                    console.log(evt);
                }
                console.log(`beforeunload ${document.location} pageY:${y}`);
                evt.preventDefault();
                if (y < 0) {
                    return evt.returnValue = "Do you want to close this page?";
                }
            }
            return false;
        });

        Auth.ensureUser();

        processAll(document.body);
        createObserver();
    }

    // impotent set BEFORE load all srcipt
    cfgHTMX();

    return {
        init,
        fancyOpen,
        replMacros,
        GetPageLines: getPageLines,
        LoadStyles: loadStyles,
        loadCSSOnce,
        AddStyles: addStyles,
        AddScript: addScript,
        loadContent,
        navigate,
        LoadJScript: loadJScript,
        LoadCSS: loadCSS,
        errorLoadResource,
        isScrollableY,
        isScrollableX,
        SetTableEvents: setTableEvents,
        site: SiteApp,
    };
})();

$(App.init);

// --- Backward-compatible global aliases ------------------------------------
// Same names, same call signatures as the pre-singleton app.js - every file
// that still calls one of these bare (forms.js, observer.js, set_clicks.js,
// setFlatPickr.js, table-row.js, generated HTML) keeps working with zero
// changes. New code should prefer App.* directly. (relogin() moved to
// user.js - it's Auth.relogin(evt) now, not App.relogin(url).)
function fancyOpen(data) {
    return App.fancyOpen(data);
}

function replMacros(url) {
    return App.replMacros(url);
}

function GetPageLines() {
    return App.GetPageLines();
}

function LoadStyles(id, styles) {
    return App.LoadStyles(id, styles);
}

function loadCSSOnce(url) {
    return App.loadCSSOnce(url);
}

function AddStyles(name, css) {
    return App.AddStyles(name, css);
}

function AddScript(name, js) {
    return App.AddScript(name, js);
}

function loadContent(url) {
    return App.loadContent(url);
}

function navigate(url, opts) {
    return App.navigate(url, opts);
}

function LoadJScript(url, asyncS, cacheS, successFunc, completeFunc) {
    return App.LoadJScript(url, asyncS, cacheS, successFunc, completeFunc);
}

function LoadCSS(url, cacheS, successFunc) {
    return App.LoadCSS(url, cacheS, successFunc);
}

function errorLoadResource(xhr, textStatus, errorThrown) {
    return App.errorLoadResource(xhr, textStatus, errorThrown);
}

function isScrollableY(node) {
    return App.isScrollableY(node);
}

function isScrollableX(node) {
    return App.isScrollableX(node);
}

function SetTableEvents() {
    return App.SetTableEvents();
}