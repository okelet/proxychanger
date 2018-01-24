package proxychangerlib

import "fmt"
import "github.com/gosexy/gettext"

func MyGettextv(message string, params ...interface{}) string {
	return fmt.Sprintf(gettext.Gettext(message), params...)
}
