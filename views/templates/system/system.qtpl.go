// This file is automatically generated by qtc from "system.qtpl".
// See https://github.com/valyala/quicktemplate for details.

//line views/templates/system/system.qtpl:1
package system

//line views/templates/system/system.qtpl:1
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

// All the text outside function templates is treated as comments,
// i.e. it is just ignored by quicktemplate compiler (`qtc`). It is for humans.
//
// предназначен для оформления выдачи сообщений управления веб-сервером.
// route - имя пути, на который надо сделать запрос и полученный ответ показать на странице
//

//line views/templates/system/system.qtpl:7
import (
	"strings"
)

//line views/templates/system/system.qtpl:11
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line views/templates/system/system.qtpl:11
func StreamAddRescanJS(qw422016 *qt422016.Writer, route []string) {
	//line views/templates/system/system.qtpl:11
	qw422016.N().S(`
    `)
	//line views/templates/system/system.qtpl:13
	arr := strings.Join(route, "','")

	//line views/templates/system/system.qtpl:14
	qw422016.N().S(`
    <script src="/request.js"></script>

    <script>
        var arr = [ '`)
	//line views/templates/system/system.qtpl:18
	qw422016.N().S(arr)
	//line views/templates/system/system.qtpl:18
	qw422016.N().S(`'  ];
        queueRequests(arr);
    </script>

`)
//line views/templates/system/system.qtpl:22
}

//line views/templates/system/system.qtpl:22
func WriteAddRescanJS(qq422016 qtio422016.Writer, route []string) {
	//line views/templates/system/system.qtpl:22
	qw422016 := qt422016.AcquireWriter(qq422016)
	//line views/templates/system/system.qtpl:22
	StreamAddRescanJS(qw422016, route)
	//line views/templates/system/system.qtpl:22
	qt422016.ReleaseWriter(qw422016)
//line views/templates/system/system.qtpl:22
}

//line views/templates/system/system.qtpl:22
func AddRescanJS(route []string) string {
	//line views/templates/system/system.qtpl:22
	qb422016 := qt422016.AcquireByteBuffer()
	//line views/templates/system/system.qtpl:22
	WriteAddRescanJS(qb422016, route)
	//line views/templates/system/system.qtpl:22
	qs422016 := string(qb422016.B)
	//line views/templates/system/system.qtpl:22
	qt422016.ReleaseByteBuffer(qb422016)
	//line views/templates/system/system.qtpl:22
	return qs422016
//line views/templates/system/system.qtpl:22
}