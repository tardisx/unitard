package unitard

import (
	"bytes"
	"strings"
	"testing"
)

func TestTemplate(t *testing.T) {
	u := Unit{
		name:          "test_unit",
		binary:        "/fullpath/to/foobar",
		systemCtlPath: "/who/cares",
		unitFilePath:  "/doesnt/matter",
	}

	buff := bytes.NewBuffer(nil) // create empty buffer

	err := u.writeTemplate(buff)
	if err != nil {
		t.Errorf("failed to write template: %s", err)
	}
	t.Logf("template:\n%s", buff.String())

	if !strings.Contains(buff.String(), "Description=test_unit") {
		t.Error("template does not contain description?")
	}
	if !strings.Contains(buff.String(), "ExecStart=/fullpath/to/foobar") {
		t.Error("template does not contain exec start?")
	}
}

func TestCheckName(t *testing.T) {
	validNames := []string{
		"test_unit",
		"leotard123",
		"winger_01",
	}
	invalidNames := []string{
		"no way",
		"doesn't_work",
		"C:/dev/null",
		"/no/slashes",
	}

	for _, v := range validNames {
		if !checkName(v) {
			t.Errorf("%s not valid but should be", v)
		}
	}
	for _, i := range invalidNames {
		if checkName(i) {
			t.Errorf("%s  valid but should be", i)
		}
	}

}
