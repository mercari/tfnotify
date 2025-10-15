package template

import (
	htmltemplate "html/template"
	texttemplate "text/template"

	"github.com/Masterminds/sprig/v3"
)

func TxtFuncMap() texttemplate.FuncMap {
	// delete some functions for security reason
	funcs := sprig.TxtFuncMap()
	delete(funcs, "env")
	delete(funcs, "expandenv")
	delete(funcs, "getHostByName")
	return funcs
}

func FuncMap() htmltemplate.FuncMap {
	// delete some functions for security reason
	funcs := sprig.FuncMap()
	delete(funcs, "env")
	delete(funcs, "expandenv")
	delete(funcs, "getHostByName")
	return funcs
}
