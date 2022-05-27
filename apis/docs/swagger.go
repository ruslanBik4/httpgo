package docs

import (
	"github.com/go-openapi/loads"
	"github.com/go-openapi/loads/fmts"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/runtime/middleware/untyped"
	"github.com/ruslanBik4/logs"
	"log"
	"net/http"
)

func init() {
	loads.AddLoader(fmts.YAMLMatcher, fmts.YAMLDoc)
}

func LoadSpec() {
	specDoc, err := loads.JSONSpec("http://localhost:9091/apis?json")
	if err != nil {
		logs.ErrorLog(err)
		return
	}

	logs.StatusLog("Would be serving: %+v", specDoc.Spec().Info)
	// our spec doesn't have application/json in the consumes or produces
	// so we need to clear those settings out
	api := untyped.NewAPI(specDoc).WithoutJSONDefaults()

	// register serializers
	mediaType := "application/json"
	api.DefaultConsumes = mediaType
	api.DefaultProduces = mediaType
	api.RegisterConsumer(mediaType, runtime.JSONConsumer())
	api.RegisterProducer(mediaType, runtime.JSONProducer())

	for name, item := range specDoc.Spec().Paths.Paths {
		if item.Get != nil {
			//logs.StatusLog("%s: %+v\n", name, item.Get.Parameters)"/api/log"
			api.RegisterOperation("GET", name, notImplemented)
		}
		if item.Post != nil {
			//logs.StatusLog("%s: %+v\n", name, item.Post.Parameters)
			api.RegisterOperation("POST", name, notImplemented)
		}
	}

	//api.RegisterOperation("POST", "/", notImplemented)

	// validate the API descriptor, to ensure we don't have any unhandled operations
	if err := api.Validate(); err != nil {
		logs.ErrorLog(err)
		return
	}

	// construct the application context for this server
	// use the loaded spec document and the api descriptor with the default router
	app := middleware.NewContext(specDoc, api, nil)

	log.Println("serving", specDoc.Spec().Info.Title, "at http://localhost:8000")
	// serve the api
	if err := http.ListenAndServe(":8000", app.APIHandler(nil)); err != nil {
		logs.ErrorLog(err)
	}
}

var notImplemented = runtime.OperationHandlerFunc(func(params interface{}) (interface{}, error) {
	logs.StatusLog("%#v\n", params)
	return middleware.NotImplemented("not implemented"), nil
})
