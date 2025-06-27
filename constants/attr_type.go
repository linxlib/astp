package constants

var AttrTypes = map[string]AttrType{
	"BODY":      AT_BODY,
	"QUERY":     AT_QUERY,
	"PATH":      AT_PATH,
	"PLAIN":     AT_PLAIN,
	"XML":       AT_XML,
	"YAML":      AT_YAML,
	"JSON":      AT_JSON,
	"FORM":      AT_FORM,
	"COOKIE":    AT_COOKIE,
	"MULTIPART": AT_MULTIPART,

	"POST":    AT_POST,
	"GET":     AT_GET,
	"PUT":     AT_PUT,
	"DELETE":  AT_DELETE,
	"PATCH":   AT_PATCH,
	"OPTIONS": AT_OPTIONS,
	"TRACE":   AT_TRACE,
	"ANY":     AT_ANY,
	"HEAD":    AT_HEAD,

	"IGNORE":     AT_IGNORE,
	"ROUTE":      AT_ROUTE,
	"CONTROLLER": AT_CONTROLLER,
	"BASE":       AT_BASE,

	"SERVICE":    AT_SERVICE,
	"ENTITY":     AT_ENTITY,
	"TABLE":      AT_TABLE,
	"DEPRECATED": AT_DEPRECATED,
}

// AttrNames maps AttrType values to their string representations
var AttrNames = map[AttrType]string{
	AT_BODY:       "BODY",
	AT_QUERY:      "QUERY",
	AT_PATH:       "PATH",
	AT_PLAIN:      "PLAIN",
	AT_XML:        "XML",
	AT_YAML:       "YAML",
	AT_JSON:       "JSON",
	AT_FORM:       "FORM",
	AT_COOKIE:     "COOKIE",
	AT_MULTIPART:  "MULTIPART",
	AT_POST:       "POST",
	AT_GET:        "GET",
	AT_PUT:        "PUT",
	AT_DELETE:     "DELETE",
	AT_PATCH:      "PATCH",
	AT_OPTIONS:    "OPTIONS",
	AT_TRACE:      "TRACE",
	AT_ANY:        "ANY",
	AT_HEAD:       "HEAD",
	AT_IGNORE:     "IGNORE",
	AT_ROUTE:      "ROUTE",
	AT_CONTROLLER: "CONTROLLER",
	AT_BASE:       "BASE",
	AT_SERVICE:    "SERVICE",
	AT_ENTITY:     "ENTITY",
	AT_TABLE:      "TABLE",
	AT_DEPRECATED: "DEPRECATED",
}

type AttrType = int

const (
	AT_NONE AttrType = iota
	// AT_BODY param
	AT_BODY //@Body
	AT_QUERY
	AT_HEADER
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
	AT_DEPRECATED
	// AT_CUSTOM
	AT_CUSTOM
)
