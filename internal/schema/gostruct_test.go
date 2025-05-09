package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hangxie/parquet-go/parquet"
	"github.com/stretchr/testify/require"

	pio "github.com/hangxie/parquet-tools/internal/io"
)

func Test_GoStructNode_String_good(t *testing.T) {
	option := pio.ReadOption{}
	uri := "../../testdata/all-types.parquet"
	pr, err := pio.NewParquetFileReader(uri, option)
	require.Nil(t, err)
	defer func() {
		_ = pr.PFile.Close()
	}()

	schemaRoot, err := NewSchemaTree(pr, SchemaOption{})
	require.Nil(t, err)
	require.NotNil(t, schemaRoot)

	typeStr, err := goStructNode{*schemaRoot}.String()
	require.Nil(t, err)

	expected, _ := os.ReadFile("../../testdata/golden/schema-all-types-go.txt")
	// golden file has prefix of "type <root node name>"
	prefix := fmt.Sprintf("type %s ", schemaRoot.Name)
	require.Equal(t, string(expected), prefix+typeStr+"\n")
}

func Test_GoStructNode_String_composite_map_key(t *testing.T) {
	option := pio.ReadOption{}
	uri := "../../testdata/map-value-map.parquet"
	pr, err := pio.NewParquetFileReader(uri, option)
	require.Nil(t, err)
	defer func() {
		_ = pr.PFile.Close()
	}()

	schemaRoot, err := NewSchemaTree(pr, SchemaOption{})
	require.Nil(t, err)
	require.NotNil(t, schemaRoot)

	mapType := parquet.ConvertedType_MAP
	// 2nd field is "Scores", whose 1st field is "Key_value", whose 1st field is map's key
	schemaRoot.Children[1].Children[0].Children[0].ConvertedType = &mapType
	_, err = goStructNode{*schemaRoot}.String()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "go struct does not support composite type as map key")
}

func Test_GoStructNode_String_composite_map_value(t *testing.T) {
	option := pio.ReadOption{}
	uri := "../../testdata/map-composite-value.parquet"
	pr, err := pio.NewParquetFileReader(uri, option)
	require.Nil(t, err)
	defer func() {
		_ = pr.PFile.Close()
	}()

	schemaRoot, err := NewSchemaTree(pr, SchemaOption{})
	require.Nil(t, err)
	require.NotNil(t, schemaRoot)

	_, err = goStructNode{*schemaRoot}.String()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "go struct does not support composite type as map value")
}

func Test_GoStructNode_String_invalid_scalar(t *testing.T) {
	option := pio.ReadOption{}
	uri := "../../testdata/good.parquet"
	pr, err := pio.NewParquetFileReader(uri, option)
	require.Nil(t, err)
	defer func() {
		_ = pr.PFile.Close()
	}()

	schemaRoot, err := NewSchemaTree(pr, SchemaOption{})
	require.Nil(t, err)
	require.NotNil(t, schemaRoot)

	// 1st field is "Shoe_brand"
	schemaRoot.Children[0].Type = nil
	_, err = goStructNode{*schemaRoot}.String()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "type not set")
}

func Test_GoStructNode_String_invalid_list(t *testing.T) {
	option := pio.ReadOption{}
	uri := "../../testdata/reinterpret-list.parquet"
	pr, err := pio.NewParquetFileReader(uri, option)
	require.Nil(t, err)
	defer func() {
		_ = pr.PFile.Close()
	}()

	schemaRoot, err := NewSchemaTree(pr, SchemaOption{})
	require.Nil(t, err)
	require.NotNil(t, schemaRoot)

	invalidType := parquet.Type(999)
	// 2nd field is "V1", whose 1st field is "List", whose 1st field is "Element"
	schemaRoot.Children[0].Children[0].Children[0].Type = &invalidType
	_, err = goStructNode{*schemaRoot}.String()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "unknown type: 999")
}

func Test_GoStructNode_String_invalid_map_key(t *testing.T) {
	option := pio.ReadOption{}
	uri := "../../testdata/reinterpret-map-key.parquet"
	pr, err := pio.NewParquetFileReader(uri, option)
	require.Nil(t, err)
	defer func() {
		_ = pr.PFile.Close()
	}()

	schemaRoot, err := NewSchemaTree(pr, SchemaOption{})
	require.Nil(t, err)
	require.NotNil(t, schemaRoot)

	invalidType := parquet.Type(999)
	// 2nd field is "V1", whose 1st field is "Key_value", whose 1st field is map's key
	schemaRoot.Children[1].Children[0].Children[0].Type = &invalidType
	_, err = goStructNode{*schemaRoot}.String()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "unknown type: 999")
}

func Test_GoStructNode_String_invalid_map_value(t *testing.T) {
	option := pio.ReadOption{}
	uri := "../../testdata/reinterpret-map-key.parquet"
	pr, err := pio.NewParquetFileReader(uri, option)
	require.Nil(t, err)
	defer func() {
		_ = pr.PFile.Close()
	}()

	schemaRoot, err := NewSchemaTree(pr, SchemaOption{})
	require.Nil(t, err)
	require.NotNil(t, schemaRoot)

	// 2nd field is "V1", whose 1st field is "Key_value", whose 3rd field is map's value
	schemaRoot.Children[1].Children[0].Children[1].Type = nil
	_, err = goStructNode{*schemaRoot}.String()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "type not set")
}

func Test_GoStructNode_String_invalid_list_element(t *testing.T) {
	option := pio.ReadOption{}
	uri := "../../testdata/list-of-list.parquet"
	pr, err := pio.NewParquetFileReader(uri, option)
	require.Nil(t, err)
	defer func() {
		_ = pr.PFile.Close()
	}()

	schemaRoot, err := NewSchemaTree(pr, SchemaOption{})
	require.Nil(t, err)
	require.NotNil(t, schemaRoot)

	_, err = goStructNode{*schemaRoot}.String()
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "go struct does not support composite type as list element in field [Parquet_go_root.Lol]")
}

func Test_go_struct_list_variant(t *testing.T) {
	buf, err := os.ReadFile("../../testdata/golden/schema-list-variants-raw.json")
	require.Nil(t, err)

	se := SchemaNode{}
	require.Nil(t, json.Unmarshal(buf, &se))

	schemaRoot := goStructNode{se}
	actual, err := schemaRoot.String()
	require.Nil(t, err)

	buf, err = os.ReadFile("../../testdata/golden/schema-list-variants-go.txt")
	require.Nil(t, err)
	// un-gofmt ...
	expected := strings.ReplaceAll(string(buf), "\t", "")
	re := regexp.MustCompile(" +")
	expected = re.ReplaceAllString(expected, " ")
	expected = strings.TrimRight(expected, "\n")

	require.Equal(t, expected, actual)
}
