// Code generated by qtc from "header.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line views/templates/layouts/header.qtpl:1
package layouts

//line views/templates/layouts/header.qtpl:1
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line views/templates/layouts/header.qtpl:1
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line views/templates/layouts/header.qtpl:1
func StreamHeaderHTML(qw422016 *qt422016.Writer, TopMenu Menu) {
//line views/templates/layouts/header.qtpl:1
	qw422016.N().S(`
<header class="main-header">
<div class="topline">
    <nav class="topline-navbar">
        `)
//line views/templates/layouts/header.qtpl:5
	TopMenu.StreamRenderMenu(qw422016, "top-mnu-list", "top-mnu-item")
//line views/templates/layouts/header.qtpl:5
	qw422016.N().S(`
    </nav>
    <div class="topline-btns">
        <a href="/user/signout/" id="bLogOut" onclick="return logOut(this);">
            <span class="glyphicon-log-out"> </span>
        </a>
        <a href="#" id="burger" class="brg-mnu"></a>
        <a id="bLogin" href="/show/forms/?name=signin" class="navbar-link btn-login" title="откроется в модальном окне" >Авторизоваться
        <span class="glyphicon-info-sign"> </span></a>
    </div>
</div>
</header>
`)
//line views/templates/layouts/header.qtpl:17
}

//line views/templates/layouts/header.qtpl:17
func WriteHeaderHTML(qq422016 qtio422016.Writer, TopMenu Menu) {
//line views/templates/layouts/header.qtpl:17
	qw422016 := qt422016.AcquireWriter(qq422016)
//line views/templates/layouts/header.qtpl:17
	StreamHeaderHTML(qw422016, TopMenu)
//line views/templates/layouts/header.qtpl:17
	qt422016.ReleaseWriter(qw422016)
//line views/templates/layouts/header.qtpl:17
}

//line views/templates/layouts/header.qtpl:17
func HeaderHTML(TopMenu Menu) string {
//line views/templates/layouts/header.qtpl:17
	qb422016 := qt422016.AcquireByteBuffer()
//line views/templates/layouts/header.qtpl:17
	WriteHeaderHTML(qb422016, TopMenu)
//line views/templates/layouts/header.qtpl:17
	qs422016 := string(qb422016.B)
//line views/templates/layouts/header.qtpl:17
	qt422016.ReleaseByteBuffer(qb422016)
//line views/templates/layouts/header.qtpl:17
	return qs422016
//line views/templates/layouts/header.qtpl:17
}
