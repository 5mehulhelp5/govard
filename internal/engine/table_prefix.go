package engine

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	tablePrefixPattern        = regexp.MustCompile(`^[A-Za-z0-9_]*$`)
	magentoEnvTablePrefixExpr = regexp.MustCompile(`(?i)['"]table_prefix['"]\s*=>\s*['"]([^'"]*)['"]`)
)

func NormalizeTablePrefix(prefix string) string {
	return strings.TrimSpace(prefix)
}

func ValidateTablePrefix(prefix string) bool {
	return tablePrefixPattern.MatchString(NormalizeTablePrefix(prefix))
}

func FrameworkSupportsTablePrefix(framework string) bool {
	switch normalizeFrameworkManifestKey(framework) {
	case "magento2", "magento1", "openmage":
		return true
	default:
		return false
	}
}

func DetectMagentoTablePrefix(root string, framework string) string {
	if !FrameworkSupportsTablePrefix(framework) {
		return ""
	}
	switch normalizeFrameworkManifestKey(framework) {
	case "magento2":
		return DetectMagento2TablePrefix(root)
	case "magento1", "openmage":
		return DetectMagento1TablePrefix(root)
	default:
		return ""
	}
}

func DetectMagento2TablePrefix(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "app", "etc", "env.php"))
	if err != nil {
		return ""
	}
	matches := magentoEnvTablePrefixExpr.FindStringSubmatch(string(data))
	if len(matches) != 2 {
		return ""
	}
	return NormalizeTablePrefix(matches[1])
}

func DetectMagento1TablePrefix(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "app", "etc", "local.xml"))
	if err != nil {
		return ""
	}

	var localXML struct {
		Global struct {
			Resources struct {
				DB struct {
					TablePrefix string `xml:"table_prefix"`
				} `xml:"db"`
			} `xml:"resources"`
		} `xml:"global"`
	}
	if err := xml.Unmarshal(data, &localXML); err != nil {
		return ""
	}
	return NormalizeTablePrefix(localXML.Global.Resources.DB.TablePrefix)
}
