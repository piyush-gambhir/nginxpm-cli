package output

import (
	"fmt"
	"io"

	"go.yaml.in/yaml/v3"
)

type YAMLFormatter struct{}

func (f *YAMLFormatter) Format(w io.Writer, data interface{}, _ *TableDef) error {
	out, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}
	_, err = fmt.Fprint(w, string(out))
	return err
}
