package constants

var AttrTypes = map[string]AttrType{
	"BODY":       AT_BODY,
	"QUERY":      AT_QUERY,
	"PATH":       AT_PATH,
	"PLAIN":      AT_PLAIN,
	"XML":        AT_XML,
	"YAML":       AT_YAML,
	"JSON":       AT_JSON,
	"FORM":       AT_FORM,
	"COOKIE":     AT_COOKIE,
	"MULTIPART":  AT_MULTIPART,
	"POST":       AT_POST,
	"GET":        AT_GET,
	"PUT":        AT_PUT,
	"DELETE":     AT_DELETE,
	"PATCH":      AT_PATCH,
	"OPTIONS":    AT_OPTIONS,
	"TRACE":      AT_TRACE,
	"ANY":        AT_ANY,
	"HEAD":       AT_HEAD,
	"IGNORE":     AT_IGNORE,
	"ROUTE":      AT_ROUTE,
	"CONTROLLER": AT_CONTROLLER,
	"BASE":       AT_BASE,
	"SERVICE":    AT_SERVICE,
	"ENTITY":     AT_ENTITY,
	"TABLE":      AT_TABLE,
}

type AttrType = int

const (
	AT_NONE AttrType = iota
	// AT_BODY param
	AT_BODY //@Body
	AT_QUERY
	AT_PATH
	AT_PLAIN
	AT_XML
	AT_YAML
	AT_JSON
	AT_FORM
	AT_COOKIE
	AT_MULTIPART
	// AT_POST http method
	AT_POST //@POST
	AT_GET
	AT_PUT
	AT_DELETE
	AT_PATCH
	AT_OPTIONS
	AT_TRACE
	AT_ANY
	AT_HEAD
	// AT_IGNORE route
	AT_IGNORE
	AT_ROUTE      // @Route /test
	AT_CONTROLLER //@Controller or @Ctl
	AT_BASE       //@Base /api

	// AT_SERVICE inject
	AT_SERVICE
	AT_ENTITY
	AT_TABLE //@Table

	// AT_CUSTOM
	AT_CUSTOM
)
